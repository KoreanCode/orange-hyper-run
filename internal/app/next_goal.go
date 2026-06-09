package app

import (
	"path/filepath"
	"strings"
)

func readinessWithPacketNextGoal(root string, state projectState, derived goalState, readiness readinessState) readinessState {
	if !packetNextGoalCanGuide(state, derived, readiness) {
		return readiness
	}
	goal, ok := packetRecommendedNextGoal(root, state)
	if !ok {
		return readiness
	}
	readiness.NextPressure.RecommendedGoal = goal
	readiness.NextPressure.Reason = packetNextGoalReason(root, state, goal, readiness)
	return readiness
}

func packetNextGoalCanGuide(state projectState, derived goalState, readiness readinessState) bool {
	if strings.TrimSpace(state.CurrentGoalID) == "" || derived.State != "completed" {
		return false
	}
	if readiness.NextPressure.Axis == "stage_advancement" || readiness.StageGate.Advancement.Candidate {
		return false
	}
	if state.AutoContinue && runUntilReached(state, readiness) {
		return false
	}
	if strings.TrimSpace(readiness.NextPressure.RecommendedGoal) == "" {
		return true
	}
	if sustainedQualityOngoing(readiness) {
		return true
	}
	return genericReadinessRecommendedGoal(readiness.NextPressure.RecommendedGoal)
}

func sustainedQualityOngoing(readiness readinessState) bool {
	return readiness.NextPressure.Axis == "sustained_quality" &&
		readiness.NextPressure.Status == "ongoing" &&
		readiness.StageGate.CurrentStage == readiness.StageGate.NextStage
}

func genericReadinessRecommendedGoal(goal string) bool {
	normalized := normalizeSentence(goal)
	return hasAny(normalized,
		"run active quality checks",
		"reduce one small operational",
		"reduce one repeated validation",
		"next focused quality improvement",
		"continue the next focused quality",
	)
}

func packetRecommendedNextGoal(root string, state projectState) (string, bool) {
	nextText := readIfExists(filepath.Join(root, hyperDir, "goals", state.CurrentGoalID, "next.md"))
	nextGoal := recommendedNextGoalFromText(nextText)
	if surfaceProofGapFromCurrentPacket(root, state) {
		if surfaceProofNextGoal(nextGoal) {
			return nextGoal, true
		}
		return "Create an allowed visual/accessibility surface proof for the primary user flow, or add a small project-owned browser harness that proves the same surface states.", true
	}
	if actionableRecommendedNextGoal(nextGoal) {
		return nextGoal, true
	}
	return "", false
}

func recommendedNextGoalFromText(text string) string {
	lines := usefulSectionLines(text, "Recommended Next Goal")
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(oneLine(strings.Join(lines, " ")))
}

func actionableRecommendedNextGoal(goal string) bool {
	normalized := normalizeSentence(goal)
	if normalized == "" || isPlaceholder(goal) {
		return false
	}
	return !hasAny(normalized,
		"review stage advancement",
		"review stage readiness",
		"review target proof",
		"review sustained quality advancement",
	)
}

func surfaceProofNextGoal(goal string) bool {
	normalized := normalizeSentence(goal)
	return normalized != "" &&
		hasAny(normalized, "surface proof", "visual", "accessibility", "a11y", "browser", "screenshot") &&
		hasAny(normalized, "create", "add", "run", "resolve", "capture", "prove", "proof", "harness")
}

func surfaceProofGapFromCurrentPacket(root string, state projectState) bool {
	if strings.TrimSpace(state.CurrentGoalID) == "" {
		return false
	}
	goalDir := filepath.Join(root, hyperDir, "goals", state.CurrentGoalID)
	text := strings.Join([]string{
		sectionBody(readIfExists(filepath.Join(goalDir, "evidence.md")), "Surface Proof Evidence"),
		sectionBody(readIfExists(filepath.Join(goalDir, "evidence.md")), "Pressure Signals"),
		sectionBody(readIfExists(filepath.Join(goalDir, "evidence.md")), "Blocker"),
		sectionBody(readIfExists(filepath.Join(goalDir, "next.md")), "Recommended Next Goal"),
		sectionBody(readIfExists(filepath.Join(goalDir, "next.md")), "Learn Notes"),
	}, "\n")
	normalized := normalizeSentence(text)
	if !hasAny(normalized, "surface proof", "browser", "screenshot", "visual", "accessibility", "a11y") {
		return false
	}
	return hasAny(normalized,
		"gap",
		"blocked",
		"policy",
		"not captured",
		"not verified",
		"still needs",
		"future packets need",
		"need an allowed",
	)
}

func packetNextGoalReason(root string, state projectState, goal string, readiness readinessState) string {
	if surfaceProofGapFromCurrentPacket(root, state) {
		nextText := readIfExists(filepath.Join(root, hyperDir, "goals", state.CurrentGoalID, "next.md"))
		if surfaceProofNextGoal(recommendedNextGoalFromText(nextText)) {
			return "The last packet left a surface-proof gap; run the concrete next.md surface-proof recommendation before general sustained-quality work."
		}
		return "The last packet left a surface-proof gap; prioritize an allowed visual/accessibility proof before the broader next.md recommendation or generic sustained-quality work."
	}
	return "Use the completed packet's next.md recommendation as the next concrete runtime episode before falling back to generic readiness pressure."
}
