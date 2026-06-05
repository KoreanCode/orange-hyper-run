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
	if terminalPacketState(derived.State) {
		return plannedNextPacket{
			Action:   "stop",
			Command:  "hyper status --short",
			Reason:   firstNonBlank(derived.Reason, "Runtime packet stopped: "+derived.State),
			Terminal: true,
		}
	}
	if state.AutoContinue && runUntilReached(state, readiness) {
		return plannedNextPacket{
			Action:   "stop",
			Command:  "hyper status --short",
			Reason:   firstNonBlank(statusActionReason(state, derived, readiness, growth), "Run-until target proof complete: "+state.RunUntil),
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
		command := nextRunCommand(state, readiness.NextPressure.RecommendedGoal)
		return plannedNextPacket{
			Action:  "run",
			Command: command,
			Reason:  readiness.NextPressure.Reason,
		}
	}
	command := nextRunCommand(state, "")
	return plannedNextPacket{
		Action:  "run",
		Command: command,
		Reason:  firstNonBlank(statusActionReason(state, derived, readiness, growth), "Continue with the next smallest runtime packet."),
	}
}

func writeNextPacketPlan(root string, state projectState, derived goalState, readiness readinessState, growth growthState) (plannedNextPacket, *hyperError) {
	plan := buildNextPacketPlan(state, derived, readiness, growth)
	body := renderNextPacketPlan(state, readiness, plan)
	if review := nextPacketReviewFindings(root, state, plan); review != "" {
		body = strings.Replace(body, "## Guard", review+"## Guard", 1)
	}
	if err := writeText(filepath.Join(root, hyperDir, "next-packet.md"), body); err != nil {
		return plan, err
	}
	return plan, nil
}

func nextPacketReviewFindings(root string, state projectState, plan plannedNextPacket) string {
	if plan.Action != "complete-current" {
		return ""
	}
	findings := finishGateReviewFindings(root, state.CurrentGoalID)
	if len(findings) == 0 {
		return ""
	}
	lines := []string{"## Current Review Findings", ""}
	for _, finding := range findings {
		lines = append(lines, "- "+finding)
	}
	if note := finishGateReviewRepeatNote(root, state.CurrentGoalID); note != "" {
		lines = append(lines, "", "- "+note)
	}
	lines = append(lines, "", "")
	return strings.Join(lines, "\n")
}

func renderNextPacketPlan(state projectState, readiness readinessState, plan plannedNextPacket) string {
	lines := []string{
		"# Next Packet Plan",
		"",
		"Mode: " + nextPacketMode(state),
		"Action: " + plan.Action,
		"Command: " + plan.Command,
		"Reason: " + plan.Reason,
		"Readiness gate: " + readinessGateSummary(readiness),
		"Readiness pressure: " + readinessPressureSummary(readiness),
		"",
	}
	if plan.Action == "advance" {
		lines = append(lines, stageAdvancementReviewLines(readiness, state)...)
		lines = append(lines, "")
	}
	lines = append(lines,
		"## Guard",
		"",
		nextPacketGuard(state, plan),
		"",
	)
	if progress := nextPacketProgressGuard(state, plan); progress != "" {
		lines = append(lines,
			"## Progress Guard",
			"",
			progress,
			"",
		)
	}
	lines = append(lines,
		"## Codex Desktop Continuation",
		"",
		nextPacketCodexContinuation(state, plan),
		"",
	)
	return strings.Join(lines, "\n")
}

func nextPacketMode(state projectState) string {
	if !state.AutoContinue {
		return "single packet"
	}
	mode := "auto"
	if state.RunUntil != "" {
		mode += " until " + state.RunUntil
	}
	return mode
}

func nextPacketGuard(state projectState, plan plannedNextPacket) string {
	switch plan.Action {
	case "advance":
		if stageAdvanceAutoAuthorized(state) {
			return "Run `hyper advance` only after the Stage Advancement Review shows ready proof and no blocking gaps; the active auto target authorizes continuing toward " + state.RunUntil + "."
		}
		return "Do not run `hyper advance` unless the user accepts the stage change."
	case "complete-current":
		return "Do not create a new runtime packet; fix the current packet evidence, next notes, and review findings before running `hyper complete`."
	case "run":
		return "Create the next runtime packet only after the current packet has passed the finish gate and completed."
	case "stop":
		if terminalPacketState(state.Status) {
			return "Runtime packet is " + state.Status + ". Stop automatic continuation and resolve the blocker or user-waiting condition before starting more work."
		}
		if state.RunTargetSource == planTargetStageSource {
			return "Run-until target proof is complete. Raise or remove `plan.md` Target Stage before starting more work."
		}
		return "Run-until target proof is complete. Review status before choosing a new target."
	default:
		return "Review `hyper status --short` before continuing."
	}
}

func nextPacketProgressGuard(state projectState, plan plannedNextPacket) string {
	if !state.AutoContinue {
		return ""
	}
	switch plan.Action {
	case "run":
		return "Continue only if the command creates a new runtime packet or the next plan changes stage, readiness pressure, action, or command. If the same command repeats without new evidence or stage movement, stop and report the loop risk."
	case "advance":
		return "Continue only if `hyper advance` changes `plan.md` Current Stage and the refreshed next-packet plan changes action or command. If `hyper advance` leaves the same plan in place, stop and run `hyper doctor`."
	case "complete-current":
		return "Retry only after evidence.md or next.md changes directly address the review findings. If the same findings repeat after a fix attempt, stop and report the repeated finish-gate failure."
	case "stop":
		if terminalPacketState(state.Status) {
			return "Do not continue automatically after stop. Report the " + state.Status + " packet state and wait until the blocker or user-waiting condition is resolved."
		}
		return "Do not continue automatically after stop. Report the target-proof-complete state and wait for a higher target or manual follow-up."
	default:
		return "Continue only while each step produces new evidence, a new runtime packet, a stage change, or a changed next-packet plan."
	}
}

func nextPacketProgressGuardLine(state projectState, plan plannedNextPacket) string {
	progress := nextPacketProgressGuard(state, plan)
	if strings.TrimSpace(progress) == "" {
		return ""
	}
	return "Progress guard: " + compactText(progress, 220)
}

func nextPacketCodexContinuation(state projectState, plan plannedNextPacket) string {
	switch plan.Action {
	case "advance":
		if stageAdvanceAutoAuthorized(state) {
			return "Continue by running `hyper advance`, then read the refreshed `.hyper/next-packet.md` and follow only that planned command."
		}
		return "Pause here. Tell the user the stage gate is ready and run `hyper advance` only after the user accepts the stage change."
	case "complete-current":
		return "Stay in the current runtime packet. Fix evidence, next notes, and review findings, then run `hyper complete` again."
	case "run":
		return "Continue automatically by running the command above, then read the newly generated runtime packet and execute it checkpoint by checkpoint."
	case "stop":
		if terminalPacketState(state.Status) {
			return "Stop the auto loop. Report the " + state.Status + " runtime packet state and wait for user input or a deliberate manual follow-up."
		}
		if state.RunTargetSource == planTargetStageSource {
			return "Stop the auto loop. Report that the plan target proof is complete and wait for the user to raise `plan.md` Target Stage, remove it for manual work, or choose an override target."
		}
		return "Stop the auto loop. Report the current status and wait for the user to choose a new target or a manual follow-up packet."
	default:
		return "Review `hyper status --short` before continuing."
	}
}

func terminalPacketState(state string) bool {
	switch strings.TrimSpace(state) {
	case "blocked", "waiting_user":
		return true
	default:
		return false
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

func nextRunCommand(state projectState, focus string) string {
	if state.AutoContinue && state.RunTargetSource != planTargetStageSource {
		return autoRunCommand(state, focus)
	}
	parts := []string{"hyper", "run"}
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
	currentRank := stageRank(current)
	targetRank := stageRank(target)
	if currentRank == 0 || targetRank == 0 {
		return false
	}
	if currentRank > targetRank {
		return true
	}
	if currentRank < targetRank {
		return false
	}
	return targetStageProofComplete(target, readiness)
}

func targetStageProofComplete(target string, readiness readinessState) bool {
	if readiness.StageGate.Status != "ready" {
		return false
	}
	return normalizeRuntimeStage(readiness.StageGate.CurrentStage) == normalizeRuntimeStage(target)
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
	value = strings.Join(strings.Fields(value), " ")
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
