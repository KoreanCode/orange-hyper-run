package app

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func initHyper(fsys fsRoot) (commandOutput, *hyperError) {
	root := fsys.root()
	planResult, err := ensurePlanForInit(root)
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
	if err := ensureCodexDesktopRules(root); err != nil {
		return commandOutput{}, err
	}
	planCandidatePath, err := maybeWritePlanImportCandidates(root, planResult.Body)
	if err != nil {
		return commandOutput{}, err
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
	episode := compileGoalEpisode("GOAL-0000", "", planResult.Body, runOptions{}, nil, growth, readiness)
	now := nowISO()
	planHash := hashText(planResult.Body)
	stage := firstNonBlank(episode.Stage, "Tiny MVP")
	existing := readStateIfExists(root)
	hasActiveGoal := strings.TrimSpace(existing.CurrentGoalID) != ""

	state := existing
	if hasActiveGoal {
		state.Project = firstNonBlank(state.Project, episode.Plan["Product"], "Unknown project")
		state.Stage = firstNonBlank(state.Stage, stage)
		state.ExecutionAdapter = firstNonBlank(state.ExecutionAdapter, defaultExecutionAdapter())
		state.PlanPath = planFile
		state.PlanHash = planHash
		state.UpdatedAt = now
		state = applyPlanTargetToState(state, episode.Plan)
	} else {
		state = projectState{
			Project:          firstNonBlank(episode.Plan["Product"], "Unknown project"),
			Stage:            stage,
			Status:           "initialized",
			ExecutionAdapter: defaultExecutionAdapter(),
			PlanPath:         planFile,
			PlanHash:         planHash,
			UpdatedAt:        now,
		}
		state = applyPlanTargetToState(state, episode.Plan)
	}
	if err := writeJSON(filepath.Join(root, hyperDir, "state.json"), state); err != nil {
		return commandOutput{}, err
	}

	event := map[string]any{
		"type":           initEventType(planResult.Created, hasActiveGoal),
		"path":           planFile,
		"plan_hash":      planHash,
		"active_goal_id": nullableString(state.CurrentGoalID),
		"created_at":     nowISO(),
	}
	if err := appendJSONL(filepath.Join(root, hyperDir, "logs", "project.jsonl"), event); err != nil {
		return commandOutput{}, err
	}
	if err := insertEvent(db, event); err != nil {
		return commandOutput{}, err
	}

	lines := []string{
		"Project: " + state.Project,
		"Stage: " + state.Stage,
		statusRunTargetLine(state),
		"Stage contract: " + stageGrowthContract(state.Stage),
		"Status: " + state.Status,
		"Method: " + growthRuntimeDefinition,
		"Protocol: " + runtimeProtocolDefinition,
		"Readiness gate: " + readinessGateSummary(readiness),
		"Readiness pressure: " + readinessPressureSummary(readiness),
		"Pressure ledger: " + growthLoopStateSummary(growth),
		"Plan file: " + planFile,
		"Hyper dir: " + hyperDir + "/",
	}
	if hasActiveGoal {
		lines = append(lines, "Active runtime packet preserved: "+state.CurrentGoalID)
	}
	if planCandidatePath != "" {
		lines = append(lines, "Plan import candidates: "+planCandidatePath)
	}
	nextRun := "  hyper run [focus]"
	if state.AutoContinue && state.RunTargetSource == planTargetStageSource {
		nextRun = "  hyper run"
	}
	next := []string{"  Fill in plan.md", nextRun}
	if hasActiveGoal {
		next = []string{"  hyper resume"}
	}
	lines = append(lines,
		"",
		initSummary(planResult, hasActiveGoal),
		"",
		"Next:",
	)
	lines = append(lines, next...)
	lines = append(lines, "", "Codex Desktop:", "  $hyper run", "")
	return stdout(strings.Join(lines, "\n")), nil
}

func runHyper(fsys fsRoot, opts runOptions) (commandOutput, *hyperError) {
	root := fsys.root()
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
	if err := ensureCodexDesktopRules(root); err != nil {
		return commandOutput{}, err
	}
	planCandidatePath, err := maybeWritePlanImportCandidates(root, planResult.Body)
	if err != nil {
		return commandOutput{}, err
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

	previous := readStateIfExists(root)
	opts, err = applyDefaultRunTarget(opts, plan, previous)
	if err != nil {
		return commandOutput{}, err
	}
	focus := opts.Focus
	if blocked := blockingActiveGoal(root, previous); blocked != "" {
		return commandOutput{}, newError(blocked, 2)
	}
	autoLearn, err := learnGoalFromState(root, previous, db, "auto_learn_completed", false)
	if err != nil {
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
	if opts.AutoContinue && strings.TrimSpace(focus) == "" && terminalPacketState(previous.Status) && strings.TrimSpace(previous.CurrentGoalID) != "" {
		derived := deriveCurrentGoalState(root, previous.CurrentGoalID)
		if terminalPacketState(derived.State) {
			stopState := runUntilStopState(previous, opts, planResult.Body, readiness)
			if err := writeJSON(filepath.Join(root, hyperDir, "state.json"), stopState); err != nil {
				return commandOutput{}, err
			}
			nextPlan, err := writeNextPacketPlan(root, stopState, derived, readiness, growth)
			if err != nil {
				return commandOutput{}, err
			}
			if err := recordNoPacketRun(root, db, stopState, "terminal_packet_stop", nextPlan, readiness, autoLearn); err != nil {
				return commandOutput{}, err
			}
			return stdout(strings.Join(compactNonEmptyLines([]string{
				"Runtime packet stopped: " + previous.CurrentGoalID,
				"State: " + derived.State,
				"Reason: " + derived.Reason,
				"Run mode: " + formatRunMode(opts),
				runTargetSourceLine(opts),
				"Auto learn: " + formatAutoLearn(autoLearn),
				"Readiness gate: " + readinessGateSummary(readiness),
				"Readiness pressure: " + readinessPressureSummary(readiness),
				"Planned action: " + nextPlan.Action,
				"Next action: " + nextPacketActionDisplay(nextPlan),
				"Why: " + nextPlan.Reason,
				"Continuation guard: " + compactText(nextPacketGuard(stopState, nextPlan), 220),
				nextPacketProgressGuardLine(stopState, nextPlan),
				"Next packet plan: " + displayRelPath(hyperDir, "next-packet.md"),
				"",
				"No runtime packet created.",
				"",
				"Next:",
				"  " + nextPacketActionDisplay(nextPlan),
				"",
			}), "\n")), nil
		}
	}
	if opts.AutoContinue && strings.TrimSpace(opts.RunUntil) != "" {
		stopState := runUntilStopState(previous, opts, planResult.Body, readiness)
		if runUntilReached(stopState, readiness) {
			if err := writeJSON(filepath.Join(root, hyperDir, "state.json"), stopState); err != nil {
				return commandOutput{}, err
			}
			nextPlan, err := writeNextPacketPlan(root, stopState, runUntilStopDerived(stopState), readiness, growth)
			if err != nil {
				return commandOutput{}, err
			}
			if err := recordNoPacketRun(root, db, stopState, "auto_target_reached", nextPlan, readiness, autoLearn); err != nil {
				return commandOutput{}, err
			}
			return stdout(strings.Join(compactNonEmptyLines([]string{
				"Run-until target proof complete: " + opts.RunUntil,
				"Stage: " + normalizeRuntimeStage(firstNonBlank(readiness.Stage, stopState.Stage)),
				"Run mode: " + formatRunMode(opts),
				runTargetSourceLine(opts),
				"Auto learn: " + formatAutoLearn(autoLearn),
				"Readiness gate: " + readinessGateSummary(readiness),
				"Readiness pressure: " + readinessPressureSummary(readiness),
				"Planned action: " + nextPlan.Action,
				"Next action: " + nextPacketActionDisplay(nextPlan),
				"Why: " + nextPlan.Reason,
				"Continuation guard: " + compactText(nextPacketGuard(stopState, nextPlan), 220),
				nextPacketProgressGuardLine(stopState, nextPlan),
				"Next packet plan: " + displayRelPath(hyperDir, "next-packet.md"),
				"",
				"No runtime packet created.",
				"",
				"Next:",
				"  " + nextPacketActionDisplay(nextPlan),
				"",
			}), "\n")), nil
		}
		if readiness.NextPressure.Axis == "stage_advancement" || readiness.StageGate.Advancement.Candidate {
			advanceState := runUntilStopState(previous, opts, planResult.Body, readiness)
			if err := writeJSON(filepath.Join(root, hyperDir, "state.json"), advanceState); err != nil {
				return commandOutput{}, err
			}
			derived := currentStateConsistency(root, advanceState).Derived
			if strings.TrimSpace(derived.State) == "" {
				derived = goalState{State: firstNonBlank(advanceState.Status, "completed"), Reason: "Stage gate is ready."}
			}
			nextPlan, err := writeNextPacketPlan(root, advanceState, derived, readiness, growth)
			if err != nil {
				return commandOutput{}, err
			}
			if err := recordNoPacketRun(root, db, advanceState, "stage_gate_ready", nextPlan, readiness, autoLearn); err != nil {
				return commandOutput{}, err
			}
			return stdout(strings.Join(compactNonEmptyLines([]string{
				"Stage gate ready: " + readinessGateSummary(readiness),
				"Stage: " + normalizeRuntimeStage(firstNonBlank(readiness.Stage, advanceState.Stage)),
				"Run mode: " + formatRunMode(opts),
				runTargetSourceLine(opts),
				"Auto learn: " + formatAutoLearn(autoLearn),
				"Readiness pressure: " + readinessPressureSummary(readiness),
				"Planned action: " + nextPlan.Action,
				"Next action: " + nextPacketActionDisplay(nextPlan),
				"Why: " + nextPlan.Reason,
				"Continuation guard: " + compactText(nextPacketGuard(advanceState, nextPlan), 220),
				nextPacketProgressGuardLine(advanceState, nextPlan),
				"Next packet plan: " + displayRelPath(hyperDir, "next-packet.md"),
				"",
				"No runtime packet created.",
				"",
				"Next:",
				"  " + nextPacketActionDisplay(nextPlan),
				"",
			}), "\n")), nil
		}
	}

	runID, err := nextID(db, "runs", "RUN")
	if err != nil {
		return commandOutput{}, err
	}
	goalID, err := nextID(db, "goals", "GOAL")
	if err != nil {
		return commandOutput{}, err
	}
	now := nowISO()
	episode := compileGoalEpisode(goalID, focus, planResult.Body, opts, nil, growth, readiness)
	similar, err := findSimilarContext(db, buildSimilarityQuery(episode.Plan, episode, focus), 5)
	if err != nil {
		return commandOutput{}, err
	}
	episode = compileGoalEpisode(goalID, focus, planResult.Body, opts, similar, growth, readiness)
	summary := fmt.Sprintf("Created %s for %s", goalID, episode.Stage)

	goalDir := filepath.Join(root, hyperDir, "goals", goalID)
	for name, body := range map[string]string{
		"goal.md":     episode.Docs.Goal,
		"tasks.md":    episode.Docs.Tasks,
		"evidence.md": episode.Docs.Evidence,
		"review.md":   episode.Docs.Review,
		"next.md":     episode.Docs.Next,
	} {
		if err := writeText(filepath.Join(goalDir, name), body); err != nil {
			return commandOutput{}, err
		}
	}

	planHash := hashText(planResult.Body)
	state := projectState{
		Project:          firstNonBlank(episode.Plan["Product"], "Unknown project"),
		Stage:            episode.Stage,
		Status:           "active",
		ActiveRunID:      runID,
		CurrentGoalID:    goalID,
		CurrentGoalPath:  fmt.Sprintf(".hyper/goals/%s/goal.md", goalID),
		ExecutionAdapter: defaultExecutionAdapter(),
		PlanPath:         planFile,
		PlanHash:         planHash,
		Focus:            focus,
		AutoContinue:     opts.AutoContinue,
		RunUntil:         opts.RunUntil,
		RunTargetSource:  opts.RunTargetSource,
		UpdatedAt:        now,
	}
	handoff := createExecutionHandoff(runID, goalID, state.AutoContinue)

	if err := insertRun(db, runID, episode.Objective, episode.Stage, "active", now, goalID, summary); err != nil {
		return commandOutput{}, err
	}
	if err := insertGoal(db, goalID, runID, episode, "active", now); err != nil {
		return commandOutput{}, err
	}

	events := []map[string]any{
		{"type": "run_started", "run_id": runID, "objective": episode.Objective, "stage": episode.Stage, "created_at": now},
		{"type": "plan_loaded", "run_id": runID, "path": planFile, "plan_hash": planHash, "created_at": nowISO()},
		{
			"type":               "auto_learn_checked",
			"run_id":             runID,
			"previous_run_id":    nullableString(autoLearn.RunID),
			"previous_goal_id":   nullableString(autoLearn.GoalID),
			"previous_state":     autoLearn.State,
			"inserted_memories":  autoLearn.Inserted,
			"skipped":            autoLearn.Skipped,
			"reason":             autoLearn.Reason,
			"growth_pressures":   visibleGrowthPressureCount(growth.Pressures),
			"growth_candidates":  visibleGrowthCandidateCount(growth.Candidates),
			"readiness_gate":     readiness.StageGate.Status,
			"readiness_pressure": readiness.NextPressure.Axis,
			"created_at":         nowISO(),
		},
		{"type": "similar_context_retrieved", "run_id": runID, "goal_id": goalID, "count": len(similar), "created_at": nowISO()},
		{"type": "goal_created", "run_id": runID, "goal_id": goalID, "path": state.CurrentGoalPath, "summary": summary, "created_at": nowISO()},
		{"type": handoff.EventType, "run_id": runID, "goal_id": goalID, "adapter": handoff.Adapter, "created_at": nowISO()},
	}
	logPath := filepath.Join(root, hyperDir, "logs", runID+".jsonl")
	for _, event := range events {
		if err := appendJSONL(logPath, event); err != nil {
			return commandOutput{}, err
		}
		if err := insertEvent(db, event); err != nil {
			return commandOutput{}, err
		}
	}
	if err := writeJSON(filepath.Join(root, hyperDir, "state.json"), state); err != nil {
		return commandOutput{}, err
	}

	lines := []string{
		"Project: " + state.Project,
		"Stage: " + episode.Stage,
		"Stage contract: " + stageGrowthContract(episode.Stage),
		"Run: " + runID,
		"Runtime packet: " + goalID,
		"Run mode: " + formatRunMode(opts),
		runTargetSourceLine(opts),
		"Auto learn: " + formatAutoLearn(autoLearn),
		"Readiness gate: " + readinessGateSummary(readiness),
		"Readiness pressure: " + readinessPressureSummary(readiness),
		"Pressure ledger: " + growthLoopStateSummary(growth),
		fmt.Sprintf("Similar context: %d", len(similar)),
		"Runtime packet file: " + state.CurrentGoalPath,
	}
	lines = append(lines, missingTargetStageAdvisory(opts, plan)...)
	if planCandidatePath != "" {
		lines = append(lines, "Plan import candidates: "+planCandidatePath)
	}
	lines = compactNonEmptyLines(lines)
	lines = append(lines,
		"",
		"Loaded plan.md as the product brief.",
		"",
		renderExecutionHandoff(handoff),
		"",
	)
	return stdout(strings.Join(lines, "\n")), nil
}

func statusHyper(fsys fsRoot, args []string) (commandOutput, *hyperError) {
	short, optionErr := parseStatusOptions(args)
	if optionErr != nil {
		return commandOutput{}, optionErr
	}
	root := fsys.root()
	statePath := filepath.Join(root, hyperDir, "state.json")
	if !exists(statePath) {
		return stdout("No Hyper Run state found. Start with `hyper init`.\n"), nil
	}
	state, err := readState(statePath)
	if err != nil {
		return commandOutput{}, err
	}
	state = refreshStateFromPlanForStatus(root, state)
	derived := deriveCurrentGoalState(root, state.CurrentGoalID)
	if failed, ok := failedFinishGateGoalState(root, state.CurrentGoalID); ok {
		derived = failed
		state.Status = "active"
	}
	runs, goals := statusDBCounts(root)
	growth := growthStateForStatus(root)
	readiness := readinessStateForStatus(root, growth)
	readiness = readinessWithPacketNextGoal(root, state, derived, readiness)
	refresh := statusRefreshFor(root, state)
	if short {
		lines := statusShortLinesWithRefresh(state, derived, readiness, growth, refresh)
		lines = appendStatusVerifiedEvidence(lines, root, state.CurrentGoalID, true)
		lines = appendStatusReviewFindings(lines, root, state.CurrentGoalID, derived)
		return stdout(strings.Join(lines, "\n")), nil
	}
	lines := statusDashboardLinesWithRefresh(state, derived, readiness, growth, runs, goals, refresh)
	lines = appendStatusVerifiedEvidence(lines, root, state.CurrentGoalID, false)
	lines = appendStatusReviewFindings(lines, root, state.CurrentGoalID, derived)
	return stdout(strings.Join(lines, "\n")), nil
}

func appendStatusVerifiedEvidence(lines []string, root, goalID string, short bool) []string {
	if strings.TrimSpace(goalID) == "" {
		return lines
	}
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if short {
		lines = append(lines, verifiedEvidenceShortLine(root, goalID))
		return append(lines, "")
	}
	lines = append(lines, verifiedEvidenceDashboardLines(root, goalID)...)
	return append(lines, "")
}

func appendStatusReviewFindings(lines []string, root, goalID string, derived goalState) []string {
	if !isFailedFinishGateReason(derived.Reason) {
		return lines
	}
	findings := finishGateReviewFindings(root, goalID)
	if len(findings) == 0 {
		return lines
	}
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	lines = append(lines, "Review findings:")
	for _, finding := range findings {
		lines = append(lines, "  - "+finding)
	}
	if note := finishGateReviewRepeatNote(root, goalID); note != "" {
		lines = append(lines, "  - "+note)
	}
	return append(lines, "")
}

func statusDBCounts(root string) (int, int) {
	fsRuns, fsGoals := statusFilesystemCounts(root)
	if !exists(filepath.Join(root, hyperDir, "hyper.sqlite")) {
		return fsRuns, fsGoals
	}
	db, err := openDB(root)
	if err != nil {
		return fsRuns, fsGoals
	}
	defer db.Close()
	runs, runErr := countRows(db, "runs")
	if runErr != nil {
		runs = fsRuns
	}
	goals, goalErr := countRows(db, "goals")
	if goalErr != nil {
		goals = fsGoals
	}
	if fsRuns > runs {
		runs = fsRuns
	}
	if fsGoals > goals {
		goals = fsGoals
	}
	return runs, goals
}

func statusFilesystemCounts(root string) (int, int) {
	runs := 0
	if entries, err := os.ReadDir(filepath.Join(root, hyperDir, "logs")); err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if !entry.IsDir() && strings.HasPrefix(name, "RUN-") && strings.HasSuffix(name, ".jsonl") {
				runs++
			}
		}
	}
	goals := 0
	if entries, err := os.ReadDir(filepath.Join(root, hyperDir, "goals")); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "GOAL-") {
				goals++
			}
		}
	}
	return runs, goals
}

func completeHyper(fsys fsRoot) (commandOutput, *hyperError) {
	root := fsys.root()
	statePath := filepath.Join(root, hyperDir, "state.json")
	if !exists(statePath) {
		return commandOutput{}, newError("No Hyper Run state found. Start with `hyper init`.", 2)
	}
	state, err := readState(statePath)
	if err != nil {
		return commandOutput{}, err
	}
	if strings.TrimSpace(state.CurrentGoalID) == "" {
		return commandOutput{}, newError("No active runtime packet found. Start with `hyper run`.", 2)
	}
	derived := deriveCurrentGoalState(root, state.CurrentGoalID)
	if derived.State == "active" {
		goalDir := strings.TrimSuffix(state.CurrentGoalPath, "goal.md")
		return commandOutput{}, newError("Current runtime packet is still active.\n\nUpdate "+goalDir+"evidence.md and "+goalDir+"next.md, or run `hyper resume` to continue it.", 2)
	}
	if planBody := readIfExists(filepath.Join(root, planFile)); strings.TrimSpace(planBody) != "" {
		plan := parsePlan(planBody)
		if err := validatePlanStageFields(plan); err != nil {
			return commandOutput{}, err
		}
		state = applyPlanTargetToState(state, plan)
	}
	readinessForGate := readReadinessStateIfExists(root)
	if readinessForGate.Version == 0 {
		readinessForGate = readinessStateForStatus(root, growthStateForStatus(root))
	}
	finishGate, finishErr := runFinishGate(root, state, derived, readinessForGate)
	if finishErr != nil {
		if finishGate.Status == "failed" {
			var failedDerived goalState
			if failed, ok := failedFinishGateGoalState(root, state.CurrentGoalID); ok {
				failedDerived = failed
			} else {
				failedDerived = goalState{
					State:  "active",
					Reason: "Finish gate failed. Fix review.md findings, then rerun the agent finish gate.",
				}
			}
			growthForPlan := growthStateForStatus(root)
			readinessForPlan := readinessStateForStatus(root, growthForPlan)
			if readinessForPlan.Version == 0 {
				readinessForPlan = readinessForGate
			}
			nextPlan, nextErr := writeNextPacketPlan(root, state, failedDerived, readinessForPlan, growthForPlan)
			if nextErr != nil {
				return commandOutput{}, nextErr
			}
			if !strings.Contains(finishErr.Message, "Next packet plan:") {
				finishErr.Message += strings.Join([]string{
					"",
					"",
					"Planned action: " + nextPlan.Action,
					"Continuation guard: " + compactText(nextPacketGuard(state, nextPlan), 220),
					nextPacketProgressGuardLine(state, nextPlan),
					"Next packet plan: " + displayRelPath(hyperDir, "next-packet.md"),
				}, "\n")
			}
		}
		return commandOutput{}, finishErr
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

	result, err := learnGoalFromState(root, state, db, "runtime_packet_completed", true)
	if err != nil {
		return commandOutput{}, err
	}
	growth, err := updateGrowthState(root, db)
	if err != nil {
		return commandOutput{}, err
	}
	readiness := readReadinessStateIfExists(root)
	if planBody := readIfExists(filepath.Join(root, planFile)); strings.TrimSpace(planBody) != "" {
		readiness, err = updateReadinessState(root, planBody, growth)
		if err != nil {
			return commandOutput{}, err
		}
		state = applyPlanTargetToState(state, parsePlan(planBody))
	}
	finishGate.Review = renderFinishGateReview(finishGate, state, derived, readiness)
	if err := writeText(filepath.Join(root, hyperDir, "goals", state.CurrentGoalID, "review.md"), finishGate.Review); err != nil {
		return commandOutput{}, err
	}
	now := nowISO()
	if err := updateRunAndGoalStatus(db, state.ActiveRunID, state.CurrentGoalID, derived.State, now); err != nil {
		return commandOutput{}, err
	}
	state.Status = derived.State
	state.UpdatedAt = now
	if err := writeJSON(statePath, state); err != nil {
		return commandOutput{}, err
	}
	event := map[string]any{
		"type":               "runtime_packet_closed",
		"run_id":             state.ActiveRunID,
		"goal_id":            state.CurrentGoalID,
		"state":              derived.State,
		"reason":             derived.Reason,
		"inserted_memories":  result.Inserted,
		"readiness_gate":     readiness.StageGate.Status,
		"readiness_pressure": readiness.NextPressure.Axis,
		"created_at":         nowISO(),
	}
	if err := insertEvent(db, event); err != nil {
		return commandOutput{}, err
	}
	if err := appendJSONL(filepath.Join(root, hyperDir, "logs", state.ActiveRunID+".jsonl"), event); err != nil {
		return commandOutput{}, err
	}

	nextPlan, err := writeNextPacketPlan(root, state, derived, readiness, growth)
	if err != nil {
		return commandOutput{}, err
	}
	line := "Memory files updated."
	if result.MemoryCount == 0 {
		line = "No learnable signal yet."
	}
	nextReason := nextPlan.Reason
	lines := []string{
		"Completed runtime packet: " + state.CurrentGoalID,
		"State: " + derived.State,
		"Reason: " + derived.Reason,
		"Finish gate: " + finishGate.Status,
		"Proof: " + proofStatusSummary(derived, readiness),
		fmt.Sprintf("Candidate memories: %d", result.MemoryCount),
		fmt.Sprintf("Inserted memories: %d", result.Inserted),
		"Memory quality: " + formatMemoryQuality(result),
		fmt.Sprintf("Growth pressures: %d", visibleGrowthPressureCount(growth.Pressures)),
		fmt.Sprintf("Capability candidates: %d", visibleGrowthCandidateCount(growth.Candidates)),
		"Pressure ledger: " + growthLoopStateSummary(growth),
		"Readiness gate: " + readinessGateSummary(readiness),
		"Readiness pressure: " + readinessPressureSummary(readiness),
		"Planned action: " + nextPlan.Action,
		"Next action: " + nextPacketActionDisplay(nextPlan),
		"Why: " + nextReason,
		"Continuation guard: " + compactText(nextPacketGuard(state, nextPlan), 220),
	}
	if progressLine := nextPacketProgressGuardLine(state, nextPlan); progressLine != "" {
		lines = append(lines, progressLine)
	}
	lines = append(lines,
		"Next packet plan: "+displayRelPath(hyperDir, "next-packet.md"),
		line,
		"",
	)
	lines = append(lines, nextPacketCommandBlock(nextPlan, "hyper status --short")...)
	lines = append(lines, "")
	return stdout(strings.Join(lines, "\n")), nil
}

func resumeHyper(fsys fsRoot) (commandOutput, *hyperError) {
	root := fsys.root()
	statePath := filepath.Join(root, hyperDir, "state.json")
	if !exists(statePath) {
		return stdout("No Hyper Run state found. Start with `hyper init`.\n"), nil
	}
	state, err := readState(statePath)
	if err != nil {
		return commandOutput{}, err
	}
	if planBody := readIfExists(filepath.Join(root, planFile)); strings.TrimSpace(planBody) != "" {
		plan := parsePlan(planBody)
		if err := validatePlanStageFields(plan); err != nil {
			return commandOutput{}, err
		}
		state = applyPlanTargetToState(state, plan)
	}
	if strings.TrimSpace(state.CurrentGoalID) == "" {
		return stdout("No active runtime packet found. Start with `hyper run`.\n"), nil
	}
	handoff := createExecutionHandoff(state.ActiveRunID, state.CurrentGoalID, state.AutoContinue)
	lines := []string{
		fmt.Sprintf("Resuming %s at %s.", state.ActiveRunID, state.CurrentGoalID),
		"",
		renderExecutionHandoff(handoff),
		"",
	}
	if failed, ok := failedFinishGateGoalState(root, state.CurrentGoalID); ok {
		lines = append(lines, renderFailedFinishGateResumeBlock(root, state.CurrentGoalID, failed)...)
	}
	return stdout(strings.Join(lines, "\n")), nil
}

func renderFailedFinishGateResumeBlock(root, goalID string, failed goalState) []string {
	lines := []string{
		"Finish gate failed. Fix this same runtime packet before starting new work.",
		"Reason: " + failed.Reason,
		"Review file: " + displayRelPath(hyperDir, "goals", goalID, "review.md"),
	}
	if findings := finishGateReviewFindings(root, goalID); len(findings) > 0 {
		lines = append(lines, "", "Current review findings:")
		for _, finding := range findings {
			lines = append(lines, "  - "+finding)
		}
		if note := finishGateReviewRepeatNote(root, goalID); note != "" {
			lines = append(lines, "  - "+note)
		}
	}
	lines = append(lines,
		"",
		"Next:",
		"  update "+displayRelPath(hyperDir, "goals", goalID, "evidence.md"),
		"  update "+displayRelPath(hyperDir, "goals", goalID, "next.md"),
		"  rerun the agent finish gate internally",
		"",
	)
	return lines
}

func blockingActiveGoal(root string, state projectState) string {
	if strings.TrimSpace(state.CurrentGoalID) == "" {
		return ""
	}
	if failed, ok := failedFinishGateGoalState(root, state.CurrentGoalID); ok {
		lines := []string{
			"Current runtime packet has failed the finish gate: " + state.CurrentGoalID,
			"Reason: " + failed.Reason,
		}
		if findings := finishGateReviewFindings(root, state.CurrentGoalID); len(findings) > 0 {
			lines = append(lines, "", "Current review findings:")
			for _, finding := range findings {
				lines = append(lines, "  - "+finding)
			}
			if note := finishGateReviewRepeatNote(root, state.CurrentGoalID); note != "" {
				lines = append(lines, "  - "+note)
			}
		}
		lines = append(lines,
			"",
			"Fix the same packet before creating another one:",
			"  update "+displayRelPath(hyperDir, "goals", state.CurrentGoalID, "evidence.md"),
			"  update "+displayRelPath(hyperDir, "goals", state.CurrentGoalID, "next.md"),
			"  rerun the agent finish gate internally",
		)
		return strings.Join(lines, "\n")
	}
	derived := deriveCurrentGoalState(root, state.CurrentGoalID)
	if strings.TrimSpace(state.Status) != "" && strings.TrimSpace(derived.State) != "" && state.Status != "active" && state.Status != derived.State {
		return strings.Join([]string{
			"Current runtime packet state is inconsistent: " + state.CurrentGoalID,
			"State file: " + state.Status,
			"Evidence state: " + derived.State,
			"",
			"Repair it before creating another packet:",
			"  hyper status --short",
			"  hyper repair",
			"  rerun the agent finish gate if the packet still needs closure",
		}, "\n")
	}
	if state.Status != "" && state.Status != "active" {
		return ""
	}
	path := state.CurrentGoalPath
	if strings.TrimSpace(path) == "" {
		path = fmt.Sprintf(".hyper/goals/%s/goal.md", state.CurrentGoalID)
	}
	if derived.State != "active" {
		return strings.Join([]string{
			"Current runtime packet has not passed the finish gate yet: " + state.CurrentGoalID,
			"Evidence state: " + derived.State,
			"Reason: " + derived.Reason,
			"",
			"Finish it before creating another packet:",
			"  rerun the agent finish gate",
			"  if the finish gate fails, fix " + strings.TrimSuffix(path, "goal.md") + "review.md",
			"  then rerun the agent finish gate again",
		}, "\n")
	}
	return strings.Join([]string{
		"Current runtime packet is still active: " + state.CurrentGoalID,
		"Reason: " + derived.Reason,
		"",
		"Finish it before creating another packet:",
		"  hyper resume",
		"  update " + strings.TrimSuffix(path, "goal.md") + "evidence.md",
		"  update " + strings.TrimSuffix(path, "goal.md") + "next.md",
		"  rerun the agent finish gate internally",
	}, "\n")
}

func runUntilStopState(previous projectState, opts runOptions, planBody string, readiness readinessState) projectState {
	plan := parsePlan(planBody)
	state := previous
	state.Project = firstNonBlank(state.Project, readinessProductName(plan), "Unknown project")
	state.Stage = normalizeRuntimeStage(firstNonBlank(readiness.Stage, state.Stage))
	state.Status = firstNonBlank(state.Status, "completed")
	state.PlanPath = planFile
	state.PlanHash = hashText(planBody)
	state.AutoContinue = true
	state.RunUntil = opts.RunUntil
	state.RunTargetSource = opts.RunTargetSource
	state.UpdatedAt = nowISO()
	return state
}

func runUntilStopDerived(state projectState) goalState {
	return goalState{
		State:  firstNonBlank(state.Status, "completed"),
		Reason: "Run-until target proof is complete.",
	}
}

func nextCommandBlock(commands ...string) []string {
	lines := []string{"Next:"}
	seen := map[string]bool{}
	for _, command := range commands {
		command = strings.TrimSpace(command)
		if command == "" || seen[command] {
			continue
		}
		seen[command] = true
		lines = append(lines, "  "+command)
	}
	return lines
}

func recordNoPacketRun(root string, db *sql.DB, state projectState, reason string, nextPlan plannedNextPacket, readiness readinessState, autoLearn learnResult) *hyperError {
	event := map[string]any{
		"type":                "run_skipped",
		"run_id":              nullableString(state.ActiveRunID),
		"goal_id":             nullableString(state.CurrentGoalID),
		"reason":              reason,
		"stage":               state.Stage,
		"auto_continue":       state.AutoContinue,
		"run_until":           nullableString(state.RunUntil),
		"run_target_source":   nullableString(state.RunTargetSource),
		"next_action":         nextPlan.Action,
		"next_command":        nextPlan.Command,
		"readiness_gate":      readinessGateSummary(readiness),
		"readiness_pressure":  readinessPressureSummary(readiness),
		"auto_learn_state":    autoLearn.State,
		"auto_learn_inserted": autoLearn.Inserted,
		"created_at":          nowISO(),
	}
	if err := insertEvent(db, event); err != nil {
		return err
	}
	if err := appendJSONL(filepath.Join(root, hyperDir, "logs", "project.jsonl"), event); err != nil {
		return err
	}
	return nil
}

func learnCurrentGoal(fsys fsRoot) (commandOutput, *hyperError) {
	root := fsys.root()
	statePath := filepath.Join(root, hyperDir, "state.json")
	if !exists(statePath) {
		return commandOutput{}, newError("No Hyper Run state found. Start with `hyper init`.", 2)
	}
	state, err := readState(statePath)
	if err != nil {
		return commandOutput{}, err
	}
	db, err := openDB(root)
	if err != nil {
		return commandOutput{}, err
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		return commandOutput{}, err
	}
	result, err := learnGoalFromState(root, state, db, "micro_learn_completed", true)
	if err != nil {
		return commandOutput{}, err
	}
	growth, err := updateGrowthState(root, db)
	if err != nil {
		return commandOutput{}, err
	}
	readiness := readReadinessStateIfExists(root)
	if planBody := readIfExists(filepath.Join(root, planFile)); strings.TrimSpace(planBody) != "" {
		readiness, err = updateReadinessState(root, planBody, growth)
		if err != nil {
			return commandOutput{}, err
		}
	}
	line := "Memory files updated."
	if result.MemoryCount == 0 {
		line = "No learnable signal yet."
	}
	return stdout(strings.Join([]string{
		"Learn scope: micro",
		"Learn role: extract repeated needs, failures, and proofs so future packets change boundaries, validation, readiness, and capability candidates",
		"Run: " + result.RunID,
		"Runtime packet: " + result.GoalID,
		"Runtime packet state: " + result.State,
		"Reason: " + result.Reason,
		fmt.Sprintf("Candidate memories: %d", result.MemoryCount),
		fmt.Sprintf("Inserted memories: %d", result.Inserted),
		"Memory quality: " + formatMemoryQuality(result),
		fmt.Sprintf("Growth pressures: %d", visibleGrowthPressureCount(growth.Pressures)),
		fmt.Sprintf("Capability candidates: %d", visibleGrowthCandidateCount(growth.Candidates)),
		"Pressure ledger: " + growthLoopStateSummary(growth),
		"Readiness gate: " + readinessGateSummary(readiness),
		"Readiness pressure: " + readinessPressureSummary(readiness),
		line,
		"",
	}, "\n")), nil
}
