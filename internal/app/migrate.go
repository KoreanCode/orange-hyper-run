package app

import (
	"database/sql"
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
	refreshedMemories, err := refreshLegacyMemoryQuality(db)
	if err != nil {
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
	nextPacketMessage := "not updated; no completed runtime packet state found"
	if state, stateErr := readState(filepath.Join(root, hyperDir, "state.json")); stateErr == nil {
		consistency := currentStateConsistency(root, state)
		if consistency.Consistent {
			stateMessage = "state.json is consistent"
			if consistency.Derived.State == "active" {
				nextPacketMessage = "unchanged while the current runtime packet is active"
			} else {
				nextPlan, nextErr := writeNextPacketPlan(root, state, consistency.Derived, readiness, growth)
				if nextErr != nil {
					return commandOutput{}, nextErr
				}
				nextPacketMessage = filepath.Join(hyperDir, "next-packet.md") + " (" + nextPlan.Action + ")"
			}
		} else if consistency.Repairable {
			stateMessage = "state.json needs repair; run `hyper repair`"
			nextPacketMessage = "not updated; run `hyper repair` first"
		} else {
			stateMessage = "state.json mismatch is not repairable while packet is active"
			nextPacketMessage = "not updated while packet state is inconsistent"
		}
	}
	return stdout(strings.Join([]string{
		"Hyper Run Migration",
		"",
		fmt.Sprintf("Learn quality gate: refreshed %d legacy memory quality value(s)", refreshedMemories),
		"Growth state: refreshed",
		fmt.Sprintf("Visible pressures: %d -> %d", visibleGrowthPressureCount(before.Pressures), visibleGrowthPressureCount(growth.Pressures)),
		fmt.Sprintf("Visible candidates: %d -> %d", visibleGrowthCandidateCount(before.Candidates), visibleGrowthCandidateCount(growth.Candidates)),
		"Readiness gate: " + readinessGateSummary(readiness),
		"State consistency: " + stateMessage,
		"Next packet plan: " + nextPacketMessage,
		"",
		"Next:",
		"  hyper doctor",
		"  hyper status",
		"",
	}, "\n")), nil
}

func refreshLegacyMemoryQuality(db *sql.DB) (int, *hyperError) {
	rows, err := db.Query(`select id, kind, text, coalesce(confidence, 0) from memories where stale_at is null and (quality is null or trim(quality) = '')`)
	if err != nil {
		return 0, dbError(err)
	}
	defer rows.Close()
	type update struct {
		id      int64
		quality string
	}
	updates := []update{}
	for rows.Next() {
		var id int64
		var kind, text string
		var confidence float64
		if err := rows.Scan(&id, &kind, &text, &confidence); err != nil {
			return 0, dbError(err)
		}
		updates = append(updates, update{id: id, quality: memoryQuality(kind, text, firstNonZeroFloat(confidence, 0.7))})
	}
	if err := rows.Err(); err != nil {
		return 0, dbError(err)
	}
	for _, item := range updates {
		if _, err := db.Exec(`update memories set quality = ? where id = ?`, item.quality, item.id); err != nil {
			return 0, dbError(err)
		}
	}
	return len(updates), nil
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
