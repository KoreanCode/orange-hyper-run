package app

import (
	"fmt"
	"path/filepath"
	"strings"
)

func initHyper(fsys fsRoot) (commandOutput, *hyperError) {
	root := fsys.root()
	if err := ensureProjectLayout(root); err != nil {
		return commandOutput{}, err
	}
	if err := ensureCodexDesktopRules(root); err != nil {
		return commandOutput{}, err
	}
	planResult, err := ensurePlanForInit(root)
	if err != nil {
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
	episode := compileGoalEpisode("GOAL-0000", "", planResult.Body, nil, growth, readiness)
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
	next := []string{"  Fill in plan.md", "  hyper run [focus]"}
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

func runHyper(fsys fsRoot, focus string) (commandOutput, *hyperError) {
	root := fsys.root()
	planResult, err := requirePlanForRun(root)
	if err != nil {
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

	runID, err := nextID(db, "runs", "RUN")
	if err != nil {
		return commandOutput{}, err
	}
	goalID, err := nextID(db, "goals", "GOAL")
	if err != nil {
		return commandOutput{}, err
	}
	now := nowISO()
	episode := compileGoalEpisode(goalID, focus, planResult.Body, nil, growth, readiness)
	similar, err := findSimilarContext(db, buildSimilarityQuery(episode.Plan, episode, focus), 5)
	if err != nil {
		return commandOutput{}, err
	}
	episode = compileGoalEpisode(goalID, focus, planResult.Body, similar, growth, readiness)
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
	handoff := createExecutionHandoff(runID, goalID)
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
		UpdatedAt:        now,
	}

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
		"Auto learn: " + formatAutoLearn(autoLearn),
		"Readiness gate: " + readinessGateSummary(readiness),
		"Readiness pressure: " + readinessPressureSummary(readiness),
		"Pressure ledger: " + growthLoopStateSummary(growth),
		fmt.Sprintf("Similar context: %d", len(similar)),
		"Runtime packet file: " + state.CurrentGoalPath,
	}
	if planCandidatePath != "" {
		lines = append(lines, "Plan import candidates: "+planCandidatePath)
	}
	lines = append(lines,
		"",
		"Loaded plan.md as the product brief.",
		"",
		renderExecutionHandoff(handoff),
		"",
	)
	return stdout(strings.Join(lines, "\n")), nil
}

func statusHyper(fsys fsRoot) (commandOutput, *hyperError) {
	root := fsys.root()
	statePath := filepath.Join(root, hyperDir, "state.json")
	if !exists(statePath) {
		return stdout("No Hyper Run state found. Start with `hyper init`.\n"), nil
	}
	state, err := readState(statePath)
	if err != nil {
		return commandOutput{}, err
	}
	derived := deriveCurrentGoalState(root, state.CurrentGoalID)
	runs, goals := statusDBCounts(root)
	growth := readGrowthStateIfExists(root)
	readiness := readinessStateForStatus(root, growth)
	lines := statusDashboardLines(state, derived, readiness, growth, runs, goals)
	return stdout(strings.Join(lines, "\n")), nil
}

func statusDBCounts(root string) (int, int) {
	if !exists(filepath.Join(root, hyperDir, "hyper.sqlite")) {
		return 0, 0
	}
	db, err := openDB(root)
	if err != nil {
		return 0, 0
	}
	defer db.Close()
	runs, runErr := countRows(db, "runs")
	if runErr != nil {
		runs = 0
	}
	goals, goalErr := countRows(db, "goals")
	if goalErr != nil {
		goals = 0
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

	line := "Memory files updated."
	if result.MemoryCount == 0 {
		line = "No learnable signal yet."
	}
	return stdout(strings.Join([]string{
		"Completed runtime packet: " + state.CurrentGoalID,
		"State: " + derived.State,
		"Reason: " + derived.Reason,
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
		"Next:",
		"  hyper status",
		"  hyper run [next focus]",
		"",
	}, "\n")), nil
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
	if strings.TrimSpace(state.CurrentGoalID) == "" {
		return stdout("No active runtime packet found. Start with `hyper run`.\n"), nil
	}
	handoff := createExecutionHandoff(state.ActiveRunID, state.CurrentGoalID)
	return stdout(strings.Join([]string{
		fmt.Sprintf("Resuming %s at %s.", state.ActiveRunID, state.CurrentGoalID),
		"",
		renderExecutionHandoff(handoff),
		"",
	}, "\n")), nil
}

func blockingActiveGoal(root string, state projectState) string {
	if strings.TrimSpace(state.CurrentGoalID) == "" {
		return ""
	}
	if state.Status != "" && state.Status != "active" {
		return ""
	}
	derived := deriveCurrentGoalState(root, state.CurrentGoalID)
	if derived.State != "active" {
		return ""
	}
	path := state.CurrentGoalPath
	if strings.TrimSpace(path) == "" {
		path = fmt.Sprintf(".hyper/goals/%s/goal.md", state.CurrentGoalID)
	}
	return strings.Join([]string{
		"Current runtime packet is still active: " + state.CurrentGoalID,
		"Reason: " + derived.Reason,
		"",
		"Finish it before creating another packet:",
		"  hyper resume",
		"  update " + strings.TrimSuffix(path, "goal.md") + "evidence.md",
		"  update " + strings.TrimSuffix(path, "goal.md") + "next.md",
		"  hyper complete",
	}, "\n")
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
