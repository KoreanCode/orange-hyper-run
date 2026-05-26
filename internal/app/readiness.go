package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const readinessStateVersion = 1

type readinessDimensionDef struct {
	ID       string
	Name     string
	Keywords []string
	Gap      string
}

type readinessEvidenceRecord struct {
	Axis    string
	GoalID  string
	Text    string
	Status  string
	Quality string
}

type referenceBenchmarkEvidence struct {
	Category              string
	References            string
	BaselineExpectations  string
	CurrentComparison     string
	BelowBaselineGaps     string
	AboveBaselineStrength string
	Decision              string
	ReferenceCount        int
}

func updateReadinessState(root, planBody string, growth growthState) (readinessState, *hyperError) {
	evidence, err := loadReadinessEvidence(root, readinessDimensionDefs())
	if err != nil {
		return readinessState{}, err
	}
	state := deriveReadinessState(parsePlan(planBody), growth, evidence)
	state.UpdatedAt = nowISO()
	if err := materializeReadinessValidatorCandidates(root, state); err != nil {
		return readinessState{}, err
	}
	if err := writeJSON(filepath.Join(root, hyperDir, "readiness", "state.json"), state); err != nil {
		return readinessState{}, err
	}
	return state, nil
}

func readReadinessStateIfExists(root string) readinessState {
	var state readinessState
	body, err := os.ReadFile(filepath.Join(root, hyperDir, "readiness", "state.json"))
	if err != nil {
		return state
	}
	_ = json.Unmarshal(body, &state)
	return state
}

func readinessStateForStatus(root string, growth growthState) readinessState {
	planBody := readIfExists(filepath.Join(root, planFile))
	if strings.TrimSpace(planBody) == "" {
		return readReadinessStateIfExists(root)
	}
	evidence, err := loadReadinessEvidence(root, readinessDimensionDefs())
	if err != nil {
		return readReadinessStateIfExists(root)
	}
	state := deriveReadinessState(parsePlan(planBody), growth, evidence)
	state.UpdatedAt = nowISO()
	return state
}

func deriveReadinessState(plan map[string]string, growth growthState, evidence []readinessEvidenceRecord) readinessState {
	stage := normalizeRuntimeStage(firstRuntimeValue(plan["Current Stage"], "Tiny MVP"))
	dimensions := readinessDimensions(plan, growth, evidence)
	gate := readinessGateForStage(stage, dimensions)
	return readinessState{
		Version:      readinessStateVersion,
		Stage:        stage,
		Dimensions:   dimensions,
		StageGate:    gate,
		NextPressure: selectReadinessPressure(plan, stage, dimensions, gate),
	}
}

func readinessDimensions(plan map[string]string, growth growthState, evidence []readinessEvidenceRecord) []readinessDimension {
	corpus := readinessCorpus(plan, growth)
	dimensions := make([]readinessDimension, 0, len(readinessDimensionDefs()))
	for _, def := range readinessDimensionDefs() {
		status, score, evidenceText := readinessDimensionStatus(def, plan, growth, evidence, corpus)
		dimensions = append(dimensions, readinessDimension{
			ID:       def.ID,
			Name:     def.Name,
			Status:   status,
			Score:    score,
			Evidence: evidenceText,
			Gap:      def.Gap,
		})
	}
	return dimensions
}

func readinessDimensionDefs() []readinessDimensionDef {
	return []readinessDimensionDef{
		{ID: "product_completeness", Name: "Product completeness", Keywords: []string{"product", "mvp", "success criteria", "target users"}, Gap: "The product slice is still too vague to measure."},
		{ID: "core_ux", Name: "Core UX", Keywords: []string{"flow", "user flow", "screen", "ui", "ux", "browser", "click", "task", "chat", "message"}, Gap: "The primary user flow is not yet proven usable."},
		{ID: "persistence", Name: "Data persistence", Keywords: []string{"persist", "persistence", "storage", "database", "sqlite", "mysql", "postgres", "postgresql", "db", "sql", "localstorage", "reload", "save"}, Gap: "User data durability has not been proven."},
		{ID: "error_handling", Name: "Error handling", Keywords: []string{"error", "empty", "loading", "failure", "fallback", "blocked", "edge case"}, Gap: "Failure, empty, or edge states are not yet handled."},
		{ID: "validation_coverage", Name: "Validation coverage", Keywords: []string{"test", "smoke", "validation", "validate", "playwright", "go test", "npm run", "pytest"}, Gap: "The primary behavior does not have repeatable validation evidence."},
		{ID: "security_baseline", Name: "Security baseline", Keywords: []string{"security", "privacy", "permission", "rate limit", "secret", "session", "token", "telemetry", "misuse", "data handling"}, Gap: "Basic security, privacy, and misuse boundaries are not yet explicit."},
		{ID: "deployment_readiness", Name: "Deployment readiness", Keywords: []string{"deploy", "release", "production", "server", "docker", "github actions", "continuous integration", "hosted"}, Gap: "The project is not yet proven runnable outside the local development path."},
		{ID: "operations_docs", Name: "Operations and docs", Keywords: []string{"readme", "docs", "runbook", "rollback", "logs", "monitor", "environment"}, Gap: "Operational notes, setup, rollback, or handoff docs are not sufficient."},
		{ID: "maintainability", Name: "Maintainability", Keywords: []string{"refactor", "cleanup", "component", "module", "architecture", "helper", "table-driven"}, Gap: "The codebase has not accumulated enough maintainability evidence."},
		{ID: "reference_benchmark", Name: "Reference benchmark", Keywords: []string{"reference", "benchmark", "baseline", "category", "comparison", "comparable", "above baseline", "below baseline"}, Gap: "Reference comparison has not proven category baseline and differentiating strength."},
		{ID: "sustained_quality", Name: "Sustained quality", Keywords: []string{"sustained", "repeated evidence", "active validator", "active harness", "repeated pressure"}, Gap: "Sustained quality needs repeated runtime evidence and an active validator or equivalent reusable quality structure."},
	}
}

func readinessDimensionStatus(def readinessDimensionDef, plan map[string]string, growth growthState, evidenceRecords []readinessEvidenceRecord, corpus string) (string, int, string) {
	record, hasRecord := readinessEvidenceForAxis(evidenceRecords, def.ID)
	if def.ID == "sustained_quality" {
		covered, emerging, evidence := sustainedQualityGrowthEvidence(growth)
		if covered {
			return "covered", 2, evidence
		}
		if hasRecord {
			return "emerging", 1, fmt.Sprintf("%s readiness evidence needs an actual active validator, active harness, or equivalent active capability: %s", record.GoalID, record.Text)
		}
		if emerging {
			return "emerging", 1, evidence
		}
		if corpusMentionsReadinessAxis(corpus, def) {
			return "emerging", 1, "plan.md or learned context mentions this readiness axis."
		}
		return "missing", 0, def.Gap
	}
	if hasRecord {
		if record.Status == "covered" {
			return "covered", 2, fmt.Sprintf("%s readiness evidence: %s", record.GoalID, record.Text)
		}
	}
	if def.ID == "product_completeness" {
		status, score, evidence := productCompletenessFromPlan(plan)
		if status == "covered" {
			return status, score, evidence
		}
		if hasRecord {
			return "emerging", 1, fmt.Sprintf("%s readiness evidence needs stronger proof for %s: %s", record.GoalID, record.Quality, record.Text)
		}
		if status == "emerging" {
			return status, score, evidence
		}
		return "missing", 0, "plan.md does not yet define a measurable product slice."
	}

	covered, emerging, evidence := growthEvidenceForDimension(growth, def)
	if covered {
		return "covered", 2, evidence
	}
	if hasRecord {
		return "emerging", 1, fmt.Sprintf("%s readiness evidence needs stronger proof for %s: %s", record.GoalID, record.Quality, record.Text)
	}
	if emerging {
		return "emerging", 1, evidence
	}
	if corpusMentionsReadinessAxis(corpus, def) {
		return "emerging", 1, "plan.md or learned context mentions this readiness axis."
	}
	return "missing", 0, def.Gap
}

func productCompletenessFromPlan(plan map[string]string) (string, int, string) {
	product := firstRuntimeValue(plan["Product"])
	mvp := firstRuntimeValue(plan["MVP"])
	success := firstRuntimeValue(plan["Success Criteria"])
	if product != "" && mvp != "" && success != "" {
		return "covered", 2, "plan.md defines product, MVP, and success criteria."
	}
	if product != "" || mvp != "" {
		return "emerging", 1, "plan.md has partial product or MVP context."
	}
	return "missing", 0, "plan.md does not yet define a measurable product slice."
}

func loadReadinessEvidence(root string, defs []readinessDimensionDef) ([]readinessEvidenceRecord, *hyperError) {
	goalsDir := filepath.Join(root, hyperDir, "goals")
	entries, err := os.ReadDir(goalsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []readinessEvidenceRecord{}, nil
		}
		return nil, ioError(err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	records := []readinessEvidenceRecord{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		goalID := entry.Name()
		body := readIfExists(filepath.Join(goalsDir, goalID, "evidence.md"))
		for _, line := range usefulSectionLines(body, "Readiness Evidence") {
			record, ok := parseReadinessEvidenceLine(goalID, line, defs)
			if ok {
				records = append(records, record)
			}
		}
		for _, line := range usefulSectionLines(body, "Validation") {
			record, ok := parseReadinessEvidenceLine(goalID, line, defs)
			if ok {
				records = append(records, record)
				continue
			}
			records = append(records, inferReadinessEvidenceFromValidationLine(goalID, line)...)
		}
		for _, line := range usefulSectionLines(body, "Surface Proof Evidence") {
			record, ok := parseReadinessEvidenceLine(goalID, line, defs)
			if ok {
				records = append(records, record)
				continue
			}
			records = append(records, inferReadinessEvidenceFromSurfaceLine(goalID, line)...)
		}
		records = append(records, inferReadinessEvidenceFromReferenceBenchmark(goalID, usefulSectionLines(body, "Reference Benchmark Evidence"))...)
	}
	return records, nil
}

func parseReadinessEvidenceLine(goalID, line string, defs []readinessDimensionDef) (readinessEvidenceRecord, bool) {
	text := oneLine(line)
	if !usefulReadinessEvidence(text) {
		return readinessEvidenceRecord{}, false
	}
	if label, value, ok := strings.Cut(text, ":"); ok {
		axis := readinessAxisForLabel(label, defs)
		value = strings.TrimSpace(value)
		if axis != "" && usefulReadinessEvidence(value) {
			return readinessEvidenceRecordForAxis(goalID, axis, value), true
		}
	}
	if label, value, ok := strings.Cut(text, " - "); ok {
		axis := readinessAxisForLabel(label, defs)
		value = strings.TrimSpace(value)
		if axis != "" && usefulReadinessEvidence(value) {
			return readinessEvidenceRecordForAxis(goalID, axis, value), true
		}
	}
	return readinessEvidenceRecord{}, false
}

func inferReadinessEvidenceFromValidationLine(goalID, line string) []readinessEvidenceRecord {
	text := oneLine(line)
	if !usefulReadinessEvidence(text) {
		return nil
	}
	records := []readinessEvidenceRecord{}
	for _, axis := range []string{"validation_coverage", "core_ux"} {
		covered, _ := readinessEvidenceQuality(axis, text)
		if covered {
			records = append(records, readinessEvidenceRecordForAxis(goalID, axis, text))
		}
	}
	return records
}

func inferReadinessEvidenceFromSurfaceLine(goalID, line string) []readinessEvidenceRecord {
	if !surfaceProofReadinessInferenceAllowed(line) {
		return nil
	}
	text := surfaceProofValue(line)
	if !usefulReadinessEvidence(text) || !looksLikeSurfaceProof(text) {
		return nil
	}
	records := []readinessEvidenceRecord{}
	for _, axis := range []string{"core_ux", "validation_coverage", "error_handling"} {
		covered, _ := readinessEvidenceQuality(axis, text)
		if covered {
			records = append(records, readinessEvidenceRecordForAxis(goalID, axis, text))
		}
	}
	return records
}

func surfaceProofReadinessInferenceAllowed(line string) bool {
	text := oneLine(line)
	label, _, ok := strings.Cut(text, ":")
	if !ok {
		return true
	}
	switch compactReadinessLabel(label) {
	case "surfacerisksorgaps", "surfacerisk", "surfacegaps":
		return false
	default:
		return true
	}
}

func surfaceProofValue(line string) string {
	text := oneLine(line)
	if label, value, ok := strings.Cut(text, ":"); ok {
		compact := compactReadinessLabel(label)
		switch compact {
		case "targetsurface", "primaryuseraction", "stateschecked", "viewports", "evidence", "surfacerisksorgaps", "surfacerisk", "surfacegaps", "browsersmoke", "viewportproof":
			return strings.TrimSpace(value)
		}
	}
	return text
}

func looksLikeSurfaceProof(text string) bool {
	normalized := strings.ToLower(text)
	return hasAny(normalized, "surface", "screen", "route", "viewport", "mobile", "desktop", "browser", "screenshot", "smoke", "click", "primary action", "flow", "state") &&
		hasAny(normalized, "passed", "verified", "checked", "captured", "screenshot", "smoke", "browser", "viewport")
}

func readinessEvidenceRecordForAxis(goalID, axis, text string) readinessEvidenceRecord {
	covered, quality := readinessEvidenceQuality(axis, text)
	status := "emerging"
	if covered {
		status = "covered"
	}
	return readinessEvidenceRecord{Axis: axis, GoalID: goalID, Text: text, Status: status, Quality: quality}
}

func productCompletenessEvidenceCovered(normalized string) bool {
	productSurface := hasAny(normalized,
		"product", "mvp", "slice", "flow", "api", "endpoint", "command", "feature", "behavior", "primary", "user can",
	)
	measurableProof := hasAny(normalized,
		"success", "criteria", "target", "measurable", "defined", "proved", "proven", "verified", "works", "creates", "returns", "lists",
	)
	concreteBehavior := hasAny(normalized,
		"create", "list", "send", "open", "read", "write", "complete", "delete", "login", "sign up", "note", "chat", "pin", "task",
	)
	return productSurface && measurableProof && concreteBehavior
}

func coreUXEvidenceCovered(normalized string) bool {
	visualSurfaceProof := hasAny(normalized, "browser", "screenshot", "viewport", "mobile", "desktop", "screen", "surface", "user interface", "page", "button", "panel", "route")
	userActionProof := hasAny(normalized, "flow", "click", "create", "add", "edit", "complete", "delete", "send", "navigate", "reload", "primary action", "state")
	screenProof := hasAny(normalized, "smoke", "screenshot", "browser", "verified", "passed", "checked", "captured") &&
		visualSurfaceProof &&
		userActionProof
	if screenProof {
		return true
	}
	actionProof := hasAny(normalized, "create", "list", "send", "complete", "read", "write", "post", "get", "run", "execute", "start", "invoke", "return", "returns", "returned", "print", "prints", "printed", "output", "primary flow", "primary command", "run command")
	resultProof := hasAny(normalized, "verified", "passed", "proved", "proven", "works", "test", "httptest", "smoke", "exit code 0", "output matched", "expected output", "returned")
	apiOrCLIProof := hasAny(normalized, "api", "endpoint", "cli", "command", "http", "route") &&
		actionProof &&
		resultProof
	return apiOrCLIProof
}

func usefulReadinessEvidence(text string) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	if normalized == "" || isPlaceholder(normalized) {
		return false
	}
	return !weakReadinessEvidence(normalized)
}

func weakReadinessEvidence(normalized string) bool {
	for _, phrase := range []string{"not yet", "not enough", "cannot", "could not", "unable"} {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	for _, phrase := range []string{
		"missing evidence",
		"missing proof",
		"missing for",
		"is missing",
		"are missing",
		"not captured",
		"not proven",
		"not handled",
	} {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	if strings.Contains(normalized, "blocked") && !hasAny(normalized, "not blocked", "none blocking", "no blocker") {
		return true
	}
	if strings.Contains(normalized, "failed") && !hasAny(normalized, "failed before", "previously failed", "failure state") {
		return true
	}
	return false
}

func readinessAxisForLabel(label string, defs []readinessDimensionDef) string {
	compact := compactReadinessLabel(label)
	aliases := map[string]string{
		"product":             "product_completeness",
		"productcompleteness": "product_completeness",
		"mvp":                 "product_completeness",
		"core":                "core_ux",
		"coreux":              "core_ux",
		"ux":                  "core_ux",
		"flow":                "core_ux",
		"data":                "persistence",
		"datapersistence":     "persistence",
		"persistence":         "persistence",
		"storage":             "persistence",
		"errors":              "error_handling",
		"error":               "error_handling",
		"errorhandling":       "error_handling",
		"edgecases":           "error_handling",
		"validation":          "validation_coverage",
		"validationcoverage":  "validation_coverage",
		"tests":               "validation_coverage",
		"test":                "validation_coverage",
		"security":            "security_baseline",
		"securitybaseline":    "security_baseline",
		"privacy":             "security_baseline",
		"privacyboundary":     "security_baseline",
		"dataprivacy":         "security_baseline",
		"datahandling":        "security_baseline",
		"misuse":              "security_baseline",
		"misuseboundary":      "security_baseline",
		"deployment":          "deployment_readiness",
		"deploymentreadiness": "deployment_readiness",
		"deploy":              "deployment_readiness",
		"ops":                 "operations_docs",
		"operations":          "operations_docs",
		"operationsdocs":      "operations_docs",
		"docs":                "operations_docs",
		"maintainability":     "maintainability",
		"maintenance":         "maintainability",
		"codequality":         "maintainability",
		"reference":           "reference_benchmark",
		"references":          "reference_benchmark",
		"benchmark":           "reference_benchmark",
		"referencebenchmark":  "reference_benchmark",
		"baseline":            "reference_benchmark",
		"comparison":          "reference_benchmark",
		"sustainedquality":    "sustained_quality",
		"sustained":           "sustained_quality",
		"activevalidator":     "sustained_quality",
		"activeharness":       "sustained_quality",
	}
	if axis := aliases[compact]; axis != "" {
		return axis
	}
	for _, def := range defs {
		if compact == compactReadinessLabel(def.ID) || compact == compactReadinessLabel(def.Name) {
			return def.ID
		}
	}
	return ""
}

func compactReadinessLabel(label string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(label) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func readinessEvidenceForAxis(records []readinessEvidenceRecord, axis string) (readinessEvidenceRecord, bool) {
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].Axis == axis && records[i].Status == "covered" {
			return records[i], true
		}
	}
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].Axis == axis {
			return records[i], true
		}
	}
	return readinessEvidenceRecord{}, false
}

func readinessEvidenceQuality(axis, text string) (bool, string) {
	normalized := strings.ToLower(text)
	switch axis {
	case "product_completeness":
		return productCompletenessEvidenceCovered(normalized),
			"measurable product, MVP, target, or success criteria"
	case "core_ux":
		return coreUXEvidenceCovered(normalized),
			"browser, screenshot, smoke, or verified primary-flow evidence"
	case "persistence":
		return hasAny(normalized, "persist", "reload", "restart", "saved", "save", "survive", "stored", "stores", "created", "re-read", "reread", "read back", "reads it back", "confirmed", "row") &&
				hasAny(normalized, "sqlite", "mysql", "postgres", "postgresql", "database", " db ", "db check", "sql", "localstorage", "local storage", "storage", "json", ".json", ".txt", "file", "disk", "filesystem"),
			"MySQL, SQLite, DB, file, JSON, reload, restart, storage, or database evidence"
	case "error_handling":
		return hasAny(normalized, "empty", "error", "loading", "fallback", "failure", "edge", "missing argument", "missing input", "missing name", "missing required", "missing state", "missing file", "missing data", "corrupt", "corrupted", "invalid input", "invalid command", "unknown command", "required field", "required input") &&
				hasAny(normalized, "handled", "covered", "verified", "tested", "implemented", "works", "rejected", "proves", "proved", "passed"),
			"empty, loading, error, failure, fallback, or edge-state evidence"
	case "validation_coverage":
		return hasAny(normalized, "smoke", "playwright", "go test", "npm run", "pytest", "build", "command", "validation", "browser", "screenshot", "`") &&
				hasAny(normalized, "passed", "repeatable", "covered", "verified", "proved", "proven", "works"),
			"repeatable command, build, test, smoke, or coverage evidence"
	case "security_baseline":
		return securityBaselineEvidenceCovered(normalized),
			"security, privacy, permission, token, session, telemetry, data-handling, or misuse-boundary evidence"
	case "deployment_readiness":
		return deploymentEvidenceCovered(normalized),
			"deploy, hosted URL, release, build, artifact, zip, file smoke, Docker, or CI evidence"
	case "operations_docs":
		return operationsDocsEvidenceCovered(normalized),
			"README, docs, setup, runbook, rollback, smoke path, stop conditions, or environment evidence"
	case "maintainability":
		return hasAny(normalized, "refactor", "cleanup", "component", "module", "architecture", "helper", "table-driven", "test", "extracted", "reduced", "documented", "document", "documents", "handoff", "maintenance", "synchronized", "sync"),
			"refactor, modularity, test, helper, cleanup, documentation, handoff, or maintenance evidence"
	case "reference_benchmark":
		return referenceBenchmarkEvidenceQuality(normalized)
	case "sustained_quality":
		return sustainedQualityEvidenceCovered(normalized),
			"active validator, active harness, or active capability evidence"
	default:
		return len(strings.Fields(normalized)) >= 4, "specific evidence for this readiness axis"
	}
}

func inferReadinessEvidenceFromReferenceBenchmark(goalID string, lines []string) []readinessEvidenceRecord {
	if len(lines) == 0 {
		return nil
	}
	text := oneLine(strings.Join(lines, "; "))
	if !usefulReadinessEvidence(text) {
		return nil
	}
	if !hasAny(strings.ToLower(text), "reference", "benchmark", "baseline", "category", "comparison") {
		return nil
	}
	return []readinessEvidenceRecord{readinessEvidenceRecordForAxis(goalID, "reference_benchmark", text)}
}

func referenceBenchmarkEvidenceQuality(text string) (bool, string) {
	fields := parseReferenceBenchmarkEvidence(text)
	missing := referenceBenchmarkMissingRequirements(fields)
	if len(missing) > 0 {
		return false, "reference benchmark needs " + strings.Join(missing, ", ")
	}
	return true, "3-5 references, category baseline, current comparison, no critical below-baseline gap, above-baseline strength, and decision"
}

func parseReferenceBenchmarkEvidence(text string) referenceBenchmarkEvidence {
	fields := referenceBenchmarkEvidence{}
	currentField := ""
	for _, chunk := range referenceBenchmarkChunks(text) {
		label, value, ok := strings.Cut(chunk, ":")
		if !ok {
			appendReferenceBenchmarkContinuation(&fields, currentField, chunk)
			continue
		}
		value = strings.TrimSpace(value)
		compactLabel := compactReadinessLabel(label)
		field := referenceBenchmarkFieldForLabel(compactLabel)
		if value == "" || (isPlaceholder(value) && compactLabel != "belowbaselinegaps" && compactLabel != "belowbaselinegap" && compactLabel != "gaps") {
			if field != "" {
				currentField = field
			}
			continue
		}
		if field == "" {
			appendReferenceBenchmarkUnknownLabel(&fields, currentField, label, value)
			continue
		}
		currentField = field
		setReferenceBenchmarkField(&fields, field, value)
	}
	fields.ReferenceCount = countReferenceItems(fields.References)
	return fields
}

func referenceBenchmarkFieldForLabel(compactLabel string) string {
	switch compactLabel {
	case "category":
		return "category"
	case "reference", "references", "namedreference", "namedreferences":
		return "references"
	case "baseline", "baselineexpectations", "expectations", "categorybaseline":
		return "baseline"
	case "currentcomparison", "comparison":
		return "comparison"
	case "belowbaselinegaps", "belowbaselinegap", "gaps", "nocriticalbelowbaselinegap", "nocorecategorybaselinegap":
		return "below_gaps"
	case "abovebaselinestrength", "abovebaselinestrengths", "strength", "differentiatingstrength":
		return "above_strength"
	case "decision":
		return "decision"
	default:
		return ""
	}
}

func setReferenceBenchmarkField(fields *referenceBenchmarkEvidence, field, value string) {
	switch field {
	case "category":
		fields.Category = value
	case "references":
		fields.References = appendBenchmarkValue(fields.References, value)
	case "baseline":
		fields.BaselineExpectations = appendBenchmarkValue(fields.BaselineExpectations, value)
	case "comparison":
		fields.CurrentComparison = appendBenchmarkValue(fields.CurrentComparison, value)
	case "below_gaps":
		fields.BelowBaselineGaps = appendBenchmarkValue(fields.BelowBaselineGaps, value)
	case "above_strength":
		fields.AboveBaselineStrength = appendBenchmarkValue(fields.AboveBaselineStrength, value)
	case "decision":
		fields.Decision = appendBenchmarkValue(fields.Decision, value)
	}
}

func appendReferenceBenchmarkContinuation(fields *referenceBenchmarkEvidence, currentField, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	setReferenceBenchmarkField(fields, currentField, value)
}

func appendReferenceBenchmarkUnknownLabel(fields *referenceBenchmarkEvidence, currentField, label, value string) {
	label = strings.TrimSpace(label)
	value = strings.TrimSpace(value)
	if currentField == "references" {
		if referenceCountPrefix(label) {
			fields.References = appendBenchmarkValue(fields.References, value)
			return
		}
		fields.References = appendBenchmarkValue(fields.References, label)
		return
	}
	setReferenceBenchmarkField(fields, currentField, strings.TrimSpace(label+": "+value))
}

func appendBenchmarkValue(existing, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return existing
	}
	if strings.TrimSpace(existing) == "" {
		return value
	}
	return existing + "; " + value
}

func referenceCountPrefix(label string) bool {
	normalized := normalizeSentence(label)
	return hasAny(normalized, "total", "selected") || strings.ContainsAny(normalized, "0123456789")
}

func referenceBenchmarkChunks(text string) []string {
	raw := strings.FieldsFunc(text, func(r rune) bool {
		return r == '\n' || r == ';'
	})
	chunks := []string{}
	for _, chunk := range raw {
		trimmed := strings.TrimSpace(strings.TrimLeft(chunk, "-*0123456789. "))
		if trimmed != "" {
			chunks = append(chunks, trimmed)
		}
	}
	return chunks
}

func referenceBenchmarkMissingRequirements(fields referenceBenchmarkEvidence) []string {
	missing := []string{}
	if strings.TrimSpace(fields.Category) == "" {
		missing = append(missing, "category")
	}
	if fields.ReferenceCount < 3 || fields.ReferenceCount > 5 {
		missing = append(missing, "3-5 named references")
	}
	if !specificBenchmarkField(fields.BaselineExpectations) {
		missing = append(missing, "baseline expectations")
	}
	if !currentComparisonCovered(fields.CurrentComparison) {
		missing = append(missing, "current comparison using below/meets/above baseline")
	}
	if !noCriticalBelowBaselineGap(fields.BelowBaselineGaps) {
		missing = append(missing, "no critical below-baseline gap")
	}
	if !specificBenchmarkField(fields.AboveBaselineStrength) {
		missing = append(missing, "above-baseline strength")
	}
	if !specificBenchmarkField(fields.Decision) {
		missing = append(missing, "decision")
	}
	return missing
}

func countReferenceItems(value string) int {
	seen := map[string]bool{}
	for _, item := range splitReferenceItems(value) {
		normalized := normalizeReferenceItem(item)
		if normalized == "" || isPlaceholder(normalized) {
			continue
		}
		if hasAny(normalized, "3-5", "three to five", "named references", "comparable products", "comparable tools", "comparable apps", "tool a", "tool b", "tool c") {
			continue
		}
		seen[normalized] = true
	}
	return len(seen)
}

func splitReferenceItems(value string) []string {
	replacer := strings.NewReplacer("\n", ",", ";", ",", "|", ",", " / ", ",", " and ", ",")
	normalized := replacer.Replace(value)
	return strings.Split(normalized, ",")
}

func normalizeReferenceItem(item string) string {
	item = strings.TrimSpace(item)
	if label, value, ok := strings.Cut(item, ":"); ok && referenceCountPrefix(label) {
		item = value
	}
	item = strings.TrimSpace(strings.Trim(item, "."))
	return normalizeSentence(item)
}

func specificBenchmarkField(value string) bool {
	normalized := normalizeSentence(value)
	return normalized != "" && !isPlaceholder(normalized) && !hasAny(normalized, "pending", "todo") && len(strings.Fields(normalized)) >= 3
}

func currentComparisonCovered(value string) bool {
	normalized := normalizeSentence(value)
	return specificBenchmarkField(value) &&
		(hasAny(normalized, "below baseline", "below-baseline", "meets baseline", "meet baseline", "above baseline", "above-baseline") ||
			(strings.Contains(normalized, "baseline") && hasAny(normalized, "below", "meets", "meet", "above")))
}

func noCriticalBelowBaselineGap(value string) bool {
	normalized := normalizeSentence(value)
	if normalized == "" || isPlaceholder(normalized) {
		return normalized == "none" || strings.HasPrefix(normalized, "none critical")
	}
	return hasAny(normalized, "none", "no critical", "no core", "no below baseline", "no below-baseline", "not blocked", "none blocking")
}

func securityBaselineEvidenceCovered(normalized string) bool {
	hasSecurityBoundary := hasAny(normalized,
		"security", "privacy", "permission", "rate limit", "secret", "session", "token", "auth", "abuse", "misuse",
		"telemetry", "data handling", "data boundary", "local only", "local-only", "no cloud", "cloud sync", "sensitive",
	)
	hasProof := hasAny(normalized,
		"documented", "verified", "implemented", "checked", "covered", "explicit", "no cloud", "no telemetry", "deleted", "delete path",
	)
	return hasSecurityBoundary && hasProof
}

func sustainedQualityEvidenceCovered(normalized string) bool {
	if hasAny(normalized, "not active", "not yet active", "not required behavior yet", "not active required behavior") {
		return false
	}
	return hasAny(normalized, "active validator", "active harness", "active capability") &&
		hasAny(normalized, "promoted", "required", "covered", "verified", "proved", "proven", "active")
}

func sustainedQualityGrowthEvidence(growth growthState) (bool, bool, string) {
	active := []string{}
	for _, candidate := range growth.Candidates {
		if candidate.Status != "active" {
			continue
		}
		if candidate.Kind == "validator" || candidate.Kind == "harness" {
			active = append(active, candidate.Kind+" "+candidate.Name)
		}
	}
	if len(active) > 0 {
		sort.Strings(active)
		return true, true, "Active quality structures prove repeated quality pressure became required behavior: " + strings.Join(active, ", ") + "."
	}
	for _, candidate := range growth.Candidates {
		if candidate.Status == "promotable" || candidate.Status == "repeated" {
			return false, true, "Repeated quality pressure exists but has not become active required behavior yet: " + candidate.Name
		}
	}
	for _, pressure := range growth.Pressures {
		if pressure.GoalCount >= growthRepeatedSignalGoals && (pressure.Effect == "validation" || pressure.Effect == "harness") {
			return false, true, "Repeated quality pressure exists but has not crossed the active threshold yet: " + pressure.Signal
		}
	}
	return false, false, ""
}

func deploymentEvidenceCovered(normalized string) bool {
	deploymentTarget := hasAny(normalized,
		"deploy", "deployed", "deployment", "url", "https://", "http://", "build", "release", "hosted", "docker",
		"github actions", "continuous integration", "ci pipeline", "ci check", "ci passed",
		"artifact", "zip", "dist/", "file://", "static server", "server check", "packaged", "package", "parity",
		"binary", "executable", "outside the development", "outside development", "smoke command",
	)
	deploymentProof := hasAny(normalized,
		"passed", "available", "hosted", "deployed", "built", "released", "verified", "validated", "proved", "proven",
		"verifies", "validates", "creates", "creation", "created", "served", "extracted", "smoke", "parity", "ran",
	)
	return deploymentTarget && deploymentProof
}

func operationsDocsEvidenceCovered(normalized string) bool {
	docsTarget := hasAny(normalized,
		"readme", "docs", "document", "demo-release", "setup", "runbook", "rollback", "logs", "monitor", "environment",
		"handoff", "tester", "smoke path", "run command", "package command", "stop condition", "stop conditions",
	)
	docsProof := hasAny(normalized,
		"documented", "documents", "updated", "verified", "covered", "cover", "covers", "written", "records", "defines", "includes",
	)
	return docsTarget && docsProof
}

func growthEvidenceForDimension(growth growthState, def readinessDimensionDef) (bool, bool, string) {
	if def.ID == "sustained_quality" {
		return sustainedQualityGrowthEvidence(growth)
	}
	for _, pressure := range growth.Pressures {
		if !pressureMatchesReadiness(pressure, def) {
			continue
		}
		evidence := fmt.Sprintf("Learned %s signal: %s", pressure.PressureType, pressure.Signal)
		if pressure.State == "repeated" || pressure.GoalCount >= growthRepeatedSignalGoals {
			return true, true, evidence
		}
		return false, true, evidence
	}
	for _, signal := range growth.RuntimeBehavior.ValidationSignals {
		normalized := strings.ToLower(signal)
		if !readinessSignalDefersAxis(normalized, def.ID) && hasAny(normalized, def.Keywords...) {
			return true, true, "Active runtime behavior references this readiness axis."
		}
	}
	return false, false, ""
}

func pressureMatchesReadiness(pressure growthPressure, def readinessDimensionDef) bool {
	signal := strings.ToLower(pressure.Signal + " " + pressure.PressureType + " " + pressure.Effect)
	if readinessSignalDefersAxis(signal, def.ID) {
		return false
	}
	switch def.ID {
	case "validation_coverage":
		return pressure.Effect == "validation" || hasAny(signal, def.Keywords...)
	case "error_handling":
		return pressure.Effect == "stop_condition" || hasAny(signal, def.Keywords...)
	case "maintainability":
		return pressure.Effect == "implementation" || hasAny(signal, def.Keywords...)
	case "sustained_quality":
		return pressure.GoalCount >= growthRepeatedSignalGoals && (pressure.Effect == "validation" || pressure.Effect == "harness")
	default:
		return hasAny(signal, def.Keywords...)
	}
}

func corpusMentionsReadinessAxis(corpus string, def readinessDimensionDef) bool {
	if readinessSignalDefersAxis(corpus, def.ID) {
		return false
	}
	return hasAny(corpus, def.Keywords...)
}

func readinessSignalDefersAxis(normalized, axis string) bool {
	switch axis {
	case "core_ux":
		return hasAny(normalized, "before adding ui", "before adding persistence or ui", "without ui", "no ui", "not adding ui")
	case "persistence":
		return hasAny(normalized,
			"before adding persistence",
			"without persistence",
			"no persistence",
			"not persisted",
			"not persistent",
			"in-memory",
			"in memory",
		)
	case "deployment_readiness":
		return hasAny(normalized,
			"local only",
			"local-only",
			"local and in-memory",
			"local and in memory",
			"outside deployment scope",
			"not deployed",
			"no deployment",
			"without deployment",
		)
	default:
		return false
	}
}

func readinessCorpus(plan map[string]string, growth growthState) string {
	parts := []string{}
	for _, key := range []string{"Product", "Target Users", "MVP", "Build Style", "Non-goals", "Constraints", "Success Criteria", "Current Focus"} {
		parts = append(parts, plan[key])
	}
	for _, pressure := range growth.Pressures {
		parts = append(parts, pressure.Signal, pressure.PressureType, pressure.Effect)
	}
	parts = append(parts, growth.RuntimeBehavior.ValidationSignals...)
	return strings.ToLower(strings.Join(parts, "\n"))
}

func readinessGateForStage(stage string, dimensions []readinessDimension) readinessStageGate {
	current, next, axes, evidence := readinessGateDefinition(stage)
	blocking := []string{}
	dims := readinessDimensionMap(dimensions)
	for _, axis := range axes {
		dim := dims[axis]
		if dim.ID == "" {
			continue
		}
		if dim.Status != "covered" {
			blocking = append(blocking, fmt.Sprintf("%s: %s", dim.Name, dim.Gap))
		}
	}
	status := "ready"
	if len(blocking) > 0 {
		status = "not_ready"
	}
	advancement := readinessStageAdvancement(current, next, status, evidence)
	return readinessStageGate{
		CurrentStage:     current,
		NextStage:        next,
		Status:           status,
		RequiredAxes:     axes,
		BlockingGaps:     blocking,
		RequiredEvidence: evidence,
		Advancement:      advancement,
	}
}

func readinessStageAdvancement(current, next, status string, evidence []string) stageAdvancementPolicy {
	if current == next {
		return stageAdvancementPolicy{
			Candidate:        false,
			Recommendation:   current + " is the current operating stage. Continue with the next focused quality packet instead of advancing stage.",
			PlanChange:       "",
			RequiredEvidence: evidence,
		}
	}
	if status != "ready" {
		return stageAdvancementPolicy{
			Candidate:        false,
			Recommendation:   "Do not advance stage yet. Close the blocking readiness gaps first.",
			PlanChange:       "",
			RequiredEvidence: evidence,
		}
	}
	return stageAdvancementPolicy{
		Candidate:        true,
		Recommendation:   fmt.Sprintf("%s gate is ready. Recommend updating plan.md Current Stage to %s after evidence review.", current, next),
		PlanChange:       "Current Stage -> " + next,
		RequiredEvidence: evidence,
	}
}

func readinessGateDefinition(stage string) (string, string, []string, []string) {
	normalized := normalizeLabel(stage)
	if strings.Contains(normalized, "sustained") {
		return "Sustained Service Quality", "Sustained Service Quality",
			[]string{"validation_coverage", "operations_docs", "maintainability", "sustained_quality"},
			[]string{
				"Active validators, harnesses, or equivalent reusable quality structures continue to pass or have explicit blockers.",
				"Operational handoff, rollback, and recovery notes stay current.",
				"Maintainability evidence shows repeated friction is reduced before feature breadth.",
				"Sustained quality remains protected by repeated runtime evidence and active required behavior.",
			}
	}
	if strings.Contains(normalized, "service") || strings.Contains(normalized, "production") {
		return "Service Quality", "Sustained Service Quality",
			[]string{"validation_coverage", "security_baseline", "deployment_readiness", "operations_docs", "maintainability", "reference_benchmark", "sustained_quality"},
			[]string{
				"Required validation or documented manual checks are repeatable.",
				"Security, privacy, and misuse boundaries are explicit and verified.",
				"Setup, release or run, rollback, and recovery paths are documented and checked.",
				"Maintainability evidence shows the next operator can continue without hidden context.",
				"Reference benchmark evidence shows no core category-baseline gap and at least one above-baseline strength.",
				"Repeated runtime evidence has promoted an active validator, active harness, or equivalent reusable quality structure.",
			}
	}
	if strings.Contains(normalized, "beta") {
		return "Beta", "Service Quality",
			[]string{"validation_coverage", "security_baseline", "deployment_readiness", "operations_docs", "reference_benchmark"},
			[]string{"Primary flows are validated with realistic data.", "Basic security and failure boundaries are handled.", "Demo or deployment path is documented.", "Reference benchmark evidence proves category baseline and one concrete strength."}
	}
	if strings.Contains(normalized, "usable") {
		return "Usable MVP", "Beta",
			[]string{"core_ux", "persistence", "error_handling", "validation_coverage"},
			[]string{"Primary flow is usable without manual data edits.", "Persistence and edge states are proven.", "Primary validation is repeatable."}
	}
	return "Tiny MVP", "Usable MVP",
		[]string{"product_completeness", "core_ux", "validation_coverage"},
		[]string{"Product and MVP slice are measurable.", "One core user flow works locally.", "Minimal validation evidence exists."}
}

func selectReadinessPressure(plan map[string]string, stage string, dimensions []readinessDimension, gate readinessStageGate) readinessPressure {
	dims := readinessDimensionMap(dimensions)
	if readinessPressureShouldFollowGateOrder(gate) {
		for _, axis := range gate.RequiredAxes {
			dim := dims[axis]
			if dim.ID != "" && dim.Status != "covered" {
				return readinessPressureForDimension(plan, stage, dim, gate)
			}
		}
	} else {
		for _, axis := range gate.RequiredAxes {
			dim := dims[axis]
			if dim.ID != "" && dim.Status == "missing" {
				return readinessPressureForDimension(plan, stage, dim, gate)
			}
		}
		for _, axis := range gate.RequiredAxes {
			dim := dims[axis]
			if dim.ID != "" && dim.Status != "covered" {
				return readinessPressureForDimension(plan, stage, dim, gate)
			}
		}
	}
	if gate.CurrentStage == gate.NextStage {
		return readinessPressure{
			Axis:             "sustained_quality",
			AxisName:         "Sustained quality",
			Status:           "ongoing",
			Reason:           gate.CurrentStage + " is active; continue the next focused quality improvement instead of advancing stage.",
			RecommendedGoal:  readinessSustainedOngoingGoal(plan),
			WorkBoundary:     "Stay in sustained operation: reduce one repeated validation, operational, or maintainability friction without broad feature expansion.",
			ValidationSignal: "Run active validators, active harnesses, or the safest equivalent quality check and record the result.",
		}
	}
	for _, dim := range dimensions {
		if dim.Status != "covered" {
			return readinessPressureForDimension(plan, stage, dim, gate)
		}
	}
	return readinessPressureForDimension(plan, stage, dims["maintainability"], gate)
}

func readinessPressureShouldFollowGateOrder(gate readinessStageGate) bool {
	return len(gate.RequiredAxes) > 0
}

func readinessPressureForDimension(plan map[string]string, stage string, dim readinessDimension, gate readinessStageGate) readinessPressure {
	if dim.ID == "" {
		dim = readinessDimension{ID: "maintainability", Name: "Maintainability", Status: "covered", Gap: "Keep the codebase easy to continue."}
	}
	if gate.Advancement.Candidate {
		return readinessPressure{
			Axis:             "stage_advancement",
			AxisName:         "Stage advancement",
			Status:           "candidate",
			Reason:           gate.Advancement.Recommendation,
			RecommendedGoal:  "Review readiness evidence, then run `hyper advance` if the stage change to " + gate.NextStage + " is accepted.",
			WorkBoundary:     "Do not run `hyper advance` until the user accepts the stage advancement.",
			ValidationSignal: "Record the stage advancement decision, or the reason for staying in the current stage.",
		}
	}
	goal := readinessRecommendedGoal(plan, stage, dim.ID)
	workBoundary := "Prioritize the smallest service-readiness step for " + dim.Name + ": " + dim.Gap
	validationSignal := "Capture readiness evidence for " + dim.Name + " in evidence.md."
	if dim.ID == "reference_benchmark" {
		workBoundary = "Compare the current result against 3-5 named category references before adding feature breadth; close only the strongest critical below-baseline gap if one is found."
		validationSignal = "Fill Reference Benchmark Evidence with named references, baseline expectations, current comparison, below-baseline gaps, above-baseline strength, and decision."
	} else if dim.ID == "sustained_quality" {
		workBoundary = "Do not claim sustained quality from one good packet. Repeat the highest-value validation or operational proof until it becomes active required behavior."
		validationSignal = "Record repeated packet evidence and the active validator, active harness, or equivalent reusable quality structure that now protects the service."
	}
	return readinessPressure{
		Axis:             dim.ID,
		AxisName:         dim.Name,
		Status:           dim.Status,
		Reason:           fmt.Sprintf("%s is %s for the %s -> %s gate.", dim.Name, dim.Status, gate.CurrentStage, gate.NextStage),
		RecommendedGoal:  goal,
		WorkBoundary:     workBoundary,
		ValidationSignal: validationSignal,
	}
}

func readinessRecommendedGoal(plan map[string]string, stage, axis string) string {
	product := readinessProductName(plan)
	mvp := firstRuntimeValue(plan["MVP"], plan["Current Focus"], "the primary user flow")
	switch axis {
	case "product_completeness":
		return fmt.Sprintf("Clarify the smallest measurable %s product slice and success criteria.", product)
	case "core_ux":
		return fmt.Sprintf("Implement the smallest usable %s core flow: %s", product, oneLine(mvp))
	case "persistence":
		return fmt.Sprintf("Make the primary %s flow persist real user data across restart or reload.", product)
	case "error_handling":
		return fmt.Sprintf("Handle empty, failure, and edge states for the primary %s flow.", product)
	case "validation_coverage":
		return fmt.Sprintf("Add or run repeatable validation for the primary %s behavior.", product)
	case "security_baseline":
		return fmt.Sprintf("Define and implement the smallest security baseline for %s.", product)
	case "deployment_readiness":
		return fmt.Sprintf("Prove %s can run outside the local development path.", product)
	case "operations_docs":
		return fmt.Sprintf("Document setup, operation, and rollback notes for %s.", product)
	case "maintainability":
		return fmt.Sprintf("Reduce the highest-friction code path so %s can keep growing.", product)
	case "reference_benchmark":
		return fmt.Sprintf("Compare %s against 3-5 named category references, define the baseline, and close the strongest critical below-baseline gap if one exists.", product)
	case "sustained_quality":
		return fmt.Sprintf("Repeat the highest-value %s quality proof until active validation or an equivalent reusable quality structure is justified.", product)
	default:
		return fmt.Sprintf("Advance %s toward %s readiness.", product, stage)
	}
}

func readinessSustainedOngoingGoal(plan map[string]string) string {
	return fmt.Sprintf("Run active quality checks and reduce one small operational, validation, or maintainability friction for %s.", readinessProductName(plan))
}

func readinessProductName(plan map[string]string) string {
	product := firstRuntimeValue(plan["Product"], "the product")
	if before, after, ok := strings.Cut(product, " is "); ok && strings.TrimSpace(before) != "" && strings.TrimSpace(after) != "" {
		if len(strings.Fields(before)) <= 6 {
			return strings.TrimSpace(before)
		}
	}
	for _, sep := range []string{".", ","} {
		if before, _, ok := strings.Cut(product, sep); ok && strings.TrimSpace(before) != "" {
			return strings.TrimSpace(before)
		}
	}
	return product
}

func readinessDimensionMap(dimensions []readinessDimension) map[string]readinessDimension {
	result := map[string]readinessDimension{}
	for _, dim := range dimensions {
		result[dim.ID] = dim
	}
	return result
}

func readinessGateSummary(readiness readinessState) string {
	if readiness.Version == 0 {
		return "not recorded"
	}
	return fmt.Sprintf("%s -> %s (%s)", readiness.StageGate.CurrentStage, readiness.StageGate.NextStage, readiness.StageGate.Status)
}

func readinessPressureSummary(readiness readinessState) string {
	if readiness.Version == 0 || readiness.NextPressure.AxisName == "" {
		return "not selected"
	}
	return fmt.Sprintf("%s: %s", readiness.NextPressure.AxisName, readiness.NextPressure.Reason)
}

func readinessListSummary(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ", ")
}

func materializeReadinessValidatorCandidates(root string, state readinessState) *hyperError {
	for _, spec := range readinessValidatorSpecsForStage(state.Stage) {
		if err := writeReadinessValidatorCandidate(root, spec, state); err != nil {
			return err
		}
	}
	return nil
}

type readinessValidatorSpec struct {
	Name   string
	Axis   string
	Signal string
}

func readinessValidatorSpecsForStage(stage string) []readinessValidatorSpec {
	normalized := normalizeLabel(stage)
	specs := []readinessValidatorSpec{}
	if strings.Contains(normalized, "beta") {
		specs = append(specs,
			readinessValidatorSpec{Name: "validator-beta-primary-flow-smoke", Axis: "validation_coverage", Signal: "Run the repeatable primary-flow smoke test before beta handoff."},
			readinessValidatorSpec{Name: "validator-beta-security-baseline", Axis: "security_baseline", Signal: "Check basic auth, session, token, permission, or abuse boundaries before beta handoff."},
			readinessValidatorSpec{Name: "validator-beta-deploy-check", Axis: "deployment_readiness", Signal: "Verify the demo or deployment path before beta handoff."},
		)
	}
	if strings.Contains(normalized, "service") || strings.Contains(normalized, "production") {
		specs = append(specs,
			readinessValidatorSpec{Name: "validator-service-required-checks", Axis: "validation_coverage", Signal: "Run required service-quality validation before service handoff."},
			readinessValidatorSpec{Name: "validator-service-security-baseline", Axis: "security_baseline", Signal: "Verify security baseline and misuse boundaries before service handoff."},
			readinessValidatorSpec{Name: "validator-service-deploy-rollback", Axis: "deployment_readiness", Signal: "Verify deployment, rollback, and release evidence before service handoff."},
			readinessValidatorSpec{Name: "validator-service-operations-docs", Axis: "operations_docs", Signal: "Verify setup, operations, logs, and rollback docs before service handoff."},
		)
	}
	return specs
}

func writeReadinessValidatorCandidate(root string, spec readinessValidatorSpec, state readinessState) *hyperError {
	activePath := filepath.Join(root, hyperDir, "capabilities", "active", "validator", spec.Name+".md")
	if exists(activePath) {
		return nil
	}
	body := strings.Join([]string{
		"# " + spec.Name,
		"",
		"Status: candidate",
		"Kind: validator",
		"Pressure type: readiness_validator",
		"Stage: " + state.Stage,
		"Axis: " + spec.Axis,
		"",
		"## Reason",
		"",
		"Stage-specific service-quality validator candidate. Do not enforce this validator until it is promoted to active.",
		"",
		"## When Required",
		"",
		"Not required yet. This becomes required only after repeated evidence shows the project needs this validator for " + spec.Axis + ".",
		"",
		"## How To Run",
		"",
		"Run or create the smallest repeatable check that proves: " + spec.Signal,
		"",
		"## Evidence Required",
		"",
		"- Command output, URL, screenshot, or release proof\n- Runtime packet ID\n- Reason if the check is blocked",
		"",
		"## Required Behavior",
		"",
		spec.Signal,
		"",
		"## Pressure",
		"",
		"- Signal: " + spec.Signal,
		"- Stage gate: " + state.StageGate.CurrentStage + " -> " + state.StageGate.NextStage,
		"",
		"## Activation Rule",
		"",
		"Keep this as a quiet candidate until repeated evidence shows the project needs this validator as required behavior.",
		"",
	}, "\n")
	if err := writeIfMissing(filepath.Join(root, hyperDir, "validators", "generated", spec.Name+".md"), body); err != nil {
		return err
	}
	return writeIfMissing(filepath.Join(root, hyperDir, "capabilities", "candidates", "validator", spec.Name+".md"), body)
}
