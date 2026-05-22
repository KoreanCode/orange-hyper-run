package app

import (
	"fmt"
	"path/filepath"
	"strings"
)

func refreshStateFromPlanForStatus(root string, state projectState) projectState {
	planBody := readIfExists(filepath.Join(root, planFile))
	if strings.TrimSpace(planBody) == "" {
		return state
	}
	plan := parsePlan(planBody)
	if staleProjectName(state.Project) {
		state.Project = readinessProductName(plan)
	}
	if strings.TrimSpace(state.Stage) == "" {
		state.Stage = normalizeRuntimeStage(firstRuntimeValue(plan["Current Stage"], "Tiny MVP"))
	}
	return state
}

func staleProjectName(project string) bool {
	normalized := strings.ToLower(strings.TrimSpace(project))
	return normalized == "" || normalized == "unknown project" || normalized == "the product"
}

func statusDashboardLines(state projectState, derived goalState, readiness readinessState, growth growthState, runs, goals int) []string {
	project := compactText(firstNonBlank(state.Project, "Unknown project"), 120)
	stage := normalizeRuntimeStage(firstNonBlank(state.Stage, readiness.Stage, "Unknown stage"))
	lines := []string{
		"Hyper Run Status",
		"",
		"Project: " + project,
		"Stage: " + stage,
		"Stage contract: " + stageGrowthContract(stage),
		"Method: " + growthRuntimeDefinition,
		"Protocol: " + runtimeProtocolDefinition,
		"Pressure ledger: " + growthLoopStateSummary(growth),
		"Proof: " + proofStatusSummary(derived, readiness),
		"Next proof gap: " + nextProofGap(readiness),
		"Principles: " + growthPrinciplesLine(),
		"Status: " + displayProjectStatus(state, derived),
		"Runtime packet state: " + derived.State,
		"Runtime packet reason: " + derived.Reason,
	}
	runLabel := "Last run"
	packetLabel := "Last runtime packet"
	if state.Status == "active" {
		runLabel = "Active run"
		packetLabel = "Current runtime packet"
	}
	lines = append(lines,
		runLabel+": "+state.ActiveRunID,
		packetLabel+": "+state.CurrentGoalID,
		"Runtime packet file: "+state.CurrentGoalPath,
		"",
	)
	lines = append(lines, statusActionLines(state, derived, readiness, growth)...)
	lines = append(lines, "")
	lines = append(lines, pressureDashboardLines(growth)...)
	lines = append(lines, "")
	lines = append(lines, readinessDashboardLines(readiness)...)
	lines = append(lines,
		"",
		"Next:",
		"  "+statusNextCommand(state, derived, readiness),
		"",
		fmt.Sprintf("Runs recorded: %d", runs),
		fmt.Sprintf("Runtime packets recorded: %d", goals),
		fmt.Sprintf("Growth pressures: %d", visibleGrowthPressureCount(growth.Pressures)),
		fmt.Sprintf("Capability candidates: %d", visibleGrowthCandidateCount(growth.Candidates)),
		"Updated: "+state.UpdatedAt,
		"",
	)
	return lines
}

func proofStatusSummary(derived goalState, readiness readinessState) string {
	if readiness.Version == 0 {
		return "not recorded"
	}
	functional := "pending"
	if derived.State == "completed" {
		functional = "covered"
	} else if derived.State == "blocked" {
		functional = "blocked"
	}
	return fmt.Sprintf("functional %s, surface %s, operational %s",
		functional,
		proofAxisStatus(readiness, "core_ux"),
		proofAxisStatus(readiness, "validation_coverage"),
	)
}

func proofAxisStatus(readiness readinessState, axis string) string {
	for _, dim := range readiness.Dimensions {
		if dim.ID == axis {
			return firstNonBlank(dim.Status, "missing")
		}
	}
	return "missing"
}

func nextProofGap(readiness readinessState) string {
	if readiness.Version == 0 {
		return "not selected"
	}
	switch {
	case proofAxisStatus(readiness, "core_ux") != "covered":
		return "surface proof for the primary user flow"
	case proofAxisStatus(readiness, "validation_coverage") != "covered":
		return "repeatable validation proof"
	case readiness.NextPressure.AxisName != "":
		return readiness.NextPressure.AxisName
	default:
		return "none"
	}
}

func statusActionLines(state projectState, derived goalState, readiness readinessState, growth growthState) []string {
	lines := []string{"Action:"}
	lines = append(lines, "  Next action: "+statusNextCommand(state, derived, readiness))
	lines = append(lines, "  Why now: "+statusActionReason(state, derived, readiness, growth))
	lines = append(lines, "  Do not do yet: "+statusDoNotDoYet(state, derived, readiness, growth))
	return lines
}

func statusActionReason(state projectState, derived goalState, readiness readinessState, growth growthState) string {
	if derived.State == "active" {
		return "The current runtime packet is still open; evidence and next.md decide what the project learns."
	}
	if strings.TrimSpace(state.Status) != "" && strings.TrimSpace(state.Status) != strings.TrimSpace(derived.State) {
		return "The packet evidence says " + derived.State + " while state.json still says " + state.Status + "; repair before trusting automation."
	}
	if readiness.StageGate.Advancement.Candidate {
		return readiness.StageGate.Advancement.Recommendation
	}
	if readiness.NextPressure.Reason != "" {
		return readiness.NextPressure.Reason
	}
	if visibleGrowthPressureCount(growth.Pressures) > 0 {
		return "The pressure ledger has project-specific signals that should shape the next packet."
	}
	return firstNonBlank(derived.Reason, "No runtime packet is active.")
}

func statusDoNotDoYet(state projectState, derived goalState, readiness readinessState, growth growthState) string {
	if derived.State == "active" {
		return "Do not start another `hyper run` until this packet is completed or blocked."
	}
	if strings.TrimSpace(state.Status) != "" && strings.TrimSpace(state.Status) != strings.TrimSpace(derived.State) {
		return "Do not create another packet until `hyper repair` or `hyper complete` reconciles state."
	}
	if readiness.StageGate.Status == "not_ready" {
		return "Do not advance " + readiness.StageGate.CurrentStage + " until blocking readiness gaps are closed."
	}
	if readiness.StageGate.Advancement.Candidate {
		return "Do not run `hyper advance` unless the user accepts the stage advancement."
	}
	if visibleGrowthCandidateCount(growth.Candidates) > 0 && activeStructureCount(growth.Candidates) == 0 {
		return "Do not treat candidates as active harnesses or validators before promotion."
	}
	return "Do not add broad structure unless repeated evidence creates pressure for it."
}

func displayProjectStatus(state projectState, derived goalState) string {
	projectStatus := strings.TrimSpace(state.Status)
	derivedStatus := strings.TrimSpace(derived.State)
	if projectStatus == "" || derivedStatus == "" || projectStatus == derivedStatus {
		return firstNonBlank(projectStatus, derivedStatus, "unknown")
	}
	if projectStatus == "active" && derivedStatus == "active" {
		return projectStatus
	}
	return derivedStatus + " (state.json: " + projectStatus + ")"
}

func pressureDashboardLines(growth growthState) []string {
	lines := []string{"Pressure Ledger:"}
	pressures := visibleGrowthPressures(growth.Pressures)
	if len(pressures) == 0 {
		lines = append(lines, "  Top pressures: none")
	} else {
		lines = append(lines, "  Top pressures:")
		for i, pressure := range pressures {
			if i >= 3 {
				break
			}
			lines = append(lines, fmt.Sprintf("    - %s/%s: %s", pressure.State, pressure.PressureType, compactText(pressure.Signal, 120)))
		}
	}
	candidates := visibleGrowthCandidates(growth.Candidates)
	if len(candidates) == 0 {
		lines = append(lines, "  Candidate structures: none")
		return lines
	}
	lines = append(lines, "  Candidate structures:")
	for i, candidate := range candidates {
		if i >= 3 {
			break
		}
		lines = append(lines, fmt.Sprintf("    - %s (%s, %s, evidence %d)", displayGrowthCandidateName(candidate), candidate.Kind, candidate.Status, candidate.EvidenceCount))
	}
	if len(candidates) > 3 {
		lines = append(lines, fmt.Sprintf("    - ... %d more", len(candidates)-3))
	}
	return lines
}

func visibleGrowthPressures(pressures []growthPressure) []growthPressure {
	filtered := []growthPressure{}
	for _, pressure := range pressures {
		if visibleGrowthPressure(pressure) {
			filtered = append(filtered, pressure)
		}
	}
	return filtered
}

func visibleGrowthCandidates(candidates []growthCandidate) []growthCandidate {
	filtered := []growthCandidate{}
	for _, candidate := range candidates {
		if visibleGrowthCandidate(candidate) {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

func displayGrowthCandidateName(candidate growthCandidate) string {
	name := strings.TrimSpace(candidate.Name)
	prefix := candidateDisplayPrefix(candidate)
	if command := inferredCommandForSignal(candidate.Signal); command != "" && prefix != "" {
		return prefix + "-" + slugify(command)
	}
	return firstNonBlank(name, candidate.Kind, "candidate")
}

func candidateDisplayPrefix(candidate growthCandidate) string {
	name := strings.ToLower(strings.TrimSpace(candidate.Name))
	for _, prefix := range []string{"validator", "preflight", "skill", "harness"} {
		if strings.HasPrefix(name, prefix+"-") || name == prefix {
			return prefix
		}
	}
	return strings.ToLower(strings.TrimSpace(candidate.Kind))
}

func readinessDashboardLines(readiness readinessState) []string {
	if readiness.Version == 0 {
		return []string{"Readiness: not recorded"}
	}
	covered := []string{}
	emerging := []string{}
	missing := []string{}
	for _, dim := range readiness.Dimensions {
		switch dim.Status {
		case "covered":
			covered = append(covered, dim.Name)
		case "emerging":
			emerging = append(emerging, dim.Name)
		default:
			missing = append(missing, dim.Name)
		}
	}
	lines := []string{
		"Readiness gate: " + readinessGateSummary(readiness),
		"Readiness pressure: " + readinessPressureSummary(readiness),
		"Readiness:",
		"  Gate: " + readinessGateSummary(readiness),
		"  Next pressure: " + readinessPressureSummary(readiness),
		"  Covered axes: " + readinessListSummary(covered),
		"  Emerging axes: " + readinessListSummary(emerging),
		"  Missing axes: " + readinessListSummary(missing),
	}
	if len(readiness.StageGate.BlockingGaps) > 0 {
		lines = append(lines, "  Blocking gaps:")
		for _, gap := range readiness.StageGate.BlockingGaps {
			lines = append(lines, "    - "+compactText(gap, 140))
		}
	} else {
		lines = append(lines, "  Blocking gaps: none")
	}
	if readiness.StageGate.Advancement.Recommendation != "" {
		lines = append(lines, "  Stage advancement: "+compactText(readiness.StageGate.Advancement.Recommendation, 160))
	}
	if readiness.NextPressure.Axis == "stage_advancement" || readiness.StageGate.Advancement.Candidate {
		lines = append(lines, "  Recommended action: hyper advance")
	} else if readiness.NextPressure.RecommendedGoal != "" {
		lines = append(lines, "  Recommended run: hyper run \""+compactText(readiness.NextPressure.RecommendedGoal, 120)+"\"")
	}
	return lines
}

func statusNextCommand(state projectState, derived goalState, readiness readinessState) string {
	if strings.TrimSpace(state.Status) != "" && strings.TrimSpace(derived.State) != "" && strings.TrimSpace(state.Status) != strings.TrimSpace(derived.State) {
		return "hyper repair"
	}
	if strings.TrimSpace(state.CurrentGoalID) == "" {
		return "hyper run [focus]"
	}
	if derived.State == "active" {
		return "update " + strings.TrimSuffix(state.CurrentGoalPath, "goal.md") + "evidence.md and next.md, then run `hyper complete`"
	}
	if readiness.NextPressure.Axis == "stage_advancement" || readiness.StageGate.Advancement.Candidate {
		return "hyper advance"
	}
	if readiness.NextPressure.RecommendedGoal != "" {
		return "hyper run \"" + compactText(readiness.NextPressure.RecommendedGoal, 120) + "\""
	}
	return "hyper run [next focus]"
}
