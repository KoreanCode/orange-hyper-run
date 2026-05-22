package app

import (
	"fmt"
	"path/filepath"
	"strings"
)

type stateConsistency struct {
	HasState      bool
	HasGoal       bool
	ProjectStatus string
	Derived       goalState
	Consistent    bool
	Repairable    bool
	Reason        string
}

func currentStateConsistency(root string, state projectState) stateConsistency {
	projectStatus := strings.TrimSpace(state.Status)
	goalID := strings.TrimSpace(state.CurrentGoalID)
	if goalID == "" {
		return stateConsistency{
			HasState:      true,
			ProjectStatus: projectStatus,
			Derived:       goalState{State: "initialized", Reason: "No current runtime packet recorded."},
			Consistent:    projectStatus == "" || projectStatus == "initialized",
			Repairable:    false,
			Reason:        "No current runtime packet recorded.",
		}
	}
	derived := deriveCurrentGoalState(root, goalID)
	consistent := projectStatus == "" || projectStatus == derived.State
	repairable := !consistent && derived.State != "active"
	reason := "state.json matches the current runtime packet."
	if !consistent {
		reason = fmt.Sprintf("state.json says %s, but %s is %s.", firstNonBlank(projectStatus, "unknown"), goalID, derived.State)
	}
	return stateConsistency{
		HasState:      true,
		HasGoal:       true,
		ProjectStatus: projectStatus,
		Derived:       derived,
		Consistent:    consistent,
		Repairable:    repairable,
		Reason:        reason,
	}
}

func repairHyper(fsys fsRoot) (commandOutput, *hyperError) {
	root := fsys.root()
	statePath := filepath.Join(root, hyperDir, "state.json")
	if !exists(statePath) {
		return commandOutput{}, newError("No Hyper Run state found. Start with `hyper init`.", 2)
	}
	state, err := readState(statePath)
	if err != nil {
		return commandOutput{}, err
	}
	consistency := currentStateConsistency(root, state)
	if !consistency.Repairable {
		lines := []string{
			"Hyper Run Repair",
			"",
			"State: no repair needed",
			"Reason: " + consistency.Reason,
			"",
		}
		if !consistency.Consistent && consistency.Derived.State == "active" {
			lines[2] = "State: repair blocked"
			lines = append(lines[:4], append([]string{
				"Current runtime packet is still active. Update evidence.md and next.md before repair.",
			}, lines[4:]...)...)
		}
		return stdout(strings.Join(lines, "\n")), nil
	}

	db, err := openDB(root)
	if err != nil {
		return commandOutput{}, err
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		return commandOutput{}, err
	}
	now := nowISO()
	oldStatus := state.Status
	state.Status = consistency.Derived.State
	state.UpdatedAt = now
	if err := updateRunAndGoalStatus(db, state.ActiveRunID, state.CurrentGoalID, consistency.Derived.State, now); err != nil {
		return commandOutput{}, err
	}
	if err := writeJSON(statePath, state); err != nil {
		return commandOutput{}, err
	}
	event := map[string]any{
		"type":        "state_repaired",
		"run_id":      state.ActiveRunID,
		"goal_id":     state.CurrentGoalID,
		"from_status": oldStatus,
		"to_status":   state.Status,
		"reason":      consistency.Reason,
		"created_at":  nowISO(),
	}
	if err := insertEvent(db, event); err != nil {
		return commandOutput{}, err
	}
	if strings.TrimSpace(state.ActiveRunID) != "" {
		if err := appendJSONL(filepath.Join(root, hyperDir, "logs", state.ActiveRunID+".jsonl"), event); err != nil {
			return commandOutput{}, err
		}
	}
	return stdout(strings.Join([]string{
		"Hyper Run Repair",
		"",
		"State: repaired",
		"Runtime packet: " + state.CurrentGoalID,
		"From: " + firstNonBlank(oldStatus, "unknown"),
		"To: " + state.Status,
		"Reason: " + consistency.Derived.Reason,
		"",
		"Next:",
		"  hyper status",
		"",
	}, "\n")), nil
}
