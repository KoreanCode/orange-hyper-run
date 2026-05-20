package main

import (
	"database/sql"
	"path/filepath"
	"strings"
)

func learnGoalFromState(root string, state projectState, db *sql.DB, eventType string, recordWhenEmpty bool) (learnResult, *hyperError) {
	goalID := state.CurrentGoalID
	runID := state.ActiveRunID
	if strings.TrimSpace(goalID) == "" || strings.TrimSpace(runID) == "" {
		return learnResult{Skipped: true, Reason: "No active runtime packet recorded.", State: "none", RunID: runID, GoalID: goalID}, nil
	}
	goalDir := filepath.Join(root, hyperDir, "goals", goalID)
	evidenceText := readIfExists(filepath.Join(goalDir, "evidence.md"))
	nextText := readIfExists(filepath.Join(goalDir, "next.md"))
	derived := deriveGoalState(evidenceText, nextText)
	memories := memoriesForDerivedState(derived, goalID, evidenceText, nextText)
	if len(memories) == 0 && !recordWhenEmpty {
		return learnResult{Skipped: true, Reason: derived.Reason, State: derived.State, RunID: runID, GoalID: goalID}, nil
	}
	inserted := 0
	for _, mem := range memories {
		ok, err := insertMemoryIfNew(db, mem)
		if err != nil {
			return learnResult{}, err
		}
		if ok {
			inserted++
			if err := appendMemoryMarkdown(root, mem); err != nil {
				return learnResult{}, err
			}
		}
	}
	event := map[string]any{"type": eventType, "run_id": runID, "goal_id": goalID, "state": derived.State, "inserted_memories": inserted, "created_at": nowISO()}
	if err := insertEvent(db, event); err != nil {
		return learnResult{}, err
	}
	if err := appendJSONL(filepath.Join(root, hyperDir, "logs", runID+".jsonl"), event); err != nil {
		return learnResult{}, err
	}
	return learnResult{Skipped: false, Reason: derived.Reason, State: derived.State, RunID: runID, GoalID: goalID, Inserted: inserted, MemoryCount: len(memories)}, nil
}

func buildSimilarityQuery(plan map[string]string, ep episode, focus string) string {
	return joinNonEmpty([]string{focus, ep.Objective, ep.Stage, ep.BuildStyle, ep.Scope, plan["Product"], plan["MVP"], plan["Current Focus"], plan["Target Users"]}, "\n")
}
