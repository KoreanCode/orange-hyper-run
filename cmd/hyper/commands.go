package main

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
	stage := firstNonBlank(episode.Plan["Current Stage"], "Tiny MVP")
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
		"Status: " + state.Status,
		"Readiness gate: " + readinessGateSummary(readiness),
		"Readiness pressure: " + readinessPressureSummary(readiness),
		"Plan file: " + planFile,
		"Hyper dir: " + hyperDir + "/",
	}
	if hasActiveGoal {
		lines = append(lines, "Active runtime packet preserved: "+state.CurrentGoalID)
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
			"growth_pressures":   len(growth.Pressures),
			"growth_candidates":  len(growth.Candidates),
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

	return stdout(strings.Join([]string{
		"Project: " + state.Project,
		"Stage: " + episode.Stage,
		"Run: " + runID,
		"Runtime packet: " + goalID,
		"Auto learn: " + formatAutoLearn(autoLearn),
		"Readiness gate: " + readinessGateSummary(readiness),
		"Readiness pressure: " + readinessPressureSummary(readiness),
		fmt.Sprintf("Similar context: %d", len(similar)),
		"Runtime packet file: " + state.CurrentGoalPath,
		"",
		"Loaded plan.md as the product brief.",
		"",
		renderExecutionHandoff(handoff),
		"",
	}, "\n")), nil
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
	db, err := openDB(root)
	if err != nil {
		return commandOutput{}, err
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		return commandOutput{}, err
	}
	derived := deriveCurrentGoalState(root, state.CurrentGoalID)
	runs, err := countRows(db, "runs")
	if err != nil {
		return commandOutput{}, err
	}
	goals, err := countRows(db, "goals")
	if err != nil {
		return commandOutput{}, err
	}
	growth := readGrowthStateIfExists(root)
	readiness := readinessStateForStatus(root, growth)
	lines := []string{
		"Project: " + state.Project,
		"Stage: " + state.Stage,
		"Status: " + state.Status,
		"Runtime packet state: " + derived.State,
		"Runtime packet reason: " + derived.Reason,
	}
	lines = append(lines, readinessStatusLines(readiness)...)
	lines = append(lines,
		"Active run: "+state.ActiveRunID,
		"Current runtime packet: "+state.CurrentGoalID,
		"Runtime packet file: "+state.CurrentGoalPath,
		fmt.Sprintf("Runs recorded: %d", runs),
		fmt.Sprintf("Runtime packets recorded: %d", goals),
		fmt.Sprintf("Growth pressures: %d", len(growth.Pressures)),
		fmt.Sprintf("Capability candidates: %d", len(growth.Candidates)),
		"Updated: "+state.UpdatedAt,
		"",
	)
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
		"Learn role: durable decisions, reusable patterns, blockers/failures, and constraints",
		"Run: " + result.RunID,
		"Runtime packet: " + result.GoalID,
		"Runtime packet state: " + result.State,
		"Reason: " + result.Reason,
		fmt.Sprintf("Candidate memories: %d", result.MemoryCount),
		fmt.Sprintf("Inserted memories: %d", result.Inserted),
		fmt.Sprintf("Growth pressures: %d", len(growth.Pressures)),
		fmt.Sprintf("Capability candidates: %d", len(growth.Candidates)),
		"Readiness gate: " + readinessGateSummary(readiness),
		"Readiness pressure: " + readinessPressureSummary(readiness),
		line,
		"",
	}, "\n")), nil
}
