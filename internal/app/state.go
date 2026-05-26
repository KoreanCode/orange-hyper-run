package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func readStateIfExists(root string) projectState {
	state, _ := readState(filepath.Join(root, hyperDir, "state.json"))
	return state
}

func readState(path string) (projectState, *hyperError) {
	var state projectState
	body, err := os.ReadFile(path)
	if err != nil {
		return state, ioError(err)
	}
	if err := json.Unmarshal(body, &state); err != nil {
		return state, newError(err.Error(), 1)
	}
	return state, nil
}

func initEventType(created, hasActiveGoal bool) string {
	if created {
		return "project_initialized"
	}
	if hasActiveGoal {
		return "project_init_checked"
	}
	return "project_reinitialized"
}

func initSummary(plan planResult, hasActiveGoal bool) string {
	if hasActiveGoal {
		return "Loaded existing Hyper Run state and preserved the active runtime packet."
	}
	if plan.Created {
		return "Created blank plan.md. Fill it in before creating the first runtime packet."
	}
	return "Loaded existing plan.md and refreshed Hyper Run project state."
}

func formatAutoLearn(result learnResult) string {
	if result.Skipped {
		return fmt.Sprintf("skipped (%s)", result.Reason)
	}
	return fmt.Sprintf("%s, inserted %d", result.State, result.Inserted)
}

func formatMemoryQuality(result learnResult) string {
	if len(result.Quality) == 0 {
		if rejected := formatRejectedMemoryQuality(result); rejected != "none" {
			return "rejected " + rejected
		}
		return "none"
	}
	order := []string{"durable", "weak", "passive", "one_off"}
	parts := []string{}
	for _, key := range order {
		if result.Quality[key] > 0 {
			parts = append(parts, fmt.Sprintf("%s %d", key, result.Quality[key]))
		}
	}
	if len(parts) == 0 {
		if rejected := formatRejectedMemoryQuality(result); rejected != "none" {
			return "rejected " + rejected
		}
		return "none"
	}
	if rejected := formatRejectedMemoryQuality(result); rejected != "none" {
		parts = append(parts, "rejected "+rejected)
	}
	return strings.Join(parts, ", ")
}

func formatRejectedMemoryQuality(result learnResult) string {
	if len(result.Rejected) == 0 {
		return "none"
	}
	order := []string{"noisy", "passive", "one_off", "invalid"}
	parts := []string{}
	for _, key := range order {
		if result.Rejected[key] > 0 {
			parts = append(parts, fmt.Sprintf("%s %d", key, result.Rejected[key]))
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}
