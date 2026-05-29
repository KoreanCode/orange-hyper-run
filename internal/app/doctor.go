package app

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/KoreanCode/orange-hyper-run/internal/buildinfo"
)

type doctorCheck struct {
	Name   string
	Status string
	Detail string
}

func doctorHyper(fsys fsRoot) (commandOutput, *hyperError) {
	root := fsys.root()
	checks := []doctorCheck{}
	executable, execErr := os.Executable()
	if execErr != nil || strings.TrimSpace(executable) == "" {
		checks = append(checks, doctorCheck{"Executable", "WARN", "could not resolve current executable"})
	} else {
		checks = append(checks, doctorCheck{"Executable", "OK", executable})
		if path, err := exec.LookPath("hyper"); err == nil {
			status := "OK"
			detail := path
			if filepath.Clean(path) != filepath.Clean(executable) {
				status = "WARN"
				detail = fmt.Sprintf("PATH resolves %s, current executable is %s", path, executable)
			}
			checks = append(checks, doctorCheck{"PATH", status, detail})
		} else {
			checks = append(checks, doctorCheck{"PATH", "WARN", "`hyper` is not found on PATH"})
		}
	}
	checks = append(checks, doctorCheck{"Version", "OK", buildinfo.Version + " (" + runtime.GOOS + "/" + runtime.GOARCH + ")"})
	checks = append(checks, doctorCheck{"Update URL", "OK", resolveUpdateURL("")})

	planPath := filepath.Join(root, planFile)
	if !exists(planPath) {
		checks = append(checks, doctorCheck{"plan.md", "FAIL", "missing; run `hyper init` first"})
	} else {
		planBody := readIfExists(planPath)
		plan := parsePlan(planBody)
		missing := missingPlanFields(plan)
		if len(missing) > 0 {
			checks = append(checks, doctorCheck{"plan.md", "WARN", "missing or sparse fields: " + strings.Join(missing, ", ")})
		} else {
			checks = append(checks, doctorCheck{"plan.md", "OK", "product brief is present"})
		}
		checks = append(checks, doctorPlanTargetCheck(plan))
	}

	hyperPath := filepath.Join(root, hyperDir)
	if !exists(hyperPath) {
		checks = append(checks, doctorCheck{".hyper", "FAIL", "missing; run `hyper init`"})
	} else {
		checks = append(checks, doctorCheck{".hyper", "OK", hyperPath})
	}

	checks = append(checks, doctorStateChecks(root)...)
	checks = append(checks, doctorGrowthMigrationCheck(root))
	checks = append(checks, doctorReadinessStateCheck(root))
	checks = append(checks, doctorNextPacketPlanCheck(root))
	checks = append(checks, doctorSignatureCheck())
	checks = append(checks, doctorDBCheck(root))
	checks = append(checks, doctorCodexChecks(root)...)

	lines := []string{
		"Hyper Run Doctor",
		"",
		"Workspace: " + root,
	}
	for _, check := range checks {
		lines = append(lines, fmt.Sprintf("[%s] %s: %s", check.Status, check.Name, check.Detail))
	}
	lines = append(lines, "", doctorSummary(checks))
	if actions := doctorActionLines(checks); len(actions) > 0 {
		lines = append(lines, "", "Next:")
		for _, action := range actions {
			lines = append(lines, "  "+action)
		}
	}
	lines = append(lines, "")
	return stdout(strings.Join(lines, "\n")), nil
}

func doctorPlanTargetCheck(plan map[string]string) doctorCheck {
	value := firstRuntimeValue(plan["Target Stage"])
	if value == "" {
		return doctorCheck{"Target Stage", "OK", "not set; `hyper run` uses single-packet mode unless --until is provided"}
	}
	target, err := normalizeRunUntilTarget(value)
	if err != nil {
		return doctorCheck{"Target Stage", "FAIL", "invalid `" + value + "`; use tiny-mvp, usable-mvp, beta, service-quality, or sustained-service-quality"}
	}
	return doctorCheck{"Target Stage", "OK", target + " from plan.md"}
}

func doctorSignatureCheck() doctorCheck {
	if _, err := exec.LookPath("cosign"); err == nil {
		return doctorCheck{"Signature verification", "OK", "cosign available for release signature checks"}
	}
	if signatureVerificationRequired() {
		return doctorCheck{"Signature verification", "FAIL", "HYPER_RUN_VERIFY_SIGNATURE requires cosign, but cosign is not on PATH"}
	}
	return doctorCheck{"Signature verification", "OK", "optional cosign not installed; checksum verification remains active"}
}

func missingPlanFields(plan map[string]string) []string {
	required := []string{"Product", "MVP", "Current Stage", "Build Style", "Success Criteria"}
	missing := []string{}
	for _, field := range required {
		if firstRuntimeValue(plan[field]) == "" {
			missing = append(missing, field)
		}
	}
	return missing
}

func doctorStateChecks(root string) []doctorCheck {
	path := filepath.Join(root, hyperDir, "state.json")
	if !exists(path) {
		return []doctorCheck{{"state.json", "WARN", "missing; run `hyper init` or `hyper run`"}}
	}
	state, err := readState(path)
	if err != nil {
		return []doctorCheck{{"state.json", "FAIL", err.Message}}
	}
	consistency := currentStateConsistency(root, state)
	stateStatus := "OK"
	stateDetail := "status=" + firstNonBlank(state.Status, "unknown")
	if !consistency.Consistent {
		stateStatus = "WARN"
		stateDetail = consistency.Reason + " Run `hyper repair`."
	}
	checks := []doctorCheck{{"state.json", stateStatus, stateDetail}}
	if stageCheck := doctorStageSourceCheck(root, state); stageCheck.Name != "" {
		checks = append(checks, stageCheck)
	}
	if strings.TrimSpace(state.CurrentGoalID) == "" {
		checks = append(checks, doctorCheck{"Current packet", "OK", "none active"})
		return checks
	}
	status := "OK"
	if consistency.Derived.State == "active" {
		status = "WARN"
	}
	checks = append(checks, doctorCheck{"Current packet", status, state.CurrentGoalID + " is " + consistency.Derived.State + "; " + consistency.Derived.Reason})
	return checks
}

func doctorStageSourceCheck(root string, state projectState) doctorCheck {
	refresh := stageSourceRefresh(root, state)
	if !refresh.Needed {
		return doctorCheck{}
	}
	return doctorCheck{"Stage source", "WARN", refresh.Reason}
}

func doctorGrowthMigrationCheck(root string) doctorCheck {
	growth := readGrowthStateIfExists(root)
	if growth.Version == 0 {
		return doctorCheck{"Growth migration", "OK", "no growth state yet"}
	}
	if growthHasUnstoredManualActiveCapability(root, growth) {
		return doctorCheck{"Growth migration", "WARN", "active capability files are not reflected in stored growth state; run `hyper migrate`"}
	}
	if growthMigrationNeeded(growth) {
		return doctorCheck{"Growth migration", "WARN", "legacy or noisy growth entries found; run `hyper migrate`"}
	}
	return doctorCheck{"Growth migration", "OK", "growth state uses current rules"}
}

func doctorReadinessStateCheck(root string) doctorCheck {
	stored := readReadinessStateIfExists(root)
	if stored.Version == 0 {
		return doctorCheck{"Readiness state", "OK", "no readiness state yet"}
	}
	if !exists(filepath.Join(root, planFile)) {
		return doctorCheck{"Readiness state", "WARN", "plan.md missing; cannot refresh readiness"}
	}
	current := readinessStateForStatus(root, growthStateForStatus(root))
	if current.Version == 0 {
		return doctorCheck{"Readiness state", "OK", "readiness state is present"}
	}
	if !sameReadinessForDoctor(stored, current) {
		return doctorCheck{
			Name:   "Readiness state",
			Status: "WARN",
			Detail: "stored " + readinessDoctorSummary(stored) + "; current evidence resolves " + readinessDoctorSummary(current) + ". Run `hyper migrate`.",
		}
	}
	return doctorCheck{"Readiness state", "OK", "readiness state is current"}
}

func doctorNextPacketPlanCheck(root string) doctorCheck {
	path := filepath.Join(root, hyperDir, "next-packet.md")
	statePath := filepath.Join(root, hyperDir, "state.json")
	if !exists(statePath) {
		return doctorCheck{"Next packet plan", "OK", "no runtime state yet"}
	}
	state, err := readState(statePath)
	if err != nil {
		return doctorCheck{"Next packet plan", "WARN", "cannot inspect state.json: " + err.Message}
	}
	state = applyPlanTargetFromRoot(root, state)
	consistency := currentStateConsistency(root, state)
	if !consistency.Consistent {
		return doctorCheck{"Next packet plan", "WARN", "cannot verify until state.json is repaired"}
	}
	if consistency.Derived.State == "active" {
		return doctorCheck{"Next packet plan", "OK", "not required while the current runtime packet is active"}
	}
	if refresh := statusRefreshFor(root, state); statusRefreshActionable(state, consistency.Derived, refresh) {
		return doctorCheck{"Next packet plan", "WARN", "cannot trust next-packet until refresh completes: " + refresh.Reason}
	}
	if !exists(path) {
		return doctorCheck{"Next packet plan", "WARN", "missing; run `hyper migrate` or complete the current packet again"}
	}
	growth := growthStateForStatus(root)
	readiness := readinessStateForStatus(root, growth)
	expected := buildNextPacketPlan(state, consistency.Derived, readiness, growth)
	body := readIfExists(path)
	actualAction := nextPacketPlanAction(body)
	if actualAction == "" {
		return doctorCheck{"Next packet plan", "WARN", "missing Action; run `hyper migrate`"}
	}
	if actualAction != expected.Action {
		return doctorCheck{"Next packet plan", "WARN", "expected action `" + expected.Action + "`, found `" + actualAction + "`; run `hyper migrate`"}
	}
	actual := nextPacketPlanCommand(body)
	if actual == "" {
		return doctorCheck{"Next packet plan", "WARN", "missing Command; run `hyper migrate`"}
	}
	if actual != expected.Command {
		return doctorCheck{"Next packet plan", "WARN", "expected `" + expected.Command + "`, found `" + actual + "`; run `hyper migrate`"}
	}
	if issue := nextPacketPlanShapeIssue(expected, body); issue != "" {
		return doctorCheck{"Next packet plan", "WARN", issue + "; run `hyper migrate`"}
	}
	return doctorCheck{"Next packet plan", "OK", displayRelPath(hyperDir, "next-packet.md") + " matches current state"}
}

func nextPacketPlanCommand(body string) string {
	return nextPacketPlanField(body, "Command:")
}

func nextPacketPlanAction(body string) string {
	return nextPacketPlanField(body, "Action:")
}

func nextPacketPlanField(body, name string) string {
	for _, line := range strings.Split(body, "\n") {
		if _, value, ok := strings.Cut(strings.TrimSpace(line), name); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func nextPacketPlanShapeIssue(expected plannedNextPacket, body string) string {
	if !strings.Contains(body, "## Guard") {
		return "missing Guard section"
	}
	if !strings.Contains(body, "## Codex Desktop Continuation") {
		return "missing Codex Desktop Continuation"
	}
	if expected.Action == "advance" && !strings.Contains(body, "## Stage Advancement Review") {
		return "missing Stage Advancement Review"
	}
	return ""
}

func sameReadinessForDoctor(a, b readinessState) bool {
	if a.Stage != b.Stage || a.StageGate.Status != b.StageGate.Status || a.NextPressure.Axis != b.NextPressure.Axis {
		return false
	}
	aDims := readinessDimensionMap(a.Dimensions)
	bDims := readinessDimensionMap(b.Dimensions)
	for _, id := range doctorRelevantReadinessAxes(a, b) {
		aDim := aDims[id]
		bDim := bDims[id]
		if aDim.ID == "" && bDim.ID == "" {
			continue
		}
		if bDim.Status != aDim.Status {
			return false
		}
	}
	return true
}

func doctorRelevantReadinessAxes(states ...readinessState) []string {
	seen := map[string]bool{}
	axes := []string{}
	add := func(axis string) {
		axis = strings.TrimSpace(axis)
		if axis == "" || seen[axis] {
			return
		}
		seen[axis] = true
		axes = append(axes, axis)
	}
	for _, state := range states {
		for _, axis := range state.StageGate.RequiredAxes {
			add(axis)
		}
		add(state.NextPressure.Axis)
	}
	return axes
}

func readinessDoctorSummary(readiness readinessState) string {
	return readinessGateSummary(readiness) + " / pressure " + firstNonBlank(readiness.NextPressure.Axis, "none")
}

func doctorDBCheck(root string) doctorCheck {
	path := filepath.Join(root, hyperDir, "hyper.sqlite")
	if !exists(path) {
		return doctorCheck{"SQLite", "WARN", "missing .hyper/hyper.sqlite"}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return doctorCheck{"SQLite", "FAIL", err.Error()}
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return doctorCheck{"SQLite", "FAIL", err.Error()}
	}
	return doctorCheck{"SQLite", "OK", path}
}

func doctorCodexChecks(root string) []doctorCheck {
	files := []string{
		"AGENTS.md",
		filepath.Join(".agents", "skills", "hyper", "SKILL.md"),
		filepath.Join(".agents", "skills", "hyper-run", "SKILL.md"),
		filepath.Join(hyperDir, "codex-desktop.md"),
		filepath.Join(hyperDir, "commands", "hyper-run.md"),
	}
	missing := []string{}
	for _, file := range files {
		if !exists(filepath.Join(root, file)) {
			missing = append(missing, file)
		}
	}
	if len(missing) > 0 {
		return []doctorCheck{{"Codex Desktop routing", "WARN", "missing: " + strings.Join(missing, ", ")}}
	}
	return []doctorCheck{{"Codex Desktop routing", "OK", "$hyper run files are installed"}}
}

func doctorSummary(checks []doctorCheck) string {
	warn := 0
	fail := 0
	for _, check := range checks {
		switch check.Status {
		case "WARN":
			warn++
		case "FAIL":
			fail++
		}
	}
	if fail > 0 {
		return fmt.Sprintf("Summary: %d failure(s), %d warning(s). Fix failures before relying on Hyper Run.", fail, warn)
	}
	if warn > 0 {
		return fmt.Sprintf("Summary: 0 failures, %d warning(s). Hyper Run is usable, but the warnings should be cleaned up.", warn)
	}
	return "Summary: all checks passed."
}

func doctorActionLines(checks []doctorCheck) []string {
	actions := []string{}
	seen := map[string]bool{}
	for _, check := range checks {
		if check.Status != "WARN" && check.Status != "FAIL" {
			continue
		}
		action := doctorActionForCheck(check)
		if action == "" || seen[action] {
			continue
		}
		seen[action] = true
		actions = append(actions, action)
	}
	return actions
}

func doctorActionForCheck(check doctorCheck) string {
	name := strings.ToLower(strings.TrimSpace(check.Name))
	detail := strings.ToLower(strings.TrimSpace(check.Detail))
	switch name {
	case "path":
		if strings.Contains(detail, "not found") {
			return "Add Hyper Run's install directory to PATH, then run `hyper version`."
		}
		return "Run `which hyper`; remove or reorder the older binary so the shell uses the expected Hyper Run executable."
	case "plan.md":
		if strings.Contains(detail, "missing") {
			return "Run `hyper init`, fill in plan.md, then run `hyper doctor` again."
		}
		return "Fill the missing plan.md fields, then run `hyper status --short`."
	case ".hyper":
		return "Run `hyper init` to recreate project-local Hyper Run files."
	case "state.json":
		if strings.Contains(detail, "run `hyper repair`") {
			return "Run `hyper repair`, then run `hyper doctor` again."
		}
		return "Run `hyper init` or `hyper run [focus]` to create runtime state."
	case "stage source":
		return "Run `hyper migrate`, then run `hyper doctor` again."
	case "current packet":
		return "Finish the current packet: update evidence.md and next.md, then run `hyper complete`."
	case "growth migration", "readiness state":
		return "Run `hyper migrate`, then run `hyper doctor` again."
	case "next packet plan":
		if strings.Contains(detail, "repair") {
			return "Run `hyper repair`, then run `hyper doctor` again."
		}
		return "Run `hyper migrate`, then run `hyper doctor` again."
	case "signature verification":
		return "Install `cosign` or unset `HYPER_RUN_VERIFY_SIGNATURE`."
	case "sqlite":
		if strings.Contains(detail, "missing") {
			return "Run `hyper migrate`; if SQLite is still missing, run `hyper init`."
		}
		return "Fix the SQLite error, then run `hyper doctor` again."
	case "codex desktop routing":
		return "Run `hyper init` to reinstall Codex Desktop routing files."
	default:
		return "Fix " + check.Name + ", then run `hyper doctor` again."
	}
}
