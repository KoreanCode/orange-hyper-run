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
		{ID: "security_baseline", Name: "Security baseline", Keywords: []string{"security", "permission", "rate limit", "secret", "session", "token"}, Gap: "Basic security and misuse boundaries are not yet explicit."},
		{ID: "deployment_readiness", Name: "Deployment readiness", Keywords: []string{"deploy", "release", "production", "server", "docker", "ci", "hosted"}, Gap: "The project is not yet proven runnable outside the local development path."},
		{ID: "operations_docs", Name: "Operations and docs", Keywords: []string{"readme", "docs", "runbook", "rollback", "logs", "monitor", "environment"}, Gap: "Operational notes, setup, rollback, or handoff docs are not sufficient."},
		{ID: "maintainability", Name: "Maintainability", Keywords: []string{"refactor", "cleanup", "component", "module", "architecture", "helper", "table-driven"}, Gap: "The codebase has not accumulated enough maintainability evidence."},
	}
}

func readinessDimensionStatus(def readinessDimensionDef, plan map[string]string, growth growthState, evidenceRecords []readinessEvidenceRecord, corpus string) (string, int, string) {
	record, hasRecord := readinessEvidenceForAxis(evidenceRecords, def.ID)
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
	if hasAny(corpus, def.Keywords...) {
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

func readinessEvidenceRecordForAxis(goalID, axis, text string) readinessEvidenceRecord {
	covered, quality := readinessEvidenceQuality(axis, text)
	status := "emerging"
	if covered {
		status = "covered"
	}
	return readinessEvidenceRecord{Axis: axis, GoalID: goalID, Text: text, Status: status, Quality: quality}
}

func usefulReadinessEvidence(text string) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	if normalized == "" || isPlaceholder(normalized) {
		return false
	}
	return !hasAny(normalized, "not yet", "not enough", "missing", "blocked", "failed", "cannot", "could not", "unable")
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
		return hasAny(normalized, "product", "mvp", "slice") &&
				hasAny(normalized, "success", "criteria", "target", "measurable", "defined"),
			"measurable product, MVP, target, or success criteria"
	case "core_ux":
		return hasAny(normalized, "smoke", "screenshot", "browser", "verified", "passed") &&
				hasAny(normalized, "flow", "click", "create", "add", "edit", "complete", "delete", "send", "navigate", "reload"),
			"browser, screenshot, smoke, or verified primary-flow evidence"
	case "persistence":
		return hasAny(normalized, "persist", "reload", "restart", "saved", "survive", "stored", "created", "re-read", "reread", "confirmed", "row") &&
				hasAny(normalized, "sqlite", "mysql", "postgres", "postgresql", "database", " db ", "db check", "sql", "localstorage", "local storage", "storage"),
			"MySQL, SQLite, DB, reload, restart, storage, or database evidence"
	case "error_handling":
		return hasAny(normalized, "empty", "error", "loading", "fallback", "failure", "edge") &&
				hasAny(normalized, "handled", "covered", "verified", "tested", "implemented", "works"),
			"empty, loading, error, failure, fallback, or edge-state evidence"
	case "validation_coverage":
		return hasAny(normalized, "smoke", "playwright", "go test", "npm run", "pytest", "build", "command", "validation", "`") &&
				hasAny(normalized, "passed", "repeatable", "covered", "verified"),
			"repeatable command, build, test, smoke, or coverage evidence"
	case "security_baseline":
		return hasAny(normalized, "security", "permission", "rate limit", "secret", "session", "token", "auth", "abuse") &&
				hasAny(normalized, "documented", "verified", "implemented", "checked", "covered"),
			"security, permission, token, session, rate-limit, or abuse-boundary evidence"
	case "deployment_readiness":
		return hasAny(normalized, "deploy", "deployed", "url", "https://", "http://", "build", "release", "hosted", "docker", "ci") &&
				hasAny(normalized, "passed", "available", "hosted", "deployed", "built", "released", "verified"),
			"deploy, hosted URL, release, build, Docker, or CI evidence"
	case "operations_docs":
		return hasAny(normalized, "readme", "docs", "setup", "runbook", "rollback", "logs", "monitor", "environment") &&
				hasAny(normalized, "documented", "updated", "verified", "covered", "written"),
			"README, docs, setup, runbook, rollback, logs, or environment evidence"
	case "maintainability":
		return hasAny(normalized, "refactor", "cleanup", "component", "module", "architecture", "helper", "table-driven", "test", "extracted", "reduced", "documented"),
			"refactor, modularity, test, helper, cleanup, or architecture evidence"
	default:
		return len(strings.Fields(normalized)) >= 4, "specific evidence for this readiness axis"
	}
}

func growthEvidenceForDimension(growth growthState, def readinessDimensionDef) (bool, bool, string) {
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
		if hasAny(strings.ToLower(signal), def.Keywords...) {
			return true, true, "Active runtime behavior references this readiness axis."
		}
	}
	return false, false, ""
}

func pressureMatchesReadiness(pressure growthPressure, def readinessDimensionDef) bool {
	signal := strings.ToLower(pressure.Signal + " " + pressure.PressureType + " " + pressure.Effect)
	switch def.ID {
	case "validation_coverage":
		return pressure.Effect == "validation" || hasAny(signal, def.Keywords...)
	case "error_handling":
		return pressure.Effect == "stop_condition" || hasAny(signal, def.Keywords...)
	case "maintainability":
		return pressure.Effect == "implementation" || hasAny(signal, def.Keywords...)
	default:
		return hasAny(signal, def.Keywords...)
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
	if strings.Contains(normalized, "service") || strings.Contains(normalized, "production") {
		return "Service Quality", "Sustained Service Quality",
			[]string{"validation_coverage", "security_baseline", "deployment_readiness", "operations_docs", "maintainability"},
			[]string{"Required validation passes.", "Security and misuse boundaries are explicit.", "Deployment, rollback, and operational handoff are documented."}
	}
	if strings.Contains(normalized, "beta") {
		return "Beta", "Service Quality",
			[]string{"validation_coverage", "security_baseline", "deployment_readiness", "operations_docs"},
			[]string{"Primary flows are validated with realistic data.", "Basic security and failure boundaries are handled.", "Demo or deployment path is documented."}
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
	for _, dim := range dimensions {
		if dim.Status != "covered" {
			return readinessPressureForDimension(plan, stage, dim, gate)
		}
	}
	return readinessPressureForDimension(plan, stage, dims["maintainability"], gate)
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
			RecommendedGoal:  "Review readiness evidence and update plan.md Current Stage to " + gate.NextStage + " if the evidence is accepted.",
			WorkBoundary:     "Do not auto-edit plan.md. Review evidence first, then update Current Stage only when the user accepts the stage advancement.",
			ValidationSignal: "Record the stage advancement decision or reason for staying in the current stage.",
		}
	}
	goal := readinessRecommendedGoal(plan, stage, dim.ID)
	return readinessPressure{
		Axis:             dim.ID,
		AxisName:         dim.Name,
		Status:           dim.Status,
		Reason:           fmt.Sprintf("%s is %s for the %s -> %s gate.", dim.Name, dim.Status, gate.CurrentStage, gate.NextStage),
		RecommendedGoal:  goal,
		WorkBoundary:     "Prioritize the smallest service-readiness step for " + dim.Name + ": " + dim.Gap,
		ValidationSignal: "Capture readiness evidence for " + dim.Name + " in evidence.md.",
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
	default:
		return fmt.Sprintf("Advance %s toward %s readiness.", product, stage)
	}
}

func readinessProductName(plan map[string]string) string {
	product := firstRuntimeValue(plan["Product"], "the product")
	if before, after, ok := strings.Cut(product, " is "); ok && strings.TrimSpace(before) != "" && strings.TrimSpace(after) != "" {
		if len(strings.Fields(before)) <= 4 {
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
