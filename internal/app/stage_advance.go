package app

import "strings"

func stageAdvancementReviewLines(readiness readinessState) []string {
	if readiness.Version == 0 {
		return []string{
			"## Stage Advancement Review",
			"",
			"- Gate: not recorded",
			"- Decision: do not advance without current readiness evidence.",
		}
	}
	lines := []string{
		"## Stage Advancement Review",
		"",
		"- Current stage: " + readiness.StageGate.CurrentStage,
		"- Recommended next stage: " + readiness.StageGate.NextStage,
		"- Plan change: " + firstNonBlank(readiness.StageGate.Advancement.PlanChange, "none"),
		"- Required proof covered: " + stageAdvanceRequiredProofSummary(readiness),
		"- Blocking gaps: " + stageAdvanceBlockingGapSummary(readiness),
	}
	if readiness.StageGate.Advancement.Candidate {
		lines = append(lines, "- User decision required: accept before running `hyper advance`.")
	} else {
		lines = append(lines, "- User decision required: keep working until blocking gaps are closed.")
	}
	return lines
}

func stageAdvanceRequiredProofSummary(readiness readinessState) string {
	if len(readiness.StageGate.RequiredAxes) == 0 {
		return "none"
	}
	dims := readinessDimensionMap(readiness.Dimensions)
	names := []string{}
	for _, axis := range readiness.StageGate.RequiredAxes {
		dim := dims[axis]
		name := firstNonBlank(dim.Name, axis)
		status := firstNonBlank(dim.Status, "unknown")
		names = append(names, name+" ("+status+")")
	}
	return strings.Join(names, ", ")
}

func stageAdvanceBlockingGapSummary(readiness readinessState) string {
	if len(readiness.StageGate.BlockingGaps) == 0 {
		return "none"
	}
	return strings.Join(readiness.StageGate.BlockingGaps, "; ")
}

func stageAdvanceRunTargetSummary(state projectState) string {
	if strings.TrimSpace(state.RunUntil) == "" {
		return "single packet"
	}
	if strings.TrimSpace(state.RunTargetSource) == "" {
		return state.RunUntil
	}
	return state.RunUntil + " (" + state.RunTargetSource + ")"
}
