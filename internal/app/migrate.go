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
	planBody := readIfExists(filepath.Join(root, planFile))
	if strings.TrimSpace(planBody) != "" {
		if err := validatePlanStageFields(parsePlan(planBody)); err != nil {
			return commandOutput{}, err
		}
	}
	db, err := openDB(root)
	if err != nil {
		return commandOutput{}, err
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		return commandOutput{}, err
	}
	if err := ensureProjectLayout(root); err != nil {
		return commandOutput{}, err
	}
	if err := ensureCodexDesktopRules(root); err != nil {
		return commandOutput{}, err
	}
	refreshedMemories, err := refreshLegacyMemoryQuality(db)
	if err != nil {
		return commandOutput{}, err
	}
	staledMemories, err := staleNoisyMemoryRecords(db)
	if err != nil {
		return commandOutput{}, err
	}
	if staledMemories > 0 {
		if err := rewriteMemoryMarkdownFiles(root, db); err != nil {
			return commandOutput{}, err
		}
	}
	before := readGrowthStateIfExists(root)
	growth, err := updateGrowthState(root, db)
	if err != nil {
		return commandOutput{}, err
	}
	readiness := readReadinessStateIfExists(root)
	if strings.TrimSpace(planBody) != "" {
		readiness, err = updateReadinessState(root, planBody, growth)
		if err != nil {
			return commandOutput{}, err
		}
	}
	stateMessage := "not checked"
	plannedAction := "not available"
	nextAction := "hyper status"
	nextPacketMessage := "not updated; no completed runtime packet state found"
	if state, stateErr := readState(filepath.Join(root, hyperDir, "state.json")); stateErr == nil {
		stageRefreshed := false
		targetRefreshed := false
		if strings.TrimSpace(readiness.Stage) != "" && knownRuntimeStage(readiness.Stage) && normalizeRuntimeStage(state.Stage) != readiness.Stage {
			state.Stage = readiness.Stage
			stageRefreshed = true
		}
		if strings.TrimSpace(planBody) != "" {
			beforeTarget := state
			state = applyPlanTargetToState(state, parsePlan(planBody))
			targetRefreshed = beforeTarget.AutoContinue != state.AutoContinue ||
				beforeTarget.RunUntil != state.RunUntil ||
				beforeTarget.RunTargetSource != state.RunTargetSource
		}
		if stageRefreshed || targetRefreshed {
			if strings.TrimSpace(planBody) != "" {
				state.PlanHash = hashText(planBody)
			}
			state.UpdatedAt = nowISO()
			if err := writeJSON(filepath.Join(root, hyperDir, "state.json"), state); err != nil {
				return commandOutput{}, err
			}
		}
		consistency := currentStateConsistency(root, state)
		if consistency.Consistent {
			stateMessage = "state.json is consistent"
			if stageRefreshed {
				stateMessage += "; stage refreshed to " + readiness.Stage
			}
			if targetRefreshed {
				stateMessage += "; run target refreshed to " + migrateTargetSummary(state)
			}
			if consistency.Derived.State == "active" && !isFailedFinishGateReason(consistency.Derived.Reason) {
				plannedAction = "complete-current"
				nextAction = statusNextCommandWithRefresh(state, consistency.Derived, readiness, statusRefresh{})
				nextPacketMessage = "unchanged while the current runtime packet is active"
			} else {
				nextPlan, nextErr := writeNextPacketPlan(root, state, consistency.Derived, readiness, growth)
				if nextErr != nil {
					return commandOutput{}, nextErr
				}
				plannedAction = nextPlan.Action
				nextAction = nextPacketActionDisplay(nextPlan)
				nextPacketMessage = displayRelPath(hyperDir, "next-packet.md") + " (" + nextPlan.Action + ")"
			}
		} else if consistency.Repairable {
			stateMessage = "state.json needs repair; run `hyper repair`"
			plannedAction = "repair"
			nextAction = "hyper repair"
			nextPacketMessage = "not updated; run `hyper repair` first"
		} else {
			stateMessage = "state.json mismatch is not repairable while packet is active"
			plannedAction = "blocked"
			nextAction = "hyper status --short"
			nextPacketMessage = "not updated while packet state is inconsistent"
		}
	}
	lines := []string{
		"Hyper Run Migration",
		"",
		fmt.Sprintf("Learn quality gate: refreshed %d legacy memory quality value(s)", refreshedMemories),
		fmt.Sprintf("Learn quality gate: staled %d noisy memory record(s)", staledMemories),
		"Growth state: refreshed",
		"Codex routing: refreshed",
		fmt.Sprintf("Visible pressures: %d -> %d", visibleGrowthPressureCount(before.Pressures), visibleGrowthPressureCount(growth.Pressures)),
		fmt.Sprintf("Visible candidates: %d -> %d", visibleGrowthCandidateCount(before.Candidates), visibleGrowthCandidateCount(growth.Candidates)),
		"Readiness gate: " + readinessGateSummary(readiness),
		"State consistency: " + stateMessage,
		"Planned action: " + plannedAction,
		"Next action: " + nextAction,
		"Next packet plan: " + nextPacketMessage,
		"",
	}
	lines = append(lines, nextCommandBlock(nextAction, "hyper doctor", "hyper status --short")...)
	lines = append(lines, "")
	return stdout(strings.Join(lines, "\n")), nil
}

func migrateTargetSummary(state projectState) string {
	if strings.TrimSpace(state.RunUntil) == "" {
		return "single packet"
	}
	return state.RunUntil
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

func staleNoisyMemoryRecords(db *sql.DB) (int, *hyperError) {
	rows, err := db.Query(`select id, kind, text, coalesce(confidence, 0), coalesce(quality, '') from memories where stale_at is null order by created_at asc, id asc`)
	if err != nil {
		return 0, dbError(err)
	}
	defer rows.Close()
	ids := []int64{}
	for rows.Next() {
		var record memoryRecord
		if err := rows.Scan(&record.ID, &record.Kind, &record.Text, &record.Confidence, &record.Quality); err != nil {
			return 0, dbError(err)
		}
		if noisyPersistedMemoryRecord(record) {
			ids = append(ids, record.ID)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, dbError(err)
	}
	for _, id := range ids {
		if _, err := db.Exec(`update memories set stale_at = ? where id = ?`, nowISO(), id); err != nil {
			return 0, dbError(err)
		}
	}
	return len(ids), nil
}

func noisyPersistedMemoryRecord(record memoryRecord) bool {
	signal := memorySignal(record.Text)
	normalized := normalizeSentence(signal)
	if normalized == "" {
		return true
	}
	return isNoIssueText(normalized) || isPassiveNoChangeText(normalized) || isHyperProtocolNoiseText(normalized)
}

func rewriteMemoryMarkdownFiles(root string, db *sql.DB) *hyperError {
	rows, err := db.Query(`select id, kind, text, coalesce(confidence, 0), coalesce(quality, '') from memories where stale_at is null order by created_at asc, id asc`)
	if err != nil {
		return dbError(err)
	}
	defer rows.Close()
	type markdownMemory struct {
		kind    string
		text    string
		quality string
	}
	memories := []markdownMemory{}
	for rows.Next() {
		var record memoryRecord
		if err := rows.Scan(&record.ID, &record.Kind, &record.Text, &record.Confidence, &record.Quality); err != nil {
			return dbError(err)
		}
		if noisyPersistedMemoryRecord(record) {
			continue
		}
		quality := firstNonBlank(record.Quality, memoryQuality(record.Kind, record.Text, firstNonZeroFloat(record.Confidence, 0.7)), "weak")
		memories = append(memories, markdownMemory{kind: record.Kind, text: record.Text, quality: quality})
	}
	if err := rows.Err(); err != nil {
		return dbError(err)
	}
	files := map[string]struct {
		title string
		lines []string
	}{
		"decision":   {title: "Decisions"},
		"pattern":    {title: "Patterns"},
		"failure":    {title: "Failures"},
		"constraint": {title: "Constraints"},
	}
	for _, mem := range memories {
		entry, ok := files[mem.kind]
		if !ok {
			continue
		}
		entry.lines = append(entry.lines, "- ["+mem.quality+"] "+mem.text)
		files[mem.kind] = entry
	}
	for kind, entry := range files {
		rel := ""
		switch kind {
		case "decision":
			rel = ".hyper/memories/decisions.md"
		case "pattern":
			rel = ".hyper/memories/patterns.md"
		case "failure":
			rel = ".hyper/memories/failures.md"
		case "constraint":
			rel = ".hyper/memories/constraints.md"
		}
		body := "# " + entry.title + "\n\n"
		if len(entry.lines) > 0 {
			body += strings.Join(entry.lines, "\n") + "\n"
		}
		if err := writeText(filepath.Join(root, rel), body); err != nil {
			return err
		}
	}
	return nil
}

func growthMigrationNeeded(growth growthState) bool {
	if strings.TrimSpace(growth.ActivationPolicy.Method) == "" {
		return true
	}
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
