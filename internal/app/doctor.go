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
	lines = append(lines, "", doctorSummary(checks), "")
	return stdout(strings.Join(lines, "\n")), nil
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

func doctorGrowthMigrationCheck(root string) doctorCheck {
	growth := readGrowthStateIfExists(root)
	if growth.Version == 0 {
		return doctorCheck{"Growth migration", "OK", "no growth state yet"}
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
	current := readinessStateForStatus(root, readGrowthStateIfExists(root))
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
