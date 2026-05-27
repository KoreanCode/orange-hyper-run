package app

import (
	"path/filepath"
	"strings"
)

type plannedNextPacket struct {
	Action   string
	Command  string
	Reason   string
	Terminal bool
}

func buildNextPacketPlan(state projectState, derived goalState, readiness readinessState, growth growthState) plannedNextPacket {
	if derived.State == "active" {
		return plannedNextPacket{
			Action:  "complete-current",
			Command: "hyper complete",
			Reason:  statusActionReason(state, derived, readiness, growth),
		}
	}
	if state.AutoContinue && runUntilReached(state, readiness) {
		return plannedNextPacket{
			Action:   "stop",
			Command:  "hyper status --short",
			Reason:   firstNonBlank(statusActionReason(state, derived, readiness, growth), "Run-until target reached: "+state.RunUntil),
			Terminal: true,
		}
	}
	if readiness.NextPressure.Axis == "stage_advancement" || readiness.StageGate.Advancement.Candidate {
		return plannedNextPacket{
			Action:  "advance",
			Command: "hyper advance",
			Reason:  statusActionReason(state, derived, readiness, growth),
		}
	}
	if readiness.NextPressure.RecommendedGoal != "" {
		command := "hyper run " + quoteCommandArg(readiness.NextPressure.RecommendedGoal)
		if state.AutoContinue {
			command = autoRunCommand(state, readiness.NextPressure.RecommendedGoal)
		}
		return plannedNextPacket{
			Action:  "run",
			Command: command,
			Reason:  readiness.NextPressure.Reason,
		}
	}
	command := "hyper run [next focus]"
	if state.AutoContinue {
		command = autoRunCommand(state, "")
	}
	return plannedNextPacket{
		Action:  "run",
		Command: command,
		Reason:  firstNonBlank(statusActionReason(state, derived, readiness, growth), "Continue with the next smallest runtime packet."),
	}
}

func writeNextPacketPlan(root string, state projectState, derived goalState, readiness readinessState, growth growthState) (plannedNextPacket, *hyperError) {
	plan := buildNextPacketPlan(state, derived, readiness, growth)
	body := renderNextPacketPlan(state, readiness, plan)
	if err := writeText(filepath.Join(root, hyperDir, "next-packet.md"), body); err != nil {
		return plan, err
	}
	return plan, nil
}

func renderNextPacketPlan(state projectState, readiness readinessState, plan plannedNextPacket) string {
	mode := "single packet"
	if state.AutoContinue {
		mode = "auto"
		if state.RunUntil != "" {
			mode += " until " + state.RunUntil
		}
	}
	return strings.Join([]string{
		"# Next Packet Plan",
		"",
		"Mode: " + mode,
		"Action: " + plan.Action,
		"Command: " + plan.Command,
		"Reason: " + plan.Reason,
		"Readiness gate: " + readinessGateSummary(readiness),
		"Readiness pressure: " + readinessPressureSummary(readiness),
		"",
		"## Guard",
		"",
		nextPacketGuard(plan),
		"",
	}, "\n")
}

func nextPacketGuard(plan plannedNextPacket) string {
	switch plan.Action {
	case "advance":
		return "Do not run `hyper advance` unless the user accepts the stage change."
	case "complete-current":
		return "Do not create a new runtime packet; fix the current packet evidence, next notes, and review findings before running `hyper complete`."
	case "run":
		return "Create the next runtime packet only after the current packet has passed the finish gate and completed."
	case "stop":
		return "Run-until target is reached. Review status before choosing a new target."
	default:
		return "Review `hyper status --short` before continuing."
	}
}

func autoRunCommand(state projectState, focus string) string {
	parts := []string{"hyper", "run", "--auto"}
	if state.RunUntil != "" {
		parts = append(parts, "--until", quoteCommandArg(state.RunUntil))
	}
	if strings.TrimSpace(focus) != "" {
		parts = append(parts, quoteCommandArg(focus))
	}
	return strings.Join(parts, " ")
}

func runUntilReached(state projectState, readiness readinessState) bool {
	target := strings.TrimSpace(state.RunUntil)
	if target == "" {
		return false
	}
	current := normalizeRuntimeStage(firstNonBlank(readiness.Stage, state.Stage))
	return stageRank(current) >= stageRank(target)
}

func stageRank(stage string) int {
	switch normalizeRuntimeStage(stage) {
	case "Tiny MVP":
		return 1
	case "Usable MVP":
		return 2
	case "Beta":
		return 3
	case "Service Quality":
		return 4
	case "Sustained Service Quality":
		return 5
	default:
		return 0
	}
}

func quoteCommandArg(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
