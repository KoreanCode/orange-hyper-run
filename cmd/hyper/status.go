package main

import (
	"fmt"
	"strings"
)

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
		"Principles: " + growthPrinciplesLine(),
		"Status: " + state.Status,
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
	lines = append(lines, readinessDashboardLines(readiness)...)
	lines = append(lines,
		"",
		"Next:",
		"  "+statusNextCommand(state, derived, readiness),
		"",
		fmt.Sprintf("Runs recorded: %d", runs),
		fmt.Sprintf("Runtime packets recorded: %d", goals),
		fmt.Sprintf("Growth pressures: %d", len(growth.Pressures)),
		fmt.Sprintf("Capability candidates: %d", len(growth.Candidates)),
		"Updated: "+state.UpdatedAt,
		"",
	)
	return lines
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
	if readiness.NextPressure.RecommendedGoal != "" {
		lines = append(lines, "  Recommended run: hyper run \""+compactText(readiness.NextPressure.RecommendedGoal, 120)+"\"")
	}
	return lines
}

func statusNextCommand(state projectState, derived goalState, readiness readinessState) string {
	if strings.TrimSpace(state.CurrentGoalID) == "" {
		return "hyper run [focus]"
	}
	if derived.State == "active" {
		return "update " + strings.TrimSuffix(state.CurrentGoalPath, "goal.md") + "evidence.md and next.md, then run `hyper complete`"
	}
	if readiness.NextPressure.RecommendedGoal != "" {
		return "hyper run \"" + compactText(readiness.NextPressure.RecommendedGoal, 120) + "\""
	}
	return "hyper run [next focus]"
}
