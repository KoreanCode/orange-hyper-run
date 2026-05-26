package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	growthStateVersion             = 1
	growthRepeatedSignalGoals      = 2
	growthPromotableSignalGoals    = 3
	growthActiveSignalGoals        = 4
	growthHarnessStablePressures   = 3
	growthHarnessPromotableSignals = 4
	growthHarnessActiveSignals     = 5
)

type memoryRecord struct {
	ID         int64
	Kind       string
	Text       string
	Confidence float64
	Quality    string
}

type pressureAccumulator struct {
	kind            string
	pressureType    string
	signal          string
	canonicalSignal string
	effect          string
	goals           map[string]bool
	memoryCount     int
}

func updateGrowthState(root string, db *sql.DB) (growthState, *hyperError) {
	records, err := loadMemoryRecords(db)
	if err != nil {
		return growthState{}, err
	}
	previous := readGrowthStateIfExists(root)
	pressures := deriveGrowthPressures(records)
	candidates, err := materializeGrowthCandidates(root, pressures, previous)
	if err != nil {
		return growthState{}, err
	}
	runtimeBehavior, err := growthBehaviorWithActiveCapabilities(root, pressures)
	if err != nil {
		return growthState{}, err
	}
	state := growthState{
		Version:         growthStateVersion,
		UpdatedAt:       nowISO(),
		PressureLedger:  pressureLedgerFor(pressures, candidates),
		Pressures:       pressures,
		RuntimeBehavior: runtimeBehavior,
		Candidates:      candidates,
		Thresholds: growthThresholds{
			RepeatedSignalGoals:      growthRepeatedSignalGoals,
			PromotableSignalGoals:    growthPromotableSignalGoals,
			ActiveSignalGoals:        growthActiveSignalGoals,
			HarnessStablePressures:   growthHarnessStablePressures,
			HarnessPromotableSignals: growthHarnessPromotableSignals,
			HarnessActiveSignals:     growthHarnessActiveSignals,
		},
	}
	if err := writeJSON(filepath.Join(root, hyperDir, "growth", "state.json"), state); err != nil {
		return growthState{}, err
	}
	return state, nil
}

func readGrowthStateIfExists(root string) growthState {
	var state growthState
	body, err := os.ReadFile(filepath.Join(root, hyperDir, "growth", "state.json"))
	if err != nil {
		return state
	}
	_ = json.Unmarshal(body, &state)
	return state
}

func loadMemoryRecords(db *sql.DB) ([]memoryRecord, *hyperError) {
	rows, err := db.Query(`select id, kind, text, coalesce(confidence, 0), coalesce(quality, '') from memories where stale_at is null order by created_at asc, id asc`)
	if err != nil {
		return nil, dbError(err)
	}
	defer rows.Close()
	records := []memoryRecord{}
	for rows.Next() {
		var record memoryRecord
		if err := rows.Scan(&record.ID, &record.Kind, &record.Text, &record.Confidence, &record.Quality); err != nil {
			return nil, dbError(err)
		}
		if strings.TrimSpace(record.Quality) == "" {
			record.Quality = memoryQuality(record.Kind, record.Text, firstNonZeroFloat(record.Confidence, 0.7))
		}
		records = append(records, record)
	}
	return records, nil
}

func deriveGrowthPressures(records []memoryRecord) []growthPressure {
	accs := []*pressureAccumulator{}
	for _, record := range records {
		if !growthRecordAllowed(record) {
			continue
		}
		signal := memorySignal(record.Text)
		if signal == "" || isNoisyGrowthSignal(signal) {
			continue
		}
		kind := strings.ToLower(strings.TrimSpace(record.Kind))
		pressureType, effect := growthClassification(kind, signal)
		canonical := canonicalPressureSignal(signal)
		acc := findPressureAccumulator(accs, pressureType, canonical)
		if acc == nil {
			acc = &pressureAccumulator{
				kind:            kind,
				pressureType:    pressureType,
				signal:          signal,
				canonicalSignal: canonical,
				effect:          effect,
				goals:           map[string]bool{},
			}
			accs = append(accs, acc)
		}
		acc.memoryCount++
		acc.goals[memoryGoalID(record.Text)] = true
	}

	pressures := make([]growthPressure, 0, len(accs))
	for _, acc := range accs {
		sources := sortedGoalIDs(acc.goals)
		goalCount := len(sources)
		if goalCount == 0 {
			goalCount = 1
		}
		state := "observed"
		if goalCount >= growthRepeatedSignalGoals {
			state = "repeated"
		}
		pressures = append(pressures, growthPressure{
			Kind:            acc.kind,
			PressureType:    acc.pressureType,
			Signal:          acc.signal,
			CanonicalSignal: acc.canonicalSignal,
			Effect:          acc.effect,
			State:           state,
			GoalCount:       goalCount,
			MemoryCount:     acc.memoryCount,
			Score:           growthScore(goalCount, acc.memoryCount),
			Sources:         sources,
		})
	}
	sort.Slice(pressures, func(i, j int) bool {
		if pressures[i].Score == pressures[j].Score {
			if pressures[i].Kind == pressures[j].Kind {
				return pressures[i].Signal < pressures[j].Signal
			}
			return pressures[i].Kind < pressures[j].Kind
		}
		return pressures[i].Score > pressures[j].Score
	})
	if len(pressures) > 24 {
		pressures = pressures[:24]
	}
	return pressures
}

func growthRecordAllowed(record memoryRecord) bool {
	quality := strings.ToLower(strings.TrimSpace(record.Quality))
	if memoryQualityIsIgnored(quality) {
		return false
	}
	signal := memorySignal(record.Text)
	if signal == "" || isNoisyGrowthSignal(signal) {
		return false
	}
	if quality == "weak" {
		return usefulValidationSignal(signal) || hasAny(normalizeSentence(signal), "before every", "before each", "repeatable")
	}
	return true
}

func findPressureAccumulator(accs []*pressureAccumulator, pressureType, canonical string) *pressureAccumulator {
	for _, acc := range accs {
		if acc.pressureType != pressureType {
			continue
		}
		if tokenJaccardString(acc.canonicalSignal, canonical) >= 0.72 {
			return acc
		}
	}
	return nil
}

func growthScore(goalCount, memoryCount int) float64 {
	return float64(goalCount) + float64(memoryCount-goalCount)*0.25
}

func memoryGoalID(text string) string {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return "project"
	}
	field := strings.Trim(fields[0], " :")
	if strings.HasPrefix(field, "GOAL-") {
		return field
	}
	return "project"
}

func memorySignal(text string) string {
	signal := oneLine(text)
	if signal == "" {
		return ""
	}
	if goalID := memoryGoalID(signal); goalID != "project" {
		signal = strings.TrimSpace(strings.TrimPrefix(signal, goalID))
	}
	prefixes := []string{
		"decisions:",
		"readiness evidence:",
		"reusable patterns:",
		"learn decision:",
		"learn pattern:",
		"learn constraint:",
		"learn failure:",
		"surface proof evidence:",
		"validated:",
		"next runtime episode:",
		"blocked:",
		"waiting for user:",
	}
	for {
		changed := false
		lower := strings.ToLower(signal)
		for _, prefix := range prefixes {
			if strings.HasPrefix(lower, prefix) {
				if prefix == "validated:" || prefix == "next runtime episode:" {
					return ""
				}
				signal = strings.TrimSpace(signal[len(prefix):])
				changed = true
				break
			}
		}
		if !changed {
			break
		}
	}
	if isPlaceholder(signal) {
		return ""
	}
	return signal
}

func isNoisyGrowthSignal(signal string) bool {
	normalized := normalizeSentence(signal)
	if isPlaceholder(normalized) {
		return true
	}
	if isNoIssueText(normalized) || isPassiveNoChangeText(normalized) {
		return true
	}
	tokens := pressureTokens(signal)
	if len(tokens) < 2 {
		return true
	}
	noise := map[string]bool{
		"done": true, "fixed": true, "updated": true, "complete": true,
	}
	return noise[normalized]
}

func canonicalPressureSignal(signal string) string {
	tokens := pressureTokens(signal)
	if len(tokens) == 0 {
		return ""
	}
	sort.Strings(tokens)
	return strings.Join(tokens, " ")
}

func pressureTokens(signal string) []string {
	replacements := map[string]string{
		"each":        "every",
		"tests":       "test",
		"testing":     "test",
		"validated":   "validate",
		"validating":  "validate",
		"credentials": "credential",
		"services":    "service",
		"packets":     "packet",
	}
	stops := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true, "by": true,
		"for": true, "from": true, "in": true, "into": true, "is": true, "it": true, "of": true,
		"on": true, "or": true, "the": true, "this": true, "that": true, "to": true, "with": true,
		"before": true, "after": true, "when": true, "where": true, "should": true, "must": true,
	}
	seen := map[string]bool{}
	tokens := []string{}
	fields := strings.FieldsFunc(strings.ToLower(signal), func(r rune) bool {
		return !(r == '_' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r >= '가' && r <= '힣')
	})
	for _, field := range fields {
		if replacement := replacements[field]; replacement != "" {
			field = replacement
		}
		if len([]rune(field)) < 2 || stops[field] {
			continue
		}
		if !seen[field] {
			tokens = append(tokens, field)
			seen[field] = true
		}
	}
	return tokens
}

func tokenJaccardString(left, right string) float64 {
	return tokenJaccard(tokenSet(strings.Fields(left)), tokenSet(strings.Fields(right)))
}

func tokenJaccard(left, right map[string]bool) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	intersection := 0
	union := map[string]bool{}
	for token := range left {
		union[token] = true
		if right[token] {
			intersection++
		}
	}
	for token := range right {
		union[token] = true
	}
	return float64(intersection) / float64(len(union))
}

func tokenSet(tokens []string) map[string]bool {
	result := map[string]bool{}
	for _, token := range tokens {
		result[token] = true
	}
	return result
}

func growthClassification(kind, signal string) (string, string) {
	switch kind {
	case "decision":
		return "stable_decision", "work_boundary"
	case "constraint":
		return "recurring_constraint", "work_boundary"
	case "failure":
		return "recurring_failure", "stop_condition"
	case "pattern":
		if isSurfaceValidationPattern(signal) {
			return "surface_validation", "validation"
		}
		if isValidationPattern(signal) {
			return "repeated_validation", "validation"
		}
		return "implementation_pattern", "implementation"
	default:
		return "context", "context"
	}
}

func isValidationPattern(signal string) bool {
	normalized := strings.ToLower(signal)
	return hasAny(normalized, "test", "build", "smoke", "validate", "validation", "playwright", "browser", "go test", "npm run", "pytest")
}

func isSurfaceValidationPattern(signal string) bool {
	normalized := strings.ToLower(signal)
	return hasAny(normalized, "surface", "screen", "route", "viewport", "mobile", "desktop", "screenshot", "browser", "visual", "responsive", "accessibility", "focus", "keyboard") &&
		hasAny(normalized, "smoke", "verified", "passed", "checked", "captured", "screenshot", "validation", "primary action", "flow")
}

func sortedGoalIDs(goals map[string]bool) []string {
	sources := make([]string, 0, len(goals))
	for goal := range goals {
		sources = append(sources, goal)
	}
	sort.Strings(sources)
	return sources
}

func growthBehaviorFromPressures(pressures []growthPressure) growthBehavior {
	behavior := growthBehavior{
		WorkBoundary:      []string{},
		ValidationSignals: []string{},
		StopConditions:    []string{},
	}
	for _, pressure := range pressures {
		switch pressure.Effect {
		case "work_boundary":
			if len(behavior.WorkBoundary) >= 4 {
				continue
			}
			switch pressure.Kind {
			case "decision":
				behavior.WorkBoundary = append(behavior.WorkBoundary, growthLine("Carry forward", pressure, "learned decision"))
			case "constraint":
				behavior.WorkBoundary = append(behavior.WorkBoundary, growthLine("Respect", pressure, "learned constraint"))
			}
		case "validation":
			if len(behavior.ValidationSignals) < 3 {
				behavior.ValidationSignals = append(behavior.ValidationSignals, growthLine("Reuse", pressure, "validation pattern"))
			}
		case "stop_condition":
			if len(behavior.StopConditions) < 3 {
				behavior.StopConditions = append(behavior.StopConditions, growthLine("Stop early if this appears again", pressure, "known failure"))
			}
		}
	}
	return behavior
}

func growthBehaviorWithActiveCapabilities(root string, pressures []growthPressure) (growthBehavior, *hyperError) {
	behavior := growthBehaviorFromPressures(pressures)
	validators, err := activeValidatorCapabilities(root)
	if err != nil {
		return behavior, err
	}
	seen := map[string]bool{}
	for _, signal := range behavior.ValidationSignals {
		seen[normalizeLabel(signal)] = true
	}
	for _, validator := range validators {
		line := fmt.Sprintf("- Required active validator %s: %s", validator.Name, validator.Signal)
		key := normalizeLabel(line)
		if seen[key] {
			continue
		}
		behavior.ValidationSignals = append(behavior.ValidationSignals, line)
		seen[key] = true
	}
	return behavior, nil
}

type activeCapability struct {
	Kind    string
	Name    string
	Signal  string
	Path    string
	Managed bool
}

func activeValidatorCapabilities(root string) ([]activeCapability, *hyperError) {
	dir := filepath.Join(root, hyperDir, "capabilities", "active", "validator")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []activeCapability{}, nil
		}
		return nil, ioError(err)
	}
	validators := []activeCapability{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, ioError(readErr)
		}
		validator, ok := parseActiveValidatorCapability(entry.Name(), string(body))
		if ok {
			validator.Path = path
			validator.Managed = managedCapabilityFile(string(body))
			validators = append(validators, validator)
		}
	}
	sort.Slice(validators, func(i, j int) bool {
		if validators[i].Name == validators[j].Name {
			return validators[i].Signal < validators[j].Signal
		}
		return validators[i].Name < validators[j].Name
	})
	return validators, nil
}

func activeCapabilities(root string) ([]activeCapability, *hyperError) {
	dir := filepath.Join(root, hyperDir, "capabilities", "active")
	kindEntries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []activeCapability{}, nil
		}
		return nil, ioError(err)
	}
	capabilities := []activeCapability{}
	for _, kindEntry := range kindEntries {
		if !kindEntry.IsDir() {
			continue
		}
		kind := kindEntry.Name()
		entries, readErr := os.ReadDir(filepath.Join(dir, kind))
		if readErr != nil {
			return nil, ioError(readErr)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			path := filepath.Join(dir, kind, entry.Name())
			body, bodyErr := os.ReadFile(path)
			if bodyErr != nil {
				return nil, ioError(bodyErr)
			}
			capability, ok := parseActiveCapability(kind, entry.Name(), string(body))
			if ok {
				capability.Path = path
				capability.Managed = managedCapabilityFile(string(body))
				capabilities = append(capabilities, capability)
			}
		}
	}
	sort.Slice(capabilities, func(i, j int) bool {
		if capabilities[i].Kind == capabilities[j].Kind {
			if capabilities[i].Name == capabilities[j].Name {
				return capabilities[i].Signal < capabilities[j].Signal
			}
			return capabilities[i].Name < capabilities[j].Name
		}
		return capabilities[i].Kind < capabilities[j].Kind
	})
	return capabilities, nil
}

func parseActiveValidatorCapability(filename, body string) (activeCapability, bool) {
	return parseActiveCapability("validator", filename, body)
}

func parseActiveCapability(kind, filename, body string) (activeCapability, bool) {
	status := capabilityField(body, "Status")
	if status != "" && normalizeLabel(status) != "active" {
		return activeCapability{}, false
	}
	name := firstNonBlank(markdownTitle(body), strings.TrimSuffix(filename, filepath.Ext(filename)))
	signal := firstNonBlank(
		capabilityField(body, "Signal"),
		firstSectionLine(body, "Required Behavior"),
		firstSectionLine(body, "Validation"),
	)
	if name == "" || signal == "" {
		return activeCapability{}, false
	}
	return activeCapability{Kind: kind, Name: name, Signal: oneLine(signal)}, true
}

func managedCapabilityFile(body string) bool {
	return capabilityField(body, "Pressure type") != "" || capabilityField(body, "Evidence count") != ""
}

func markdownTitle(body string) string {
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if title, ok := strings.CutPrefix(trimmed, "# "); ok {
			return strings.TrimSpace(title)
		}
	}
	return ""
}

func capabilityField(body, label string) string {
	prefix := strings.ToLower(label) + ":"
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimSpace(strings.TrimLeft(trimmed, "-*"))
		if !strings.HasPrefix(strings.ToLower(trimmed), prefix) {
			continue
		}
		return strings.TrimSpace(trimmed[len(prefix):])
	}
	return ""
}

func growthLine(verb string, pressure growthPressure, label string) string {
	prefix := "learned"
	if pressure.State == "repeated" {
		prefix = "repeated"
	}
	label = strings.ReplaceAll(label, "learned", prefix)
	return fmt.Sprintf("- %s %s: %s", verb, label, pressure.Signal)
}

func materializeGrowthCandidates(root string, pressures []growthPressure, previous growthState) ([]growthCandidate, *hyperError) {
	candidates := []growthCandidate{}
	seen := map[string]bool{}
	for _, pressure := range pressures {
		if pressure.GoalCount < growthRepeatedSignalGoals {
			continue
		}
		switch pressure.Effect {
		case "validation":
			prefix := validatorCandidatePrefix(pressure)
			reason := "Repeated validation pressure crossed the validator threshold."
			if pressure.PressureType == "surface_validation" {
				reason = "Repeated surface proof pressure crossed the validator threshold."
			}
			candidate := growthCandidateForPressure("validator", prefix, "validators", reason, pressure)
			if err := writeGrowthCandidate(root, candidate, pressure); err != nil {
				return nil, err
			}
			if !seen[candidate.LifecyclePath] {
				candidates = append(candidates, candidate)
				seen[candidate.LifecyclePath] = true
			}
		case "implementation":
			candidate := growthCandidateForPressure("skill", "skill", "skills", "Repeated implementation pressure crossed the skill threshold.", pressure)
			if err := writeGrowthCandidate(root, candidate, pressure); err != nil {
				return nil, err
			}
			if !seen[candidate.LifecyclePath] {
				candidates = append(candidates, candidate)
				seen[candidate.LifecyclePath] = true
			}
		case "stop_condition":
			candidate := growthCandidateForPressure("validator", "preflight", "validators", "Repeated failure pressure crossed the preflight threshold.", pressure)
			if err := writeGrowthCandidate(root, candidate, pressure); err != nil {
				return nil, err
			}
			if !seen[candidate.LifecyclePath] {
				candidates = append(candidates, candidate)
				seen[candidate.LifecyclePath] = true
			}
		}
	}
	if harnessPressureReady(pressures) {
		pressure := aggregateHarnessPressure(pressures)
		candidate := harnessCandidateForPressure(pressure)
		if err := writeGrowthCandidate(root, candidate, pressure); err != nil {
			return nil, err
		}
		if !seen[candidate.LifecyclePath] {
			candidates = append(candidates, candidate)
			seen[candidate.LifecyclePath] = true
		}
	}
	active, activeErr := activeCapabilities(root)
	if activeErr != nil {
		return nil, activeErr
	}
	for _, capability := range active {
		if capability.Managed {
			continue
		}
		candidate := growthCandidateForActiveCapability(capability)
		seenKey := firstNonBlank(candidate.LifecyclePath, candidate.Kind+"\x00"+candidate.Name)
		if !seen[seenKey] {
			candidates = append(candidates, candidate)
			seen[seenKey] = true
		}
	}
	retired, err := retiredGrowthCandidates(root, previous, candidates)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, retired...)
	return candidates, nil
}

func validatorCandidatePrefix(pressure growthPressure) string {
	if pressure.PressureType != "surface_validation" {
		return "validator"
	}
	normalized := strings.ToLower(pressure.Signal)
	if hasAny(normalized, "responsive", "breakpoint", "overflow", "clipped", "text fit", "layout shift") {
		return "validator-responsive-check"
	}
	if hasAny(normalized, "accessibility", "focus", "keyboard", "contrast", "aria") {
		return "validator-accessibility-check"
	}
	return "validator-visual-smoke"
}

func growthCandidateForPressure(kind, prefix, generatedDir, reason string, pressure growthPressure) growthCandidate {
	name := growthCandidateName(prefix, pressure)
	status := capabilityStatusForEvidence(pressure.GoalCount)
	return growthCandidate{
		Kind:                kind,
		Name:                name,
		Status:              status,
		GeneratedPath:       filepath.Join(hyperDir, generatedDir, "generated", name+".md"),
		LifecyclePath:       filepath.Join(hyperDir, "capabilities", lifecycleBucket(status), kind, name+".md"),
		Reason:              reason,
		Signal:              pressure.Signal,
		PressureType:        pressure.PressureType,
		Sources:             pressure.Sources,
		EvidenceCount:       pressure.GoalCount,
		RepeatedThreshold:   growthRepeatedSignalGoals,
		PromotionThreshold:  growthPromotableSignalGoals,
		ActivationThreshold: growthActiveSignalGoals,
	}
}

func growthCandidateName(prefix string, pressure growthPressure) string {
	if command := inferredCommandForSignal(pressure.Signal); command != "" {
		return prefix + "-" + slugify(command)
	}
	return prefix + "-" + slugify(cleanCandidateSignal(pressure.Signal))
}

func cleanCandidateSignal(signal string) string {
	cleaned := oneLine(signal)
	prefixes := []string{
		"validation pattern:",
		"readiness evidence:",
		"surface proof evidence:",
		"pressure signals:",
		"learn pattern:",
		"pattern:",
		"proof -",
		"proof:",
	}
	for {
		lower := strings.ToLower(strings.TrimSpace(cleaned))
		changed := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(lower, prefix) {
				cleaned = strings.TrimSpace(cleaned[len(prefix):])
				changed = true
				break
			}
		}
		if !changed {
			break
		}
	}
	return cleaned
}

func harnessCandidateForPressure(pressure growthPressure) growthCandidate {
	status := harnessStatusForPressure(pressure.MemoryCount)
	name := "harness-growth-candidate"
	return growthCandidate{
		Kind:                "harness",
		Name:                name,
		Status:              status,
		GeneratedPath:       filepath.Join(hyperDir, "harnesses", "generated", name+".md"),
		LifecyclePath:       filepath.Join(hyperDir, "capabilities", lifecycleBucket(status), "harness", name+".md"),
		Reason:              "Multiple repeated pressures crossed the harness threshold.",
		Signal:              pressure.Signal,
		PressureType:        pressure.PressureType,
		Sources:             pressure.Sources,
		EvidenceCount:       pressure.GoalCount,
		RepeatedThreshold:   growthHarnessStablePressures,
		PromotionThreshold:  growthHarnessPromotableSignals,
		ActivationThreshold: growthHarnessActiveSignals,
	}
}

func growthCandidateForActiveCapability(capability activeCapability) growthCandidate {
	return growthCandidate{
		Kind:                capability.Kind,
		Name:                capability.Name,
		Status:              "active",
		LifecyclePath:       capability.Path,
		Reason:              "Active capability file is installed in the project.",
		Signal:              capability.Signal,
		PressureType:        "active_capability",
		EvidenceCount:       growthActiveSignalGoals,
		RepeatedThreshold:   growthRepeatedSignalGoals,
		PromotionThreshold:  growthPromotableSignalGoals,
		ActivationThreshold: growthActiveSignalGoals,
	}
}

func capabilityStatusForEvidence(goalCount int) string {
	switch {
	case goalCount >= growthActiveSignalGoals:
		return "active"
	case goalCount >= growthPromotableSignalGoals:
		return "promotable"
	case goalCount >= growthRepeatedSignalGoals:
		return "repeated"
	default:
		return "observed"
	}
}

func harnessStatusForPressure(stablePressureCount int) string {
	switch {
	case stablePressureCount >= growthHarnessActiveSignals:
		return "active"
	case stablePressureCount >= growthHarnessPromotableSignals:
		return "promotable"
	default:
		return "repeated"
	}
}

func lifecycleBucket(status string) string {
	switch status {
	case "active":
		return "active"
	case "retired":
		return "retired"
	default:
		return "candidates"
	}
}

func retiredGrowthCandidates(root string, previous growthState, current []growthCandidate) ([]growthCandidate, *hyperError) {
	currentKeys := map[string]bool{}
	for _, candidate := range current {
		currentKeys[candidate.Kind+"\x00"+candidate.Name] = true
	}
	retired := []growthCandidate{}
	seen := map[string]bool{}
	for _, candidate := range previous.Candidates {
		key := candidate.Kind + "\x00" + candidate.Name
		if currentKeys[key] || seen[key] || candidate.Status == "retired" {
			continue
		}
		candidate.Status = "retired"
		candidate.LifecyclePath = filepath.Join(hyperDir, "capabilities", "retired", candidate.Kind, candidate.Name+".md")
		candidate.Reason = "Source pressure no longer appears in current growth state."
		if err := writeGrowthCandidate(root, candidate, growthPressure{
			Kind:            candidate.Kind,
			PressureType:    candidate.PressureType,
			Signal:          candidate.Signal,
			CanonicalSignal: canonicalPressureSignal(candidate.Signal),
			Effect:          "retired",
			State:           "retired",
			GoalCount:       candidate.EvidenceCount,
			MemoryCount:     candidate.EvidenceCount,
			Sources:         candidate.Sources,
		}); err != nil {
			return nil, err
		}
		retired = append(retired, candidate)
		seen[key] = true
	}
	return retired, nil
}

func harnessPressureReady(pressures []growthPressure) bool {
	stable := 0
	hasValidation := false
	for _, pressure := range pressures {
		if pressure.GoalCount < growthRepeatedSignalGoals {
			continue
		}
		if pressure.Effect == "validation" {
			hasValidation = true
		}
		if pressure.Effect == "validation" || pressure.Effect == "implementation" || pressure.Effect == "work_boundary" {
			stable++
		}
	}
	return hasValidation && stable >= growthHarnessStablePressures
}

func aggregateHarnessPressure(pressures []growthPressure) growthPressure {
	sources := map[string]bool{}
	stablePressureCount := 0
	for _, pressure := range pressures {
		if pressure.GoalCount < growthRepeatedSignalGoals {
			continue
		}
		stablePressureCount++
		for _, source := range pressure.Sources {
			sources[source] = true
		}
	}
	return growthPressure{
		Kind:            "harness",
		PressureType:    "harness_emergence",
		Signal:          "Promote repeated decisions, validation patterns, and constraints into a project-specific harness candidate.",
		CanonicalSignal: "harness emergence",
		Effect:          "harness",
		State:           harnessStatusForPressure(stablePressureCount),
		GoalCount:       len(sources),
		MemoryCount:     stablePressureCount,
		Score:           growthScore(len(sources), stablePressureCount),
		Sources:         sortedGoalIDs(sources),
	}
}

func writeGrowthCandidate(root string, candidate growthCandidate, pressure growthPressure) *hyperError {
	body := strings.Join([]string{
		"# " + candidate.Name,
		"",
		"Status: " + candidate.Status,
		"Kind: " + candidate.Kind,
		"Pressure type: " + candidate.PressureType,
		fmt.Sprintf("Evidence count: %d", candidate.EvidenceCount),
		fmt.Sprintf("Repeated threshold: %d", candidate.RepeatedThreshold),
		fmt.Sprintf("Promotion threshold: %d", candidate.PromotionThreshold),
		fmt.Sprintf("Activation threshold: %d", candidate.ActivationThreshold),
		"",
		"## Reason",
		"",
		candidate.Reason,
		"",
		"## When Required",
		"",
		candidateWhenRequired(candidate, pressure),
		"",
		"## How To Run",
		"",
		candidateHowToRun(candidate, pressure),
		"",
		"## Evidence Required",
		"",
		candidateEvidenceRequired(candidate, pressure),
		"",
		"## Required Behavior",
		"",
		candidateRequiredBehavior(candidate, pressure),
		"",
		"## Pressure",
		"",
		"- Kind: " + pressure.Kind,
		"- Effect: " + pressure.Effect,
		"- Signal: " + pressure.Signal,
		"- Sources: " + strings.Join(pressure.Sources, ", "),
		"",
		"## Activation Rule",
		"",
		candidateActivationRule(candidate),
		"",
	}, "\n")
	if err := removeConflictingLifecycleCopies(root, candidate); err != nil {
		return err
	}
	if strings.TrimSpace(candidate.GeneratedPath) != "" {
		generatedPath := filepath.Join(root, candidate.GeneratedPath)
		if candidate.Status == "retired" {
			if err := os.Remove(generatedPath); err != nil && !os.IsNotExist(err) {
				return ioError(err)
			}
		} else if err := writeText(generatedPath, body); err != nil {
			return err
		}
	}
	if err := writeText(filepath.Join(root, candidate.LifecyclePath), body); err != nil {
		return err
	}
	return nil
}

func candidateWhenRequired(candidate growthCandidate, pressure growthPressure) string {
	if candidate.Status != "active" {
		return "Not required yet. Treat this as a candidate until repeated evidence proves the project needs it."
	}
	switch candidate.Kind {
	case "validator":
		return "Required before closing a runtime packet that touches this pressure: " + compactText(pressure.Signal, 140)
	case "skill":
		return "Use when a future runtime packet repeats this implementation pressure: " + compactText(pressure.Signal, 140)
	case "harness":
		return "Use when multiple repeated validators, skills, or constraints need one stable project-specific structure."
	default:
		return "Use when future work repeats this pressure."
	}
}

func candidateHowToRun(candidate growthCandidate, pressure growthPressure) string {
	if command := inferredCommandForSignal(pressure.Signal); command != "" {
		return "`" + command + "`"
	}
	switch candidate.Kind {
	case "validator":
		return "Run the smallest repeatable check that proves this signal, then paste the command output into evidence.md."
	case "skill":
		return "Apply this guidance during implementation and record the changed files plus validation evidence."
	case "harness":
		return "Review active/repeated capabilities and create a project-specific harness only when one command or workflow would reduce repeated setup."
	default:
		return "Record the action taken and the proof in evidence.md."
	}
}

func inferredCommandForSignal(signal string) string {
	if command := firstBacktickCommand(signal); command != "" {
		return command
	}
	normalized := strings.ToLower(signal)
	switch {
	case strings.Contains(normalized, "go test"):
		return "go test ./..."
	case strings.Contains(normalized, "npm run smoke:persistence"):
		return "npm run smoke:persistence"
	case strings.Contains(normalized, "npm run smoke:api"):
		return "npm run smoke:api"
	case strings.Contains(normalized, "npm run build"):
		return "npm run build"
	case strings.Contains(normalized, "pytest"):
		return "pytest"
	}
	return ""
}

func candidateEvidenceRequired(candidate growthCandidate, pressure growthPressure) string {
	switch candidate.Kind {
	case "validator":
		return "- Command output or smoke result\n- Runtime packet ID\n- Created/read record ID, URL, screenshot, or equivalent proof when available"
	case "skill":
		return "- Runtime packet ID\n- Changed files\n- Validation result showing the guidance helped without adding avoidable process"
	case "harness":
		return "- At least one repeated validation pressure\n- Multiple repeated implementation or boundary pressures\n- Evidence that a shared harness would reduce repeated setup"
	default:
		return "- Runtime packet ID\n- Evidence that the pressure repeated"
	}
}

func candidateRequiredBehavior(candidate growthCandidate, pressure growthPressure) string {
	switch candidate.Kind {
	case "validator":
		return "Before `hyper complete`, prove this behavior or record why it is blocked: " + compactText(pressure.Signal, 160)
	case "skill":
		return "Keep this implementation guidance in mind when the same pressure appears: " + compactText(pressure.Signal, 160)
	case "harness":
		return "Only consolidate repeated validators, skills, and constraints after the project has enough evidence that the structure will be reused."
	default:
		return compactText(pressure.Signal, 160)
	}
}

func candidateActivationRule(candidate growthCandidate) string {
	if candidate.Status == "active" {
		return "This capability is active because it crossed the activation threshold. Keep it active only while future evidence continues to support it."
	}
	return "Do not treat this file as active behavior yet. Promote it only after future runtime packets keep confirming the same pressure and the project needs a stable structure."
}

func firstBacktickCommand(value string) string {
	before, after, ok := strings.Cut(value, "`")
	_ = before
	if !ok {
		return ""
	}
	command, _, ok := strings.Cut(after, "`")
	if !ok {
		return ""
	}
	command = strings.TrimSpace(command)
	if command == "" {
		return ""
	}
	return command
}

func removeConflictingLifecycleCopies(root string, candidate growthCandidate) *hyperError {
	lifecyclePath := filepath.Join(root, candidate.LifecyclePath)
	for _, bucket := range []string{"candidates", "active", "retired"} {
		path := filepath.Join(root, hyperDir, "capabilities", bucket, candidate.Kind, candidate.Name+".md")
		keep := path == lifecyclePath
		if keep {
			continue
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return ioError(err)
		}
	}
	return nil
}

func slugify(value string) string {
	value = strings.ToLower(value)
	var builder strings.Builder
	lastHyphen := false
	for _, r := range value {
		isASCIIAlphaNum := r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
		if isASCIIAlphaNum {
			builder.WriteRune(r)
			lastHyphen = false
			continue
		}
		if !lastHyphen && builder.Len() > 0 {
			builder.WriteByte('-')
			lastHyphen = true
		}
		if builder.Len() >= 60 {
			break
		}
	}
	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		return "signal-" + hashText(value)[:8]
	}
	return slug
}
