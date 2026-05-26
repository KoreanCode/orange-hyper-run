package app

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
	augmentPlanAliases(result, body)
	return result
}

func augmentPlanAliases(plan map[string]string, body string) {
	augmentInlinePlanAliases(plan, body)
	for heading, value := range plan {
		canonical := canonicalPlanKey(heading)
		if canonical == "" {
			continue
		}
		setPlanAliasIfMissing(plan, canonical, value)
	}
	if firstRuntimeValue(plan["Product"]) == "" {
		setPlanAliasIfMissing(plan, "Product", firstMarkdownHeading(body, "# "))
	}
}

func augmentInlinePlanAliases(plan map[string]string, body string) {
	lines := strings.Split(body, "\n")
	inFence := false
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence || line == "" {
			continue
		}
		label, value, ok := splitInlinePlanField(line)
		if !ok {
			continue
		}
		canonical := canonicalPlanKey(label)
		if canonical == "" {
			continue
		}
		if strings.TrimSpace(value) == "" {
			value = followingInlinePlanValue(lines, i+1)
		}
		if inlineProductBriefCanFillMVP(label, plan) {
			setPlanAliasIfMissing(plan, "MVP", value)
			continue
		}
		setPlanAliasIfMissing(plan, canonical, value)
	}
}

func inlineProductBriefCanFillMVP(label string, plan map[string]string) bool {
	switch compactPlanHeading(label) {
	case "productbrief", "brief", "productdefinition", "servicedefinition":
		return firstRuntimeValue(plan["Product"]) != "" && firstRuntimeValue(plan["MVP"]) == ""
	default:
		return false
	}
}

func splitInlinePlanField(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	line = strings.TrimLeft(line, "#")
	line = strings.TrimSpace(line)
	line = strings.TrimLeft(line, "-*")
	line = strings.TrimSpace(line)
	index := strings.Index(line, ":")
	if index <= 0 {
		return "", "", false
	}
	label := strings.TrimSpace(line[:index])
	if label == "" || len([]rune(label)) > 48 {
		return "", "", false
	}
	return label, strings.TrimSpace(line[index+1:]), true
}

func followingInlinePlanValue(lines []string, start int) string {
	values := []string{}
	inFence := false
	for i := start; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
			inFence = !inFence
			if len(values) == 0 {
				continue
			}
		}
		if inFence {
			values = append(values, line)
			continue
		}
		if line == "" {
			if len(values) == 0 {
				continue
			}
			break
		}
		if strings.HasPrefix(line, "#") {
			break
		}
		if label, _, ok := splitInlinePlanField(line); ok && canonicalPlanKey(label) != "" {
			break
		}
		values = append(values, line)
	}
	return strings.Join(values, "\n")
}

func canonicalPlanKey(heading string) string {
	normalized := compactPlanHeading(heading)
	aliases := map[string]string{
		"product":           "Product",
		"productbrief":      "Product",
		"brief":             "Product",
		"productdefinition": "Product",
		"service":           "Product",
		"servicedefinition": "Product",
		"project":           "Product",
		"projectname":       "Product",
		"name":              "Product",
		"oneliner":          "Product",
		"제품":                "Product",
		"제품정의":              "Product",
		"서비스":               "Product",
		"서비스정의":             "Product",
		"프로젝트명":             "Product",
		"한줄소개":              "Product",
		"targetusers":       "Target Users",
		"users":             "Target Users",
		"타깃사용자":             "Target Users",
		"타겟사용자":             "Target Users",
		"대상사용자":             "Target Users",
		"mvp":               "MVP",
		"mvpgoal":           "MVP",
		"mvpscope":          "MVP",
		"mvp목표":             "MVP",
		"mvp범위":             "MVP",
		"mvp핵심범위":           "MVP",
		"첫검증상품":             "MVP",
		"currentstage":      "Current Stage",
		"stage":             "Current Stage",
		"phase":             "Current Stage",
		"단계":                "Current Stage",
		"현재단계":              "Current Stage",
		"현재스테이지":            "Current Stage",
		"스테이지":              "Current Stage",
		"페이즈":               "Current Stage",
		"buildstyle":        "Build Style",
		"stack":             "Build Style",
		"technicalstack":    "Build Style",
		"기술스택":              "Build Style",
		"기술선택":              "Build Style",
		"기본기술선택":            "Build Style",
		"모바일앱개발방향":          "Build Style",
		"nongoals":          "Non-goals",
		"non-goals":         "Non-goals",
		"제외범위":              "Non-goals",
		"mvp제외범위":           "Non-goals",
		"나중에만들것":            "Non-goals",
		"constraints":       "Constraints",
		"risks":             "Constraints",
		"제약":                "Constraints",
		"제약사항":              "Constraints",
		"리스크":               "Constraints",
		"법적운영리스크":           "Constraints",
		"successcriteria":   "Success Criteria",
		"successmetrics":    "Success Criteria",
		"successsignals":    "Success Criteria",
		"successsignal":     "Success Criteria",
		"validation":        "Success Criteria",
		"validationplan":    "Success Criteria",
		"성공지표":              "Success Criteria",
		"성공기준":              "Success Criteria",
		"완료기준":              "Success Criteria",
		"검증":                "Success Criteria",
		"검증방법":              "Success Criteria",
		"currentfocus":      "Current Focus",
		"priority":          "Current Focus",
		"priorities":        "Current Focus",
		"우선순위":              "Current Focus",
		"반드시먼저만들것":          "Current Focus",
	}
	return aliases[normalized]
}

func compactPlanHeading(heading string) string {
	heading = strings.ToLower(strings.TrimSpace(heading))
	heading = strings.TrimLeft(heading, "0123456789.:-_ )(")
	var builder strings.Builder
	for _, r := range heading {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r >= '가' && r <= '힣' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func setPlanAliasIfMissing(plan map[string]string, key, value string) {
	if firstRuntimeValue(plan[key]) != "" {
		return
	}
	value = planAliasValue(key, value)
	if value == "" {
		return
	}
	plan[key] = value
}

func planAliasValue(key, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if key == "Product" {
		return firstUsefulPlanLine(value)
	}
	return value
}

func firstUsefulPlanLine(value string) string {
	for _, line := range strings.Split(value, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "###") || strings.HasPrefix(trimmed, "|") || strings.HasPrefix(trimmed, ">") {
			continue
		}
		trimmed = strings.TrimSpace(strings.Trim(trimmed, "*_`"))
		if isPlaceholder(trimmed) {
			continue
		}
		return trimmed
	}
	return ""
}

func updatePlanCurrentStage(body, nextStage string) (string, bool) {
	nextStage = strings.TrimSpace(nextStage)
	if nextStage == "" {
		return body, false
	}
	lines := strings.Split(body, "\n")
	headingIndex := -1
	endIndex := len(lines)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "## ") {
			continue
		}
		heading := strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
		if canonicalPlanKey(heading) == "Current Stage" {
			headingIndex = i
			break
		}
	}
	if headingIndex == -1 {
		if updated, changed, found := updateInlinePlanCurrentStage(body, nextStage); found {
			return updated, changed
		}
		trimmed := strings.TrimRight(body, "\n")
		if trimmed != "" {
			trimmed += "\n\n"
		}
		return trimmed + "## Current Stage\n\n" + nextStage + "\n", true
	}
	for i := headingIndex + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "## ") {
			endIndex = i
			break
		}
	}
	current := strings.TrimSpace(strings.Join(lines[headingIndex+1:endIndex], "\n"))
	if strings.EqualFold(current, nextStage) || normalizeRuntimeStage(current) == nextStage {
		return body, false
	}
	replacement := []string{lines[headingIndex], "", nextStage, ""}
	updated := append([]string{}, lines[:headingIndex]...)
	updated = append(updated, replacement...)
	for endIndex < len(lines) && strings.TrimSpace(lines[endIndex]) == "" {
		endIndex++
	}
	updated = append(updated, lines[endIndex:]...)
	out := strings.Join(updated, "\n")
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	return out, true
}

func updateInlinePlanCurrentStage(body, nextStage string) (string, bool, bool) {
	lines := strings.Split(body, "\n")
	inFence := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		label, value, ok := splitInlinePlanField(line)
		if !ok || canonicalPlanKey(label) != "Current Stage" {
			continue
		}
		current := strings.TrimSpace(value)
		if strings.EqualFold(current, nextStage) || normalizeRuntimeStage(current) == nextStage {
			return body, false, true
		}
		index := strings.Index(line, ":")
		if index < 0 {
			return body, false, true
		}
		lines[i] = strings.TrimRight(line[:index+1], " ") + " " + nextStage
		out := strings.Join(lines, "\n")
		if !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		return out, true, true
	}
	return body, false, false
}

func firstMarkdownHeading(body, prefix string) string {
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) && !strings.HasPrefix(trimmed, prefix+"#") {
			heading := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
			if genericPlanTitle(heading) {
				continue
			}
			return heading
		}
	}
	return ""
}

func genericPlanTitle(value string) bool {
	switch compactPlanHeading(value) {
	case "plan", "productplan", "projectplan", "serviceplan", "기획서", "제품기획서", "프로젝트기획서", "서비스기획서":
		return true
	default:
		return false
	}
}

func compileGoalEpisode(goalID, focus, planBody string, similar []similarContext, growth growthState, readiness readinessState) episode {
	plan := parsePlan(planBody)
	stage := normalizeRuntimeStage(firstRuntimeValue(plan["Current Stage"], "Tiny MVP"))
	buildStyle := firstRuntimeValue(plan["Build Style"], "Detect from project")
	product := readinessProductName(plan)
	objective := runtimeObjective(focus, plan, stage, product, readiness)
	validation := applyReadinessValidation(applyGrowthValidation(applyStageValidation(validationForBuildStyle(buildStyle), stage), growth), readiness)
	stopCondition := runtimeStopCondition(plan, stage, growth, readiness)
	scope := runtimeWorkBoundary(objective, stage, plan, growth, readiness)
	nonGoals := firstRuntimeValue(plan["Non-goals"], "No explicit non-goals recorded in plan.md.")
	docs := episodeDocs{
		Goal:     buildGoalDoc(goalID, objective, focus, plan, stage, buildStyle, scope, validation, stopCondition, similar, growth, readiness),
		Tasks:    buildTasksDoc(goalID, buildStyle, stage, readiness, growth),
		Evidence: buildEvidenceDoc(goalID, stage, readiness, growth),
		Review:   fmt.Sprintf("# %s Review\n\n## Result\n\nPending.\n\n## Issues\n\nPending.\n", goalID),
		Next:     buildNextDoc(goalID, readiness),
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

func buildNextDoc(goalID string, readiness readinessState) string {
	return fmt.Sprintf(`# %s Next

## Recommended Next Goal

Pending.

## Learn Notes

%s

- Decision: Pending.
- Pattern: Pending.
- Constraint: Pending.
- Failure: Pending.
`, goalID, nextLearnNotesGuidance(readiness))
}

func nextLearnNotesGuidance(readiness readinessState) string {
	if readiness.NextPressure.Axis == "reference_benchmark" || readinessGateHasAxis(readiness, "reference_benchmark") {
		return "Write only durable reference signals that should change future packets: category baseline, benchmark rule, accepted tradeoff, repeated below-baseline gap, or comparison-driven constraint. Do not record one-off reference names unless they change future work boundaries, validation, stop conditions, readiness, or capability candidates. Leave a line as Pending. or remove it when there is no reusable signal."
	}
	return "Write only durable signals that should change a future runtime packet. Leave a line as Pending. or remove it when there is no reusable signal."
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
		normalized == "none in this run" ||
		normalized == "none this run" ||
		normalized == "none blocking" ||
		normalized == "nothing blocking" ||
		normalized == "no blocker" ||
		normalized == "no blockers" ||
		normalized == "no technical blocker" ||
		normalized == "no blocker found" ||
		normalized == "no blockers found" ||
		normalized == "no blocking issue" ||
		normalized == "no blocking issues" ||
		normalized == "no current blocker" ||
		normalized == "no current blockers" ||
		normalized == "no blocker in this episode" ||
		normalized == "no blockers in this episode" ||
		normalized == "no blocker in this run" ||
		normalized == "no blockers in this run" ||
		normalized == "no runtime blocker" ||
		normalized == "no runtime blockers" ||
		normalized == "no remaining blocker" ||
		normalized == "no remaining blockers" ||
		normalized == "no blocker remains" ||
		normalized == "no blockers remain" ||
		normalized == "no failure" ||
		normalized == "no failures" ||
		normalized == "none critical" ||
		normalized == "no critical gap" ||
		normalized == "no critical gaps" ||
		normalized == "no failure in this episode" ||
		normalized == "no failures in this episode" ||
		normalized == "no failure in this run" ||
		normalized == "no failures in this run" ||
		normalized == "no failure this run" ||
		normalized == "no failures this run" ||
		normalized == "clear: implementation and validation completed for this packet" ||
		normalized == "clear implementation and validation completed for this packet" ||
		normalized == "implementation and validation completed for this packet" {
		return true
	}
	return strings.HasPrefix(normalized, "no blocker for this episode") ||
		strings.HasPrefix(normalized, "no blockers for this episode") ||
		strings.HasPrefix(normalized, "no blocker for this packet") ||
		strings.HasPrefix(normalized, "no blockers for this packet") ||
		strings.HasPrefix(normalized, "no technical blocker") ||
		strings.HasPrefix(normalized, "none blocking for") ||
		strings.HasPrefix(normalized, "none in this run") ||
		strings.HasPrefix(normalized, "none this run") ||
		strings.HasPrefix(normalized, "nothing blocking for") ||
		strings.HasPrefix(normalized, "no blocker found for") ||
		strings.HasPrefix(normalized, "no blockers found for") ||
		strings.HasPrefix(normalized, "no blocking issue for") ||
		strings.HasPrefix(normalized, "no blocking issues for") ||
		strings.HasPrefix(normalized, "no current blocker for") ||
		strings.HasPrefix(normalized, "no current blockers for") ||
		strings.HasPrefix(normalized, "no runtime blocker was found") ||
		strings.HasPrefix(normalized, "no remaining blocker for this packet") ||
		strings.HasPrefix(normalized, "no remaining blockers for this packet") ||
		strings.HasPrefix(normalized, "no blocker remains for this packet") ||
		strings.HasPrefix(normalized, "no blockers remain for this packet") ||
		strings.HasPrefix(normalized, "no failure for this episode") ||
		strings.HasPrefix(normalized, "no failures for this episode") ||
		strings.HasPrefix(normalized, "no failure in this run") ||
		strings.HasPrefix(normalized, "no failures in this run") ||
		strings.HasPrefix(normalized, "none critical for") ||
		strings.HasPrefix(normalized, "no critical gap for") ||
		strings.HasPrefix(normalized, "no critical gaps for") ||
		strings.HasPrefix(normalized, "no core category-baseline gap") ||
		strings.HasPrefix(normalized, "clear: implementation and validation completed") ||
		strings.HasPrefix(normalized, "clear implementation and validation completed") ||
		strings.HasPrefix(normalized, "implementation and validation completed for this packet")
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
		{name: "Sustained Service Quality", patterns: []string{"sustained service quality", "sustained quality"}},
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
	if readiness.NextPressure.Axis == "reference_benchmark" {
		lines = append(lines,
			"- Do not add broad feature work until Reference Benchmark Evidence proves the category baseline.",
			"- Select 3-5 named references, define baseline expectations, compare below/meets/above baseline, and identify the strongest core gap.",
			"- If a critical below-baseline gap exists, implement only the smallest fix for that gap or leave a blocked decision; do not advance the stage.",
			"- If no critical gap exists, record the above-baseline strength and recommend stage advancement or the next sustained-service pressure.",
		)
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

func runtimeStopCondition(plan map[string]string, stage string, growth growthState, readiness readinessState) string {
	base := stageDoneCondition(stage)
	if criteria := firstRuntimeValue(plan["Success Criteria"]); criteria != "" && !sameStopCondition(criteria, base) {
		base = "- Plan success criteria: " + compactText(criteria, 240) + "\n" + base
	}
	return applyReadinessStopConditions(applyGrowthStopConditions(base, growth), readiness)
}

func sameStopCondition(criteria, base string) bool {
	criteria = normalizeSentence(criteria)
	base = normalizeSentence(base)
	return criteria == "" || criteria == base || strings.Contains(base, criteria)
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
	if strings.Contains(normalized, "sustained") {
		return strings.Join([]string{
			"- Active validators, harnesses, or equivalent reusable quality structures continue to pass or have explicit blockers.",
			"- Repeated failures or friction are converted into the next focused quality packet.",
			"- Operational, validation, and maintainability evidence stays current without broad feature expansion.",
			"- next.md identifies the next sustained-service improvement, not another stage advancement.",
		}, "\n")
	}
	if strings.Contains(normalized, "service") || strings.Contains(normalized, "production") {
		return strings.Join([]string{
			"- Required validation, security, deployment, operations, and maintainability evidence is captured.",
			"- Setup, release, rollback, and recovery paths are repeatable from documented commands or artifacts.",
			"- Reference benchmark evidence shows no critical category-baseline gap and at least one above-baseline strength.",
			"- No critical blocker remains without an owner, next action, or explicit stop condition.",
			"- next.md identifies the highest-value sustained-service improvement instead of broad feature work.",
		}, "\n")
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
	if strings.Contains(normalized, "sustained") {
		return "Keep the service healthy through repeated quality evidence, active validators or harnesses, and friction reduction before adding breadth."
	}
	if strings.Contains(normalized, "service") || strings.Contains(normalized, "production") {
		return "Close operational and reference acceptance criteria first: repeatable validation, security/privacy boundaries, release and rollback proof, operator docs, maintainability evidence, and category-baseline comparison before feature breadth."
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
	if strings.Contains(normalized, "sustained") {
		return "Sustained Service Quality validation should run active validators or harnesses, record any blocker, and convert repeated failure into the next focused quality packet."
	}
	if strings.Contains(normalized, "service") || strings.Contains(normalized, "production") {
		return "Service Quality validation should prove the service can be set up, checked, released or run, rolled back, handed off, and compared against category references from documented commands, artifacts, or benchmark notes; run active validators or record why each required check is blocked."
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

func buildGoalDoc(goalID, objective, focus string, plan map[string]string, stage, buildStyle, workBoundary, validation, stopCondition string, similar []similarContext, growth growthState, readiness readinessState) string {
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

## Stage Runtime Behavior

%s

## Active Capabilities

%s

## Growth Principles

%s

## Execution Contract

%s

## Proof Contract

%s

## Work Boundary

%s

## Validation Signals

%s

## Evidence Required

- Command output or reason validation could not run
- Readiness evidence in axis-slot format, for example "%s: proof"
- Active capability evidence when required validators are present
- Surface proof evidence when this packet changes a user-facing screen or flow
- Changed file summary
- Decisions that should persist into future runs
- Reusable patterns that should guide similar future work
- Repeated need, failure, or proof that should influence future structure
- Blocker, constraint, or failure signal when applicable
- Screenshot path when applicable
- Learn Notes must contain only durable signals: decision, pattern, constraint, or failure that should change a future packet.

## Done Checklist

%s

## Stop When

%s
`, goalID, runtimeContinuation(similar), objective, product, stage, stageContract, targetUsers, runtimeProtocolDefinition, growthLoopDefinition, buildStyle, currentFocus, buildStageGateDoc(readiness), stageRuntimeBehaviorDoc(stage, buildStyle, readiness), activeCapabilitiesDoc(growth), formatGrowthPrinciples(), executionContractDoc(stage, readiness, growth), proofContractDoc(stage, buildStyle, readiness), workBoundary, validation, readinessEvidenceExampleAxis(readiness), doneChecklistDoc(stage, readiness, growth), stopCondition)
}

func executionContractDoc(stage string, readiness readinessState, growth growthState) string {
	lines := []string{
		"- Work one coherent episode only; do not start a second runtime packet inside this packet.",
		"- Prefer the smallest reversible implementation that moves the current readiness pressure.",
		"- If validation fails twice for the same reason, stop and record the failure instead of broadening scope.",
		"- Close the packet with evidence.md, next.md, and `hyper complete`; do not create the next packet first.",
	}
	if readiness.StageGate.Advancement.Candidate {
		lines = append(lines, "- Gate-ready packets may recommend `hyper advance`, but must not silently change plan.md stage.")
	}
	if activeStructureCount(growth.Candidates) > 0 {
		lines = append(lines, "- Active capabilities are required behavior for this packet unless explicitly blocked with a reason.")
	}
	if strings.Contains(normalizeLabel(stage), "service") || strings.Contains(normalizeLabel(stage), "production") {
		lines = append(lines, "- Service Quality packets require deployment, security, docs, operational, or reference benchmark evidence when those surfaces are touched.")
	}
	return strings.Join(lines, "\n")
}

func doneChecklistDoc(stage string, readiness readinessState, growth growthState) string {
	lines := []string{
		"- evidence.md contains real validation output or a concrete blocked reason.",
		"- evidence.md names the changed files and the user-visible or operational behavior changed.",
		"- next.md recommends exactly one next runtime episode.",
		"- Learn Notes avoid summaries and keep only reusable project signals.",
	}
	if readiness.NextPressure.AxisName != "" && readiness.NextPressure.Axis != "stage_advancement" {
		lines = append(lines, "- Readiness Evidence includes concrete proof for "+readiness.NextPressure.AxisName+".")
	}
	if activeStructureCount(growth.Candidates) > 0 {
		lines = append(lines, "- Active Capability Evidence shows each active validator, skill, or harness ran or why it was blocked.")
	}
	if strings.Contains(normalizeLabel(stage), "beta") || strings.Contains(normalizeLabel(stage), "service") {
		lines = append(lines, "- Stop conditions cover failure, regression, and missing credential cases found during this packet.")
	}
	if referenceBenchmarkRequired(stage, readiness) {
		lines = append(lines, "- Reference Benchmark Evidence lists 3-5 references, baseline expectations, current comparison, below-baseline gaps, above-baseline strength, and the next pressure.")
	}
	return strings.Join(lines, "\n")
}

func proofContractDoc(stage, buildStyle string, readiness readinessState) string {
	lines := []string{
		"- Functional Proof: prove the smallest useful behavior works for this runtime packet.",
		"- Surface Proof: required only when user-facing screens or flows change; prove a target user can understand the screen, take the primary action, and see the result or recovery state.",
		"- Operational Proof: prove the safest available build, test, smoke, setup, or handoff path is repeatable, or document why it is blocked.",
	}
	normalizedBuild := normalizeLabel(buildStyle)
	if hasAny(normalizedBuild, "web", "local app", "game", "desktop") {
		lines = append(lines,
			"- Surface proof should name the affected surface, primary user action, checked states, viewport(s), screenshot or browser smoke evidence, and remaining surface gaps.",
		)
	}
	normalizedStage := normalizeLabel(stage)
	if strings.Contains(normalizedStage, "tiny") && strings.Contains(normalizedStage, "mvp") {
		lines = append(lines, "- Tiny MVP surface proof can be manual browser smoke plus screenshot evidence for one core flow; do not create visual regression or accessibility harnesses yet.")
	} else if strings.Contains(normalizedStage, "usable") && strings.Contains(normalizedStage, "mvp") {
		lines = append(lines, "- Usable MVP surface proof should cover the primary flow states touched by this packet and record mobile or desktop gaps that should become future pressure.")
	} else if strings.Contains(normalizedStage, "beta") {
		lines = append(lines, "- Beta surface proof may create visual smoke, accessibility, or responsive-check candidates only after repeated evidence.")
	} else if strings.Contains(normalizedStage, "service") || strings.Contains(normalizedStage, "production") {
		lines = append(lines, "- Service Quality surface proof should run active visual or accessibility validators when promoted, or document why they are blocked.")
		lines = append(lines, "- Service Quality reference proof should compare the current result with 3-5 category references and block advancement when a core user or operator expectation is below baseline.")
	}
	if readiness.NextPressure.Axis == "core_ux" {
		lines = append(lines, "- Current readiness pressure is Core UX, so surface proof should directly support that axis.")
	}
	if readiness.NextPressure.Axis == "validation_coverage" {
		lines = append(lines, "- Current readiness pressure is Validation coverage, so any surface proof should include repeatable browser smoke or command evidence when available.")
	}
	return strings.Join(lines, "\n")
}

func stageRuntimeBehaviorDoc(stage, buildStyle string, readiness readinessState) string {
	lines := []string{
		"- Build style: " + compactText(firstRuntimeValue(buildStyle, "Detect from project"), 140),
		"- Stage behavior: " + compactText(firstRuntimeValue(stageRuntimeBoundary(stage), stageGrowthContract(stage)), 180),
		"- Done means: " + compactText(strings.ReplaceAll(stageDoneCondition(stage), "\n", " "), 220),
	}
	if readiness.NextPressure.AxisName != "" {
		lines = append(lines, "- This packet should move readiness pressure: "+readiness.NextPressure.AxisName)
	}
	if readiness.StageGate.Advancement.Candidate {
		lines = append(lines, "- Gate is ready; only recommend stage advancement, do not silently edit plan.md.")
	}
	return strings.Join(lines, "\n")
}

func activeCapabilitiesDoc(growth growthState) string {
	candidates := visibleGrowthCandidates(growth.Candidates)
	lines := []string{}
	requiredActiveValidators := map[string]bool{}
	for _, signal := range growth.RuntimeBehavior.ValidationSignals {
		if name := requiredActiveValidatorName(signal); name != "" {
			requiredActiveValidators[name] = true
		}
	}
	for _, candidate := range candidates {
		if candidate.Status != "active" {
			continue
		}
		if candidate.Kind == "validator" && requiredActiveValidators[candidate.Name] {
			continue
		}
		lines = append(lines, fmt.Sprintf("- Active %s %s: %s", candidate.Kind, displayGrowthCandidateName(candidate), compactText(candidate.Signal, 180)))
	}
	for _, signal := range growth.RuntimeBehavior.ValidationSignals {
		if strings.Contains(signal, "Required active validator") {
			lines = append(lines, signal)
		}
	}
	if len(lines) == 0 {
		return "- None active. Candidate structures are informational until promoted."
	}
	return strings.Join(lines, "\n")
}

func requiredActiveValidatorName(signal string) string {
	fields := strings.Fields(strings.TrimPrefix(strings.TrimSpace(signal), "- "))
	if len(fields) < 4 {
		return ""
	}
	if fields[0] == "Required" && fields[1] == "active" && fields[2] == "validator" {
		return strings.TrimSuffix(fields[3], ":")
	}
	return ""
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
		lines = append(lines, "- Gate requirement: "+compactText(evidence, 160))
	}
	return strings.Join(lines, "\n")
}

func buildTasksDoc(goalID, buildStyle, stage string, readiness readinessState, growth growthState) string {
	browserTask := ""
	if hasAny(normalizeLabel(buildStyle), "web", "local app", "game") {
		browserTask = "- [ ] Capture browser screenshot and console evidence when UI changes are made\n"
	}
	referenceTask := ""
	if referenceBenchmarkRequired(stage, readiness) {
		referenceTask = "- [ ] Fill Reference Benchmark Evidence with category references, baseline expectations, current comparison, and blocking gaps\n"
	}
	readinessTask := ""
	if readiness.NextPressure.AxisName != "" {
		readinessTask = "- [ ] Fill the `" + readiness.NextPressure.AxisName + ":` readiness evidence slot with concrete proof\n"
	}
	activeTask := ""
	if activeStructureCount(growth.Candidates) > 0 {
		activeTask = "- [ ] Run or explicitly block every active capability listed in goal.md\n"
	}
	return fmt.Sprintf("# %s Tasks\n\n- [ ] Read plan.md and this runtime packet\n- [ ] Inspect current project structure and recent Hyper evidence\n- [ ] Confirm the stage behavior for `%s`\n- [ ] Implement the smallest coherent step toward the current episode\n- [ ] Run validation or record why validation is blocked\n%s%s%s%s- [ ] Update evidence.md with validation, readiness evidence, active capability evidence, pressure signals, changed files, decisions, reusable patterns, and blockers\n- [ ] Write next.md with exactly one recommended next runtime episode and durable Learn Notes only\n- [ ] Run `hyper complete`; if the finish gate fails, fix this same packet using review.md\n", goalID, stage, browserTask, referenceTask, readinessTask, activeTask)
}

func buildEvidenceDoc(goalID, stage string, readiness readinessState, growth growthState) string {
	return fmt.Sprintf("# %s Evidence\n\n## Validation\n\nPending.\n\n## Readiness Evidence\n\n%s\n\n## Surface Proof Evidence\n\n- Target surface: Pending.\n- Primary user action: Pending.\n- States checked: Pending.\n- Viewports: Pending.\n- Evidence: Pending.\n- Surface risks or gaps: Pending.\n\n%s## Active Capability Evidence\n\n%s\n\n## Pressure Signals\n\nPending.\n\n## Changed Files\n\nPending.\n\n## Decisions\n\nPending.\n\n## Reusable Patterns\n\nPending.\n\n## Learn Quality Gate\n\n- Keep as memory only if it should change future work boundary, validation, stop conditions, readiness, or capability candidates.\n- Do not record one-off progress, file lists, generic summaries, or \"none\" statements as Learn signals.\n\n## Blocker\n\nPending.\n\n## Notes\n\nPending.\n", goalID, readinessEvidenceTemplate(readiness), referenceBenchmarkEvidenceTemplate(stage, readiness), activeCapabilityEvidenceTemplate(growth))
}

func activeCapabilityEvidenceTemplate(growth growthState) string {
	lines := []string{}
	for _, candidate := range visibleGrowthCandidates(growth.Candidates) {
		if candidate.Status != "active" {
			continue
		}
		name := displayGrowthCandidateName(candidate)
		signal := compactText(firstNonBlank(candidate.Signal, candidate.Reason), 160)
		if signal == "" {
			lines = append(lines, "- "+name+": Pending. Run or explicitly block this active "+candidate.Kind+".")
			continue
		}
		lines = append(lines, "- "+name+": Pending. Required behavior: "+signal)
	}
	if len(lines) == 0 {
		return "Pending."
	}
	return strings.Join(lines, "\n")
}

func referenceBenchmarkEvidenceTemplate(stage string, readiness readinessState) string {
	if !referenceBenchmarkRequired(stage, readiness) {
		return ""
	}
	return strings.Join([]string{
		"## Reference Benchmark Evidence",
		"",
		"- Category: Pending.",
		"- References: Pending. List 3-5 comparable products, tools, apps, or workflows.",
		"- Baseline expectations: Pending. Name the category expectations a real user or operator would assume.",
		"- Current comparison: Pending. Mark each important area as below baseline, meets baseline, or above baseline.",
		"- Below-baseline gaps: Pending. Block Service Quality if any core user or operator expectation is below baseline.",
		"- Above-baseline strength: Pending. Name at least one concrete strength versus the references.",
		"- Decision: Pending. State whether Service Quality is allowed or blocked, and what the next pressure should be.",
		"",
	}, "\n")
}

func serviceQualityStage(stage string) bool {
	normalized := normalizeLabel(stage)
	return strings.Contains(normalized, "service") || strings.Contains(normalized, "production")
}

func referenceBenchmarkRequired(stage string, readiness readinessState) bool {
	if readiness.Version != 0 {
		return readiness.NextPressure.Axis == "reference_benchmark"
	}
	normalized := normalizeLabel(stage)
	if strings.Contains(normalized, "beta") || serviceQualityStage(stage) {
		return true
	}
	for _, axis := range readiness.StageGate.RequiredAxes {
		if axis == "reference_benchmark" {
			return true
		}
	}
	return false
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
