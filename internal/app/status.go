package app

import (
	"fmt"
	"path/filepath"
	"strings"
)

func parseStatusOptions(args []string) (bool, *hyperError) {
	short := false
	for _, arg := range args {
		switch strings.TrimSpace(arg) {
		case "", "--full":
		case "--short", "-s":
			short = true
		default:
			return false, newError("Unknown status option: "+arg+"\n\nUsage:\n  hyper status\n  hyper status --short", 2)
		}
	}
	return short, nil
}

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
	return statusDashboardLinesWithRefresh(state, derived, readiness, growth, runs, goals, statusRefresh{})
}

func statusDashboardLinesWithRefresh(state projectState, derived goalState, readiness readinessState, growth growthState, runs, goals int, refresh statusRefresh) []string {
	project := compactText(firstNonBlank(state.Project, "Unknown project"), 120)
	stage := normalizeRuntimeStage(firstNonBlank(state.Stage, readiness.Stage, "Unknown stage"))
	lines := []string{
		"Hyper Run Status",
		"",
		"Project: " + project,
		"Stage: " + stage,
		"Run mode: " + stateRunMode(state),
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
	if refresh.Needed {
		lines = append(lines, "State refresh: needed - "+refresh.Reason)
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
	lines = append(lines, statusActionLinesWithRefresh(state, derived, readiness, growth, refresh)...)
	lines = append(lines, "")
	lines = append(lines, pressureDashboardLines(growth)...)
	lines = append(lines, "")
	lines = append(lines, readinessDashboardLines(readiness)...)
	lines = append(lines,
		"",
		"Next:",
		"  "+statusNextCommandWithRefresh(state, derived, readiness, refresh),
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

func statusShortLines(state projectState, derived goalState, readiness readinessState, growth growthState) []string {
	return statusShortLinesWithRefresh(state, derived, readiness, growth, statusRefresh{})
}

func statusShortLinesWithRefresh(state projectState, derived goalState, readiness readinessState, growth growthState, refresh statusRefresh) []string {
	project := compactText(firstNonBlank(state.Project, "Unknown project"), 80)
	stage := normalizeRuntimeStage(firstNonBlank(state.Stage, readiness.Stage, "Unknown stage"))
	next := statusNextCommandWithRefresh(state, derived, readiness, refresh)
	lines := []string{
		"Hyper Run Status",
		"Project: " + project,
		"Stage: " + stage,
		"Mode: " + stateRunMode(state),
		"Gate: " + readinessGateSummary(readiness),
		"Proof: " + proofStatusSummary(derived, readiness),
		"Packet: " + shortPacketSummary(state, derived),
		"Next: " + next,
		"Why: " + statusActionReasonWithRefresh(state, derived, readiness, growth, refresh),
	}
	if refresh.Needed {
		lines = append(lines, "Refresh: "+refresh.Reason)
	}
	if benchmark := referenceBenchmarkShortStatus(readiness); benchmark != "" {
		lines = append(lines, "Benchmark: "+benchmark)
	}
	if gap := statusShortGap(readiness); gap != "" {
		lines = append(lines, "Gap: "+gap)
	}
	if guard := statusShortGuardWithRefresh(state, derived, readiness, growth, refresh); guard != "" {
		lines = append(lines, "Guard: "+guard)
	}
	lines = append(lines, "")
	return lines
}

func stateRunMode(state projectState) string {
	if !state.AutoContinue {
		return "single packet"
	}
	if state.RunUntil != "" {
		return "auto until " + state.RunUntil
	}
	return "auto"
}

func shortPacketSummary(state projectState, derived goalState) string {
	goalID := firstNonBlank(state.CurrentGoalID, "none")
	if strings.TrimSpace(derived.State) == "" {
		return goalID
	}
	return goalID + " (" + derived.State + ")"
}

func statusShortGap(readiness readinessState) string {
	if readiness.Version == 0 {
		return ""
	}
	if readiness.StageGate.CurrentStage == readiness.StageGate.NextStage && readiness.StageGate.Status == "ready" {
		return ""
	}
	if readiness.StageGate.Advancement.Candidate {
		return "none; stage advancement is ready"
	}
	if readiness.NextPressure.Axis != "" && readiness.NextPressure.Axis != "stage_advancement" {
		dim := readinessDimensionMap(readiness.Dimensions)[readiness.NextPressure.Axis]
		if dim.ID != "" {
			return compactText(readiness.NextPressure.AxisName+": "+firstNonBlank(dim.Gap, dim.Evidence, readiness.NextPressure.Reason), 120)
		}
		return compactText(readinessPressureSummary(readiness), 120)
	}
	if len(readiness.StageGate.BlockingGaps) > 0 {
		return compactText(readiness.StageGate.BlockingGaps[0], 120)
	}
	if gap := nextProofGap(readiness); gap != "" && gap != "none" {
		return gap
	}
	return ""
}

func statusShortGuardWithRefresh(state projectState, derived goalState, readiness readinessState, growth growthState, refresh statusRefresh) string {
	if statusRefreshActionable(state, derived, refresh) {
		return "run `hyper migrate` before advancing or starting another packet"
	}
	warning := statusDoNotDoYet(state, derived, readiness, growth)
	if strings.HasPrefix(warning, "Do not add broad structure") {
		return ""
	}
	if derived.State == "active" || (strings.TrimSpace(state.Status) != "" && strings.TrimSpace(state.Status) != strings.TrimSpace(derived.State)) {
		return warning
	}
	if readiness.StageGate.Advancement.Candidate {
		return "accept the stage change before running `hyper advance`"
	}
	return warning
}

type statusRefresh struct {
	Needed bool
	Reason string
}

func statusRefreshFor(root string) statusRefresh {
	growth := readGrowthStateIfExists(root)
	if growth.Version != 0 {
		if growthHasUnstoredManualActiveCapability(root, growth) {
			return statusRefresh{Needed: true, Reason: "active capability files are not reflected in stored growth state; run `hyper migrate`"}
		}
		if growthMigrationNeeded(growth) {
			return statusRefresh{Needed: true, Reason: "legacy or noisy growth entries found; run `hyper migrate`"}
		}
	}
	stored := readReadinessStateIfExists(root)
	if stored.Version == 0 || !exists(filepath.Join(root, planFile)) {
		return statusRefresh{}
	}
	current := readinessStateForStatus(root, growthStateForStatus(root))
	if current.Version != 0 && !sameReadinessForDoctor(stored, current) {
		return statusRefresh{Needed: true, Reason: "stored readiness differs from current evidence; run `hyper migrate`"}
	}
	return statusRefresh{}
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
	parts := []string{"functional " + functional}
	if proofAxisVisible(readiness, "core_ux") {
		parts = append(parts, "surface "+proofAxisStatus(readiness, "core_ux"))
	}
	if proofAxisVisible(readiness, "validation_coverage") {
		parts = append(parts, "operational "+proofAxisStatus(readiness, "validation_coverage"))
	}
	summary := strings.Join(parts, ", ")
	if referenceBenchmarkRelevant(readiness) {
		summary += ", benchmark " + proofAxisStatus(readiness, "reference_benchmark")
	}
	return summary
}

func proofAxisVisible(readiness readinessState, axis string) bool {
	status := proofAxisStatus(readiness, axis)
	return readinessAxisRequired(readiness, axis) || status == "covered" || readiness.NextPressure.Axis == axis
}

func readinessAxisRequired(readiness readinessState, axis string) bool {
	for _, required := range readiness.StageGate.RequiredAxes {
		if required == axis {
			return true
		}
	}
	return false
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
	if readiness.StageGate.CurrentStage == readiness.StageGate.NextStage && readiness.StageGate.Status == "ready" {
		return "none"
	}
	switch {
	case readinessAxisRequired(readiness, "core_ux") && proofAxisStatus(readiness, "core_ux") != "covered":
		return "surface proof for the primary user flow"
	case readinessAxisRequired(readiness, "validation_coverage") && proofAxisStatus(readiness, "validation_coverage") != "covered":
		return "repeatable validation proof"
	case readiness.NextPressure.AxisName != "":
		return readiness.NextPressure.AxisName
	default:
		return "none"
	}
}

func statusActionLinesWithRefresh(state projectState, derived goalState, readiness readinessState, growth growthState, refresh statusRefresh) []string {
	lines := []string{"Action:"}
	lines = append(lines, "  Next action: "+statusNextCommandWithRefresh(state, derived, readiness, refresh))
	lines = append(lines, "  Why now: "+statusActionReasonWithRefresh(state, derived, readiness, growth, refresh))
	lines = append(lines, "  Do not do yet: "+statusDoNotDoYetWithRefresh(state, derived, readiness, growth, refresh))
	return lines
}

func statusActionReason(state projectState, derived goalState, readiness readinessState, growth growthState) string {
	return statusActionReasonWithRefresh(state, derived, readiness, growth, statusRefresh{})
}

func statusActionReasonWithRefresh(state projectState, derived goalState, readiness readinessState, growth growthState, refresh statusRefresh) string {
	if statusRefreshActionable(state, derived, refresh) {
		return "Project state needs refresh before trusting the next action: " + refresh.Reason
	}
	if derived.State == "active" {
		if isFailedFinishGateReason(derived.Reason) {
			return derived.Reason
		}
		return "The current runtime packet is still open; evidence and next.md decide what the project learns."
	}
	if strings.TrimSpace(state.Status) != "" && strings.TrimSpace(state.Status) != strings.TrimSpace(derived.State) {
		return "The packet evidence says " + derived.State + " while state.json still says " + state.Status + "; repair before trusting automation."
	}
	if state.AutoContinue && runUntilReached(state, readiness) {
		return "Auto target " + state.RunUntil + " is reached; review status before choosing a new target or manual next run."
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
	return statusDoNotDoYetWithRefresh(state, derived, readiness, growth, statusRefresh{})
}

func statusDoNotDoYetWithRefresh(state projectState, derived goalState, readiness readinessState, growth growthState, refresh statusRefresh) string {
	if statusRefreshActionable(state, derived, refresh) {
		return "Do not advance or start another packet until `hyper migrate` refreshes growth and readiness state."
	}
	if derived.State == "active" {
		if isFailedFinishGateReason(derived.Reason) {
			return "Do not start another `hyper run`; fix review.md findings in the same packet and run `hyper complete` again."
		}
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
		return growthCandidateNameForCommand(prefix, command)
	}
	return firstNonBlank(name, candidate.Kind, "candidate")
}

func candidateDisplayPrefix(candidate growthCandidate) string {
	name := strings.ToLower(strings.TrimSpace(candidate.Name))
	for _, prefix := range []string{
		"validator-responsive-check",
		"validator-accessibility-check",
		"validator-visual-smoke",
		"validator",
		"preflight",
		"skill",
		"harness",
	} {
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
	for _, dim := range visibleReadinessDimensions(readiness) {
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
	if benchmark := referenceBenchmarkDashboardStatus(readiness); benchmark != "" {
		lines = append(lines, "  "+benchmark)
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

func visibleReadinessDimensions(readiness readinessState) []readinessDimension {
	required := readinessRequiredAxisMap(readiness)
	if len(required) == 0 {
		return readiness.Dimensions
	}
	visible := []readinessDimension{}
	for _, dim := range readiness.Dimensions {
		if dim.ID == "reference_benchmark" && !referenceBenchmarkRelevant(readiness) {
			continue
		}
		if dim.Status != "missing" || required[dim.ID] || dim.ID == readiness.NextPressure.Axis {
			visible = append(visible, dim)
		}
	}
	return visible
}

func readinessRequiredAxisMap(readiness readinessState) map[string]bool {
	required := map[string]bool{}
	for _, axis := range readiness.StageGate.RequiredAxes {
		required[axis] = true
	}
	return required
}

func referenceBenchmarkRelevant(readiness readinessState) bool {
	return readinessRequiredAxisMap(readiness)["reference_benchmark"] || readiness.NextPressure.Axis == "reference_benchmark"
}

func referenceBenchmarkShortStatus(readiness readinessState) string {
	if !referenceBenchmarkRelevant(readiness) {
		return ""
	}
	dim, ok := readinessDimensionMap(readiness.Dimensions)["reference_benchmark"]
	if !ok {
		return "missing"
	}
	return dim.Status + " - " + compactText(firstNonBlank(dim.Evidence, dim.Gap), 100)
}

func referenceBenchmarkDashboardStatus(readiness readinessState) string {
	if !referenceBenchmarkRelevant(readiness) {
		return ""
	}
	dim, ok := readinessDimensionMap(readiness.Dimensions)["reference_benchmark"]
	if !ok {
		return "Reference benchmark: missing"
	}
	return "Reference benchmark: " + dim.Status + " - " + compactText(firstNonBlank(dim.Evidence, dim.Gap), 140)
}

func statusNextCommandWithRefresh(state projectState, derived goalState, readiness readinessState, refresh statusRefresh) string {
	if statusRefreshActionable(state, derived, refresh) {
		return "hyper migrate"
	}
	if strings.TrimSpace(state.Status) != "" && strings.TrimSpace(derived.State) != "" && strings.TrimSpace(state.Status) != strings.TrimSpace(derived.State) {
		return "hyper repair"
	}
	if strings.TrimSpace(state.CurrentGoalID) == "" {
		if state.AutoContinue && runUntilReached(state, readiness) {
			return "hyper status --short"
		}
		return "hyper run [focus]"
	}
	if derived.State == "active" {
		return "update " + strings.TrimSuffix(state.CurrentGoalPath, "goal.md") + "evidence.md and next.md, then run `hyper complete`"
	}
	if state.AutoContinue && runUntilReached(state, readiness) {
		return "hyper status --short"
	}
	if readiness.NextPressure.Axis == "stage_advancement" || readiness.StageGate.Advancement.Candidate {
		return "hyper advance"
	}
	if readiness.NextPressure.RecommendedGoal != "" {
		if state.AutoContinue {
			return autoRunCommand(state, readiness.NextPressure.RecommendedGoal)
		}
		return "hyper run \"" + compactText(readiness.NextPressure.RecommendedGoal, 120) + "\""
	}
	if state.AutoContinue {
		return autoRunCommand(state, "")
	}
	return "hyper run [next focus]"
}

func statusRefreshActionable(state projectState, derived goalState, refresh statusRefresh) bool {
	if !refresh.Needed {
		return false
	}
	if derived.State == "active" {
		return false
	}
	if strings.TrimSpace(state.Status) != "" && strings.TrimSpace(derived.State) != "" && strings.TrimSpace(state.Status) != strings.TrimSpace(derived.State) {
		return false
	}
	return true
}
