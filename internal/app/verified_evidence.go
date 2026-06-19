package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const verifiedEvidenceEventType = "verified_command"

type verifiedEvidenceRecord struct {
	ID                    string   `json:"id"`
	Type                  string   `json:"type"`
	Status                string   `json:"status"`
	Axis                  string   `json:"axis,omitempty"`
	Name                  string   `json:"name,omitempty"`
	Command               []string `json:"command"`
	CommandLine           string   `json:"command_line"`
	CWD                   string   `json:"cwd"`
	RunID                 string   `json:"run_id,omitempty"`
	GoalID                string   `json:"goal_id,omitempty"`
	StartedAt             string   `json:"started_at"`
	FinishedAt            string   `json:"finished_at"`
	DurationMillis        int64    `json:"duration_millis"`
	ExitCode              int      `json:"exit_code"`
	CommitSHA             string   `json:"commit_sha"`
	WorktreeStatusSHA256  string   `json:"worktree_status_sha256"`
	StdoutSHA256          string   `json:"stdout_sha256"`
	StderrSHA256          string   `json:"stderr_sha256"`
	StdoutBytes           int      `json:"stdout_bytes"`
	StderrBytes           int      `json:"stderr_bytes"`
	StdoutPath            string   `json:"stdout_path,omitempty"`
	StderrPath            string   `json:"stderr_path,omitempty"`
	RecordPath            string   `json:"record_path"`
	RecordedBy            string   `json:"recorded_by"`
	ReadinessEvidenceText string   `json:"readiness_evidence_text"`
}

type verifyOptions struct {
	Axis    string
	Name    string
	Command []string
}

type verifiedEvidenceGoalSummary struct {
	GoalID       string
	Total        int
	Passed       int
	Failed       int
	Newest       verifiedEvidenceRecord
	LatestFailed verifiedEvidenceRecord
}

func verifyHyper(fsys fsRoot, args []string) (commandOutput, *hyperError) {
	root := fsys.root()
	if err := ensureProjectLayout(root); err != nil {
		return commandOutput{}, err
	}
	opts, err := parseVerifyOptions(args)
	if err != nil {
		return commandOutput{}, err
	}
	state := readStateIfExists(root)
	record, stdoutText, stderrText, runErr := runVerifiedCommand(root, state, opts)
	if recordErr := persistVerifiedEvidence(root, state, record, stdoutText, stderrText); recordErr != nil {
		return commandOutput{}, recordErr
	}
	out := renderVerifiedEvidenceOutput(record)
	if runErr != nil {
		return stdout(out), newError(fmt.Sprintf("Verified command failed with exit code %d. Record: %s", record.ExitCode, record.RecordPath), recordExitCode(record.ExitCode))
	}
	return stdout(out), nil
}

func parseVerifyOptions(args []string) (verifyOptions, *hyperError) {
	opts := verifyOptions{Axis: "validation_coverage"}
	commandIndex := -1
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		if arg == "--" {
			commandIndex = i + 1
			break
		}
		switch arg {
		case "--axis":
			if i+1 >= len(args) {
				return opts, newError("hyper verify requires a value after --axis.", 2)
			}
			opts.Axis = strings.TrimSpace(args[i+1])
			i++
		case "--name":
			if i+1 >= len(args) {
				return opts, newError("hyper verify requires a value after --name.", 2)
			}
			opts.Name = strings.TrimSpace(args[i+1])
			i++
		default:
			return opts, newError("hyper verify options must appear before `--`.\n\n"+commandUsage("verify"), 2)
		}
	}
	if commandIndex == -1 || commandIndex >= len(args) {
		return opts, newError("hyper verify requires `-- <command> [args...]`.", 2)
	}
	opts.Command = append([]string{}, args[commandIndex:]...)
	if strings.TrimSpace(opts.Command[0]) == "" {
		return opts, newError("hyper verify requires a non-empty command after `--`.", 2)
	}
	opts.Axis = normalizeVerifyAxis(opts.Axis)
	if opts.Axis == "" {
		return opts, newError("hyper verify --axis must match a readiness axis such as validation_coverage, core_ux, sustained_quality, operations_docs, or maintainability.", 2)
	}
	if strings.TrimSpace(opts.Name) == "" {
		opts.Name = strings.Join(opts.Command, " ")
	}
	return opts, nil
}

func normalizeVerifyAxis(axis string) string {
	axis = strings.TrimSpace(axis)
	if axis == "" {
		return "validation_coverage"
	}
	if match := readinessAxisForLabel(axis, readinessDimensionDefs()); match != "" {
		return match
	}
	compact := compactReadinessLabel(strings.ReplaceAll(axis, "_", " "))
	for _, def := range readinessDimensionDefs() {
		if compact == compactReadinessLabel(def.ID) || compact == compactReadinessLabel(def.Name) {
			return def.ID
		}
	}
	return ""
}

func runVerifiedCommand(root string, state projectState, opts verifyOptions) (verifiedEvidenceRecord, string, string, error) {
	start := time.Now()
	startedAt := start.UTC().Format("2006-01-02T15:04:05.000Z")
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd := exec.Command(opts.Command[0], opts.Command[1:]...)
	cmd.Dir = root
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	runErr := cmd.Run()
	finished := time.Now()
	exitCode := 0
	status := "passed"
	if runErr != nil {
		status = "failed"
		exitCode = commandExitCode(runErr)
	}
	stdoutText := stdoutBuf.String()
	stderrText := stderrBuf.String()
	recordID := nextVerifiedEvidenceID(root)
	recordRel := displayRelPath(hyperDir, "verified-evidence", recordID+".json")
	stdoutRel := displayRelPath(hyperDir, "verified-evidence", recordID+".stdout.txt")
	stderrRel := displayRelPath(hyperDir, "verified-evidence", recordID+".stderr.txt")
	commandLine := strings.Join(opts.Command, " ")
	record := verifiedEvidenceRecord{
		ID:                    recordID,
		Type:                  verifiedEvidenceEventType,
		Status:                status,
		Axis:                  opts.Axis,
		Name:                  opts.Name,
		Command:               append([]string{}, opts.Command...),
		CommandLine:           commandLine,
		CWD:                   root,
		RunID:                 state.ActiveRunID,
		GoalID:                state.CurrentGoalID,
		StartedAt:             startedAt,
		FinishedAt:            finished.UTC().Format("2006-01-02T15:04:05.000Z"),
		DurationMillis:        finished.Sub(start).Milliseconds(),
		ExitCode:              exitCode,
		CommitSHA:             gitCommitSHA(root),
		WorktreeStatusSHA256:  hashText(gitStatusShort(root)),
		StdoutSHA256:          hashText(stdoutText),
		StderrSHA256:          hashText(stderrText),
		StdoutBytes:           len([]byte(stdoutText)),
		StderrBytes:           len([]byte(stderrText)),
		StdoutPath:            stdoutRel,
		StderrPath:            stderrRel,
		RecordPath:            recordRel,
		RecordedBy:            "hyper verify",
		ReadinessEvidenceText: verifiedReadinessEvidenceText(opts.Axis, commandLine, status, exitCode, recordID),
	}
	return record, stdoutText, stderrText, runErr
}

func persistVerifiedEvidence(root string, state projectState, record verifiedEvidenceRecord, stdoutText, stderrText string) *hyperError {
	dir := filepath.Join(root, hyperDir, "verified-evidence")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ioError(err)
	}
	// Re-run is not acceptable for evidence, so write the buffers captured during
	// command execution through the paths embedded in the record.
	if err := writeText(filepath.Join(root, filepath.FromSlash(record.StdoutPath)), stdoutText); err != nil {
		return err
	}
	if err := writeText(filepath.Join(root, filepath.FromSlash(record.StderrPath)), stderrText); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(root, filepath.FromSlash(record.RecordPath)), record); err != nil {
		return err
	}
	event := verifiedEvidenceEvent(record)
	if err := appendJSONL(filepath.Join(root, hyperDir, "logs", "verified-evidence.jsonl"), event); err != nil {
		return err
	}
	if strings.TrimSpace(state.ActiveRunID) != "" {
		if err := appendJSONL(filepath.Join(root, hyperDir, "logs", state.ActiveRunID+".jsonl"), event); err != nil {
			return err
		}
	}
	db, err := openDB(root)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		return err
	}
	return insertEvent(db, event)
}

func verifiedEvidenceEvent(record verifiedEvidenceRecord) map[string]any {
	return map[string]any{
		"type":                    verifiedEvidenceEventType,
		"id":                      record.ID,
		"status":                  record.Status,
		"axis":                    record.Axis,
		"name":                    record.Name,
		"command":                 record.Command,
		"command_line":            record.CommandLine,
		"run_id":                  record.RunID,
		"goal_id":                 record.GoalID,
		"created_at":              record.FinishedAt,
		"started_at":              record.StartedAt,
		"finished_at":             record.FinishedAt,
		"duration_millis":         record.DurationMillis,
		"exit_code":               record.ExitCode,
		"commit_sha":              record.CommitSHA,
		"worktree_status_sha256":  record.WorktreeStatusSHA256,
		"stdout_sha256":           record.StdoutSHA256,
		"stderr_sha256":           record.StderrSHA256,
		"stdout_bytes":            record.StdoutBytes,
		"stderr_bytes":            record.StderrBytes,
		"record_path":             record.RecordPath,
		"stdout_path":             record.StdoutPath,
		"stderr_path":             record.StderrPath,
		"readiness_evidence_text": record.ReadinessEvidenceText,
	}
}

func renderVerifiedEvidenceOutput(record verifiedEvidenceRecord) string {
	lines := []string{
		"Verified evidence: " + record.ID,
		"Status: " + record.Status,
		fmt.Sprintf("Exit code: %d", record.ExitCode),
		"Command: " + record.CommandLine,
		"Axis: " + record.Axis,
		"Goal: " + firstNonBlank(record.GoalID, "none"),
		"Run: " + firstNonBlank(record.RunID, "none"),
		"Record: " + record.RecordPath,
		"Stdout: " + record.StdoutPath,
		"Stderr: " + record.StderrPath,
		"Stdout SHA256: " + record.StdoutSHA256,
		"Stderr SHA256: " + record.StderrSHA256,
		"Commit SHA: " + record.CommitSHA,
		"Worktree status SHA256: " + record.WorktreeStatusSHA256,
	}
	return strings.Join(lines, "\n")
}

func nextVerifiedEvidenceID(root string) string {
	records, _ := filepath.Glob(filepath.Join(root, hyperDir, "verified-evidence", "VE-*.json"))
	maxID := 0
	for _, path := range records {
		base := strings.TrimSuffix(filepath.Base(path), ".json")
		number := strings.TrimPrefix(base, "VE-")
		value, err := strconv.Atoi(number)
		if err == nil && value > maxID {
			maxID = value
		}
	}
	return fmt.Sprintf("VE-%04d", maxID+1)
}

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if ok := errors.As(err, &exitErr); ok {
		return exitErr.ExitCode()
	}
	return 1
}

func recordExitCode(exitCode int) int {
	if exitCode <= 0 {
		return 1
	}
	if exitCode > 125 {
		return 1
	}
	return exitCode
}

func gitCommitSHA(root string) string {
	out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func gitStatusShort(root string) string {
	out, err := exec.Command("git", "-C", root, "status", "--short").Output()
	if err != nil {
		return "unknown"
	}
	return string(out)
}

func loadVerifiedEvidenceRecords(root string) []verifiedEvidenceRecord {
	paths, err := filepath.Glob(filepath.Join(root, hyperDir, "verified-evidence", "VE-*.json"))
	if err != nil {
		return nil
	}
	sort.Strings(paths)
	records := []verifiedEvidenceRecord{}
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var record verifiedEvidenceRecord
		if err := json.Unmarshal(body, &record); err != nil {
			continue
		}
		if record.ID == "" {
			record.ID = strings.TrimSuffix(filepath.Base(path), ".json")
		}
		if record.RecordPath == "" {
			record.RecordPath = displayRelPath(hyperDir, "verified-evidence", record.ID+".json")
		}
		records = append(records, record)
	}
	return records
}

func verifiedReadinessEvidenceRecords(root, goalID string, defs []readinessDimensionDef) []readinessEvidenceRecord {
	records := []readinessEvidenceRecord{}
	for _, record := range loadVerifiedEvidenceRecords(root) {
		if !verifiedEvidenceGoalMatches(record, goalID) || record.Status != "passed" || record.ExitCode != 0 {
			continue
		}
		axis := normalizeVerifyAxis(record.Axis)
		if axis == "" {
			axis = "validation_coverage"
		}
		if !readinessAxisExists(axis, defs) {
			continue
		}
		text := firstNonBlank(record.ReadinessEvidenceText, verifiedReadinessEvidenceText(axis, record.CommandLine, record.Status, record.ExitCode, record.ID))
		records = append(records, readinessEvidenceRecordForAxis(record.GoalID, axis, text))
		if axis != "validation_coverage" {
			validationText := verifiedReadinessEvidenceText("validation_coverage", record.CommandLine, record.Status, record.ExitCode, record.ID)
			records = append(records, readinessEvidenceRecordForAxis(record.GoalID, "validation_coverage", validationText))
		}
	}
	return records
}

func readinessAxisExists(axis string, defs []readinessDimensionDef) bool {
	for _, def := range defs {
		if def.ID == axis {
			return true
		}
	}
	return false
}

func verifiedEvidenceGoalMatches(record verifiedEvidenceRecord, goalID string) bool {
	goalID = strings.TrimSpace(goalID)
	return goalID == "" || strings.TrimSpace(record.GoalID) == goalID
}

func goalHasPassedVerifiedEvidence(root, goalID string) bool {
	for _, record := range loadVerifiedEvidenceRecords(root) {
		if verifiedEvidenceGoalMatches(record, goalID) && record.Status == "passed" && record.ExitCode == 0 {
			return true
		}
	}
	return false
}

func activeValidatorVerifiedEvidenceCovers(root, goalID string, capability activeCapability) bool {
	if capability.Kind != "validator" {
		return false
	}
	expectedCommand := normalizeSentence(inferredCommandForSignal(capability.Signal))
	if expectedCommand == "" {
		return false
	}
	for _, record := range loadVerifiedEvidenceRecords(root) {
		if !verifiedEvidenceGoalMatches(record, goalID) || record.Status != "passed" || record.ExitCode != 0 {
			continue
		}
		if strings.Contains(normalizeSentence(record.CommandLine), expectedCommand) {
			return true
		}
	}
	return false
}

func verifiedEvidenceSummaryForGoal(root, goalID string) verifiedEvidenceGoalSummary {
	summary := verifiedEvidenceGoalSummary{GoalID: strings.TrimSpace(goalID)}
	for _, record := range loadVerifiedEvidenceRecords(root) {
		if !verifiedEvidenceGoalMatches(record, goalID) {
			continue
		}
		summary.Total++
		switch record.Status {
		case "passed":
			if record.ExitCode == 0 {
				summary.Passed++
			} else {
				summary.Failed++
				summary.LatestFailed = record
			}
		case "failed":
			summary.Failed++
			summary.LatestFailed = record
		}
		summary.Newest = record
	}
	return summary
}

func verifiedEvidenceShortLine(root, goalID string) string {
	summary := verifiedEvidenceSummaryForGoal(root, goalID)
	goal := firstNonBlank(summary.GoalID, "current packet")
	if summary.Total == 0 {
		return "Verified Evidence: " + goal + " has no records yet"
	}
	line := fmt.Sprintf("Verified Evidence: %s %d record(s); passed %d, failed %d; newest %s",
		goal,
		summary.Total,
		summary.Passed,
		summary.Failed,
		verifiedEvidenceRecordStatusPhrase(summary.Newest),
	)
	if summary.Failed > 0 && summary.LatestFailed.ID != summary.Newest.ID {
		line += "; latest failed " + verifiedEvidenceRecordStatusPhrase(summary.LatestFailed)
	}
	return line
}

func verifiedEvidenceDashboardLines(root, goalID string) []string {
	summary := verifiedEvidenceSummaryForGoal(root, goalID)
	goal := firstNonBlank(summary.GoalID, "current packet")
	lines := []string{"Verified Evidence:", "  Current packet: " + goal}
	if summary.Total == 0 {
		return append(lines, "  Records: none yet")
	}
	lines = append(lines,
		fmt.Sprintf("  Records: %d total, %d passed, %d failed", summary.Total, summary.Passed, summary.Failed),
		"  Newest: "+verifiedEvidenceRecordStatusPhrase(summary.Newest),
		"  Record: "+summary.Newest.RecordPath,
	)
	if summary.Failed > 0 {
		lines = append(lines, "  Latest failure: "+verifiedEvidenceRecordStatusPhrase(summary.LatestFailed))
	}
	return lines
}

func doctorVerifiedEvidenceCheck(root string) doctorCheck {
	state := readStateIfExists(root)
	goalID := strings.TrimSpace(state.CurrentGoalID)
	if goalID == "" {
		return doctorCheck{"Verified Evidence", "OK", "no current packet"}
	}
	summary := verifiedEvidenceSummaryForGoal(root, goalID)
	if summary.Total == 0 {
		return doctorCheck{"Verified Evidence", "OK", "no records for " + goalID + " yet"}
	}
	detail := fmt.Sprintf("%s records=%d passed=%d failed=%d; newest %s",
		goalID,
		summary.Total,
		summary.Passed,
		summary.Failed,
		verifiedEvidenceRecordStatusPhrase(summary.Newest),
	)
	status := "OK"
	if summary.Failed > 0 {
		status = "WARN"
		if summary.LatestFailed.ID != summary.Newest.ID {
			detail += "; latest failed " + verifiedEvidenceRecordStatusPhrase(summary.LatestFailed)
		}
	}
	return doctorCheck{"Verified Evidence", status, detail}
}

func verifiedEvidenceRecordStatusPhrase(record verifiedEvidenceRecord) string {
	if strings.TrimSpace(record.ID) == "" {
		return "none"
	}
	command := compactText(firstNonBlank(record.CommandLine, strings.Join(record.Command, " ")), 90)
	displayStatus := firstNonBlank(record.Status, "unknown")
	if record.ExitCode != 0 {
		displayStatus = "failed"
	}
	phrase := record.ID + " " + displayStatus
	if command != "" {
		phrase += " `" + command + "`"
	}
	if displayStatus == "failed" {
		phrase += fmt.Sprintf(" exit %d", record.ExitCode)
	}
	return phrase
}

func verifiedReadinessEvidenceText(axis, commandLine, status string, exitCode int, recordID string) string {
	return fmt.Sprintf("Verified Evidence %s executed CLI command `%s` with status %s and exit code %d.", recordID, commandLine, status, exitCode)
}
