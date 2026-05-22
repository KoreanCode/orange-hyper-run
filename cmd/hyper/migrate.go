package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func migrateHyper(fsys fsRoot) (commandOutput, *hyperError) {
	root := fsys.root()
	if !exists(filepath.Join(root, hyperDir)) {
		return commandOutput{}, newError("No Hyper Run project found. Start with `hyper init`.", 2)
	}
	db, err := openDB(root)
	if err != nil {
		return commandOutput{}, err
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		return commandOutput{}, err
	}
	before := readGrowthStateIfExists(root)
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
	stateMessage := "not checked"
	if state, stateErr := readState(filepath.Join(root, hyperDir, "state.json")); stateErr == nil {
		consistency := currentStateConsistency(root, state)
		if consistency.Consistent {
			stateMessage = "state.json is consistent"
		} else if consistency.Repairable {
			stateMessage = "state.json needs repair; run `hyper repair`"
		} else {
			stateMessage = "state.json mismatch is not repairable while packet is active"
		}
	}
	return stdout(strings.Join([]string{
		"Hyper Run Migration",
		"",
		"Growth state: refreshed",
		fmt.Sprintf("Visible pressures: %d -> %d", visibleGrowthPressureCount(before.Pressures), visibleGrowthPressureCount(growth.Pressures)),
		fmt.Sprintf("Visible candidates: %d -> %d", visibleGrowthCandidateCount(before.Candidates), visibleGrowthCandidateCount(growth.Candidates)),
		"Readiness gate: " + readinessGateSummary(readiness),
		"State consistency: " + stateMessage,
		"",
		"Next:",
		"  hyper doctor",
		"  hyper status",
		"",
	}, "\n")), nil
}

func growthMigrationNeeded(growth growthState) bool {
	for _, pressure := range growth.Pressures {
		if !visibleGrowthPressure(pressure) {
			return true
		}
	}
	for _, candidate := range growth.Candidates {
		if !visibleGrowthCandidate(candidate) && candidate.Status != "retired" {
			return true
		}
		if visibleGrowthCandidate(candidate) && displayGrowthCandidateName(candidate) != strings.TrimSpace(candidate.Name) {
			return true
		}
	}
	return false
}
