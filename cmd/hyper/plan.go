package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func requirePlanForRun(root string) (planResult, *hyperError) {
	path := filepath.Join(root, planFile)
	if exists(path) {
		body, err := os.ReadFile(path)
		if err != nil {
			return planResult{}, ioError(err)
		}
		return planResult{Body: string(body)}, nil
	}
	return planResult{}, newError("Missing plan.md.\n\nInitialize Hyper Run in this project first:\n  hyper init\n\n"+planTemplate(), 2)
}

func ensurePlanForInit(root string) (planResult, *hyperError) {
	path := filepath.Join(root, planFile)
	if exists(path) {
		body, err := os.ReadFile(path)
		if err != nil {
			return planResult{}, ioError(err)
		}
		return planResult{Body: string(body)}, nil
	}
	body := planTemplate()
	if err := writeText(path, body); err != nil {
		return planResult{}, err
	}
	return planResult{Body: body, Created: true}, nil
}

func planTemplate() string {
	return `# Product Plan

## Product

## Target Users

## MVP

## Current Stage

Tiny MVP

## Build Style

## Non-goals

## Constraints

## Success Criteria

## Current Focus

`
}

func parsePlan(body string) map[string]string {
	result := map[string]string{}
	current := ""
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "## ") {
			current = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			result[current] = ""
			continue
		}
		if current != "" {
			existing := result[current]
			if existing != "" {
				existing += "\n"
			}
			result[current] = strings.TrimSpace(existing + line)
		}
	}
	return result
}

func compileGoalEpisode(goalID, focus, planBody string, similar []similarContext, growth growthState, readiness readinessState) episode {
	plan := parsePlan(planBody)
	stage := normalizeRuntimeStage(firstRuntimeValue(plan["Current Stage"], "Tiny MVP"))
	buildStyle := firstRuntimeValue(plan["Build Style"], "Detect from project")
	product := readinessProductName(plan)
	objective := runtimeObjective(focus, plan, stage, product, readiness)
	validation := applyReadinessValidation(applyGrowthValidation(applyStageValidation(validationForBuildStyle(buildStyle), stage), growth), readiness)
	stopCondition := applyReadinessStopConditions(applyGrowthStopConditions(firstRuntimeValue(plan["Success Criteria"], stageDoneCondition(stage)), growth), readiness)
	scope := runtimeWorkBoundary(objective, stage, plan, growth, readiness)
	nonGoals := firstRuntimeValue(plan["Non-goals"], "No explicit non-goals recorded in plan.md.")
	docs := episodeDocs{
		Goal:     buildGoalDoc(goalID, objective, focus, plan, stage, buildStyle, scope, validation, stopCondition, similar, readiness),
		Tasks:    buildTasksDoc(goalID, buildStyle, readiness),
		Evidence: buildEvidenceDoc(goalID, readiness),
		Review:   fmt.Sprintf("# %s Review\n\n## Result\n\nPending.\n\n## Issues\n\nPending.\n", goalID),
		Next:     fmt.Sprintf("# %s Next\n\n## Recommended Next Goal\n\nPending.\n\n## Learn Notes\n\n- Decision: Pending.\n- Pattern: Pending.\n- Constraint: Pending.\n- Failure: Pending.\n", goalID),
	}
	return episode{
		Plan:          plan,
		Stage:         stage,
		BuildStyle:    buildStyle,
		Objective:     objective,
		Scope:         scope,
		NonGoals:      nonGoals,
		Validation:    validation,
		StopCondition: stopCondition,
		Docs:          docs,
	}
}

func runtimeObjective(focus string, plan map[string]string, stage, product string, readiness readinessState) string {
	focus = strings.TrimSpace(focus)
	readinessGoal := strings.TrimSpace(readiness.NextPressure.RecommendedGoal)
	if focus != "" && readinessGoal != "" && broadRuntimeFocus(focus) {
		return fmt.Sprintf("Translate `%s` into the smallest %s step for %s: %s", oneLine(focus), stage, readiness.NextPressure.AxisName, readinessGoal)
	}
	if focus != "" {
		return focus
	}
	return firstRuntimeValue(readinessGoal, plan["Current Focus"], fmt.Sprintf("Advance %s for %s", stage, product))
}

func broadRuntimeFocus(focus string) bool {
	normalized := normalizeLabel(focus)
	return hasAny(normalized,
		"service", "production", "quality", "harden", "upgrade", "improve", "polish", "complete", "finish", "better",
		"실서비스", "서비스", "품질", "고도화", "업그레이드", "완성", "개선", "베타", "프로덕션",
	)
}

func firstRuntimeValue(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" && !isPlaceholder(trimmed) {
			return trimmed
		}
	}
	return ""
}

func isPlaceholder(value string) bool {
	normalized := normalizeSentence(value)
	switch normalized {
	case "tbd", "todo", "n/a", "na", "none", "pending":
		return true
	}
	return isNoIssueText(normalized)
}

func isNoIssueText(normalized string) bool {
	if normalized == "" {
		return true
	}
	if normalized == "none in this episode" ||
		normalized == "no blocker" ||
		normalized == "no blockers" ||
		normalized == "no blocker in this episode" ||
		normalized == "no blockers in this episode" ||
		normalized == "no runtime blocker" ||
		normalized == "no runtime blockers" ||
		normalized == "no failure" ||
		normalized == "no failures" ||
		normalized == "no failure in this episode" ||
		normalized == "no failures in this episode" {
		return true
	}
	return strings.HasPrefix(normalized, "no blocker for this episode") ||
		strings.HasPrefix(normalized, "no blockers for this episode") ||
		strings.HasPrefix(normalized, "no runtime blocker was found") ||
		strings.HasPrefix(normalized, "no failure for this episode") ||
		strings.HasPrefix(normalized, "no failures for this episode")
}

func normalizeRuntimeStage(stage string) string {
	stage = firstRuntimeValue(stage)
	if stage == "" {
		return ""
	}
	normalized := normalizeLabel(stage)
	type stagePattern struct {
		name     string
		patterns []string
	}
	patterns := []stagePattern{
		{name: "Tiny MVP", patterns: []string{"tiny mvp"}},
		{name: "Usable MVP", patterns: []string{"usable mvp"}},
		{name: "Beta", patterns: []string{"beta"}},
		{name: "Service Quality", patterns: []string{"service quality", "production"}},
	}
	bestName := ""
	bestIndex := len(normalized) + 1
	for _, candidate := range patterns {
		for _, pattern := range candidate.patterns {
			index := strings.Index(normalized, pattern)
			if index >= 0 && index < bestIndex {
				bestIndex = index
				bestName = candidate.name
			}
		}
	}
	if bestName != "" {
		return bestName
	}
	return stage
}

func runtimeWorkBoundary(objective, stage string, plan map[string]string, growth growthState, readiness readinessState) string {
	mvp := firstRuntimeValue(plan["MVP"])
	nonGoals := firstRuntimeValue(plan["Non-goals"])
	constraints := firstRuntimeValue(plan["Constraints"])
	lines := []string{
		"- Do the smallest coherent implementation step that advances: " + compactText(objective, 180),
		"- Keep the work inside the current stage: " + stage,
		"- Stage contract: " + compactText(stageGrowthContract(stage), 180),
	}
	if guidance := stageRuntimeBoundary(stage); guidance != "" {
		lines = append(lines, "- "+guidance)
	}
	if mvp != "" {
		lines = append(lines, "- Use the product MVP brief as the boundary: "+compactText(mvp, 180))
	}
	if nonGoals != "" {
		lines = append(lines, "- Avoid plan non-goals: "+compactText(nonGoals, 160))
	}
	if constraints != "" {
		lines = append(lines, "- Respect constraints: "+compactText(constraints, 200))
	}
	lines = append(lines, growth.RuntimeBehavior.WorkBoundary...)
	if readiness.NextPressure.WorkBoundary != "" {
		lines = append(lines, "- "+readiness.NextPressure.WorkBoundary)
	}
	if readiness.NextPressure.Axis == "product_completeness" || readinessGateHasAxis(readiness, "product_completeness") {
		lines = append(lines, "- If the product brief is incomplete, inspect the current project and choose the smallest reversible step.")
	}
	if len(lines) == 2 {
		lines = append(lines, "- If the product brief is incomplete, inspect the current project and choose the smallest reversible step.")
	}
	return strings.Join(lines, "\n")
}

func readinessGateHasAxis(readiness readinessState, axis string) bool {
	for _, required := range readiness.StageGate.RequiredAxes {
		if required == axis {
			for _, dim := range readiness.Dimensions {
				if dim.ID == axis && dim.Status != "covered" {
					return true
				}
			}
		}
	}
	return false
}

func applyGrowthValidation(base string, growth growthState) string {
	if len(growth.RuntimeBehavior.ValidationSignals) == 0 {
		return base
	}
	return base + "\n" + strings.Join(growth.RuntimeBehavior.ValidationSignals, "\n")
}

func applyReadinessValidation(base string, readiness readinessState) string {
	if readiness.NextPressure.ValidationSignal == "" {
		return base
	}
	return base + "\n- " + readiness.NextPressure.ValidationSignal
}

func applyStageValidation(base, stage string) string {
	if signal := stageValidationSignal(stage); signal != "" {
		return base + "\n- " + signal
	}
	return base
}

func applyGrowthStopConditions(base string, growth growthState) string {
	if len(growth.RuntimeBehavior.StopConditions) == 0 {
		return base
	}
	return base + "\n" + strings.Join(growth.RuntimeBehavior.StopConditions, "\n")
}

func applyReadinessStopConditions(base string, readiness readinessState) string {
	if readiness.StageGate.Advancement.Candidate {
		return base + "\n- Do not edit plan.md Current Stage until the user accepts the stage advancement recommendation."
	}
	if readiness.StageGate.Status == "not_ready" {
		return base + "\n- Do not advance from " + readiness.StageGate.CurrentStage + " to " + readiness.StageGate.NextStage + " until stage gate evidence is captured."
	}
	return base
}

func stageDoneCondition(stage string) string {
	normalized := normalizeLabel(stage)
	if strings.Contains(normalized, "tiny") && strings.Contains(normalized, "mvp") {
		return strings.Join([]string{
			"- One core user flow works locally.",
			"- The project can be run from documented commands.",
			"- Minimal validation has passed or blockers are documented.",
			"- evidence.md and next.md are updated.",
		}, "\n")
	}
	if strings.Contains(normalized, "usable") && strings.Contains(normalized, "mvp") {
		return strings.Join([]string{
			"- Core flow is usable without manual data edits.",
			"- Empty, loading, and error states are handled for the primary path.",
			"- Validation covers the primary path.",
			"- next.md identifies the highest-value polish or beta-readiness step.",
		}, "\n")
	}
	if strings.Contains(normalized, "beta") {
		return "- Primary flows are validated against realistic data.\n- Known blockers are documented with owner or next action.\n- Release or demo readiness evidence is captured."
	}
	if strings.Contains(normalized, "service") || strings.Contains(normalized, "production") {
		return "- Required production checks pass.\n- Operational risks and rollback notes are documented.\n- Deployment or release evidence is captured."
	}
	return "- The current stage objective is advanced.\n- Validation has passed or blockers are documented.\n- evidence.md and next.md are updated."
}

func stageRuntimeBoundary(stage string) string {
	normalized := normalizeLabel(stage)
	if strings.Contains(normalized, "tiny") && strings.Contains(normalized, "mvp") {
		return "Prefer one reversible product slice over abstractions, broad polish, or generated harnesses."
	}
	if strings.Contains(normalized, "usable") && strings.Contains(normalized, "mvp") {
		return "Strengthen the primary flow with persistence, edge states, and repeatable validation before adding breadth."
	}
	if strings.Contains(normalized, "beta") {
		return "Prioritize realistic data, reliability, security, deployment, and documentation gaps over new feature breadth."
	}
	if strings.Contains(normalized, "service") || strings.Contains(normalized, "production") {
		return "Treat operations, security, rollback, required validation, and maintainability evidence as part of the implementation."
	}
	return ""
}

func stageValidationSignal(stage string) string {
	normalized := normalizeLabel(stage)
	if strings.Contains(normalized, "tiny") && strings.Contains(normalized, "mvp") {
		return "Tiny MVP validation may be a narrow local build or smoke pass, but it must prove one useful flow."
	}
	if strings.Contains(normalized, "usable") && strings.Contains(normalized, "mvp") {
		return "Usable MVP validation must cover the primary flow plus persistence or edge-state evidence touched by this packet."
	}
	if strings.Contains(normalized, "beta") {
		return "Beta validation should use realistic data and capture security, deployment, or docs evidence when those axes are touched."
	}
	if strings.Contains(normalized, "service") || strings.Contains(normalized, "production") {
		return "Service Quality validation should run required validators when active, or record why each required check is blocked."
	}
	return ""
}

func validationForBuildStyle(buildStyle string) string {
	normalized := normalizeLabel(buildStyle)
	common := "- Detect and run the safest available build, test, or smoke command."
	if hasAny(normalized, "web", "local app", "game") {
		return common + "\n- If a browser UI exists, capture screenshot evidence and check console errors.\n- If a dev server is required, document the URL and verification steps."
	}
	if hasAny(normalized, "cli", "automation") {
		return common + "\n- Run the smallest representative command or dry-run path.\n- Capture command output in evidence.md."
	}
	if strings.Contains(normalized, "api") {
		return common + "\n- Run a representative endpoint, contract, or smoke check when available.\n- Capture request/response or command evidence."
	}
	if strings.Contains(normalized, "desktop") {
		return common + "\n- Capture launch, build, or manual verification evidence when automated checks are unavailable."
	}
	return common + "\n- If validation cannot run, document the blocker in evidence.md."
}

func buildGoalDoc(goalID, objective, focus string, plan map[string]string, stage, buildStyle, workBoundary, validation, stopCondition string, similar []similarContext, readiness readinessState) string {
	currentFocus := firstRuntimeValue(strings.TrimSpace(focus), plan["Current Focus"], "Continue the current stage.")
	product := firstRuntimeValue(plan["Product"], "the current project")
	targetUsers := firstRuntimeValue(plan["Target Users"], "Not specified yet.")
	stageContract := compactText(stageGrowthContract(stage), 180)
	objective = compactText(objective, 240)
	product = compactText(product, 180)
	targetUsers = compactText(targetUsers, 180)
	buildStyle = compactText(buildStyle, 160)
	currentFocus = compactText(currentFocus, 160)
	workBoundary = compactMultiline(workBoundary, 10, 180)
	validation = compactMultiline(validation, 8, 180)
	stopCondition = compactMultiline(stopCondition, 8, 180)
	return fmt.Sprintf(`# %s Runtime Packet

## Continue From

%s

## Current Episode

%s

## Why Now

- Product: %s
- Stage: %s
- Stage contract: %s
- Target users: %s
- Runtime protocol: %s
- Growth loop: %s
- This packet exists to turn observed project state into the next required structure, not to freeze a long-lived SPEC.

## Runtime Inputs

- Build style: %s
- Current focus: %s

## Stage Gate

%s

## Growth Principles

%s

## Work Boundary

%s

## Validation Signals

%s

## Evidence Required

- Command output or reason validation could not run
- Readiness evidence in axis-slot format, for example "%s: proof"
- Active capability evidence when required validators are present
- Changed file summary
- Decisions that should persist into future runs
- Reusable patterns that should guide similar future work
- Repeated need, failure, or proof that should influence future structure
- Blocker, constraint, or failure signal when applicable
- Screenshot path when applicable

## Stop When

%s
`, goalID, runtimeContinuation(similar), objective, product, stage, stageContract, targetUsers, runtimeProtocolDefinition, growthLoopDefinition, buildStyle, currentFocus, buildStageGateDoc(readiness), formatGrowthPrinciples(), workBoundary, validation, readinessEvidenceExampleAxis(readiness), stopCondition)
}

func formatGrowthPrinciples() string {
	lines := []string{}
	for _, principle := range growthPrinciples() {
		lines = append(lines, "- "+principle)
	}
	return strings.Join(lines, "\n")
}

func buildStageGateDoc(readiness readinessState) string {
	if readiness.Version == 0 {
		return "- Readiness state has not been recorded yet."
	}
	lines := []string{
		"- Current gate: " + readiness.StageGate.CurrentStage + " -> " + readiness.StageGate.NextStage,
		"- Gate status: " + readiness.StageGate.Status,
		"- Next readiness pressure: " + readiness.NextPressure.AxisName + " (" + readiness.NextPressure.Status + ")",
		"- Pressure reason: " + readiness.NextPressure.Reason,
	}
	if readiness.StageGate.Advancement.Candidate {
		lines = append(lines, "- Stage advancement candidate: "+readiness.StageGate.Advancement.Recommendation)
		lines = append(lines, "- Stage advancement plan change: "+readiness.StageGate.Advancement.PlanChange)
	} else if readiness.StageGate.Advancement.Recommendation != "" {
		lines = append(lines, "- Stage advancement candidate: no. "+readiness.StageGate.Advancement.Recommendation)
	}
	if len(readiness.StageGate.BlockingGaps) > 0 {
		for _, gap := range readiness.StageGate.BlockingGaps {
			lines = append(lines, "- Gate gap: "+compactText(gap, 160))
		}
	}
	for _, evidence := range readiness.StageGate.RequiredEvidence {
		lines = append(lines, "- Gate evidence: "+compactText(evidence, 160))
	}
	return strings.Join(lines, "\n")
}

func buildTasksDoc(goalID, buildStyle string, readiness readinessState) string {
	browserTask := ""
	if hasAny(normalizeLabel(buildStyle), "web", "local app", "game") {
		browserTask = "- [ ] Capture browser screenshot and console evidence when UI changes are made\n"
	}
	readinessTask := ""
	if readiness.NextPressure.AxisName != "" {
		readinessTask = "- [ ] Fill the `" + readiness.NextPressure.AxisName + ":` readiness evidence slot with concrete proof\n"
	}
	return fmt.Sprintf("# %s Tasks\n\n- [ ] Read plan.md and this runtime packet\n- [ ] Inspect current project structure and recent Hyper evidence\n- [ ] Implement the smallest coherent step toward the current episode\n- [ ] Run validation or record why validation is blocked\n%s%s- [ ] Update evidence.md with validation, readiness evidence, active capability evidence, pressure signals, changed files, decisions, reusable patterns, and blockers\n- [ ] Write next.md with the next recommended runtime episode and Learn Notes\n", goalID, browserTask, readinessTask)
}

func buildEvidenceDoc(goalID string, readiness readinessState) string {
	return fmt.Sprintf("# %s Evidence\n\n## Validation\n\nPending.\n\n## Readiness Evidence\n\n%s\n\n## Active Capability Evidence\n\nPending.\n\n## Pressure Signals\n\nPending.\n\n## Changed Files\n\nPending.\n\n## Decisions\n\nPending.\n\n## Reusable Patterns\n\nPending.\n\n## Blocker\n\nPending.\n\n## Notes\n\nPending.\n", goalID, readinessEvidenceTemplate(readiness))
}

func readinessEvidenceTemplate(readiness readinessState) string {
	defs := readinessDimensionDefs()
	lines := []string{}
	if readiness.NextPressure.AxisName != "" && readiness.NextPressure.Axis != "stage_advancement" {
		lines = append(lines, readiness.NextPressure.AxisName+": Pending.")
	}
	for _, def := range defs {
		if readiness.NextPressure.Axis == def.ID {
			continue
		}
		lines = append(lines, def.Name+": Pending.")
	}
	if readiness.NextPressure.Axis == "stage_advancement" {
		lines = append(lines, "Stage advancement: Pending.")
	}
	return strings.Join(lines, "\n")
}

func readinessEvidenceExampleAxis(readiness readinessState) string {
	if readiness.NextPressure.AxisName != "" && readiness.NextPressure.Axis != "stage_advancement" {
		return readiness.NextPressure.AxisName
	}
	return "Validation coverage"
}

func runtimeContinuation(similar []similarContext) string {
	if len(similar) == 0 {
		return "- No prior runtime context matched. Start from `plan.md` and the current repository state."
	}
	return formatSimilarContext(similar)
}

func createExecutionHandoff(runID, goalID string) handoff {
	_ = runID
	goalPath := fmt.Sprintf(".hyper/goals/%s/goal.md", goalID)
	evidencePath := fmt.Sprintf(".hyper/goals/%s/evidence.md", goalID)
	nextPath := fmt.Sprintf(".hyper/goals/%s/next.md", goalID)
	return handoff{
		Adapter:           "prompt",
		EventType:         "execution_handoff_generated",
		Description:       "Prompt handoff mode. In Codex Desktop, use this as the execution payload for `$hyper run`.",
		InstructionsLabel: "Codex Desktop payload:",
		Instructions:      fmt.Sprintf("Read %s as a runtime packet and complete it checkpoint by checkpoint. Update %s, write %s, and stop early for destructive actions, missing credentials, unclear product scope, or repeated validation failure.", goalPath, evidencePath, nextPath),
	}
}

func renderExecutionHandoff(h handoff) string {
	return fmt.Sprintf("Execution adapter: %s\n%s\n\n%s\n\n%s\n", h.Adapter, h.Description, h.InstructionsLabel, h.Instructions)
}

func defaultExecutionAdapter() string {
	return "prompt"
}
