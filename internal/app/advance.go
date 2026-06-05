package app

import (
	"path/filepath"
	"strings"
)

func advanceHyper(fsys fsRoot) (commandOutput, *hyperError) {
	root := fsys.root()
	statePath := filepath.Join(root, hyperDir, "state.json")
	if !exists(statePath) {
		return commandOutput{}, newError("No Hyper Run state found. Start with `hyper init`.", 2)
	}
	planResult, err := requirePlanForRun(root)
	if err != nil {
		return commandOutput{}, err
	}
	plan := parsePlan(planResult.Body)
	if err := validatePlanStageFields(plan); err != nil {
		return commandOutput{}, err
	}
	if err := ensureProjectLayout(root); err != nil {
		return commandOutput{}, err
	}
	state, err := readState(statePath)
	if err != nil {
		return commandOutput{}, err
	}
	consistency := currentStateConsistency(root, state)
	if consistency.Derived.State == "active" {
		return commandOutput{}, newError("Current runtime packet is still active. Complete or block it before advancing the stage.", 2)
	}
	if !consistency.Consistent && !consistency.Repairable {
		return commandOutput{}, newError("Project state is not ready for stage advancement. Run `hyper doctor` or `hyper repair` first.", 2)
	}

	db, err := openDB(root)
	if err != nil {
		return commandOutput{}, err
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		return commandOutput{}, err
	}
	if err := ensureMemoryFiles(root); err != nil {
		return commandOutput{}, err
	}
	growth, err := updateGrowthState(root, db)
	if err != nil {
		return commandOutput{}, err
	}
	readiness, err := updateReadinessState(root, planResult.Body, growth)
	if err != nil {
		return commandOutput{}, err
	}
	if !readiness.StageGate.Advancement.Candidate {
		return commandOutput{}, newError(stageAdvanceNotReadyMessage(readiness), 2)
	}

	previousStage := readiness.StageGate.CurrentStage
	nextStage := readiness.StageGate.NextStage
	updatedPlan, changed := updatePlanCurrentStage(planResult.Body, nextStage)
	if !changed {
		return commandOutput{}, newError("plan.md already has Current Stage set to "+nextStage+". Run `hyper status` to review the next gate.", 2)
	}
	if err := writeText(filepath.Join(root, planFile), updatedPlan); err != nil {
		return commandOutput{}, err
	}
	updatedReadiness, err := updateReadinessState(root, updatedPlan, growth)
	if err != nil {
		return commandOutput{}, err
	}

	now := nowISO()
	oldStatus := state.Status
	repaired := false
	if consistency.Repairable {
		state.Status = consistency.Derived.State
		repaired = true
		if err := updateRunAndGoalStatus(db, state.ActiveRunID, state.CurrentGoalID, consistency.Derived.State, now); err != nil {
			return commandOutput{}, err
		}
	}
	plan = parsePlan(updatedPlan)
	state.Project = firstNonBlank(readinessProductName(plan), state.Project, "Unknown project")
	state.Stage = nextStage
	state.PlanPath = planFile
	state.PlanHash = hashText(updatedPlan)
	state = applyPlanTargetToState(state, plan)
	state.UpdatedAt = now
	if err := writeJSON(statePath, state); err != nil {
		return commandOutput{}, err
	}

	event := map[string]any{
		"type":                "stage_advanced",
		"run_id":              state.ActiveRunID,
		"goal_id":             state.CurrentGoalID,
		"from_stage":          previousStage,
		"to_stage":            nextStage,
		"plan_change":         "Current Stage -> " + nextStage,
		"state_repaired":      repaired,
		"from_status":         oldStatus,
		"to_status":           state.Status,
		"readiness_gate":      updatedReadiness.StageGate.Status,
		"readiness_next_gate": readinessGateSummary(updatedReadiness),
		"created_at":          nowISO(),
	}
	if err := insertEvent(db, event); err != nil {
		return commandOutput{}, err
	}
	if err := appendJSONL(filepath.Join(root, hyperDir, "logs", "project.jsonl"), event); err != nil {
		return commandOutput{}, err
	}
	if strings.TrimSpace(state.ActiveRunID) != "" {
		if err := appendJSONL(filepath.Join(root, hyperDir, "logs", state.ActiveRunID+".jsonl"), event); err != nil {
			return commandOutput{}, err
		}
	}

	nextPlan, nextErr := writeNextPacketPlan(root, state, consistency.Derived, updatedReadiness, growth)
	if nextErr != nil {
		return commandOutput{}, nextErr
	}

	lines := []string{
		"Hyper Run Stage Advance",
		"",
		"Stage advanced: " + previousStage + " -> " + nextStage,
		"Accepted gate: " + previousStage + " -> " + nextStage + " (ready)",
		"Updated: plan.md Current Stage -> " + nextStage,
		"Plan change: " + readiness.StageGate.Advancement.PlanChange,
		"Required proof covered: " + stageAdvanceRequiredProofSummary(readiness),
		"Run target after advance: " + stageAdvanceRunTargetSummary(state),
	}
	if repaired {
		lines = append(lines, "State repaired: "+firstNonBlank(oldStatus, "unknown")+" -> "+state.Status)
	}
	lines = append(lines,
		"Readiness gate: "+readinessGateSummary(updatedReadiness),
		"Readiness pressure: "+readinessPressureSummary(updatedReadiness),
		"Planned action: "+nextPlan.Action,
		"Next action: "+nextPlan.Command,
		"Why: "+nextPlan.Reason,
		"Continuation guard: "+compactText(nextPacketGuard(state, nextPlan), 220),
	)
	if progressLine := nextPacketProgressGuardLine(state, nextPlan); progressLine != "" {
		lines = append(lines, progressLine)
	}
	lines = append(lines,
		"Next packet plan: "+displayRelPath(hyperDir, "next-packet.md"),
		"",
		"Next:",
		"  "+nextPlan.Command,
	)
	if nextPlan.Command != "hyper status --short" {
		lines = append(lines, "  hyper status --short")
	}
	lines = append(lines, "")
	return stdout(strings.Join(lines, "\n")), nil
}

func stageAdvanceNotReadyMessage(readiness readinessState) string {
	lines := []string{
		"Stage gate is not ready.",
		"Gate: " + readinessGateSummary(readiness),
	}
	if len(readiness.StageGate.BlockingGaps) == 0 {
		lines = append(lines, "Blocking gaps: none recorded, but the advancement candidate is not active.")
	} else {
		lines = append(lines, "Blocking gaps:")
		for _, gap := range readiness.StageGate.BlockingGaps {
			lines = append(lines, "  - "+gap)
		}
	}
	lines = append(lines, "Required proof: "+stageAdvanceRequiredProofSummary(readiness))
	if readiness.StageGate.Advancement.Recommendation != "" {
		lines = append(lines, "Recommendation: "+readiness.StageGate.Advancement.Recommendation)
	}
	if readiness.NextPressure.RecommendedGoal != "" {
		lines = append(lines, "", "Next:", "  hyper run \""+compactText(readiness.NextPressure.RecommendedGoal, 120)+"\"")
	}
	return strings.Join(lines, "\n")
}
