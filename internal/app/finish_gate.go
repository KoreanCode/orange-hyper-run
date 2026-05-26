package app

import (
	"fmt"
	"path/filepath"
	"strings"
)

type finishGateResult struct {
	Status   string
	GoalID   string
	Findings []string
	Review   string
}

func runFinishGate(root string, state projectState, derived goalState, readiness readinessState) (finishGateResult, *hyperError) {
	goalID := state.CurrentGoalID
	goalDir := filepath.Join(root, hyperDir, "goals", goalID)
	evidenceText := readIfExists(filepath.Join(goalDir, "evidence.md"))
	nextText := readIfExists(filepath.Join(goalDir, "next.md"))
	result := finishGateResult{Status: "passed", GoalID: goalID}

	if derived.State == "blocked" {
		result.Status = "blocked"
		result.Findings = append(result.Findings, "Packet is closing as blocked: "+derived.Reason)
		result.Review = renderFinishGateReview(result, state, derived, readiness)
		if err := writeText(filepath.Join(goalDir, "review.md"), result.Review); err != nil {
			return result, err
		}
		return result, nil
	}

	if derived.State != "completed" {
		result.Status = "failed"
		result.Findings = append(result.Findings, "Runtime packet is not completed yet: "+derived.Reason)
	}
	if !hasNonPendingSection(evidenceText, "Validation") {
		result.Status = "failed"
		result.Findings = append(result.Findings, "Add concrete command, smoke, browser, or manual validation output under `## Validation`.")
	}
	if !hasNonPendingSection(nextText, "Recommended Next Goal") {
		result.Status = "failed"
		result.Findings = append(result.Findings, "Add the next recommended runtime episode under `## Recommended Next Goal` in `next.md`.")
	}
	if finding := readinessFinishGateFinding(state, evidenceText, readiness); finding != "" {
		result.Status = "failed"
		result.Findings = append(result.Findings, finding)
	}
	if finding := activeCapabilityFinishGateFinding(root, evidenceText); finding != "" {
		result.Status = "failed"
		result.Findings = append(result.Findings, finding)
	}

	result.Review = renderFinishGateReview(result, state, derived, readiness)
	if err := writeText(filepath.Join(goalDir, "review.md"), result.Review); err != nil {
		return result, err
	}
	if result.Status == "failed" {
		return result, newError(finishGateFailureMessage(state, result), 2)
	}
	return result, nil
}

func readinessFinishGateFinding(state projectState, evidenceText string, readiness readinessState) string {
	axis := strings.TrimSpace(readiness.NextPressure.Axis)
	axisName := strings.TrimSpace(readiness.NextPressure.AxisName)
	if axis == "" || axisName == "" || axis == "stage_advancement" || axis == "product_completeness" {
		return ""
	}
	records := readinessEvidenceRecordsFromGoalText(state.CurrentGoalID, evidenceText)
	if axis == "sustained_quality" {
		for _, record := range records {
			if record.Axis == axis {
				return ""
			}
		}
		return "Add sustained quality evidence that records repeated runtime proof or a real blocker."
	}
	for _, record := range records {
		if record.Axis == axis && record.Status == "covered" {
			return ""
		}
	}
	return "Add covered readiness evidence for `" + axisName + "`" + readinessFinishGateHint(axis) + " or record a real blocker."
}

func readinessFinishGateHint(axis string) string {
	switch axis {
	case "core_ux":
		return " (for CLI work, use evidence like `Core UX: CLI smoke passed for the primary run command and verified the expected output.`)"
	case "validation_coverage":
		return " (include the exact command and a passed, verified, or repeatable result)"
	case "error_handling":
		return " (name the empty, error, fallback, or edge state and how it was verified)"
	case "security_baseline":
		return " (name the security/privacy boundary and whether it was documented, verified, or implemented)"
	case "deployment_readiness":
		return " (name the build, artifact, URL, release, or isolated run path that was verified)"
	case "operations_docs":
		return " (name the README, runbook, setup, rollback, or smoke path that was documented)"
	case "reference_benchmark":
		return " (include category, 3-5 references, current comparison, baseline gaps, and decision)"
	case "sustained_quality":
		return " (name the active validator, active harness, or equivalent reusable quality structure)"
	default:
		return ""
	}
}

func activeCapabilityFinishGateFinding(root, evidenceText string) string {
	capabilities, err := activeCapabilities(root)
	if err != nil || len(capabilities) == 0 {
		return ""
	}
	lines := usefulSectionLines(evidenceText, "Active Capability Evidence")
	missing := []string{}
	for _, capability := range capabilities {
		if activeCapabilityEvidenceCovers(capability, lines) {
			continue
		}
		if activeValidatorValidationCovers(capability, evidenceText) {
			continue
		}
		missing = append(missing, capability.Name)
	}
	if len(missing) == 0 {
		return ""
	}
	return "Record active capability evidence for: " + strings.Join(missing, ", ")
}

func activeCapabilityEvidenceCovers(capability activeCapability, lines []string) bool {
	if len(lines) == 0 {
		return false
	}
	name := normalizeSentence(capability.Name)
	command := normalizeSentence(inferredCommandForSignal(capability.Signal))
	for _, line := range lines {
		normalized := normalizeSentence(line)
		if !credibleActiveCapabilityEvidence(normalized) {
			continue
		}
		if name != "" && strings.Contains(normalized, name) {
			return true
		}
		if command != "" && strings.Contains(normalized, command) {
			return true
		}
	}
	return false
}

func activeValidatorValidationCovers(capability activeCapability, evidenceText string) bool {
	if capability.Kind != "validator" {
		return false
	}
	command := normalizeSentence(inferredCommandForSignal(capability.Signal))
	if command == "" {
		return false
	}
	for _, fragment := range validationCommandEvidenceFragments(sectionBody(evidenceText, "Validation"), command) {
		validation := normalizeSentence(fragment)
		if !strings.Contains(validation, command) || !credibleActiveCapabilityEvidence(validation) {
			continue
		}
		if successfulValidationEvidence(validation) {
			return true
		}
	}
	return false
}

func validationCommandEvidenceFragments(body, command string) []string {
	lines := strings.Split(body, "\n")
	fragments := []string{}
	current := []string{}
	sawCommandBoundary := false
	for _, line := range lines {
		if validationCommandBoundary(line) && len(current) > 0 {
			fragments = append(fragments, strings.Join(current, "\n"))
			current = nil
		}
		if strings.TrimSpace(line) != "" || len(current) > 0 {
			current = append(current, line)
		}
		if validationCommandBoundary(line) {
			sawCommandBoundary = true
		}
	}
	if len(current) > 0 {
		fragments = append(fragments, strings.Join(current, "\n"))
	}
	if sawCommandBoundary {
		return fragments
	}
	lineFragments := []string{}
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(strings.TrimLeft(line, "-*0123456789. "))
		if trimmed == "" || isPlaceholder(trimmed) {
			continue
		}
		if strings.Contains(normalizeSentence(trimmed), command) {
			lineFragments = append(lineFragments, trimmed)
		}
	}
	return lineFragments
}

func validationCommandBoundary(line string) bool {
	trimmed := strings.TrimSpace(strings.TrimLeft(line, "-*0123456789. "))
	normalized := strings.ToLower(trimmed)
	return strings.HasPrefix(normalized, "command:") ||
		strings.HasPrefix(normalized, "$ ") ||
		strings.HasPrefix(normalized, "> ") ||
		strings.HasPrefix(normalized, "run:") ||
		strings.HasPrefix(normalized, "check:")
}

func successfulValidationEvidence(normalized string) bool {
	return hasAny(normalized,
		"passed",
		"success",
		"succeeded",
		"verified",
		"checked",
		"covered",
		"proved",
		"proven",
		"built",
		" ok ",
		"ok ./",
	)
}

func credibleActiveCapabilityEvidence(normalized string) bool {
	if normalized == "" || isPlaceholder(normalized) {
		return false
	}
	if explicitActiveCapabilityBlocker(normalized) {
		return true
	}
	if hasAny(normalized, "failed", "failure", "blocked", "warning", "warn") &&
		!hasAny(normalized, "passed", "success", "succeeded", "verified", "checked", "covered", "handled", "proved", "proven", "recovered") {
		return false
	}
	if hasAny(normalized,
		"pending",
		"todo",
		"tbd",
		"not run",
		"not executed",
		"not checked",
		"not verified",
		"not validated",
		"not yet",
		"missing",
	) {
		return false
	}
	return true
}

func explicitActiveCapabilityBlocker(normalized string) bool {
	return hasAny(normalized,
		"blocked because",
		"blocked by",
		"cannot run because",
		"could not run because",
		"unable to run because",
		"missing credential",
		"missing credentials",
		"missing token",
		"missing secret",
		"permission denied",
		"network unavailable",
		"command unavailable",
	)
}

func readinessEvidenceRecordsFromGoalText(goalID, evidenceText string) []readinessEvidenceRecord {
	defs := readinessDimensionDefs()
	records := []readinessEvidenceRecord{}
	for _, line := range usefulSectionLines(evidenceText, "Readiness Evidence") {
		if record, ok := parseReadinessEvidenceLine(goalID, line, defs); ok {
			records = append(records, record)
		}
	}
	for _, line := range usefulSectionLines(evidenceText, "Validation") {
		if record, ok := parseReadinessEvidenceLine(goalID, line, defs); ok {
			records = append(records, record)
			continue
		}
		records = append(records, inferReadinessEvidenceFromValidationLine(goalID, line)...)
	}
	for _, line := range usefulSectionLines(evidenceText, "Surface Proof Evidence") {
		if record, ok := parseReadinessEvidenceLine(goalID, line, defs); ok {
			records = append(records, record)
			continue
		}
		records = append(records, inferReadinessEvidenceFromSurfaceLine(goalID, line)...)
	}
	records = append(records, inferReadinessEvidenceFromReferenceBenchmark(goalID, usefulSectionLines(evidenceText, "Reference Benchmark Evidence"))...)
	return records
}

func renderFinishGateReview(result finishGateResult, state projectState, derived goalState, readiness readinessState) string {
	findings := "- None."
	if len(result.Findings) > 0 {
		lines := []string{}
		for _, finding := range result.Findings {
			lines = append(lines, "- "+finding)
		}
		findings = strings.Join(lines, "\n")
	}
	return strings.Join([]string{
		"# " + state.CurrentGoalID + " Review",
		"",
		"## Finish Gate",
		"",
		"Status: " + result.Status,
		"Runtime packet state: " + derived.State,
		"Reason: " + derived.Reason,
		"Readiness gate: " + readinessGateSummary(readiness),
		"",
		"## Findings",
		"",
		findings,
		"",
		"## Return Path",
		"",
		"Stay in the same runtime packet. Update `evidence.md` and `next.md`, then run `hyper complete` again.",
		"",
	}, "\n")
}

func finishGateFailureMessage(state projectState, result finishGateResult) string {
	goalDir := strings.TrimSuffix(state.CurrentGoalPath, "goal.md")
	if goalDir == "" {
		goalDir = fmt.Sprintf(".hyper/goals/%s/", state.CurrentGoalID)
	}
	lines := []string{
		"Finish gate failed for " + state.CurrentGoalID + ".",
		"",
		"Findings:",
	}
	for _, finding := range result.Findings {
		lines = append(lines, "  - "+finding)
	}
	lines = append(lines,
		"",
		"Review file: "+goalDir+"review.md",
		"",
		"Next:",
		"  update "+goalDir+"evidence.md",
		"  update "+goalDir+"next.md",
		"  hyper complete",
	)
	return strings.Join(lines, "\n")
}
