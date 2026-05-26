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
	for _, record := range records {
		if record.Axis == axis && record.Status == "covered" {
			return ""
		}
	}
	return "Add covered readiness evidence for `" + axisName + "` or record a real blocker."
}

func activeCapabilityFinishGateFinding(root, evidenceText string) string {
	validators, err := activeValidatorCapabilities(root)
	if err != nil || len(validators) == 0 {
		return ""
	}
	if hasNonPendingSection(evidenceText, "Active Capability Evidence") {
		return ""
	}
	names := []string{}
	for _, validator := range validators {
		names = append(names, validator.Name)
	}
	return "Record active capability evidence for: " + strings.Join(names, ", ")
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
