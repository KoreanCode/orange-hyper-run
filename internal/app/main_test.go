package app

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesProjectStateAndRules(t *testing.T) {
	root := t.TempDir()
	out, err := runCLI(args("init"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	assertContains(t, out.Stdout, "Project: Unknown project")
	assertContains(t, out.Stdout, "Status: initialized")
	assertContains(t, out.Stdout, "$hyper run")
	assertContains(t, out.Stdout, "Fill in plan.md")
	assertContains(t, readFile(t, filepath.Join(root, "plan.md")), "# Product Plan")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "state.json")), "initialized")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "logs", "project.jsonl")), "project_initialized")
	assertContains(t, readFile(t, filepath.Join(root, "AGENTS.md")), "$hyper run")
	assertContains(t, readFile(t, filepath.Join(root, "AGENTS.md")), "$hyper status --short")
	assertContains(t, readFile(t, filepath.Join(root, "AGENTS.md")), "$hyper migrate")
	assertContains(t, readFile(t, filepath.Join(root, ".agents", "skills", "hyper", "SKILL.md")), "name: hyper")
	assertContains(t, readFile(t, filepath.Join(root, ".agents", "skills", "hyper", "SKILL.md")), "compatibility shim")
	assertContains(t, readFile(t, filepath.Join(root, ".agents", "skills", "hyper", "SKILL.md")), "$hyper status --short")
	assertContains(t, readFile(t, filepath.Join(root, ".agents", "skills", "hyper", "SKILL.md")), "$hyper migrate")
	assertContains(t, readFile(t, filepath.Join(root, ".agents", "skills", "hyper-run", "SKILL.md")), "name: hyper-run")
	assertContains(t, readFile(t, filepath.Join(root, ".agents", "skills", "hyper-run", "SKILL.md")), "hyper run")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "codex-desktop.md")), "$hyper run")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "commands", "hyper-run.md")), "Required flow")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), `"version": 1`)
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), `"pressure_ledger"`)
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), `"No structure before pressure."`)
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "readiness", "state.json")), `"version": 1`)
}

func TestOpenDBConfiguresBusyTimeout(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, hyperDir), 0755); err != nil {
		t.Fatal(err)
	}
	db, herr := openDB(root)
	if herr != nil {
		t.Fatalf("openDB failed: %v", herr)
	}
	defer db.Close()

	var timeout int
	if err := db.QueryRow("pragma busy_timeout").Scan(&timeout); err != nil {
		t.Fatalf("busy timeout pragma failed: %v", err)
	}
	if timeout < 5000 {
		t.Fatalf("expected busy timeout >= 5000ms, got %d", timeout)
	}

	var journalMode string
	if err := db.QueryRow("pragma journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("journal mode pragma failed: %v", err)
	}
	if strings.ToLower(journalMode) != "wal" {
		t.Fatalf("expected wal journal mode, got %s", journalMode)
	}
}

func TestVersionShowsBuildAndExecutable(t *testing.T) {
	out, err := runCLI(args("version"), testRoot(t.TempDir()), fakeUpdater{})
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	assertContains(t, out.Stdout, "Version:")
	assertContains(t, out.Stdout, "Commit:")
	assertContains(t, out.Stdout, "Executable:")
	assertContains(t, out.Stdout, "Update source: github:KoreanCode/orange-hyper-run")
}

func TestSubcommandHelpDoesNotError(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{args("run", "--help"), "Usage:\n  hyper run [--auto] [--until stage] [focus]"},
		{args("status", "--help"), "Usage:\n  hyper status\n  hyper status --short"},
		{args("update", "--help"), "Usage:\n  hyper update [source]"},
	} {
		out, err := runCLI(tc.args, testRoot(t.TempDir()), fakeUpdater{})
		if err != nil {
			t.Fatalf("%v failed: %v", tc.args, err)
		}
		assertContains(t, out.Stdout, tc.want)
	}
}

func TestInitRejectsObjectiveArgument(t *testing.T) {
	root := t.TempDir()
	_, err := runCLI(args("init", "Build a tiny CRM MVP"), testRoot(root), fakeUpdater{})
	if err == nil {
		t.Fatal("expected error")
	}
	assertContains(t, err.Message, "does not take an objective")
}

func TestInitAppendsHyperRunRulesToExistingAgentsFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "AGENTS.md"), "# Existing Instructions\n\nKeep existing rules.\n")

	if _, err := runCLI(args("init"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	agents := readFile(t, filepath.Join(root, "AGENTS.md"))
	assertContains(t, agents, "Keep existing rules.")
	assertContains(t, agents, "<!-- hyper-run:start -->")
	assertContains(t, agents, "$hyper run")

	if _, err := runCLI(args("init"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("second init failed: %v", err)
	}
	agents = readFile(t, filepath.Join(root, "AGENTS.md"))
	if strings.Count(agents, "<!-- hyper-run:start -->") != 1 {
		t.Fatalf("expected one Hyper Run section, got:\n%s", agents)
	}
}

func TestInitRepairsLegacySkillFiles(t *testing.T) {
	root := t.TempDir()
	mustRun(t, root, "init")
	legacyPath := filepath.Join(root, ".agents", "skills", "hyper-run", "SKILL.md")
	writeFile(t, legacyPath, "name hyper-run\ndescription legacy invalid skill\n")
	writeFile(t, filepath.Join(root, "AGENTS.md"), "# Project Instructions\n\n<!-- hyper-run:start -->\n## Hyper Run\n\nWhen the user writes `$hyper run`, use old rules.\n<!-- hyper-run:end -->\n\nKeep this custom note.\n")

	if _, err := runCLI(args("init"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("second init failed: %v", err)
	}

	assertContains(t, readFile(t, legacyPath), "---\nname: hyper-run")
	assertContains(t, readFile(t, filepath.Join(root, ".agents", "skills", "hyper", "SKILL.md")), "---\nname: hyper")
	agents := readFile(t, filepath.Join(root, "AGENTS.md"))
	assertContains(t, agents, "$hyper-run")
	assertContains(t, agents, "thin Codex Desktop router")
	assertContains(t, agents, "Keep this custom note.")
}

func TestRunCreatesGoalAfterInit(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny CRM", "Build a tiny CRM MVP")
	out, err := runCLI(args("run"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	assertContains(t, out.Stdout, "GOAL-0001")
	assertContains(t, out.Stdout, "Auto learn: skipped")
	assertContains(t, out.Stdout, "Codex Desktop payload:")
	goal := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "goal.md"))
	assertContains(t, goal, "# GOAL-0001 Runtime Packet")
	assertContains(t, goal, "## Continue From")
	assertContains(t, goal, "## Current Episode")
	assertContains(t, goal, "Build a tiny CRM MVP")
	assertContains(t, goal, "Stage contract: Existence proof")
	assertContains(t, goal, "Growth loop: Execution -> Evidence -> Pressure Ledger -> Candidate -> Structure when proven.")
	assertContains(t, goal, "No structure before pressure.")
	assertContains(t, goal, "## Stage Gate")
	assertContains(t, goal, "Gate requirement:")
	assertNotContains(t, goal, "Gate evidence:")
	assertContains(t, goal, "## Stage Runtime Behavior")
	assertContains(t, goal, "## Active Capabilities")
	assertContains(t, goal, "## Proof Contract")
	assertContains(t, goal, "## Execution Contract")
	assertContains(t, goal, "## Done Checklist")
	assertContains(t, goal, "Functional Proof")
	assertContains(t, goal, "Surface Proof")
	assertContains(t, goal, "Operational Proof")
	assertContains(t, goal, "Next readiness pressure")
	assertContains(t, goal, "Capture readiness evidence")
	assertNotContains(t, goal, "## Scope")
	assertNotContains(t, goal, "## Non-goals")
	evidence := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"))
	assertContains(t, evidence, "## Readiness Evidence")
	assertContains(t, evidence, "Core UX: Pending.")
	assertContains(t, evidence, "## Surface Proof Evidence")
	assertContains(t, evidence, "- Target surface: Pending.")
	assertContains(t, evidence, "## Active Capability Evidence")
	assertContains(t, evidence, "## Decisions")
	assertContains(t, evidence, "## Reusable Patterns")
	assertContains(t, evidence, "## Learn Quality Gate")
	next := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"))
	assertContains(t, next, "## Learn Notes")
	assertContains(t, next, "Write only durable signals")
	assertContains(t, next, "- Decision: Pending.")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "logs", "RUN-0001.jsonl")), "goal_created")
}

func TestRunRequiresInit(t *testing.T) {
	root := t.TempDir()
	_, err := runCLI(args("run", "Build a tiny CRM MVP"), testRoot(root), fakeUpdater{})
	if err == nil {
		t.Fatal("expected error")
	}
	assertContains(t, err.Message, "hyper init")
	if err.Code != 2 {
		t.Fatalf("expected code 2, got %d", err.Code)
	}
}

func TestStatusAndResumeUseActiveState(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a tiny notes MVP")
	mustRun(t, root, "run")

	status, err := runCLI(args("status"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	assertContains(t, status.Stdout, "Active run: RUN-0001")
	assertContains(t, status.Stdout, "Current runtime packet: GOAL-0001")
	assertContains(t, status.Stdout, "Runtime packet state: active")
	assertContains(t, status.Stdout, "Stage contract:")
	assertContains(t, status.Stdout, "Method: Evidence-first project growth protocol")
	assertContains(t, status.Stdout, "Pressure ledger:")
	assertContains(t, status.Stdout, "Proof:")
	assertContains(t, status.Stdout, "Next proof gap:")
	assertContains(t, status.Stdout, "Principles:")
	assertContains(t, status.Stdout, "Readiness gate:")
	assertContains(t, status.Stdout, "Readiness pressure:")
	assertContains(t, status.Stdout, "Covered axes:")
	assertContains(t, status.Stdout, "Blocking gaps:")
	assertContains(t, status.Stdout, "Next:")
	assertContains(t, status.Stdout, "Growth pressures:")
	assertContains(t, status.Stdout, "Capability candidates:")

	resume, err := runCLI(args("resume"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("resume failed: %v", err)
	}
	assertContains(t, resume.Stdout, "Resuming RUN-0001 at GOAL-0001")
	assertContains(t, resume.Stdout, "Execution adapter: prompt")
}

func TestDoctorReportsProjectState(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a tiny notes MVP")
	mustRun(t, root, "run")

	out, err := runCLI(args("doctor"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("doctor failed: %v", err)
	}
	assertContains(t, out.Stdout, "Hyper Run Doctor")
	assertContains(t, out.Stdout, "Workspace: "+root)
	assertContains(t, out.Stdout, "plan.md")
	assertContains(t, out.Stdout, "SQLite")
	assertContains(t, out.Stdout, "Codex Desktop routing")
	assertContains(t, out.Stdout, "Current packet")
	assertContains(t, out.Stdout, "Summary:")
}

func TestDoctorWarnsWhenStoredReadinessIsStale(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a tiny notes MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`npm run build` passed and browser smoke verified the create note flow.\n\n## Readiness Evidence\n\nCore UX: Browser smoke verified create and complete flow.\nValidation coverage: `npm run build` passed and primary flow smoke test passed.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n")
	state, err := readState(filepath.Join(root, ".hyper", "state.json"))
	if err != nil {
		t.Fatalf("read state failed: %v", err)
	}
	state.Status = "completed"
	if err := writeJSON(filepath.Join(root, ".hyper", "state.json"), state); err != nil {
		t.Fatalf("write state failed: %v", err)
	}

	out, err := runCLI(args("doctor"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("doctor failed: %v", err)
	}
	assertContains(t, out.Stdout, "[WARN] Readiness state:")
	assertContains(t, out.Stdout, "Run `hyper migrate`")
}

func TestDoctorWarnsWhenNextPacketPlanIsStale(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plan.md"), "# Service Probe\n\n## Product Brief\n\nA tiny notes API.\n\n## Current Stage\n\nTiny MVP\n\n## Success Signals\n\nCreate and list one note.\n")
	mustRun(t, root, "init")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`go test ./...` passed.\n\n## Readiness Evidence\n\nProduct completeness: A tiny notes API now has a measurable create-and-list flow: `POST /notes` creates one note and `GET /notes` returns it.\nValidation coverage: `go test ./...` passed and the primary HTTP API flow is repeatable.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nDocument the API command surface.\n\n## Learn Notes\n\n- pattern: API MVPs should prove create/list with HTTP tests.\n")
	mustRun(t, root, "complete")
	writeFile(t, filepath.Join(root, ".hyper", "next-packet.md"), "# Next Packet Plan\n\nAction: advance\nCommand: hyper advance\n")

	out, err := runCLI(args("doctor"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("doctor failed: %v", err)
	}
	assertContains(t, out.Stdout, "[WARN] Next packet plan: expected `hyper run \"Implement the smallest usable A tiny notes API core flow: the primary user flow\"`, found `hyper advance`; run `hyper migrate`")
}

func TestDoctorDoesNotTrustNextPacketWhenRefreshIsNeeded(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny tasks", "Build a tiny task list MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`npm run build` passed and browser smoke passed.\n\n## Readiness Evidence\n\nCore UX: Browser smoke passed for create and complete flow.\nValidation coverage: `npm run build` passed and primary browser smoke is repeatable.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n")
	mustRun(t, root, "complete")
	stale := growthState{
		Version: 1,
		Pressures: []growthPressure{
			{State: "repeated", PressureType: "recurring_failure", Effect: "stop_condition", Signal: "None in this run.", GoalCount: 2},
		},
	}
	if err := writeJSON(filepath.Join(root, ".hyper", "growth", "state.json"), stale); err != nil {
		t.Fatalf("write stale growth failed: %v", err)
	}

	out, err := runCLI(args("doctor"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("doctor failed: %v", err)
	}
	assertContains(t, out.Stdout, "[WARN] Growth migration: legacy or noisy growth entries found; run `hyper migrate`")
	assertContains(t, out.Stdout, "[WARN] Next packet plan: cannot trust next-packet until refresh completes: legacy or noisy growth entries found; run `hyper migrate`")
}

func TestDoctorReadinessComparisonIgnoresIrrelevantFutureAxes(t *testing.T) {
	stored := readinessState{
		Stage: "Tiny MVP",
		StageGate: readinessStageGate{
			CurrentStage: "Tiny MVP",
			NextStage:    "Usable MVP",
			Status:       "ready",
			RequiredAxes: []string{"product_completeness", "core_ux", "validation_coverage"},
		},
		NextPressure: readinessPressure{Axis: "stage_advancement"},
		Dimensions: []readinessDimension{
			{ID: "product_completeness", Status: "covered"},
			{ID: "core_ux", Status: "covered"},
			{ID: "validation_coverage", Status: "covered"},
		},
	}
	current := stored
	current.Dimensions = append(current.Dimensions, readinessDimension{ID: "sustained_quality", Status: "missing"})
	if !sameReadinessForDoctor(stored, current) {
		t.Fatal("doctor should not warn when only an irrelevant future-stage axis was added")
	}
}

func TestRepairReconcilesStaleProjectState(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a tiny notes MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\nSmoke passed.\n\n## Blocker\n\nNo remaining blocker for this packet.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nContinue.\n")
	state, err := readState(filepath.Join(root, ".hyper", "state.json"))
	if err != nil {
		t.Fatalf("read state failed: %v", err)
	}
	state.Status = "blocked"
	if err := writeJSON(filepath.Join(root, ".hyper", "state.json"), state); err != nil {
		t.Fatalf("write stale state failed: %v", err)
	}

	doctor, err := runCLI(args("doctor"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("doctor failed: %v", err)
	}
	assertContains(t, doctor.Stdout, "Run `hyper repair`")

	out, err := runCLI(args("repair"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("repair failed: %v", err)
	}
	assertContains(t, out.Stdout, "State: repaired")
	assertContains(t, out.Stdout, "To: completed")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "state.json")), `"status": "completed"`)
}

func TestRepairRefreshesReadinessAndNextPacketPlan(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a tiny notes MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`npm run build` passed and browser smoke verified the create note flow.\n\n## Readiness Evidence\n\nCore UX: Browser smoke verified create and complete flow.\nValidation coverage: `npm run build` passed and primary flow smoke test passed.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n")
	state, err := readState(filepath.Join(root, ".hyper", "state.json"))
	if err != nil {
		t.Fatalf("read state failed: %v", err)
	}
	state.Status = "blocked"
	if err := writeJSON(filepath.Join(root, ".hyper", "state.json"), state); err != nil {
		t.Fatalf("write stale state failed: %v", err)
	}

	out, err := runCLI(args("repair"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("repair failed: %v", err)
	}
	assertContains(t, out.Stdout, "State: repaired")
	assertContains(t, out.Stdout, "Readiness gate: Tiny MVP -> Usable MVP (ready)")
	assertContains(t, out.Stdout, "Next action: hyper advance")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "next-packet.md")), "Action: advance")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "next-packet.md")), "Command: hyper advance")
}

func TestStatusShowsTopPressuresAndCandidateNames(t *testing.T) {
	state := projectState{Project: "Tiny CRM", Stage: "Tiny MVP", Status: "blocked", ActiveRunID: "RUN-0002", CurrentGoalID: "GOAL-0002", CurrentGoalPath: ".hyper/goals/GOAL-0002/goal.md", UpdatedAt: "now"}
	derived := goalState{State: "completed", Reason: "done"}
	growth := growthState{
		Pressures: []growthPressure{
			{State: "repeated", PressureType: "repeated_validation", Signal: "`npm run build` passed repeatedly."},
			{State: "observed", PressureType: "stable_decision", Signal: "Keep local-first storage."},
			{State: "repeated", PressureType: "implementation_pattern", Signal: "Error handling: proof - corrupted saved-state fallback remains unchanged in `loadState()`."},
		},
		Candidates: []growthCandidate{
			{Name: "validator-validation-pattern-npm-run-build-passed-vite-emitted-the-existing", Kind: "validator", Status: "promotable", Signal: "`npm run build` passed repeatedly.", EvidenceCount: 3},
			{Name: "skill-error-handling-proof-corrupted-saved-state-fallback-remains", Kind: "skill", Status: "repeated", Signal: "Error handling: proof - corrupted saved-state fallback remains unchanged in `loadState()`.", EvidenceCount: 2},
		},
	}
	out := strings.Join(statusDashboardLines(state, derived, readinessState{}, growth, 2, 2), "\n")
	assertContains(t, out, "Status: completed (state.json: blocked)")
	assertContains(t, out, "Top pressures:")
	assertContains(t, out, "repeated/repeated_validation")
	assertContains(t, out, "Candidate structures:")
	assertContains(t, out, "validator-npm-run-build")
	assertNotContains(t, out, "validator-validation-pattern-npm-run-build-passed-vite-emitted-the-existing")
	assertNotContains(t, out, "skill-error-handling-proof-corrupted-saved-state-fallback-remains")
}

func TestStatusShowsActionGuidance(t *testing.T) {
	state := projectState{Project: "Tiny CRM", Stage: "Tiny MVP", Status: "completed", ActiveRunID: "RUN-0001", CurrentGoalID: "GOAL-0001", CurrentGoalPath: ".hyper/goals/GOAL-0001/goal.md", UpdatedAt: "now"}
	derived := goalState{State: "completed", Reason: "done"}
	readiness := readinessState{
		Version: 1,
		StageGate: readinessStageGate{
			CurrentStage: "Tiny MVP",
			NextStage:    "Usable MVP",
			Status:       "not_ready",
			BlockingGaps: []string{"Core UX: not proven."},
		},
		NextPressure: readinessPressure{AxisName: "Core UX", Status: "emerging", Reason: "Core UX is emerging.", RecommendedGoal: "Prove the primary flow."},
	}
	out := strings.Join(statusDashboardLines(state, derived, readiness, growthState{}, 1, 1), "\n")
	assertContains(t, out, "Action:")
	assertContains(t, out, "Next action: hyper run \"Prove the primary flow.\"")
	assertContains(t, out, "Why now: Core UX is emerging.")
	assertContains(t, out, "Do not do yet: Do not advance Tiny MVP until blocking readiness gaps are closed.")
}

func TestStatusHighlightsReferenceBenchmarkWhenRequired(t *testing.T) {
	state := projectState{Project: "Tiny CRM", Stage: "Beta", Status: "completed", ActiveRunID: "RUN-0003", CurrentGoalID: "GOAL-0003", CurrentGoalPath: ".hyper/goals/GOAL-0003/goal.md", UpdatedAt: "now"}
	derived := goalState{State: "completed", Reason: "done"}
	readiness := readinessState{
		Version: 1,
		Stage:   "Beta",
		Dimensions: []readinessDimension{
			{ID: "core_ux", Name: "Core UX", Status: "covered", Evidence: "Browser smoke covered the primary flow."},
			{ID: "validation_coverage", Name: "Validation coverage", Status: "covered", Evidence: "`go test ./...` passed."},
			{ID: "reference_benchmark", Name: "Reference benchmark", Status: "missing", Gap: "Reference comparison has not proven category baseline and differentiating strength."},
		},
		StageGate: readinessStageGate{
			CurrentStage: "Beta",
			NextStage:    "Service Quality",
			Status:       "not_ready",
			RequiredAxes: []string{"validation_coverage", "security_baseline", "deployment_readiness", "operations_docs", "reference_benchmark"},
			BlockingGaps: []string{"Reference benchmark: Reference comparison has not proven category baseline and differentiating strength."},
		},
		NextPressure: readinessPressure{Axis: "reference_benchmark", AxisName: "Reference benchmark", Status: "missing", Reason: "Reference benchmark is missing for the Beta -> Service Quality gate.", RecommendedGoal: "Compare the service against references."},
	}

	full := strings.Join(statusDashboardLines(state, derived, readiness, growthState{}, 3, 3), "\n")
	assertContains(t, full, "Proof: functional covered, surface covered, operational covered, benchmark missing")
	assertContains(t, full, "Reference benchmark: missing - Reference comparison has not proven category baseline and differentiating strength.")
	assertContains(t, full, "Next proof gap: Reference benchmark")

	short := strings.Join(statusShortLines(state, derived, readiness, growthState{}), "\n")
	assertContains(t, short, "Benchmark: missing - Reference comparison has not proven category baseline and differentiating strength.")
	assertContains(t, short, "Gap: Reference benchmark: Reference comparison has not proven category baseline")
}

func TestStatusDoesNotReportSurfaceGapWhenCoreUXIsNotRequired(t *testing.T) {
	state := projectState{Project: "Local Build Relay", Stage: "Sustained Service Quality", Status: "completed", ActiveRunID: "RUN-0001", CurrentGoalID: "GOAL-0001", CurrentGoalPath: ".hyper/goals/GOAL-0001/goal.md", AutoContinue: true, RunUntil: "Sustained Service Quality"}
	derived := goalState{State: "completed", Reason: "done"}
	readiness := readinessState{
		Version: 1,
		Stage:   "Sustained Service Quality",
		Dimensions: []readinessDimension{
			{ID: "core_ux", Name: "Core UX", Status: "emerging", Evidence: "CLI command surface exists."},
			{ID: "validation_coverage", Name: "Validation coverage", Status: "covered", Evidence: "`go test ./...` passed."},
			{ID: "sustained_quality", Name: "Sustained quality", Status: "covered", Evidence: "Active validator is required."},
		},
		StageGate: readinessStageGate{
			CurrentStage: "Sustained Service Quality",
			NextStage:    "Sustained Service Quality",
			Status:       "ready",
			RequiredAxes: []string{"validation_coverage", "operations_docs", "maintainability", "sustained_quality"},
		},
		NextPressure: readinessPressure{Axis: "sustained_quality", AxisName: "Sustained quality", Status: "ongoing", Reason: "Continue focused quality work."},
	}

	short := strings.Join(statusShortLines(state, derived, readiness, growthState{}), "\n")
	assertContains(t, short, "Proof: functional covered, operational covered")
	assertNotContains(t, short, "surface emerging")
	assertNotContains(t, short, "surface proof for the primary user flow")
	assertNotContains(t, short, "Gap:")
}

func TestStatusDoesNotShowFutureReferenceBenchmarkBeforeRequired(t *testing.T) {
	state := projectState{Project: "Tiny Pet", Stage: "Tiny MVP", Status: "completed", ActiveRunID: "RUN-0013", CurrentGoalID: "GOAL-0013", CurrentGoalPath: ".hyper/goals/GOAL-0013/goal.md", UpdatedAt: "now"}
	derived := goalState{State: "completed", Reason: "done"}
	readiness := readinessState{
		Version: 1,
		Stage:   "Tiny MVP",
		Dimensions: []readinessDimension{
			{ID: "core_ux", Name: "Core UX", Status: "covered", Evidence: "Browser smoke covered the primary flow."},
			{ID: "validation_coverage", Name: "Validation coverage", Status: "covered", Evidence: "`go test ./...` passed."},
			{ID: "reference_benchmark", Name: "Reference benchmark", Status: "emerging", Evidence: "GOAL-0013 readiness evidence needs stronger proof for reference benchmark needs category, 3-5 named references."},
		},
		StageGate:    readinessStageGate{CurrentStage: "Tiny MVP", NextStage: "Usable MVP", Status: "ready", RequiredAxes: []string{"product_completeness", "core_ux", "validation_coverage"}, Advancement: stageAdvancementPolicy{Candidate: true}},
		NextPressure: readinessPressure{Axis: "stage_advancement", AxisName: "Stage advancement", Status: "candidate", Reason: "Tiny MVP gate is ready."},
	}

	dashboard := strings.Join(readinessDashboardLines(readiness), "\n")
	assertNotContains(t, dashboard, "Reference benchmark")
	assertNotContains(t, dashboard, "Benchmark:")
	assertNotContains(t, dashboard, "Emerging axes: Reference benchmark")

	short := strings.Join(statusShortLines(state, derived, readiness, growthState{}), "\n")
	assertContains(t, short, "Proof: functional covered, surface covered, operational covered")
	assertNotContains(t, short, "benchmark emerging")
	assertNotContains(t, short, "Benchmark:")
}

func TestStatusShortPrioritizesActivePacketGuard(t *testing.T) {
	state := projectState{Project: "LLog", Stage: "Beta", Status: "active", ActiveRunID: "RUN-0012", CurrentGoalID: "GOAL-0012", CurrentGoalPath: ".hyper/goals/GOAL-0012/goal.md", UpdatedAt: "now"}
	derived := goalState{State: "active", Reason: "Runtime packet evidence is still pending."}
	readiness := readinessState{
		Version: 1,
		Stage:   "Beta",
		Dimensions: []readinessDimension{
			{ID: "core_ux", Name: "Core UX", Status: "covered", Evidence: "Browser smoke covered the primary flow."},
			{ID: "validation_coverage", Name: "Validation coverage", Status: "covered", Evidence: "`go test ./...` passed."},
			{ID: "reference_benchmark", Name: "Reference benchmark", Status: "covered", Evidence: "GOAL-0011 readiness evidence: benchmark covered."},
		},
		StageGate: readinessStageGate{
			CurrentStage: "Beta",
			NextStage:    "Service Quality",
			Status:       "ready",
			RequiredAxes: []string{"validation_coverage", "security_baseline", "deployment_readiness", "operations_docs", "reference_benchmark"},
			Advancement:  stageAdvancementPolicy{Candidate: true, Recommendation: "Beta gate is ready."},
		},
		NextPressure: readinessPressure{Axis: "stage_advancement", AxisName: "Stage advancement", Status: "candidate", Reason: "Beta gate is ready."},
	}

	short := strings.Join(statusShortLines(state, derived, readiness, growthState{}), "\n")
	assertContains(t, short, "Next: update .hyper/goals/GOAL-0012/evidence.md and next.md, then run `hyper complete`")
	assertContains(t, short, "Guard: Do not start another `hyper run` until this packet is completed or blocked.")
	assertNotContains(t, short, "Guard: accept the stage change before running `hyper advance`")
}

func TestStatusShortGapMatchesNextReadinessPressure(t *testing.T) {
	state := projectState{Project: "Active Guard CLI", Stage: "Tiny MVP", Status: "active", ActiveRunID: "RUN-0001", CurrentGoalID: "GOAL-0001", CurrentGoalPath: ".hyper/goals/GOAL-0001/goal.md", AutoContinue: true, RunUntil: "Service Quality"}
	derived := goalState{State: "active", Reason: "Runtime packet evidence is still pending."}
	readiness := readinessState{
		Version: 1,
		Stage:   "Tiny MVP",
		Dimensions: []readinessDimension{
			{ID: "validation_coverage", Name: "Validation coverage", Status: "missing", Gap: "The primary behavior does not have repeatable validation evidence."},
		},
		StageGate: readinessStageGate{
			CurrentStage: "Tiny MVP",
			NextStage:    "Usable MVP",
			Status:       "not_ready",
			RequiredAxes: []string{"product_completeness", "core_ux", "validation_coverage"},
			BlockingGaps: []string{
				"Core UX: The primary user flow is not yet proven usable.",
				"Validation coverage: The primary behavior does not have repeatable validation evidence.",
			},
		},
		NextPressure: readinessPressure{
			Axis:     "validation_coverage",
			AxisName: "Validation coverage",
			Status:   "missing",
			Reason:   "Validation coverage is missing for the Tiny MVP -> Usable MVP gate.",
		},
	}

	short := strings.Join(statusShortLines(state, derived, readiness, growthState{}), "\n")
	assertContains(t, short, "Gap: Validation coverage: The primary behavior does not have repeatable validation evidence.")
	assertNotContains(t, short, "Gap: Core UX")
}

func TestRunBlocksPendingActiveGoal(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a tiny notes MVP")
	mustRun(t, root, "run")

	_, err := runCLI(args("run", "Start another packet"), testRoot(root), fakeUpdater{})
	if err == nil {
		t.Fatal("expected pending active goal to block next run")
	}
	assertContains(t, err.Message, "Current runtime packet is still active")
	assertContains(t, err.Message, "hyper complete")
}

func TestRunBlocksCompletedEvidenceBeforeFinishGate(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a tiny notes MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`go test ./...` passed.\n\n## Readiness Evidence\n\nValidation coverage: `go test ./...` passed and is repeatable.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nAdd the primary notes flow.\n\n## Learn Notes\n\n- Pattern: Run go test before handoff.\n")

	_, err := runCLI(args("run", "Start another packet"), testRoot(root), fakeUpdater{})
	if err == nil {
		t.Fatal("expected finish gate guard to block next run")
	}
	assertContains(t, err.Message, "has not passed the finish gate yet")
	assertContains(t, err.Message, "hyper complete")
	assertContains(t, err.Message, "review.md")
	if exists(filepath.Join(root, ".hyper", "goals", "GOAL-0002")) {
		t.Fatal("new runtime packet should not be created before the finish gate passes")
	}
}

func TestCompleteLearnsAndRefreshesReadiness(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a tiny notes MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`go test ./...` passed and primary CLI smoke passed.\n\n## Readiness Evidence\n\nCore UX: CLI smoke passed for create and complete flow.\nValidation coverage: `go test ./...` passed and is repeatable.\n\n## Decisions\n\nKeep notes local-first.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n\n## Learn Notes\n\n- Pattern: Run go test before handoff.\n")

	out, err := runCLI(args("complete"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	assertContains(t, out.Stdout, "Completed runtime packet: GOAL-0001")
	assertContains(t, out.Stdout, "State: completed")
	assertContains(t, out.Stdout, "Memory quality:")
	assertContains(t, out.Stdout, "Proof: functional covered, surface covered, operational covered")
	assertContains(t, out.Stdout, "Readiness gate: Tiny MVP -> Usable MVP (ready)")
	assertContains(t, out.Stdout, "Next action: hyper advance")
	assertContains(t, out.Stdout, "Why: Tiny MVP gate is ready.")
	assertContains(t, out.Stdout, "  hyper advance")
	assertContains(t, out.Stdout, "  hyper status --short")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "state.json")), `"status": "completed"`)
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "readiness", "state.json")), `"candidate": true`)

	status, err := runCLI(args("status"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	assertContains(t, status.Stdout, "Last run: RUN-0001")
	assertContains(t, status.Stdout, "Last runtime packet: GOAL-0001")
	assertNotContains(t, status.Stdout, "Active run: RUN-0001")
}

func TestCompleteRejectsActivePacket(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a tiny notes MVP")
	mustRun(t, root, "run")

	_, err := runCLI(args("complete"), testRoot(root), fakeUpdater{})
	if err == nil {
		t.Fatal("expected active packet to block complete")
	}
	assertContains(t, err.Message, "Current runtime packet is still active")
	assertContains(t, err.Message, ".hyper/goals/GOAL-0001/evidence.md")
	assertContains(t, err.Message, ".hyper/goals/GOAL-0001/next.md")
}

func TestCompleteRunsFinishGateBeforeLearning(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a tiny notes MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-smoke.md"), "# validator-smoke\n\nStatus: active\nKind: validator\nSignal: Run npm run smoke before completing packets.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`go test ./...` passed.\n\n## Readiness Evidence\n\nCore UX: CLI smoke passed for create and complete flow.\nValidation coverage: `go test ./...` passed and is repeatable.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nAdd the next slice.\n")

	_, err := runCLI(args("complete"), testRoot(root), fakeUpdater{})
	if err == nil {
		t.Fatal("expected finish gate to reject missing active capability evidence")
	}
	assertContains(t, err.Message, "Finish gate failed for GOAL-0001")
	assertContains(t, err.Message, "Record active capability evidence for: validator-smoke")
	review := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "review.md"))
	assertContains(t, review, "Status: failed")
	assertContains(t, review, "Stay in the same runtime packet")
	assertNotContains(t, readFile(t, filepath.Join(root, ".hyper", "state.json")), `"status": "completed"`)
}

func TestCompleteRequiresSpecificActiveCapabilityEvidence(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny CLI", "Build a tiny CLI MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-go-test.md"), "# validator-go-test\n\nStatus: active\nKind: validator\nSignal: Run go test ./... before completing packets.\n")
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "harness", "harness-growth-candidate.md"), "# harness-growth-candidate\n\nStatus: active\nKind: harness\n\n## Required Behavior\n\nRun the project-specific handoff harness before completing packets.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`go test ./...` passed.\n\n## Readiness Evidence\n\nCore UX: CLI smoke verified create and complete flow.\nValidation coverage: `go test ./...` passed and primary CLI smoke is repeatable.\n\n## Active Capability Evidence\n\nNone active.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n")

	_, err := runCLI(args("complete"), testRoot(root), fakeUpdater{})
	if err == nil {
		t.Fatal("expected active capability evidence to name or prove the validator")
	}
	assertContains(t, err.Message, "Record active capability evidence for:")
	assertContains(t, err.Message, "harness-growth-candidate")
	assertNotContains(t, err.Message, "validator-go-test")

	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`go test ./...` passed.\n\n## Readiness Evidence\n\nCore UX: CLI smoke verified create and complete flow.\nValidation coverage: `go test ./...` passed and primary CLI smoke is repeatable.\n\n## Active Capability Evidence\n\nvalidator-go-test: `go test ./...` passed.\nharness-growth-candidate: project-specific handoff harness passed.\n\n## Blocker\n\nNone blocking.\n")
	if _, err := runCLI(args("complete"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("complete should accept named active capability evidence: %v", err)
	}
}

func TestCompleteRejectsPendingActiveCapabilityTemplate(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny CLI", "Build a tiny CLI MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-go-test.md"), "# validator-go-test\n\nStatus: active\nKind: validator\nSignal: Run go test ./... before completing packets.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`npm run build` passed.\n\n## Readiness Evidence\n\nCore UX: CLI smoke verified create and complete flow.\nValidation coverage: `npm run build` passed and primary CLI smoke is repeatable.\n\n## Active Capability Evidence\n\nvalidator-go-test: Pending. Required behavior: Run go test ./... before completing packets.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n")

	_, err := runCLI(args("complete"), testRoot(root), fakeUpdater{})
	if err == nil {
		t.Fatal("expected pending active capability template to fail finish gate")
	}
	assertContains(t, err.Message, "Record active capability evidence for: validator-go-test")
}

func TestCompleteAcceptsValidationOutputForActiveValidator(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny CLI", "Build a tiny CLI MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-go-test.md"), "# validator-go-test\n\nStatus: active\nKind: validator\nSignal: Run go test ./... before completing packets.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\nCommand: `go test ./...`\n\nOutput:\n\n```text\nok ./...\n```\n\n## Readiness Evidence\n\nCore UX: CLI smoke verified create and complete flow.\nValidation coverage: `go test ./...` passed and primary CLI smoke is repeatable.\n\n## Active Capability Evidence\n\nvalidator-go-test: Pending. Required behavior: Run go test ./... before completing packets.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n")

	out, err := runCLI(args("complete"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("validation output should satisfy active validator proof: %v", err)
	}
	assertContains(t, out.Stdout, "Finish gate: passed")
}

func TestCompleteRejectsFailedValidationOutputForActiveValidator(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny CLI", "Build a tiny CLI MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-go-test.md"), "# validator-go-test\n\nStatus: active\nKind: validator\nSignal: Run go test ./... before completing packets.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\nCommand: `go test ./...`\n\nOutput:\n\n```text\nFAIL ./...\n```\n\ngo test ./... failed.\n\n## Readiness Evidence\n\nCore UX: CLI smoke verified create and complete flow.\nValidation coverage: `go test ./...` failed and needs repair.\n\n## Active Capability Evidence\n\nvalidator-go-test: Pending. Required behavior: Run go test ./... before completing packets.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nRepair the failing validation.\n")

	_, err := runCLI(args("complete"), testRoot(root), fakeUpdater{})
	if err == nil {
		t.Fatal("failed validator output must not satisfy active validator proof")
	}
	assertContains(t, err.Message, "Record active capability evidence for: validator-go-test")
}

func TestCompleteRejectsFailedActiveValidatorWhenAnotherValidationPassed(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny CLI", "Build a tiny CLI MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-go-test.md"), "# validator-go-test\n\nStatus: active\nKind: validator\nSignal: Run go test ./... before completing packets.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\nCommand: `go test ./...`\n\nOutput:\n\n```text\nFAIL ./...\n```\n\ngo test ./... failed.\n\nCommand: `npm run build`\n\nOutput:\n\n```text\nbuilt in 120ms\n```\n\nnpm run build passed.\n\n## Readiness Evidence\n\nCore UX: CLI smoke verified create and complete flow.\nValidation coverage: `npm run build` passed, but `go test ./...` failed.\n\n## Active Capability Evidence\n\nvalidator-go-test: Pending. Required behavior: Run go test ./... before completing packets.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nRepair the failing active validator.\n")

	_, err := runCLI(args("complete"), testRoot(root), fakeUpdater{})
	if err == nil {
		t.Fatal("a different passing validation command must not satisfy a failed active validator")
	}
	assertContains(t, err.Message, "Record active capability evidence for: validator-go-test")
}

func TestCompleteAcceptsExplicitActiveCapabilityBlocker(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny CLI", "Build a tiny CLI MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-go-test.md"), "# validator-go-test\n\nStatus: active\nKind: validator\nSignal: Run go test ./... before completing packets.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`go test ./...` passed.\n\n## Readiness Evidence\n\nCore UX: CLI smoke verified create and complete flow.\nValidation coverage: `go test ./...` passed and primary CLI smoke is repeatable.\n\n## Active Capability Evidence\n\nvalidator-go-test: blocked because missing credentials for the private module registry.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n")

	out, err := runCLI(args("complete"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("explicit active capability blocker should satisfy finish gate: %v", err)
	}
	assertContains(t, out.Stdout, "Finish gate: passed")
}

func TestCompleteAllowsEmergingSustainedQualityEvidence(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plan.md"), "# Product Plan\n\n## Product\n\nLocal Build Relay\n\n## Target Users\n\nDevelopers\n\n## MVP\n\nRun one handoff command.\n\n## Current Stage\n\nService Quality\n\n## Build Style\n\nGo CLI\n\n## Success Criteria\n\nEvery packet proves the handoff command.\n")
	if _, err := runCLI(args("run", "Repeat handoff validation"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`go test ./...` passed.\n\n## Readiness Evidence\n\nSustained quality: Repeated runtime evidence exists for the same handoff validation pattern, but it is not active required behavior yet.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nRepeat validation again.\n")

	if _, err := runCLI(args("complete"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("emerging sustained quality evidence should allow packet closure: %v", err)
	}
}

func TestRunAutoUntilPlansNextPacketAfterComplete(t *testing.T) {
	root := t.TempDir()
	mustRun(t, root, "init")
	writeFile(t, filepath.Join(root, "plan.md"), "# Product Plan\n\n## Product\n\nTiny CRM\n\n## Target Users\n\nSolo sellers\n\n## MVP\n\nAdd and revisit customer notes.\n\n## Current Stage\n\nUsable MVP\n\n## Build Style\n\nWeb app\n\n## Non-goals\n\nTeam collaboration\n\n## Constraints\n\nLocal first\n\n## Success Criteria\n\nPrimary customer notes flow works without manual data edits.\n\n## Current Focus\n\nImprove customer notes.\n")

	out, err := runCLI(args("run", "--auto", "--until", "service-quality", "Upgrade Tiny CRM toward service quality"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("auto run failed: %v", err)
	}
	assertContains(t, out.Stdout, "Run mode: auto until Service Quality")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "state.json")), `"auto_continue": true`)
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "state.json")), `"run_until": "Service Quality"`)

	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`npm run build` passed and browser smoke verified the primary notes flow.\n\n## Readiness Evidence\n\nCore UX: Browser smoke verified create and revisit customer notes flow.\nData persistence: SQLite database stored a created customer note and confirmed the row after reload.\nValidation coverage: `npm run build` passed and primary browser smoke is repeatable.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nHandle empty, failure, and edge states.\n")

	complete, err := runCLI(args("complete"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	assertContains(t, complete.Stdout, "Finish gate: passed")
	assertContains(t, complete.Stdout, "Next action: hyper run --auto --until \"Service Quality\" \"Handle empty, failure, and edge states for the primary Tiny CRM flow.\"")
	nextPlan := readFile(t, filepath.Join(root, ".hyper", "next-packet.md"))
	assertContains(t, nextPlan, "Mode: auto until Service Quality")
	assertContains(t, nextPlan, "Action: run")
	assertContains(t, nextPlan, "Command: hyper run --auto --until \"Service Quality\"")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "review.md")), "Status: passed")
}

func TestStatusAutoTargetReachedExplainsPause(t *testing.T) {
	state := projectState{
		Project:         "Local Clip Shelf",
		Stage:           "Service Quality",
		Status:          "completed",
		ActiveRunID:     "RUN-0001",
		CurrentGoalID:   "GOAL-0001",
		CurrentGoalPath: ".hyper/goals/GOAL-0001/goal.md",
		AutoContinue:    true,
		RunUntil:        "Service Quality",
	}
	derived := goalState{State: "completed", Reason: "done"}
	readiness := readinessState{
		Version: 1,
		Stage:   "Service Quality",
		StageGate: readinessStageGate{
			CurrentStage: "Service Quality",
			NextStage:    "Sustained Service Quality",
			Status:       "not_ready",
			BlockingGaps: []string{"Maintainability: The codebase has not accumulated enough maintainability evidence."},
		},
		NextPressure: readinessPressure{Axis: "maintainability", AxisName: "Maintainability", Status: "emerging", Reason: "Maintainability is emerging for the Service Quality -> Sustained Service Quality gate."},
	}

	short := strings.Join(statusShortLines(state, derived, readiness, growthState{}), "\n")
	assertContains(t, short, "Next: hyper status --short")
	assertContains(t, short, "Why: Auto target Service Quality is reached; review status before choosing a new target or manual next run.")
	assertNotContains(t, short, "Why: Maintainability is emerging")
}

func TestRunAutoUntilDoesNotCreatePacketAfterTargetReached(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plan.md"), "# Product Plan\n\n## Product\n\nTiny Bookmark CLI\n\n## Target Users\n\nDevelopers\n\n## MVP\n\nAdd and list one bookmark.\n\n## Current Stage\n\nTiny MVP\n\n## Build Style\n\nGo CLI\n\n## Success Criteria\n\nCommand-surface add/list flow is repeatable.\n")
	mustRun(t, root, "init")
	if _, err := runCLI(args("run", "--auto", "--until", "usable-mvp", "Build the bookmark CLI"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("auto run failed: %v", err)
	}
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`go test ./...` passed.\n\n## Readiness Evidence\n\nProduct completeness: Tiny Bookmark CLI has a measurable add/list command flow.\nCore UX: CLI command test passed for add and list behavior, proving the primary bookmark flow works from the command surface.\nValidation coverage: `go test ./...` passed and the primary CLI add/list flow is repeatable.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview Tiny MVP evidence.\n\n## Learn Notes\n\n- pattern: CLI MVPs can use command-surface proof for Core UX.\n")
	if _, err := runCLI(args("complete"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	if _, err := runCLI(args("advance"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("advance failed: %v", err)
	}

	out, err := runCLI(args("run", "--auto", "--until", "usable-mvp", "Should not continue"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("run at reached target should stop cleanly: %v", err)
	}
	assertContains(t, out.Stdout, "Run-until target already reached: Usable MVP")
	assertContains(t, out.Stdout, "No runtime packet created.")
	assertContains(t, out.Stdout, "Next action: hyper status --short")
	if exists(filepath.Join(root, ".hyper", "goals", "GOAL-0002")) {
		t.Fatal("auto run should not create another packet after the target stage is reached")
	}
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "next-packet.md")), "Action: stop")
}

func TestRunAutoUntilReachedBeforeFirstPacketKeepsHandoffConsistent(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plan.md"), "# Product Plan\n\n## Product\n\nAuto Target Guard\n\n## Target Users\n\nDevelopers\n\n## MVP\n\nOne command flow is already usable.\n\n## Current Stage\n\nUsable MVP\n\n## Build Style\n\nGo CLI\n\n## Success Criteria\n\nAuto run-until does not create work after the target stage is reached.\n")
	mustRun(t, root, "init")

	out, err := runCLI(args("run", "--auto", "--until", "usable-mvp", "Should not start"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("run at reached target should stop cleanly: %v", err)
	}
	assertContains(t, out.Stdout, "No runtime packet created.")
	if exists(filepath.Join(root, ".hyper", "goals", "GOAL-0001")) {
		t.Fatal("auto run should not create a first packet when the target stage is already reached")
	}
	status, err := runCLI(args("status", "--short"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	assertContains(t, status.Stdout, "Mode: auto until Usable MVP")
	assertContains(t, status.Stdout, "Next: hyper status --short")
	doctor, err := runCLI(args("doctor"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("doctor failed: %v", err)
	}
	assertContains(t, doctor.Stdout, "[OK] Next packet plan: .hyper/next-packet.md matches current state")
}

func TestRunAutoUntilSustainedQualityPromotesActiveValidator(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plan.md"), "# Product Plan\n\n## Product\n\nService Quality Chain\n\n## Target Users\n\nDevelopers\n\n## MVP\n\nAdd one release note and list it back.\n\n## Current Stage\n\nTiny MVP\n\n## Build Style\n\nLocal CLI\n\n## Success Criteria\n\nReach sustained quality only after repeated validation becomes active required behavior.\n")
	mustRun(t, root, "init")
	if _, err := runCLI(args("run", "--auto", "--until", "sustained-service-quality", "Drive to sustained quality"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("initial auto run failed: %v", err)
	}
	validation := "`./check.sh` passed with output: `release-note add/list/error smoke passed`."
	writeEvidence := func(goalID, readiness string) {
		writeFile(t, filepath.Join(root, ".hyper", "goals", goalID, "evidence.md"), "# "+goalID+" Evidence\n\n## Validation\n\n"+validation+"\n\n## Readiness Evidence\n\n"+readiness+"\n\n## Active Capability Evidence\n\nNo active project capability required yet.\n\n## Changed Files\n\nfixture\n\n## Decisions\n\nKeep the local CLI boundary.\n\n## Reusable Patterns\n\nUse `./check.sh` as the repeated validation path.\n\n## Blockers\n\nNone blocking.\n")
		writeFile(t, filepath.Join(root, ".hyper", "goals", goalID, "next.md"), "# "+goalID+" Next\n\n## Recommended Next Goal\n\nContinue toward sustained quality.\n\n## Learn Notes\n\n- pattern: Use `./check.sh` as the repeated validation path.\n")
	}
	complete := func(goalID string) string {
		out, err := runCLI(args("complete"), testRoot(root), fakeUpdater{})
		if err != nil {
			t.Fatalf("complete %s failed: %v", goalID, err)
		}
		assertContains(t, out.Stdout, "Finish gate: passed")
		return out.Stdout
	}
	advance := func() {
		if _, err := runCLI(args("advance"), testRoot(root), fakeUpdater{}); err != nil {
			t.Fatalf("advance failed: %v", err)
		}
	}
	nextRun := func(focus string) {
		if _, err := runCLI(args("run", "--auto", "--until", "sustained-service-quality", focus), testRoot(root), fakeUpdater{}); err != nil {
			t.Fatalf("auto run failed: %v", err)
		}
	}

	writeEvidence("GOAL-0001", "Product completeness: Service Quality Chain has a measurable add/list CLI slice.\nCore UX: CLI smoke passed for add and list behavior from the command surface.\nValidation coverage: "+validation)
	complete("GOAL-0001")
	advance()
	nextRun("Prove persistence and error handling")

	writeEvidence("GOAL-0002", "Data persistence: Text file storage saved a release note and a separate list command re-read it from disk.\nError handling: Missing argument and unknown command states are handled and verified.\nValidation coverage: "+validation)
	complete("GOAL-0002")
	advance()
	nextRun("Prove beta service quality axes")

	writeEvidence("GOAL-0003", strings.Join([]string{
		"Validation coverage: " + validation,
		"Security baseline: Local-only security and privacy boundary is documented and verified: no cloud sync, no telemetry, no secrets, no tokens, and no sessions.",
		"Deployment readiness: Release artifacts are created in `dist/` and the packaged smoke command passed outside the source command path.",
		"Operations and docs: README documents setup, run command, smoke command, rollback, recovery, and stop condition.",
		"Reference benchmark: Category: Local CLI release-note tracker; References: git-chglog, standard-version, release-it; Baseline expectations: A useful local release-note CLI should add entries, list entries, keep data local, expose a repeatable smoke command, and document rollback or recovery; Current comparison: Service Quality Chain meets baseline for local add/list, file-backed persistence, repeatable smoke validation, local-only security boundary, and rollback docs; Below baseline gaps: none critical for the local-only CLI category; Above baseline strength: Hyper Run evidence ties validation, security, release artifact, docs, and benchmark proof to stage advancement; Decision: Service Quality advancement is acceptable because no core category-baseline gap remains.",
	}, "\n"))
	complete("GOAL-0003")
	advance()
	nextRun("Promote repeated validation to active required behavior")

	writeEvidence("GOAL-0004", "Maintainability: Documented validation helper coverage in `DEVELOPMENT.md`; the maintained `./check.sh` helper keeps command validation repeatable without hidden local context and reduces future operator handoff friction.\nValidation coverage: "+validation+"\nOperations and docs: DEVELOPMENT and README documents setup, validation, rollback, recovery, and handoff constraints for the next operator.")
	out := complete("GOAL-0004")
	assertContains(t, out, "1 active structure")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-check-sh.md")), "Status: active")
	advance()

	status, err := runCLI(args("status", "--short"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	assertContains(t, status.Stdout, "Stage: Sustained Service Quality")
	assertContains(t, status.Stdout, "Next: hyper status --short")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "next-packet.md")), "Action: stop")
}

func TestStatusAutoTargetReachedDoesNotHideActivePacket(t *testing.T) {
	state := projectState{
		Project:         "Local Clip Shelf",
		Stage:           "Service Quality",
		Status:          "active",
		ActiveRunID:     "RUN-0002",
		CurrentGoalID:   "GOAL-0002",
		CurrentGoalPath: ".hyper/goals/GOAL-0002/goal.md",
		AutoContinue:    true,
		RunUntil:        "Service Quality",
	}
	derived := goalState{State: "active", Reason: "Runtime packet evidence is still pending."}
	readiness := readinessState{Version: 1, Stage: "Service Quality"}

	short := strings.Join(statusShortLines(state, derived, readiness, growthState{}), "\n")
	assertContains(t, short, "Next: update .hyper/goals/GOAL-0002/evidence.md and next.md, then run `hyper complete`")
	assertContains(t, short, "Why: The current runtime packet is still open")
}

func TestGoalStateTreatsNoRemainingBlockerAsCompleted(t *testing.T) {
	state := deriveGoalState("## Validation\n\nSmoke passed.\n\n## Blocker\n\nNo remaining blocker for this packet. Final art still needs a designer asset.\n", "## Recommended Next Goal\n\nContinue.\n")
	if state.State != "completed" {
		t.Fatalf("expected no-remaining blocker text to complete, got %+v", state)
	}
}

func TestStatusDerivesReadinessForLegacyState(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a tiny notes MVP")
	mustRun(t, root, "run")
	if err := os.Remove(filepath.Join(root, ".hyper", "readiness", "state.json")); err != nil {
		t.Fatalf("remove readiness state failed: %v", err)
	}

	status, err := runCLI(args("status"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	assertContains(t, status.Stdout, "Readiness gate:")
	assertContains(t, status.Stdout, "Covered axes:")
	assertNotContains(t, status.Stdout, "Readiness: not recorded")
}

func TestStatusRefreshesReadinessFromLatestEvidence(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny tasks", "Build a tiny task list MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`npm run build` passed and browser smoke passed.\n\n## Readiness Evidence\n\nCore UX: Browser smoke passed for create and complete flow.\nValidation coverage: `npm run build` passed and primary browser smoke is repeatable.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n")
	mustRun(t, root, "complete")

	status, err := runCLI(args("status"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	assertContains(t, status.Stdout, "Readiness gate: Tiny MVP -> Usable MVP (ready)")
	assertContains(t, status.Stdout, "Stage advancement:")
	assertContains(t, status.Stdout, "Next action: hyper advance")
	assertContains(t, status.Stdout, "Recommended action: hyper advance")
}

func TestStatusShortShowsOnlyDecisionSurface(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny tasks", "Build a tiny task list MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`npm run build` passed and browser smoke passed.\n\n## Readiness Evidence\n\nCore UX: Browser smoke passed for create and complete flow.\nValidation coverage: `npm run build` passed and primary browser smoke is repeatable.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n")
	mustRun(t, root, "complete")

	status, err := runCLI(args("status", "--short"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status --short failed: %v", err)
	}
	assertContains(t, status.Stdout, "Hyper Run Status")
	assertContains(t, status.Stdout, "Stage: Tiny MVP")
	assertContains(t, status.Stdout, "Gate: Tiny MVP -> Usable MVP (ready)")
	assertContains(t, status.Stdout, "Proof: functional covered, surface covered, operational covered")
	assertContains(t, status.Stdout, "Packet: GOAL-0001 (completed)")
	assertContains(t, status.Stdout, "Next: hyper advance")
	assertContains(t, status.Stdout, "Gap: none; stage advancement is ready")
	assertContains(t, status.Stdout, "Guard: accept the stage change before running `hyper advance`")
	assertNotContains(t, status.Stdout, "Pressure Ledger:")
	assertNotContains(t, status.Stdout, "Readiness:")
}

func TestStatusSuggestsMigrateBeforeNextActionWhenGrowthIsStale(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny tasks", "Build a tiny task list MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`npm run build` passed and browser smoke passed.\n\n## Readiness Evidence\n\nCore UX: Browser smoke passed for create and complete flow.\nValidation coverage: `npm run build` passed and primary browser smoke is repeatable.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n")
	mustRun(t, root, "complete")
	stale := growthState{
		Version: 1,
		Pressures: []growthPressure{
			{State: "repeated", PressureType: "recurring_failure", Effect: "stop_condition", Signal: "None in this run.", GoalCount: 2},
		},
	}
	if err := writeJSON(filepath.Join(root, ".hyper", "growth", "state.json"), stale); err != nil {
		t.Fatalf("write stale growth failed: %v", err)
	}

	short, err := runCLI(args("status", "--short"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status --short failed: %v", err)
	}
	assertContains(t, short.Stdout, "Next: hyper migrate")
	assertContains(t, short.Stdout, "Refresh: legacy or noisy growth entries found; run `hyper migrate`")
	assertContains(t, short.Stdout, "Guard: run `hyper migrate` before advancing or starting another packet")
	assertNotContains(t, short.Stdout, "Next: hyper advance")

	full, err := runCLI(args("status"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	assertContains(t, full.Stdout, "State refresh: needed - legacy or noisy growth entries found; run `hyper migrate`")
	assertContains(t, full.Stdout, "Next action: hyper migrate")
	assertContains(t, full.Stdout, "Do not advance or start another packet until `hyper migrate` refreshes growth and readiness state.")
}

func TestStatusDoesNotPutMigrateBeforeActivePacketCompletion(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny tasks", "Build a tiny task list MVP")
	mustRun(t, root, "run")
	stale := growthState{
		Version: 1,
		Pressures: []growthPressure{
			{State: "repeated", PressureType: "recurring_failure", Effect: "stop_condition", Signal: "None in this run.", GoalCount: 2},
		},
	}
	if err := writeJSON(filepath.Join(root, ".hyper", "growth", "state.json"), stale); err != nil {
		t.Fatalf("write stale growth failed: %v", err)
	}

	short, err := runCLI(args("status", "--short"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status --short failed: %v", err)
	}
	assertContains(t, short.Stdout, "Next: update .hyper/goals/GOAL-0001/evidence.md and next.md, then run `hyper complete`")
	assertContains(t, short.Stdout, "Refresh: legacy or noisy growth entries found; run `hyper migrate`")
	assertNotContains(t, short.Stdout, "Next: hyper migrate")
}

func TestStatusShortRejectsUnknownOption(t *testing.T) {
	_, err := runCLI(args("status", "--json"), testRoot(t.TempDir()), fakeUpdater{})
	if err == nil {
		t.Fatal("expected unknown status option to fail")
	}
	assertContains(t, err.Message, "Unknown status option: --json")
	assertContains(t, err.Message, "hyper status --short")
}

func TestInitPreservesActiveGoal(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a tiny notes MVP")
	mustRun(t, root, "run")

	out, err := runCLI(args("init"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("second init failed: %v", err)
	}
	assertContains(t, out.Stdout, "Active runtime packet preserved: GOAL-0001")
	assertContains(t, out.Stdout, "hyper resume")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "state.json")), `"current_goal_id": "GOAL-0001"`)
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "logs", "project.jsonl")), "project_init_checked")
}

func TestInitWritesPlanImportCandidates(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "docs", "PICKACHAT_PLAN.md"), "# Pickachat 기획안\n\n## 제품 한 줄 정의\n\n지도 기반 채팅 서비스입니다.\n\n## MVP\n\n지도에서 핀을 만들고 메시지를 보냅니다.\n\n## 첫 버전 완료 기준\n\n빌드와 smoke validation이 통과합니다.\n")

	out, err := runCLI(args("init"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	assertContains(t, out.Stdout, "Plan import candidates: .hyper/plan-candidates.md")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "plan-candidates.md")), "docs/PICKACHAT_PLAN.md")
}

func TestParsePlanUnderstandsKoreanProductPlans(t *testing.T) {
	plan := parsePlan(`# LLog 서비스 기획 및 앱 개발 계획서

## 1. 제품 정의

### 프로젝트명

**LLog / 엘로그**

### 한 줄 소개

LLog는 사주 기반 운세를 매일 확인하고 기록하는 운세 캘린더 앱입니다.

## 4. 타깃 사용자

20~35세 여성 사용자

## 5. MVP 목표

첫 번째 버전의 목표는 운세 확인, 캘린더, 하루 기록, 간단 리포트까지 이어지는 루프입니다.

## 11. 모바일 앱 개발 방향

React Native + Expo + TypeScript

## 13. 성공 지표

첫 사용자 테스트에서 프로필 입력부터 리포트까지 완료합니다.

## 19. 우선순위

반드시 먼저 만들 것은 온보딩, 홈, 캘린더, 기록, 리포트입니다.
`)
	if got := plan["Product"]; got != "LLog / 엘로그" {
		t.Fatalf("expected Korean product alias, got %q", got)
	}
	assertContains(t, plan["Target Users"], "20~35세")
	assertContains(t, plan["MVP"], "운세 확인")
	assertContains(t, plan["Build Style"], "React Native")
	assertContains(t, plan["Success Criteria"], "프로필 입력")
	assertContains(t, plan["Current Focus"], "온보딩")
}

func TestStatusRefreshesUnknownProjectFromPlan(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plan.md"), "# LLog 서비스 기획 및 앱 개발 계획서\n\n## 1. 제품 정의\n\n### 프로젝트명\n\n**LLog / 엘로그**\n")
	state := refreshStateFromPlanForStatus(root, projectState{Project: "Unknown project", Stage: "Tiny MVP"})
	if state.Project != "LLog / 엘로그" {
		t.Fatalf("expected status project to come from plan.md, got %q", state.Project)
	}
}

func TestAutoLearnFeedsNextGoalContext(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny CRM", "Build a tiny CRM MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\nCustomer records persisted in SQLite. go test passed.\n\n## Readiness Evidence\n\nProduct completeness: Tiny CRM has a measurable create-and-list customer record flow.\nCore UX: CLI smoke verified create and list customer records from the command surface.\nValidation coverage: go test passed and the customer persistence smoke is repeatable.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nAdd persisted customer records.\n")
	mustRun(t, root, "complete")

	out, err := runCLI(args("run", "Add persisted customer records"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	assertContains(t, out.Stdout, "Auto learn: completed, inserted 0")
	assertContains(t, out.Stdout, "Similar context: ")
	assertContains(t, strings.ToLower(readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0002", "goal.md"))), "customer records persisted")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "logs", "RUN-0001.jsonl")), "runtime_packet_completed")
}

func TestGrowthStateChangesNextRuntimePacket(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a local-first notes MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\ngo test ./... passed.\n\n## Readiness Evidence\n\nProduct completeness: Tiny notes has a measurable local note command slice.\nCore UX: CLI smoke verified the primary note command surface.\nValidation coverage: go test ./... passed and is repeatable.\n\n## Changed Files\n\ncmd/notes.go\n\n## Decisions\n\nKeep local-first storage.\n\n## Reusable Patterns\n\nRun go test before every runtime packet handoff.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nAdd note editing polish.\n\n## Learn Notes\n\n- Pattern: Run go test before every runtime packet handoff.\n- Constraint: Do not add external services without credentials.\n")
	mustRun(t, root, "complete")

	if _, err := runCLI(args("run", "Add note editing polish"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	growth := readFile(t, filepath.Join(root, ".hyper", "growth", "state.json"))
	assertContains(t, growth, "Run go test before every runtime packet handoff")
	assertContains(t, growth, "runtime_behavior")
	goal := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0002", "goal.md"))
	assertContains(t, goal, "Carry forward learned decision: Keep local-first storage.")
	assertContains(t, goal, "Respect learned constraint: Do not add external services without credentials.")
	assertContains(t, goal, "Reuse validation pattern: Run go test before every runtime packet handoff.")
	if exists(filepath.Join(root, ".hyper", "harnesses", "generated", "harness-growth-candidate.md")) {
		t.Fatal("harness candidate should not be generated before the threshold")
	}
}

func TestGrowthGeneratesValidatorCandidateAfterRepeatedPressure(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny CLI", "Build a tiny CLI MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\ngo test ./... passed.\n\n## Readiness Evidence\n\nProduct completeness: Tiny CLI has a measurable command flow.\nCore UX: CLI smoke verified the primary command surface.\nValidation coverage: go test ./... passed and is repeatable.\n\n## Changed Files\n\ncmd/app.go\n\n## Decisions\n\nPending.\n\n## Reusable Patterns\n\nRun go test before every runtime packet handoff.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nAdd CLI persistence.\n\n## Learn Notes\n\n- Pattern: Run go test before every runtime packet handoff.\n")
	mustRun(t, root, "complete")

	if _, err := runCLI(args("run", "Add CLI persistence"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0002", "evidence.md"), "# GOAL-0002 Evidence\n\n## Validation\n\ngo test ./... passed.\n\n## Readiness Evidence\n\nProduct completeness: Tiny CLI persistence keeps the measurable command flow intact.\nCore UX: CLI smoke verifies the primary command surface.\nValidation coverage: go test ./... passed and is repeatable.\n\n## Changed Files\n\ncmd/storage.go\n\n## Decisions\n\nPending.\n\n## Reusable Patterns\n\nRun go test before every runtime packet handoff.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0002", "next.md"), "# GOAL-0002 Next\n\n## Recommended Next Goal\n\nPolish CLI output.\n\n## Learn Notes\n\n- Pattern: Run go test before every runtime packet handoff.\n")
	mustRun(t, root, "complete")

	if _, err := runCLI(args("run", "Polish CLI output"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("third run failed: %v", err)
	}

	validatorPath := filepath.Join(root, ".hyper", "validators", "generated", "validator-go-test.md")
	assertContains(t, readFile(t, validatorPath), "Status: repeated")
	assertContains(t, readFile(t, validatorPath), "Repeated validation pressure")
	assertContains(t, readFile(t, validatorPath), "## When Required")
	assertContains(t, readFile(t, validatorPath), "## How To Run")
	assertContains(t, readFile(t, validatorPath), "## Evidence Required")
	assertContains(t, readFile(t, validatorPath), "`go test")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "capabilities", "candidates", "validator", "validator-go-test.md")), "Status: repeated")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), `"state": "repeated"`)
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), `"pressure_type": "repeated_validation"`)
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0003", "goal.md")), "Reuse validation pattern: Run go test before every runtime packet handoff.")
	if exists(filepath.Join(root, ".hyper", "harnesses", "generated", "harness-growth-candidate.md")) {
		t.Fatal("single repeated validation pressure should not create a harness candidate")
	}
}

func TestGrowthUsesShortCommandCandidateName(t *testing.T) {
	name := growthCandidateName("validator", growthPressure{Signal: "validation pattern: `npm run build` passed. Vite emitted an existing warning."})
	if name != "validator-npm-run-build" {
		t.Fatalf("expected short command candidate name, got %s", name)
	}
	display := displayGrowthCandidateName(growthCandidate{
		Name:   "validator-visual-smoke-npm-run-check",
		Kind:   "validator",
		Signal: "Validation coverage: proof - Image generation and `npm run check` passed.",
	})
	if display != "validator-visual-smoke-npm-run-check" {
		t.Fatalf("expected display name to preserve validator-visual-smoke prefix, got %s", display)
	}
}

func TestMigrateRefreshesLegacyGrowthCandidates(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny web", "Build a tiny web MVP")
	db, err := openDB(root)
	if err != nil {
		t.Fatalf("db open failed: %v", err)
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		t.Fatalf("schema failed: %v", err)
	}
	insertTestMemory(t, db, "pattern", "GOAL-0001 validation pattern: `npm run build` passed.")
	insertTestMemory(t, db, "pattern", "GOAL-0002 validation pattern: `npm run build` passed.")
	legacy := growthState{
		Version: 1,
		Pressures: []growthPressure{
			{State: "repeated", PressureType: "repeated_validation", Signal: "validation pattern: `npm run build` passed."},
			{State: "repeated", PressureType: "implementation_pattern", Signal: "Error handling: proof - saved-state fallback remains unchanged in `loadState()`."},
		},
		Candidates: []growthCandidate{
			{Name: "validator-validation-pattern-npm-run-build-passed", Kind: "validator", Status: "promotable", Signal: "validation pattern: `npm run build` passed.", EvidenceCount: 2},
			{Name: "skill-error-handling-proof-saved-state-fallback-remains", Kind: "skill", Status: "repeated", Signal: "Error handling: proof - saved-state fallback remains unchanged in `loadState()`.", EvidenceCount: 2},
		},
	}
	if err := writeJSON(filepath.Join(root, ".hyper", "growth", "state.json"), legacy); err != nil {
		t.Fatalf("write legacy growth failed: %v", err)
	}

	out, err := runCLI(args("migrate"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	assertContains(t, out.Stdout, "Growth state: refreshed")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), "validator-npm-run-build")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "capabilities", "candidates", "validator", "validator-npm-run-build.md")), "Status: repeated")
}

func TestMigrateRetiresLegacyNoIssueGrowthCandidates(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny web", "Build a tiny web MVP")
	db, err := openDB(root)
	if err != nil {
		t.Fatalf("db open failed: %v", err)
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		t.Fatalf("schema failed: %v", err)
	}
	insertRawTestMemory(t, db, "failure", "GOAL-0001 learn failure: None in this run.", "durable")
	insertRawTestMemory(t, db, "failure", "GOAL-0002 blocked: Clear: implementation and validation completed for this packet.", "durable")
	legacy := growthState{
		Version: 1,
		Pressures: []growthPressure{
			{State: "repeated", PressureType: "recurring_failure", Effect: "stop_condition", Signal: "None in this run.", GoalCount: 2},
		},
		Candidates: []growthCandidate{
			{Name: "preflight-none-in-this-run", Kind: "validator", Status: "repeated", Signal: "None in this run.", EvidenceCount: 2},
		},
	}
	if err := writeJSON(filepath.Join(root, ".hyper", "growth", "state.json"), legacy); err != nil {
		t.Fatalf("write legacy growth failed: %v", err)
	}
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "candidates", "validator", "preflight-none-in-this-run.md"), "Status: repeated\nSignal: None in this run.\n")

	out, err := runCLI(args("migrate"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	assertContains(t, out.Stdout, "Growth state: refreshed")
	state := readGrowthStateIfExists(root)
	if visibleGrowthPressureCount(state.Pressures) != 0 {
		t.Fatalf("expected no visible pressures after migration, got %+v", state.Pressures)
	}
	if visibleGrowthCandidateCount(state.Candidates) != 0 {
		t.Fatalf("expected no visible candidates after migration, got %+v", state.Candidates)
	}
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "capabilities", "retired", "validator", "preflight-none-in-this-run.md")), "Status: retired")
	if exists(filepath.Join(root, ".hyper", "capabilities", "candidates", "validator", "preflight-none-in-this-run.md")) {
		t.Fatal("expected no-op preflight candidate to move out of candidates")
	}
}

func TestMigrateRefreshesNextPacketPlan(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plan.md"), "# Service Probe\n\n## Product Brief\n\nA tiny notes API.\n\n## Current Stage\n\nTiny MVP\n\n## Success Signals\n\nCreate and list one note.\n")
	mustRun(t, root, "init")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`go test ./...` passed.\n\n## Readiness Evidence\n\nProduct completeness: A tiny notes API now has a measurable create-and-list flow: `POST /notes` creates one note and `GET /notes` returns it.\nValidation coverage: `go test ./...` passed and the primary HTTP API flow is repeatable.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nDocument the API command surface.\n\n## Learn Notes\n\n- pattern: API MVPs should prove create/list with HTTP tests.\n")
	mustRun(t, root, "complete")
	writeFile(t, filepath.Join(root, ".hyper", "next-packet.md"), "# Next Packet Plan\n\nAction: advance\nCommand: hyper advance\n")

	out, err := runCLI(args("migrate"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	assertContains(t, out.Stdout, "Next packet plan: .hyper/next-packet.md (run)")
	nextPacket := readFile(t, filepath.Join(root, ".hyper", "next-packet.md"))
	assertContains(t, nextPacket, "Action: run")
	assertContains(t, nextPacket, "Command: hyper run \"Implement the smallest usable A tiny notes API core flow: the primary user flow\"")
	assertNotContains(t, nextPacket, "Command: hyper advance")
}

func TestMigrateRefreshesLegacyMemoryQualityFixture(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny CLI", "Build a tiny CLI MVP")
	db, err := openDB(root)
	if err != nil {
		t.Fatalf("db open failed: %v", err)
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		t.Fatalf("schema failed: %v", err)
	}
	for _, item := range readLegacyMemoryFixture(t, "legacy-quality-gate") {
		_, err := db.Exec(`insert into memories (project_id, kind, text, source_event_ids, confidence, quality, created_at, last_used_at, stale_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?)`, "default", item.Kind, item.Text, nil, item.Confidence, item.Quality, nowISO(), nil, nil)
		if err != nil {
			t.Fatalf("insert fixture memory failed: %v", err)
		}
	}

	out, err := runCLI(args("migrate"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	assertContains(t, out.Stdout, "Learn quality gate: refreshed 3 legacy memory quality value(s)")

	state := readGrowthStateIfExists(root)
	if visibleGrowthPressureCount(state.Pressures) != 1 {
		t.Fatalf("expected one visible pressure after quality-gate migration, got %+v", state.Pressures)
	}
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), "Run go test before every runtime packet handoff")
	assertNotContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), "None in this run")

	var blank int
	if err := db.QueryRow(`select count(*) from memories where quality is null or trim(quality) = ''`).Scan(&blank); err != nil {
		t.Fatalf("count blank qualities failed: %v", err)
	}
	if blank != 0 {
		t.Fatalf("expected migration to fill legacy memory quality values, got %d blank", blank)
	}
}

func TestGrowthIgnoresPassiveReadinessProofAsSkillCandidate(t *testing.T) {
	root := t.TempDir()
	if err := ensureProjectLayout(root); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	db, err := openDB(root)
	if err != nil {
		t.Fatalf("db open failed: %v", err)
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		t.Fatalf("schema failed: %v", err)
	}

	insertTestMemory(t, db, "pattern", "GOAL-0001 readiness evidence: Error handling: proof - corrupted saved-state fallback remains unchanged in `loadState()`.")
	insertTestMemory(t, db, "pattern", "GOAL-0002 readiness evidence: Error handling: proof - corrupted saved-state fallback remains unchanged in `loadState()`.")

	state, hyperErr := updateGrowthState(root, db)
	if hyperErr != nil {
		t.Fatalf("growth failed: %v", hyperErr)
	}
	if len(state.Pressures) != 0 {
		t.Fatalf("expected passive unchanged proof to be ignored, got %+v", state.Pressures)
	}
	if len(state.Candidates) != 0 {
		t.Fatalf("expected no skill candidate for passive unchanged proof, got %+v", state.Candidates)
	}
}

func TestGrowthIgnoresStageAdvancementProtocolNoise(t *testing.T) {
	pressures := deriveGrowthPressures([]memoryRecord{
		{Kind: "decision", Text: "GOAL-0001 learn decision: Preserve current stage in `plan.md`; stage advancement remains a recommendation pending user acceptance.", Confidence: 0.75, Quality: "durable"},
		{Kind: "constraint", Text: "GOAL-0002 learn constraint: Do not edit `plan.md Current Stage` until the user accepts stage advancement.", Confidence: 0.75, Quality: "durable"},
		{Kind: "decision", Text: "GOAL-0003 decisions: Do not edit `plan.md Current Stage` in this packet; stage advancement is a recommendation pending user acceptance.", Confidence: 0.75, Quality: "durable"},
	})
	if len(pressures) != 0 {
		t.Fatalf("expected stage advancement protocol notes to stay out of growth pressure, got %+v", pressures)
	}
	memories := appendMemoryIfUseful(nil, "decision", "GOAL-0001 decisions: Preserve current stage in `plan.md`; stage advancement remains a recommendation pending user acceptance.", 0.75)
	if len(memories) != 0 {
		t.Fatalf("expected protocol note to stay out of memory, got %+v", memories)
	}
}

func TestGrowthTreatsKnownGapFailureAsImplementationPressure(t *testing.T) {
	pressures := deriveGrowthPressures([]memoryRecord{
		{Kind: "failure", Text: "GOAL-0003 learn failure: Malformed `.release_notes.json` recovery is not handled yet.", Confidence: 0.8, Quality: "durable"},
	})
	if len(pressures) != 1 {
		t.Fatalf("expected one implementation gap pressure, got %+v", pressures)
	}
	if pressures[0].PressureType != "implementation_gap" || pressures[0].Effect != "implementation" {
		t.Fatalf("expected implementation gap pressure, got %+v", pressures[0])
	}
	behavior := growthBehaviorFromPressures(pressures)
	if len(behavior.StopConditions) != 0 {
		t.Fatalf("known implementation gap should not become a stop condition, got %+v", behavior.StopConditions)
	}
}

func TestGrowthTreatsRemainingGapFailureAsImplementationPressure(t *testing.T) {
	pressures := deriveGrowthPressures([]memoryRecord{
		{Kind: "failure", Text: "GOAL-0003 learn failure: Operations docs and reference benchmark remain incomplete.", Confidence: 0.8, Quality: "durable"},
	})
	if len(pressures) != 1 {
		t.Fatalf("expected one implementation pressure, got %+v", pressures)
	}
	if pressures[0].PressureType != "implementation_gap" || pressures[0].Effect != "implementation" {
		t.Fatalf("expected remaining gap to become implementation pressure, got %+v", pressures[0])
	}
	behavior := growthBehaviorFromPressures(pressures)
	if len(behavior.StopConditions) != 0 {
		t.Fatalf("remaining gap should not become a stop condition, got %+v", behavior.StopConditions)
	}
}

func TestGrowthGroupsRepeatedValidationByCommand(t *testing.T) {
	pressures := deriveGrowthPressures([]memoryRecord{
		{Kind: "pattern", Text: "GOAL-0001 reusable patterns: Use `./check.sh` as the narrow local smoke command for the add/list flow.", Confidence: 0.75, Quality: "durable"},
		{Kind: "pattern", Text: "GOAL-0002 reusable patterns: Use `./check.sh` as the repeated validation path for add/list plus CLI edge states.", Confidence: 0.75, Quality: "durable"},
	})
	if len(pressures) != 1 {
		t.Fatalf("expected same-command validation pressure to merge, got %+v", pressures)
	}
	if pressures[0].State != "repeated" || pressures[0].GoalCount != 2 {
		t.Fatalf("expected repeated validation pressure across two goals, got %+v", pressures[0])
	}
	if pressures[0].PressureType != "repeated_validation" {
		t.Fatalf("expected repeated validation pressure, got %+v", pressures[0])
	}
}

func TestErrorHandlingEvidenceAcceptsCLIInvalidCommandStates(t *testing.T) {
	covered, _ := readinessEvidenceQuality("error_handling", "Missing argument and unknown command states are handled and verified.")
	if !covered {
		t.Fatal("CLI missing-argument and unknown-command evidence should cover error handling")
	}
}

func TestSimilarContextIgnoresProtocolNoiseMemories(t *testing.T) {
	root := t.TempDir()
	if err := ensureProjectLayout(root); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	db, err := openDB(root)
	if err != nil {
		t.Fatalf("db open failed: %v", err)
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		t.Fatalf("schema failed: %v", err)
	}
	insertRawTestMemory(t, db, "decision", "GOAL-0001 learn decision: Preserve current stage in `plan.md`; stage advancement remains a recommendation pending user acceptance.", "durable")

	similar, hyperErr := findSimilarContext(db, "stage advancement plan current stage", 5)
	if hyperErr != nil {
		t.Fatalf("similar context failed: %v", hyperErr)
	}
	if len(similar) != 0 {
		t.Fatalf("expected protocol noise memory to stay out of similar context, got %+v", similar)
	}
}

func TestGrowthClustersSignalsAndPromotesLifecycle(t *testing.T) {
	root := t.TempDir()
	if err := ensureProjectLayout(root); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	db, err := openDB(root)
	if err != nil {
		t.Fatalf("db open failed: %v", err)
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		t.Fatalf("schema failed: %v", err)
	}

	insertTestMemory(t, db, "pattern", "GOAL-0001 learn pattern: Run go test before every runtime packet handoff.")
	insertTestMemory(t, db, "pattern", "GOAL-0002 learn pattern: Run go test before each runtime handoff.")
	state, hyperErr := updateGrowthState(root, db)
	if hyperErr != nil {
		t.Fatalf("growth failed: %v", hyperErr)
	}
	if len(state.Pressures) != 1 {
		t.Fatalf("expected one clustered pressure, got %+v", state.Pressures)
	}
	if state.Pressures[0].GoalCount != 2 {
		t.Fatalf("expected two goal sources, got %+v", state.Pressures[0])
	}
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "capabilities", "candidates", "validator", "validator-go-test.md")), "Status: repeated")

	insertTestMemory(t, db, "pattern", "GOAL-0003 learn pattern: Run go test before every runtime handoff.")
	state, hyperErr = updateGrowthState(root, db)
	if hyperErr != nil {
		t.Fatalf("growth failed: %v", hyperErr)
	}
	if state.Candidates[0].Status != "promotable" {
		t.Fatalf("expected promotable candidate, got %+v", state.Candidates[0])
	}
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "capabilities", "candidates", "validator", "validator-go-test.md")), "Status: promotable")

	insertTestMemory(t, db, "pattern", "GOAL-0004 learn pattern: Run go test before each runtime packet handoff.")
	state, hyperErr = updateGrowthState(root, db)
	if hyperErr != nil {
		t.Fatalf("growth failed: %v", hyperErr)
	}
	if state.Candidates[0].Status != "active" {
		t.Fatalf("expected active candidate, got %+v", state.Candidates[0])
	}
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-go-test.md")), "Status: active")
	if exists(filepath.Join(root, ".hyper", "capabilities", "candidates", "validator", "validator-go-test.md")) {
		t.Fatal("active validator should move out of candidates")
	}

	if _, err := db.Exec(`update memories set stale_at = ? where kind = ?`, nowISO(), "pattern"); err != nil {
		t.Fatalf("stale update failed: %v", err)
	}
	state, hyperErr = updateGrowthState(root, db)
	if hyperErr != nil {
		t.Fatalf("growth failed: %v", hyperErr)
	}
	if len(state.Candidates) != 1 || state.Candidates[0].Status != "retired" {
		t.Fatalf("expected retired candidate, got %+v", state.Candidates)
	}
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "capabilities", "retired", "validator", "validator-go-test.md")), "Status: retired")
	if exists(filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-go-test.md")) {
		t.Fatal("retired validator should no longer remain active")
	}
	if exists(filepath.Join(root, ".hyper", "validators", "generated", "validator-go-test.md")) {
		t.Fatal("retired validator should no longer remain in generated validators")
	}
	assertNotContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), "Required active validator validator-go-test")
}

func TestHarnessCandidateEvidenceCountUsesStablePressureCount(t *testing.T) {
	pressure := aggregateHarnessPressure([]growthPressure{
		{Effect: "validation", GoalCount: 2, Sources: []string{"GOAL-0001", "GOAL-0002"}},
		{Effect: "implementation", GoalCount: 2, Sources: []string{"GOAL-0003", "GOAL-0004"}},
		{Effect: "work_boundary", GoalCount: 2, Sources: []string{"GOAL-0005", "GOAL-0006"}},
	})
	candidate := harnessCandidateForPressure(pressure)
	if candidate.Status != "repeated" {
		t.Fatalf("expected repeated harness candidate, got %+v", candidate)
	}
	if candidate.EvidenceCount != 3 {
		t.Fatalf("expected harness evidence count to use stable pressure count, got %+v", candidate)
	}
}

func TestHarnessCandidateNeedsMultipleNonValidationStructures(t *testing.T) {
	pressures := []growthPressure{
		{Effect: "validation", GoalCount: 2, Sources: []string{"GOAL-0001", "GOAL-0002"}},
		{Effect: "validation", GoalCount: 2, Sources: []string{"GOAL-0001", "GOAL-0002"}},
		{Effect: "validation", GoalCount: 2, Sources: []string{"GOAL-0001", "GOAL-0002"}},
		{Effect: "work_boundary", GoalCount: 2, Sources: []string{"GOAL-0001", "GOAL-0002"}},
	}
	if harnessPressureReady(pressures) {
		t.Fatal("single repeated decision plus repeated validation must not create a harness candidate")
	}
}

func TestHarnessCandidateNeedsImplementationAndBoundaryPressure(t *testing.T) {
	pressures := []growthPressure{
		{Effect: "validation", GoalCount: 4, Sources: []string{"GOAL-0001", "GOAL-0002", "GOAL-0003", "GOAL-0004"}},
		{Effect: "work_boundary", GoalCount: 4, Sources: []string{"GOAL-0001", "GOAL-0002", "GOAL-0003", "GOAL-0004"}},
		{Effect: "work_boundary", GoalCount: 4, Sources: []string{"GOAL-0001", "GOAL-0002", "GOAL-0003", "GOAL-0004"}},
	}
	if harnessPressureReady(pressures) {
		t.Fatal("repeated decisions plus validation must not create a harness without implementation pressure")
	}
}

func TestGrowthBehaviorDedupesValidationSignalsByCommand(t *testing.T) {
	behavior := growthBehaviorFromPressures([]growthPressure{
		{
			Effect:          "validation",
			Signal:          "Use `./check.sh` as the narrow local smoke command for the add/list flow.",
			CanonicalSignal: "add check command flow list local narrow sh smoke use",
		},
		{
			Effect:          "validation",
			Signal:          "Validation coverage: `./check.sh` passed and covers add, list, read-back, and `go test ./...`.",
			CanonicalSignal: "add back check coverage covers go list passed read sh test validation",
		},
	})
	if len(behavior.ValidationSignals) != 1 {
		t.Fatalf("expected same-command validation signals to dedupe, got %+v", behavior.ValidationSignals)
	}
	assertContains(t, behavior.ValidationSignals[0], "./check.sh")
}

func TestGrowthBehaviorDedupesNoHarnessBoundaryPressure(t *testing.T) {
	behavior := growthBehaviorFromPressures([]growthPressure{
		{Kind: "constraint", Effect: "work_boundary", Signal: "Do not create harnesses until repeated evidence shows the project needs one.", CanonicalSignal: "create do evidence harnesses needs not one project repeated shows until"},
		{Kind: "constraint", Effect: "work_boundary", Signal: "Do not add a harness while active validator promotion can cover the repeated smoke command.", CanonicalSignal: "active add command cover harness not promotion repeated smoke validator while"},
		{Kind: "constraint", Effect: "work_boundary", Signal: "Do not create a harness while one local smoke command still covers the required proof.", CanonicalSignal: "command covers create do harness local not one proof required smoke still while"},
		{Kind: "decision", Effect: "work_boundary", Signal: "Keep edge-state checks inside the same narrow smoke command instead of adding a separate harness.", CanonicalSignal: "adding checks command edge harness inside instead keep narrow same separate smoke state"},
	})
	if len(behavior.WorkBoundary) != 1 {
		t.Fatalf("expected overlapping no-harness constraints to dedupe, got %+v", behavior.WorkBoundary)
	}
	assertContains(t, behavior.WorkBoundary[0], "harness")
}

func TestActiveValidatorReplacesSameCommandValidationSignal(t *testing.T) {
	root := t.TempDir()
	if err := ensureProjectLayout(root); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-check-sh.md"), "# validator-check-sh\n\nStatus: active\nKind: validator\nSignal: Use `./check.sh` as the repeated validation path.\n")
	behavior, hyperErr := growthBehaviorWithActiveCapabilities(root, []growthPressure{
		{Effect: "validation", Signal: "Use `./check.sh` as the narrow local smoke command.", CanonicalSignal: "check command local narrow sh smoke use"},
	})
	if hyperErr != nil {
		t.Fatalf("growth behavior failed: %v", hyperErr)
	}
	if len(behavior.ValidationSignals) != 1 {
		t.Fatalf("expected active validator to replace same-command reuse signal, got %+v", behavior.ValidationSignals)
	}
	assertContains(t, behavior.ValidationSignals[0], "Required active validator validator-check-sh")
}

func TestDuplicateCommandCandidatesKeepStrongestLifecycle(t *testing.T) {
	root := t.TempDir()
	if err := ensureProjectLayout(root); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	pressures := []growthPressure{
		{
			Kind:         "pattern",
			PressureType: "repeated_validation",
			Signal:       "validation pattern: `./check.sh` passed.",
			Effect:       "validation",
			State:        "repeated",
			GoalCount:    4,
			MemoryCount:  4,
			Sources:      []string{"GOAL-0001", "GOAL-0002", "GOAL-0003", "GOAL-0004"},
		},
		{
			Kind:         "pattern",
			PressureType: "repeated_validation",
			Signal:       "`./check.sh` passed as active validator smoke.",
			Effect:       "validation",
			State:        "repeated",
			GoalCount:    2,
			MemoryCount:  2,
			Sources:      []string{"GOAL-0005", "GOAL-0006"},
		},
	}
	candidates, hyperErr := materializeGrowthCandidates(root, pressures, growthState{})
	if hyperErr != nil {
		t.Fatalf("materialize candidates failed: %v", hyperErr)
	}
	if len(candidates) != 1 {
		t.Fatalf("expected one deduped validator candidate, got %+v", candidates)
	}
	if candidates[0].Status != "active" {
		t.Fatalf("expected strongest active validator to win, got %+v", candidates[0])
	}
	if !exists(filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-check-sh.md")) {
		t.Fatal("active validator file should exist")
	}
	if exists(filepath.Join(root, ".hyper", "capabilities", "candidates", "validator", "validator-check-sh.md")) {
		t.Fatal("weaker duplicate validator candidate should not overwrite active validator")
	}
}

func TestHarnessCandidateRequiresEnoughSourceGoalsForActivation(t *testing.T) {
	twoGoalPressure := aggregateHarnessPressure([]growthPressure{
		{Effect: "validation", GoalCount: 2, Sources: []string{"GOAL-0003", "GOAL-0004"}},
		{Effect: "validation", GoalCount: 2, Sources: []string{"GOAL-0003", "GOAL-0004"}},
		{Effect: "validation", GoalCount: 2, Sources: []string{"GOAL-0003", "GOAL-0004"}},
		{Effect: "implementation", GoalCount: 2, Sources: []string{"GOAL-0003", "GOAL-0004"}},
		{Effect: "work_boundary", GoalCount: 2, Sources: []string{"GOAL-0003", "GOAL-0004"}},
	})
	candidate := harnessCandidateForPressure(twoGoalPressure)
	if candidate.Status != "repeated" {
		t.Fatalf("harness must not become active from many pressures in only two packets, got %+v", candidate)
	}
	assertContains(t, candidate.LifecyclePath, filepath.Join(".hyper", "capabilities", "candidates", "harness"))

	fiveGoalPressure := aggregateHarnessPressure([]growthPressure{
		{Effect: "validation", GoalCount: 5, Sources: []string{"GOAL-0001", "GOAL-0002", "GOAL-0003", "GOAL-0004", "GOAL-0005"}},
		{Effect: "validation", GoalCount: 5, Sources: []string{"GOAL-0001", "GOAL-0002", "GOAL-0003", "GOAL-0004", "GOAL-0005"}},
		{Effect: "implementation", GoalCount: 5, Sources: []string{"GOAL-0001", "GOAL-0002", "GOAL-0003", "GOAL-0004", "GOAL-0005"}},
		{Effect: "work_boundary", GoalCount: 5, Sources: []string{"GOAL-0001", "GOAL-0002", "GOAL-0003", "GOAL-0004", "GOAL-0005"}},
		{Effect: "work_boundary", GoalCount: 5, Sources: []string{"GOAL-0001", "GOAL-0002", "GOAL-0003", "GOAL-0004", "GOAL-0005"}},
	})
	candidate = harnessCandidateForPressure(fiveGoalPressure)
	if candidate.Status != "active" {
		t.Fatalf("expected active harness only after enough stable pressures and source goals, got %+v", candidate)
	}
}

func TestReadinessEvidenceDoesNotBecomeValidatorExceptValidationCoverage(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, hyperDir), 0755); err != nil {
		t.Fatal(err)
	}
	db, hyperErr := openDB(root)
	if hyperErr != nil {
		t.Fatalf("open db failed: %v", hyperErr)
	}
	defer db.Close()
	if hyperErr := ensureSchema(db); hyperErr != nil {
		t.Fatalf("schema failed: %v", hyperErr)
	}
	for _, goal := range []string{"GOAL-0001", "GOAL-0002"} {
		insertTestMemory(t, db, "pattern", goal+" readiness evidence: Security baseline: Local-only file storage is explicit, no network or telemetry exists, and sensitive words are rejected by the CLI smoke command.")
		insertTestMemory(t, db, "pattern", goal+" readiness evidence: Reference benchmark: Category: Local file-backed utility CLI; References: Git, SQLite CLI, Taskfile, Make; Baseline expectations: local commands are documented and repeatable command output exists.")
		insertTestMemory(t, db, "pattern", goal+" readiness evidence: Validation coverage: `./check.sh` passed and is repeatable.")
	}
	state, hyperErr := updateGrowthState(root, db)
	if hyperErr != nil {
		t.Fatalf("growth failed: %v", hyperErr)
	}
	for _, candidate := range state.Candidates {
		if strings.Contains(candidate.Name, "security-baseline") || strings.Contains(candidate.Name, "reference-benchmark") {
			t.Fatalf("readiness evidence for %s should not become a validator candidate: %+v", candidate.Name, state.Candidates)
		}
	}
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "capabilities", "candidates", "validator", "validator-check-sh.md")), "Status: repeated")
}

func TestCommandHandoffPatternClassifiesAsValidation(t *testing.T) {
	pressureType, effect := growthClassification("pattern", "Pattern: Run `./check.sh` before every service-quality handoff.")
	if pressureType != "repeated_validation" || effect != "validation" {
		t.Fatalf("expected command handoff pattern to be validation pressure, got %s/%s", pressureType, effect)
	}
}

func TestMemorySignalStripsPressureSignalLabels(t *testing.T) {
	got := memorySignal("GOAL-0002 pressure signals: repeated_validation: `./check.sh` passed again as the handoff smoke.")
	if got != "`./check.sh` passed again as the handoff smoke." {
		t.Fatalf("expected clean pressure signal, got %q", got)
	}

	got = memorySignal("GOAL-0003 pressure signals: service_quality_boundary: Keep security rejection and export proof in the handoff.")
	if got != "Keep security rejection and export proof in the handoff." {
		t.Fatalf("expected clean service boundary signal, got %q", got)
	}
}

func TestBacktickCodeSymbolDoesNotClassifyAsValidationCommand(t *testing.T) {
	pressureType, effect := growthClassification("pattern", "Pattern: Check `loadState()` fallback before rendering.")
	if pressureType != "implementation_pattern" || effect != "implementation" {
		t.Fatalf("expected code-symbol pattern to remain implementation pressure, got %s/%s", pressureType, effect)
	}
}

func TestDocumentationPatternDoesNotBecomeValidationSignal(t *testing.T) {
	pressureType, effect := growthClassification("pattern", "Use README setup/build/rollback sections as the operator handoff for this local CLI.")
	if pressureType != "implementation_pattern" || effect != "implementation" {
		t.Fatalf("expected documentation handoff pattern to remain implementation pressure, got %s/%s", pressureType, effect)
	}
}

func TestReferenceBenchmarkPatternDoesNotBecomeValidationSignal(t *testing.T) {
	pressureType, effect := growthClassification("pattern", "Use reference benchmark evidence to prevent stage advancement on validation alone.")
	if pressureType != "implementation_pattern" || effect != "implementation" {
		t.Fatalf("expected reference benchmark pattern to remain implementation pressure, got %s/%s", pressureType, effect)
	}
}

func TestActiveValidatorBecomesRequiredValidationSignal(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny CLI", "Build a tiny CLI MVP")
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-run-go-test.md"), "# validator-run-go-test\n\nStatus: active\nKind: validator\n\n## Pressure\n\n- Signal: Run go test ./... before handoff.\n")
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "candidates", "validator", "validator-candidate-only.md"), "# validator-candidate-only\n\nStatus: promotable\nKind: validator\n\n## Pressure\n\n- Signal: Run candidate-only smoke check.\n")

	if _, err := runCLI(args("run", "Add CLI persistence"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	goal := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "goal.md"))
	assertContains(t, goal, "## Active Capabilities")
	assertContains(t, goal, "Required active validator validator-run-go-test")
	assertContains(t, goal, "Required active validator validator-run-go-test: Run go test ./... before handoff.")
	assertNotContains(t, goal, "candidate-only smoke check")
	evidence := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"))
	assertContains(t, evidence, "## Active Capability Evidence")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), "Required active validator validator-run-go-test")
}

func TestActiveCapabilityFilesBecomeGrowthCandidates(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny CLI", "Build a tiny CLI MVP")
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-go-test.md"), "# validator-go-test\n\nStatus: active\nKind: validator\nSignal: Run go test ./... before completing packets.\n")
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "harness", "harness-growth-candidate.md"), "# harness-growth-candidate\n\nStatus: active\nKind: harness\n\n## Required Behavior\n\nRun the project-specific handoff harness before completing packets.\n")
	db, err := openDB(root)
	if err != nil {
		t.Fatalf("db open failed: %v", err)
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		t.Fatalf("schema failed: %v", err)
	}

	state, hyperErr := updateGrowthState(root, db)
	if hyperErr != nil {
		t.Fatalf("growth failed: %v", hyperErr)
	}
	if activeStructureCount(state.Candidates) != 2 {
		t.Fatalf("expected two active structures from active capability files, got %+v", state.Candidates)
	}
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), `"name": "validator-go-test"`)
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), `"name": "harness-growth-candidate"`)
}

func TestGrowthStatusOverlayPromotesManualActiveCapabilityWithoutDuplicate(t *testing.T) {
	root := t.TempDir()
	if err := ensureProjectLayout(root); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-go-test.md"), "# validator-go-test\n\nStatus: active\nKind: validator\nSignal: Run go test ./... before completing packets.\n")
	growth := growthState{
		Pressures: []growthPressure{{State: "repeated", PressureType: "repeated_validation", Effect: "validation", Signal: "Run go test before handoff.", GoalCount: 2}},
		Candidates: []growthCandidate{
			{Kind: "validator", Name: "validator-go-test", Status: "repeated", Signal: "Run go test before handoff.", LifecyclePath: filepath.Join(hyperDir, "capabilities", "candidates", "validator", "validator-go-test.md")},
		},
	}

	overlaid := growthStateWithActiveCapabilityOverlay(root, growth)
	if len(overlaid.Candidates) != 1 {
		t.Fatalf("expected one candidate after overlay, got %+v", overlaid.Candidates)
	}
	if overlaid.Candidates[0].Status != "active" {
		t.Fatalf("expected active candidate after overlay, got %+v", overlaid.Candidates[0])
	}
	if activeStructureCount(overlaid.Candidates) != 1 || overlaid.PressureLedger.ActiveStructures != 1 {
		t.Fatalf("expected active counts to refresh, got %+v", overlaid.PressureLedger)
	}
}

func TestStatusReflectsManualActiveCapabilityBeforeMigrate(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plan.md"), strings.Join([]string{
		"# Product Plan",
		"",
		"## Product",
		"",
		"Local Build Relay",
		"",
		"## Target Users",
		"",
		"Developers",
		"",
		"## MVP",
		"",
		"Run one repeatable handoff command.",
		"",
		"## Current Stage",
		"",
		"Service Quality",
		"",
		"## Build Style",
		"",
		"Go CLI",
		"",
		"## Success Criteria",
		"",
		"Every packet proves validation, release, docs, maintainability, and benchmark baseline.",
	}, "\n"))
	if _, err := runCLI(args("run", "Prepare sustained quality"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), strings.Join([]string{
		"# GOAL-0001 Evidence",
		"",
		"## Validation",
		"",
		"`go test ./...` passed and the CLI smoke command is repeatable.",
		"",
		"## Readiness Evidence",
		"",
		"Validation coverage: `go test ./...` passed and the CLI smoke command is repeatable.",
		"Security baseline: Privacy boundary verified, no cloud sync, no telemetry, no token storage, no secrets, and local-only data handling is explicit.",
		"Deployment readiness: Built the CLI binary and ran the smoke command outside the development command.",
		"Operations and docs: README handoff notes cover setup, rollback, recovery, and the smoke command.",
		"Maintainability: Table-driven validation helper keeps command checks repeatable without hidden local context.",
		"",
		"## Reference Benchmark Evidence",
		"",
		"- Category: Local developer handoff CLI.",
		"- References: GitHub CLI, Taskfile, Make.",
		"- Baseline expectations: documented command, repeatable output, rollback notes, no hidden credentials.",
		"- Current comparison: below baseline = none; meets baseline = command/test/docs/rollback; above baseline = packet evidence loop.",
		"- Below-baseline gaps: No critical below-baseline gap.",
		"- Above-baseline strength: packet evidence loop.",
		"- Decision: Service Quality proof can continue.",
		"",
		"## Blocker",
		"",
		"None blocking.",
	}, "\n"))
	if activeStructureCount(readGrowthStateIfExists(root).Candidates) != 0 {
		t.Fatal("stored growth should not know about the manual active capability yet")
	}
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-go-test.md"), "# validator-go-test\n\nStatus: active\nKind: validator\nSignal: Run go test ./... before completing packets.\n")

	short, err := runCLI(args("status", "--short"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	assertContains(t, short.Stdout, "Gate: Service Quality -> Sustained Service Quality (ready)")
	assertContains(t, short.Stdout, "Next: update .hyper/goals/GOAL-0001/evidence.md and next.md, then run `hyper complete`")

	full, err := runCLI(args("status"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	assertContains(t, full.Stdout, "Pressure ledger: 0 pressure(s), 1 candidate(s), 1 active structure(s).")
	assertContains(t, full.Stdout, "Covered axes: Product completeness, Validation coverage, Security baseline, Deployment readiness, Operations and docs, Maintainability, Reference benchmark, Sustained quality")
	if activeStructureCount(readGrowthStateIfExists(root).Candidates) != 0 {
		t.Fatal("status overlay should not mutate stored growth state")
	}

	doctor, err := runCLI(args("doctor"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("doctor failed: %v", err)
	}
	assertContains(t, doctor.Stdout, "[WARN] Growth migration: active capability files are not reflected in stored growth state; run `hyper migrate`")
	assertContains(t, doctor.Stdout, "[WARN] Readiness state:")
	assertContains(t, doctor.Stdout, "Run `hyper migrate`.")
}

func TestReadinessPressureSelectsStageGateGap(t *testing.T) {
	root := t.TempDir()
	mustRun(t, root, "init")
	writeFile(t, filepath.Join(root, "plan.md"), "# Product Plan\n\n## Product\n\nTiny CRM\n\n## Target Users\n\nSolo sellers\n\n## MVP\n\nAdd and revisit customer notes.\n\n## Current Stage\n\nUsable MVP\n\n## Build Style\n\nWeb app\n\n## Non-goals\n\nTeam collaboration\n\n## Constraints\n\nLocal first\n\n## Success Criteria\n\nPrimary customer notes flow works without manual data edits.\n\n## Current Focus\n\nImprove customer notes.\n")

	if _, err := runCLI(args("run"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	readiness := readFile(t, filepath.Join(root, ".hyper", "readiness", "state.json"))
	assertContains(t, readiness, `"current_stage": "Usable MVP"`)
	assertContains(t, readiness, `"next_stage": "Beta"`)
	assertContains(t, readiness, `"axis": "core_ux"`)
	goal := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "goal.md"))
	assertContains(t, goal, "Current gate: Usable MVP -> Beta")
	assertContains(t, goal, "Next readiness pressure: Core UX")
	assertContains(t, goal, "Implement the smallest usable Tiny CRM core flow")
	assertContains(t, goal, "Capture readiness evidence for Core UX")
}

func TestReadinessEvidenceProgressesSelectedAxis(t *testing.T) {
	root := t.TempDir()
	mustRun(t, root, "init")
	writeFile(t, filepath.Join(root, "plan.md"), "# Product Plan\n\n## Product\n\nTiny CRM\n\n## Target Users\n\nSolo sellers\n\n## MVP\n\nAdd and revisit customer notes.\n\n## Current Stage\n\nUsable MVP\n\n## Build Style\n\nWeb app\n\n## Non-goals\n\nTeam collaboration\n\n## Constraints\n\nLocal first\n\n## Success Criteria\n\nPrimary customer notes flow works without manual data edits.\n\n## Current Focus\n\nImprove customer notes.\n")

	if _, err := runCLI(args("run"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\nBrowser smoke passed.\n\n## Readiness Evidence\n\nCore UX: Browser smoke verified add and revisit customer notes flow.\nData persistence: Customer notes persist across reload using local storage.\n\n## Changed Files\n\nsrc/App.tsx\n\n## Decisions\n\nKeep storage local-first.\n\n## Reusable Patterns\n\nPending.\n\n## Blocker\n\nPending.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nHandle empty and failure states.\n\n## Learn Notes\n\n- Pattern: Record readiness evidence with an axis label.\n")
	mustRun(t, root, "complete")

	if _, err := runCLI(args("run"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	state := readReadinessStateIfExists(root)
	if got := readinessDimensionMap(state.Dimensions)["persistence"].Status; got != "covered" {
		t.Fatalf("expected persistence covered, got %s", got)
	}
	if state.NextPressure.Axis != "error_handling" {
		t.Fatalf("expected next pressure to move to error_handling, got %+v", state.NextPressure)
	}
	goal := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0002", "goal.md"))
	assertContains(t, goal, "Next readiness pressure: Error handling")
	assertNotContains(t, goal, "Next readiness pressure: Data persistence")
	assertContains(t, goal, "Handle empty, failure, and edge states")
}

func TestReadinessEvidenceQualityRules(t *testing.T) {
	defs := readinessDimensionDefs()
	weak, ok := parseReadinessEvidenceLine("GOAL-0001", "Validation coverage: tested.", defs)
	if !ok {
		t.Fatal("expected weak validation evidence to parse")
	}
	if weak.Status != "emerging" {
		t.Fatalf("expected weak validation evidence to be emerging, got %+v", weak)
	}
	strong, ok := parseReadinessEvidenceLine("GOAL-0001", "Validation coverage: `go test ./...` passed and is repeatable.", defs)
	if !ok {
		t.Fatal("expected strong validation evidence to parse")
	}
	if strong.Status != "covered" {
		t.Fatalf("expected strong validation evidence to be covered, got %+v", strong)
	}
	shellSmoke, ok := parseReadinessEvidenceLine("GOAL-0001", "Validation coverage: The shell smoke command proved the add/list/handle flow end to end with real command output.", defs)
	if !ok {
		t.Fatal("expected shell smoke validation evidence to parse")
	}
	if shellSmoke.Status != "covered" {
		t.Fatalf("expected shell smoke validation evidence to be covered, got %+v", shellSmoke)
	}
	weakUX, ok := parseReadinessEvidenceLine("GOAL-0001", "Core UX: flow exists.", defs)
	if !ok {
		t.Fatal("expected weak UX evidence to parse")
	}
	if weakUX.Status != "emerging" {
		t.Fatalf("expected weak UX evidence to be emerging, got %+v", weakUX)
	}
	strongUX, ok := parseReadinessEvidenceLine("GOAL-0001", "Core UX: Browser smoke passed for create and complete flow.", defs)
	if !ok {
		t.Fatal("expected strong UX evidence to parse")
	}
	if strongUX.Status != "covered" {
		t.Fatalf("expected strong UX evidence to be covered, got %+v", strongUX)
	}
	apiUX, ok := parseReadinessEvidenceLine("GOAL-0001", "Core UX: HTTP API test passed for create and list endpoints.", defs)
	if !ok {
		t.Fatal("expected API UX evidence to parse")
	}
	if apiUX.Status != "covered" {
		t.Fatalf("expected API UX evidence to be covered, got %+v", apiUX)
	}
	apiProduct, ok := parseReadinessEvidenceLine("GOAL-0001", "Product completeness: A tiny notes API now has a measurable create-and-list flow: `POST /notes` creates one note and `GET /notes` returns it.", defs)
	if !ok {
		t.Fatal("expected API product evidence to parse")
	}
	if apiProduct.Status != "covered" {
		t.Fatalf("expected API product evidence to be covered, got %+v", apiProduct)
	}
	inferred := inferReadinessEvidenceFromValidationLine("GOAL-0001", "`npm run check` passed.")
	if len(inferred) != 1 || inferred[0].Axis != "validation_coverage" || inferred[0].Status != "covered" {
		t.Fatalf("expected validation command to infer covered validation evidence, got %+v", inferred)
	}
	inferred = inferReadinessEvidenceFromValidationLine("GOAL-0001", "Browser validation at mobile viewport passed the core flow.")
	axes := map[string]string{}
	for _, record := range inferred {
		axes[record.Axis] = record.Status
	}
	if axes["validation_coverage"] != "covered" || axes["core_ux"] != "covered" {
		t.Fatalf("expected browser validation flow to infer validation and core UX coverage, got %+v", inferred)
	}
	weakDeploy, ok := parseReadinessEvidenceLine("GOAL-0001", "Deployment readiness: URL documented.", defs)
	if !ok {
		t.Fatal("expected weak deployment evidence to parse")
	}
	if weakDeploy.Status != "emerging" {
		t.Fatalf("expected weak deployment evidence to be emerging, got %+v", weakDeploy)
	}
	strongDeploy, ok := parseReadinessEvidenceLine("GOAL-0001", "Deployment readiness: hosted URL https://example.com verified available.", defs)
	if !ok {
		t.Fatal("expected strong deployment evidence to parse")
	}
	if strongDeploy.Status != "covered" {
		t.Fatalf("expected strong deployment evidence to be covered, got %+v", strongDeploy)
	}
	cliDeploy, ok := parseReadinessEvidenceLine("GOAL-0001", "Deployment readiness: Built the CLI binary and ran the smoke command outside the development command.", defs)
	if !ok {
		t.Fatal("expected CLI deployment evidence to parse")
	}
	if cliDeploy.Status != "covered" {
		t.Fatalf("expected CLI deployment evidence to be covered, got %+v", cliDeploy)
	}
	exportDeploy, ok := parseReadinessEvidenceLine("GOAL-0001", "Deployment readiness: `./check.sh` verifies export artifact creation outside the normal add/list path.", defs)
	if !ok {
		t.Fatal("expected export deployment evidence to parse")
	}
	if exportDeploy.Status != "covered" {
		t.Fatalf("expected export deployment evidence to be covered, got %+v", exportDeploy)
	}
	weakReference, ok := parseReadinessEvidenceLine("GOAL-0001", "Reference benchmark: Compared against three comparable project-growth CLIs; category baseline is fine and above-baseline strength exists.", defs)
	if !ok {
		t.Fatal("expected weak reference benchmark evidence to parse")
	}
	if weakReference.Status != "emerging" {
		t.Fatalf("expected weak reference benchmark evidence to be emerging, got %+v", weakReference)
	}
	staticDeploy, ok := parseReadinessEvidenceLine("GOAL-0001", "Deployment readiness: proof. Release/build artifacts are created at `dist/llog-beta-demo/index.html` and `dist/llog-beta-demo.zip` outside the development path. Validation proved the release artifacts through direct `file://` execution, isolated artifact server URL `http://127.0.0.1:4201/?artifact=1`, extracted zip release URL, artifact parity, and mobile Playwright smoke with realistic data.", defs)
	if !ok {
		t.Fatal("expected static artifact deployment evidence to parse")
	}
	if staticDeploy.Status != "covered" {
		t.Fatalf("expected static artifact deployment evidence to be covered, got %+v", staticDeploy)
	}
	opsDocs, ok := parseReadinessEvidenceLine("GOAL-0001", "Operations and docs: `demo-release.md` documents artifact creation, direct file run, static server run, smoke path, rollback, and stop conditions.", defs)
	if !ok {
		t.Fatal("expected operations docs evidence to parse")
	}
	if opsDocs.Status != "covered" {
		t.Fatalf("expected operations docs evidence to be covered, got %+v", opsDocs)
	}
	opsNotes, ok := parseReadinessEvidenceLine("GOAL-0001", "Operations and docs: README handoff notes cover setup, rollback, recovery, and the smoke command.", defs)
	if !ok {
		t.Fatal("expected operations notes evidence to parse")
	}
	if opsNotes.Status != "covered" {
		t.Fatalf("expected operations notes evidence to be covered, got %+v", opsNotes)
	}
	maintainabilityHandoff, ok := parseReadinessEvidenceLine("GOAL-0001", "Maintainability: `DEVELOPMENT.md` documents the required `./check.sh` service-quality smoke, what it proves, and the files that must stay synchronized when command behavior changes.", defs)
	if !ok {
		t.Fatal("expected maintainability handoff evidence to parse")
	}
	if maintainabilityHandoff.Status != "covered" {
		t.Fatalf("expected maintainability handoff evidence to be covered, got %+v", maintainabilityHandoff)
	}
	referenceBenchmark, ok := parseReadinessEvidenceLine("GOAL-0001", "Reference benchmark: Category: Developer CLI; References: namba-ai, pi.dev, Claude Code; Baseline expectations: install is clear and one command creates useful work context; Current comparison: setup meets baseline and evidence loop is above baseline; Below-baseline gaps: None; no critical user or operator baseline gap remains; Above-baseline strength: project-local evidence pressure; Decision: Service Quality is allowed from the benchmark perspective.", defs)
	if !ok {
		t.Fatal("expected reference benchmark evidence to parse")
	}
	if referenceBenchmark.Status != "covered" {
		t.Fatalf("expected reference benchmark evidence to be covered, got %+v", referenceBenchmark)
	}
	naturalReferenceBenchmark, ok := parseReadinessEvidenceLine("GOAL-0001", "Reference benchmark: Category: Local file-backed utility CLI; References: Git, SQLite CLI, Taskfile, Make; Baseline expectations: local commands are documented and repeatable command output exists; Current comparison: this sample meets the repeatable local CLI baseline; Below-baseline gaps: None for this smoke path; Above-baseline strength: evidence is captured before learning; Decision: Service Quality can continue from this benchmark.", defs)
	if !ok {
		t.Fatal("expected natural reference benchmark evidence to parse")
	}
	if naturalReferenceBenchmark.Status != "covered" {
		t.Fatalf("expected natural reference benchmark evidence to be covered, got %+v", naturalReferenceBenchmark)
	}
	noneCriticalReferenceBenchmark, ok := parseReadinessEvidenceLine("GOAL-0001", "Reference benchmark: Category: Local CLI release-note tracker; References: git-chglog, standard-version, release-it; Baseline expectations: local entries, local data, repeatable smoke, setup docs, and rollback docs; Current comparison: this CLI meets baseline for local add/list, file persistence, validation, setup, and rollback; Below baseline gaps: none critical for the local-only CLI category; Above baseline strength: active validator promotion is evidence-driven; Decision: Service Quality is allowed because no core category-baseline gap remains.", defs)
	if !ok {
		t.Fatal("expected none-critical reference benchmark evidence to parse")
	}
	if noneCriticalReferenceBenchmark.Status != "covered" {
		t.Fatalf("expected none-critical reference benchmark evidence to be covered, got %+v", noneCriticalReferenceBenchmark)
	}
	errorHandling, ok := parseReadinessEvidenceLine("GOAL-0001", "Error handling: Covered. Empty, loading, error, fallback, and recovery states are handled for the primary path: missing profile fields, future birth date, incomplete daily log, empty report, and storage-disabled browser fallback. Playwright verified each state at 390x844.", defs)
	if !ok {
		t.Fatal("expected error handling evidence with missing input text to parse")
	}
	if errorHandling.Status != "covered" {
		t.Fatalf("expected error handling evidence with missing input text to be covered, got %+v", errorHandling)
	}
}

func TestBetaGateAcceptsStaticDeploymentAndRunbookEvidence(t *testing.T) {
	defs := readinessDimensionDefs()
	lines := []string{
		"Validation coverage: Playwright smoke passed and HTTP check passed; primary flow validation is repeatable.",
		"Security baseline: security boundary documented and verified for local-only storage, token and session limits.",
		"Deployment readiness: proof. Release/build artifacts are created at `dist/llog-beta-demo/index.html` and `dist/llog-beta-demo.zip` outside the `prototype/` development path. Demo deployment path is documented in `demo-release.md`. Validation proved the release artifacts through direct `file://` execution, isolated artifact server URL `http://127.0.0.1:4201/?artifact=1`, extracted zip release URL `http://127.0.0.1:4202/?release=zip`, artifact parity, and mobile Playwright smoke with realistic data.",
		"Operations and docs: `demo-release.md` documents artifact creation, direct file run, static server run, smoke path, rollback, and stop conditions.",
		"Reference benchmark: Category: Static journaling app; References: Day One, Journey, Diarium; Baseline expectations: daily entry, report, setup, and handoff are understandable; Current comparison: core entry and report meet baseline, and local artifact release evidence is above baseline; Below-baseline gaps: None; no critical user or operator baseline gap remains; Above-baseline strength: local artifact release evidence; Decision: Service Quality is allowed from the benchmark perspective.",
	}
	records := []readinessEvidenceRecord{}
	for _, line := range lines {
		record, ok := parseReadinessEvidenceLine("GOAL-0010", line, defs)
		if !ok {
			t.Fatalf("expected readiness evidence to parse: %s", line)
		}
		records = append(records, record)
	}

	state := deriveReadinessState(map[string]string{
		"Current Stage": "Beta",
		"Product":       "LLog / 엘로그",
		"MVP":           "A static fortune calendar and daily log demo.",
	}, growthState{}, records)
	dims := readinessDimensionMap(state.Dimensions)
	if dims["deployment_readiness"].Status != "covered" {
		t.Fatalf("expected deployment readiness covered, got %+v", dims["deployment_readiness"])
	}
	if dims["operations_docs"].Status != "covered" {
		t.Fatalf("expected operations docs covered, got %+v", dims["operations_docs"])
	}
	if dims["reference_benchmark"].Status != "covered" {
		t.Fatalf("expected reference benchmark covered, got %+v", dims["reference_benchmark"])
	}
	if state.StageGate.Status != "ready" {
		t.Fatalf("expected Beta gate ready, got %+v", state.StageGate)
	}
	if state.NextPressure.Axis != "stage_advancement" {
		t.Fatalf("expected stage advancement next pressure, got %+v", state.NextPressure)
	}
}

func TestCompletePromotesErrorHandlingEvidenceWithMissingInputText(t *testing.T) {
	root := t.TempDir()
	mustRun(t, root, "init")
	writeFile(t, filepath.Join(root, "plan.md"), "# Product Plan\n\n## Product\n\nLLog / 엘로그\n\n## Target Users\n\nDaily diary users\n\n## MVP\n\nA user can create a profile, view fortune states, save a daily log, and revisit the report.\n\n## Current Stage\n\nUsable MVP\n\n## Build Style\n\nStatic prototype before Expo\n\n## Success Criteria\n\nPrimary flow works with persistence, edge states, and repeatable validation.\n")

	if _, err := runCLI(args("run", "Handle empty, failure, and edge states"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\nPlaywright smoke passed at mobile 390x844 and HTTP check passed.\n\n## Readiness Evidence\n\nCore UX: Browser smoke verified the primary LLog flow and state badges at mobile 390x844.\nData persistence: localStorage saved profile/log records and storage fallback was verified.\nError handling: Covered. Empty, loading, error, fallback, and recovery states are handled for the primary path: missing profile fields, future birth date, incomplete daily log, empty report, storage-disabled browser fallback, and two-step data deletion. Playwright verified each state at 390x844.\nValidation coverage: Playwright smoke passed and HTTP check passed; primary flow validation is repeatable.\n\n## Changed Files\n\nprototype/index.html\n\n## Decisions\n\nKeep sample fortune output labeled.\n\n## Reusable Patterns\n\nValidate every state and one stale-transition risk.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nCreate the first Expo app shell.\n\n## Learn Notes\n\n- Pattern: Validate both the state itself and a transition that could leave UI stale.\n")

	out, err := runCLI(args("complete"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	assertContains(t, out.Stdout, "Finish gate: passed")
	assertContains(t, out.Stdout, "Readiness gate: Usable MVP -> Beta (ready)")
	assertContains(t, out.Stdout, "Next action: hyper advance")

	readiness := readReadinessStateIfExists(root)
	dims := readinessDimensionMap(readiness.Dimensions)
	if dims["error_handling"].Status != "covered" {
		t.Fatalf("expected error handling covered, got %+v", dims["error_handling"])
	}
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "next-packet.md")), "Action: advance")
	assertNotContains(t, readFile(t, filepath.Join(root, ".hyper", "next-packet.md")), "Handle empty, failure, and edge states")
}

func TestSurfaceProofEvidenceProgressesReadiness(t *testing.T) {
	root := t.TempDir()
	mustRun(t, root, "init")
	writeFile(t, filepath.Join(root, "plan.md"), "# Product Plan\n\n## Product\n\nTiny CRM\n\n## Target Users\n\nSolo sellers\n\n## MVP\n\nCreate and revisit customer notes.\n\n## Current Stage\n\nTiny MVP\n\n## Build Style\n\nWeb app\n\n## Success Criteria\n\nOne customer note flow works locally.\n")

	if _, err := runCLI(args("run"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`npm run build` passed.\n\n## Surface Proof Evidence\n\n- Evidence: Browser smoke verified the primary action create customer note flow at mobile 390x844 and desktop 1440x900; screenshots captured and passed.\n\n## Changed Files\n\nsrc/App.tsx\n\n## Decisions\n\nPending.\n\n## Reusable Patterns\n\nPending.\n\n## Blocker\n\nPending.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage readiness.\n")
	mustRun(t, root, "complete")

	if _, err := runCLI(args("run"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	state := readReadinessStateIfExists(root)
	dims := readinessDimensionMap(state.Dimensions)
	if dims["core_ux"].Status != "covered" {
		t.Fatalf("expected surface proof to cover core UX, got %+v", dims["core_ux"])
	}
	if dims["validation_coverage"].Status != "covered" {
		t.Fatalf("expected surface proof to cover validation coverage, got %+v", dims["validation_coverage"])
	}
	status, err := runCLI(args("status"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	assertContains(t, status.Stdout, "Proof: functional pending, surface covered, operational covered")
}

func TestRepeatedSurfaceProofCreatesVisualSmokeCandidate(t *testing.T) {
	root := t.TempDir()
	if err := ensureProjectLayout(root); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	db, err := openDB(root)
	if err != nil {
		t.Fatalf("db open failed: %v", err)
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		t.Fatalf("schema failed: %v", err)
	}
	insertTestMemory(t, db, "pattern", "GOAL-0001 surface proof evidence: Browser smoke verified primary action create note flow at mobile and desktop; screenshots captured and passed.")
	insertTestMemory(t, db, "pattern", "GOAL-0002 surface proof evidence: Browser smoke verified primary action create note flow at mobile and desktop; screenshots captured and passed.")

	state, hyperErr := updateGrowthState(root, db)
	if hyperErr != nil {
		t.Fatalf("growth failed: %v", hyperErr)
	}
	if len(state.Pressures) == 0 || state.Pressures[0].PressureType != "surface_validation" {
		t.Fatalf("expected surface validation pressure, got %+v", state.Pressures)
	}
	if len(state.Candidates) == 0 || !strings.HasPrefix(state.Candidates[0].Name, "validator-visual-smoke-") {
		t.Fatalf("expected visual smoke validator candidate, got %+v", state.Candidates)
	}
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), `"pressure_type": "surface_validation"`)
}

func TestReadinessEvidenceDoesNotDowngradeCompletePlan(t *testing.T) {
	plan := map[string]string{
		"Product":          "Tiny pet widget",
		"MVP":              "A draggable pet with one care loop.",
		"Success Criteria": "A user can run it and complete one care action.",
		"Current Stage":    "Tiny MVP",
	}
	weakRecord := readinessEvidenceRecordForAxis("GOAL-0001", "product_completeness", "proof - visible canvas pet and care panel exist.")
	state := deriveReadinessState(plan, growthState{}, []readinessEvidenceRecord{weakRecord})
	dim := readinessDimensionMap(state.Dimensions)["product_completeness"]
	if dim.Status != "covered" {
		t.Fatalf("complete plan should stay covered despite weak runtime evidence, got %+v", dim)
	}
}

func TestPlanAliasesAcceptBriefAndSuccessSignals(t *testing.T) {
	plan := parsePlan("# Service Probe\n\n## Product Brief\n\nA tiny notes API.\n\n## Success Signals\n\nCreate and list one note.\n")
	if got := plan["Product"]; got != "A tiny notes API." {
		t.Fatalf("Product alias = %q", got)
	}
	if got := plan["Success Criteria"]; got != "Create and list one note." {
		t.Fatalf("Success Criteria alias = %q", got)
	}
}

func TestPlanAliasesAcceptInlineFields(t *testing.T) {
	plan := parsePlan(`# Plan

Project: Service Desk Lite
Current Stage: Tiny MVP

Product brief:
A tiny internal support queue where a teammate can create one request, see it in a list, and mark it handled.

Build Style: Thin vertical slice first.

Validation:
Use the smallest command or smoke check that proves the useful flow still works.
`)
	if got := plan["Product"]; got != "Service Desk Lite" {
		t.Fatalf("Product inline field = %q", got)
	}
	if got := plan["MVP"]; got != "A tiny internal support queue where a teammate can create one request, see it in a list, and mark it handled." {
		t.Fatalf("Product brief inline field should fill MVP boundary, got %q", got)
	}
	if got := plan["Current Stage"]; got != "Tiny MVP" {
		t.Fatalf("Current Stage inline field = %q", got)
	}
	if got := plan["Build Style"]; got != "Thin vertical slice first." {
		t.Fatalf("Build Style inline field = %q", got)
	}
	if got := plan["Success Criteria"]; got != "Use the smallest command or smoke check that proves the useful flow still works." {
		t.Fatalf("Validation inline field = %q", got)
	}
}

func TestUpdatePlanCurrentStageUpdatesInlineField(t *testing.T) {
	body := strings.Join([]string{
		"# Plan",
		"",
		"Project: Inline Stage Probe",
		"Current Stage: Tiny MVP",
		"Build Style: Local CLI",
		"",
		"Product brief:",
		"A developer can add one item and list it back locally.",
		"",
	}, "\n")
	updated, changed := updatePlanCurrentStage(body, "Usable MVP")
	if !changed {
		t.Fatal("expected inline Current Stage to change")
	}
	assertContains(t, updated, "Current Stage: Usable MVP")
	assertNotContains(t, updated, "Current Stage: Tiny MVP")
	assertNotContains(t, updated, "## Current Stage")
	plan := parsePlan(updated)
	if got := plan["Current Stage"]; got != "Usable MVP" {
		t.Fatalf("expected updated inline stage to parse, got %q", got)
	}
}

func TestRuntimePacketCombinesPlanAndStageStopConditions(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plan.md"), strings.Join([]string{
		"# Plan",
		"",
		"Project: Inline Stage Probe",
		"Current Stage: Usable MVP",
		"Build Style: Local CLI",
		"",
		"Product brief:",
		"A developer can add one item and list it back locally.",
		"",
		"Validation:",
		"A smoke command proves add/list works.",
		"",
	}, "\n"))
	if _, err := runCLI(args("init"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := runCLI(args("run", "Make the flow persistent"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	goal := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "goal.md"))
	assertContains(t, goal, "## Stop When")
	assertContains(t, goal, "- Plan success criteria: A smoke command proves add/list works.")
	assertContains(t, goal, "- Core flow is usable without manual data edits.")
	assertContains(t, goal, "- Empty, loading, and error states are handled for the primary path.")
}

func TestReadinessIgnoresDeferredStructureSignals(t *testing.T) {
	plan := map[string]string{"Current Stage": "Tiny MVP"}
	growth := growthState{Pressures: []growthPressure{
		{
			PressureType: "repeated_validation",
			Signal:       "For tiny API MVPs, prove the primary flow with `httptest` before adding persistence or UI.",
			Effect:       "validation",
			State:        "observed",
			GoalCount:    1,
		},
		{
			PressureType: "stable_decision",
			Signal:       "Keep the Tiny MVP local and in-memory.",
			Effect:       "work_boundary",
			State:        "observed",
			GoalCount:    1,
		},
	}}
	state := deriveReadinessState(plan, growth, nil)
	dims := readinessDimensionMap(state.Dimensions)
	if got := dims["persistence"].Status; got != "missing" {
		t.Fatalf("deferred persistence should stay missing, got %+v", dims["persistence"])
	}
	if got := dims["core_ux"].Status; got != "missing" {
		t.Fatalf("deferred UI should not create Core UX pressure, got %+v", dims["core_ux"])
	}
	if got := dims["deployment_readiness"].Status; got != "missing" {
		t.Fatalf("local in-memory decision should not create deployment pressure, got %+v", dims["deployment_readiness"])
	}
}

func TestBroadFocusIsRewrittenThroughReadinessPressure(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny tasks", "Build a tiny task list MVP")

	if _, err := runCLI(args("run", "실서비스 수준으로 업그레이드"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	goal := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "goal.md"))
	assertContains(t, goal, "Translate `실서비스 수준으로 업그레이드` into the smallest Tiny MVP step")
	assertContains(t, goal, "- Current focus: 실서비스 수준으로 업그레이드")
}

func TestStageAdvancementCandidateWhenGateReady(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny tasks", "Build a tiny task list MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`npm run build` passed.\n\n## Readiness Evidence\n\nCore UX: Browser smoke verified create, complete, and delete flow.\nValidation coverage: `npm run build` passed and primary flow smoke test passed.\n\n## Changed Files\n\nsrc/App.tsx\n\n## Decisions\n\nKeep local-first storage.\n\n## Reusable Patterns\n\nPending.\n\n## Blocker\n\nPending.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n\n## Learn Notes\n\n- Pattern: Record axis-labeled readiness evidence.\n")
	mustRun(t, root, "complete")

	if _, err := runCLI(args("run"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	readiness := readFile(t, filepath.Join(root, ".hyper", "readiness", "state.json"))
	assertContains(t, readiness, `"candidate": true`)
	assertContains(t, readiness, `"plan_change": "Current Stage -> Usable MVP"`)
	goal := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0002", "goal.md"))
	assertContains(t, goal, "Stage advancement candidate")
	assertContains(t, goal, "Recommend updating plan.md Current Stage to Usable MVP")
	assertContains(t, goal, "Do not run `hyper advance` until the user accepts the stage advancement")
}

func TestAdvanceUpdatesPlanWhenGateReady(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny tasks", "Build a tiny task list MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`npm run build` passed and browser smoke passed.\n\n## Readiness Evidence\n\nCore UX: Browser smoke verified create, complete, and delete flow.\nValidation coverage: `npm run build` passed and primary flow smoke test passed.\n\n## Blocker\n\nNone blocking.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n")
	mustRun(t, root, "complete")

	out, err := runCLI(args("advance"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("advance failed: %v", err)
	}
	assertContains(t, out.Stdout, "Stage advanced: Tiny MVP -> Usable MVP")
	assertContains(t, out.Stdout, "Updated: plan.md Current Stage -> Usable MVP")
	assertContains(t, out.Stdout, "Readiness gate: Usable MVP -> Beta")
	assertContains(t, out.Stdout, "Next packet plan: .hyper/next-packet.md")
	assertContains(t, readFile(t, filepath.Join(root, "plan.md")), "## Current Stage\n\nUsable MVP")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "state.json")), `"stage": "Usable MVP"`)
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "logs", "project.jsonl")), `"stage_advanced"`)
	nextPlan := readFile(t, filepath.Join(root, ".hyper", "next-packet.md"))
	assertContains(t, nextPlan, "Action: run")
	assertNotContains(t, nextPlan, "Command: hyper advance")
}

func TestAdvanceStopsAutoPlanWhenRunUntilTargetReached(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plan.md"), strings.Join([]string{
		"# Product Plan",
		"",
		"## Product",
		"",
		"Local Clip Shelf",
		"",
		"## Target Users",
		"",
		"Developers and operators",
		"",
		"## MVP",
		"",
		"Save, search, pin, and restart clipboard snippets.",
		"",
		"## Current Stage",
		"",
		"Beta",
		"",
		"## Build Style",
		"",
		"Native desktop helper with local SQLite storage.",
		"",
		"## Success Criteria",
		"",
		"Primary flow validates with realistic data and local privacy boundaries.",
	}, "\n"))

	out, err := runCLI(args("run", "--auto", "--until", "service-quality", "Prepare service quality"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("auto run failed: %v", err)
	}
	assertContains(t, out.Stdout, "Run mode: auto until Service Quality")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), strings.Join([]string{
		"# GOAL-0001 Evidence",
		"",
		"## Validation",
		"",
		"`clip-shelf smoke` passed for save, search, pin, and restart using realistic command text.",
		"",
		"## Readiness Evidence",
		"",
		"Validation coverage: `clip-shelf smoke` passed and is repeatable for save, search, pin, and restart.",
		"Security baseline: Privacy boundary verified: clipboard content stays local in SQLite, no cloud sync or telemetry, and sensitive text can be deleted locally.",
		"Deployment readiness: Packaged helper binary smoke passed outside the development command path.",
		"Operations and docs: README documents setup, local data path, delete path, rollback, and smoke command.",
		"",
		"## Reference Benchmark Evidence",
		"",
		"- Category: Local clipboard history helper.",
		"- References: Raycast Clipboard History, Alfred Clipboard History, Maccy.",
		"- Baseline expectations: Save recent text, search quickly, pin snippets, and keep local privacy boundaries clear.",
		"- Current comparison: below baseline = none for command-helper path; meets baseline = save/search/pin/restart and privacy proof; above baseline = operator command-snippet smoke.",
		"- Below-baseline gaps: No critical below-baseline gap for the command-helper path.",
		"- Above-baseline strength: Restart persistence and privacy proof are explicit.",
		"- Decision: Service Quality is allowed for the helper command path.",
		"",
		"## Changed Files",
		"",
		"Prototype helper behavior and docs.",
		"",
		"## Blocker",
		"",
		"None blocking.",
	}, "\n"))
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n")
	if _, err := runCLI(args("complete"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("complete failed: %v", err)
	}

	advance, err := runCLI(args("advance"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("advance failed: %v", err)
	}
	assertContains(t, advance.Stdout, "Stage advanced: Beta -> Service Quality")
	assertContains(t, advance.Stdout, "Next action: hyper status --short")
	assertContains(t, advance.Stdout, "Why: Auto target Service Quality is reached; review status before choosing a new target or manual next run.")
	if strings.Count(advance.Stdout, "  hyper status --short") != 1 {
		t.Fatalf("expected one status next command, got:\n%s", advance.Stdout)
	}
	nextPlan := readFile(t, filepath.Join(root, ".hyper", "next-packet.md"))
	assertContains(t, nextPlan, "Mode: auto until Service Quality")
	assertContains(t, nextPlan, "Action: stop")
	assertContains(t, nextPlan, "Command: hyper status --short")
	assertContains(t, nextPlan, "Reason: Auto target Service Quality is reached; review status before choosing a new target or manual next run.")
	assertNotContains(t, nextPlan, "Command: hyper advance")
}

func TestAdvanceToSustainedServiceQualityDoesNotLoop(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plan.md"), strings.Join([]string{
		"# Product Plan",
		"",
		"## Product",
		"",
		"Local Build Relay",
		"",
		"## Target Users",
		"",
		"Developers",
		"",
		"## MVP",
		"",
		"Run one handoff command.",
		"",
		"## Current Stage",
		"",
		"Service Quality",
		"",
		"## Build Style",
		"",
		"Go CLI",
		"",
		"## Success Criteria",
		"",
		"Every packet proves the handoff command.",
	}, "\n"))
	if _, err := runCLI(args("run", "Prepare sustained quality"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	writeFile(t, filepath.Join(root, ".hyper", "capabilities", "active", "validator", "validator-go-test.md"), "# validator-go-test\n\nStatus: active\nKind: validator\nSignal: Run go test ./... before completing packets.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), strings.Join([]string{
		"# GOAL-0001 Evidence",
		"",
		"## Validation",
		"",
		"`go test ./...` passed.",
		"",
		"## Readiness Evidence",
		"",
		"Validation coverage: `go test ./...` passed and is repeatable.",
		"Security baseline: Security boundary verified, no cloud sync, no telemetry, and no secrets.",
		"Deployment readiness: Packaged CLI smoke passed outside the development command.",
		"Operations and docs: README documents setup, rollback, and smoke command.",
		"Maintainability: Test helper keeps command validation repeatable without hidden local context.",
		"Sustained quality: Active validator validator-go-test is required and verified before every packet handoff.",
		"",
		"## Reference Benchmark Evidence",
		"",
		"- Category: Local developer handoff CLI.",
		"- References: GitHub CLI, Taskfile, Make.",
		"- Baseline expectations: documented command, repeatable output, rollback, no hidden credentials.",
		"- Current comparison: below baseline = none; meets baseline = command/test/docs/rollback; above baseline = packet evidence loop.",
		"- Below-baseline gaps: No critical below-baseline gap.",
		"- Above-baseline strength: packet evidence loop.",
		"- Decision: Service Quality proof can continue.",
		"",
		"## Active Capability Evidence",
		"",
		"validator-go-test: `go test ./...` passed.",
		"",
		"## Blocker",
		"",
		"None blocking.",
	}, "\n"))
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview sustained quality advancement.\n")
	if _, err := runCLI(args("complete"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	advance, err := runCLI(args("advance"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("advance failed: %v", err)
	}
	assertContains(t, advance.Stdout, "Stage advanced: Service Quality -> Sustained Service Quality")
	assertNotContains(t, advance.Stdout, "Next action: hyper advance")

	status, err := runCLI(args("status", "--short"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	assertContains(t, status.Stdout, "Stage: Sustained Service Quality")
	assertContains(t, status.Stdout, "Gate: Sustained Service Quality -> Sustained Service Quality")
	assertNotContains(t, status.Stdout, "Next: hyper advance")
	assertContains(t, readFile(t, filepath.Join(root, "plan.md")), "## Current Stage\n\nSustained Service Quality")
}

func TestAdvanceRejectsWhenGateNotReady(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny tasks", "Build a tiny task list MVP")

	_, err := runCLI(args("advance"), testRoot(root), fakeUpdater{})
	if err == nil {
		t.Fatal("expected advance to require a ready stage gate")
	}
	assertContains(t, err.Message, "Stage gate is not ready")
	assertContains(t, err.Message, "Core UX")
}

func TestBetaStageGeneratesValidatorCandidates(t *testing.T) {
	root := t.TempDir()
	mustRun(t, root, "init")
	writeFile(t, filepath.Join(root, "plan.md"), "# Product Plan\n\n## Product\n\nTiny CRM\n\n## Target Users\n\nSolo sellers\n\n## MVP\n\nCustomer notes with persistence and validation.\n\n## Current Stage\n\nBeta\n\n## Build Style\n\nWeb app\n\n## Non-goals\n\nEnterprise permissions\n\n## Constraints\n\nLocal first\n\n## Success Criteria\n\nPrimary flow validates against realistic data.\n\n## Current Focus\n\nPrepare beta quality.\n")

	if _, err := runCLI(args("run"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	candidate := readFile(t, filepath.Join(root, ".hyper", "validators", "generated", "validator-beta-primary-flow-smoke.md"))
	assertContains(t, candidate, "Status: candidate")
	assertContains(t, candidate, "Stage-specific service-quality validator candidate")
	assertContains(t, candidate, "## When Required")
	assertContains(t, candidate, "## Required Behavior")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "capabilities", "candidates", "validator", "validator-beta-security-baseline.md")), "Status: candidate")
	goal := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "goal.md"))
	assertContains(t, goal, "Beta validation should use realistic data")
	assertNotContains(t, goal, "Required active validator validator-beta-primary-flow-smoke")
}

func TestRuntimePacketIgnoresPlanPlaceholders(t *testing.T) {
	root := t.TempDir()
	mustRun(t, root, "init")
	writeFile(t, filepath.Join(root, "plan.md"), "# Product Plan\n\n## Product\n\nTiny chat\n\n## MVP\n\nTBD\n\n## Current Stage\n\nTiny MVP\n\n## Build Style\n\nTBD\n\n## Non-goals\n\nTBD\n\n## Constraints\n\nTBD\n\n## Success Criteria\n\nTBD\n\n## Current Focus\n\nShip the first usable chat slice\n")

	if _, err := runCLI(args("run"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	goal := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "goal.md"))
	assertContains(t, goal, "# GOAL-0001 Runtime Packet")
	assertContains(t, goal, "- Build style: Detect from project")
	assertContains(t, goal, "- If the product brief is incomplete, inspect the current project")
	assertNotContains(t, goal, "\nTBD\n")
}

func TestInternalLearnStoresFailure(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny billing", "Build a tiny billing MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\nStatus: blocked\nReason: Missing Stripe key\n")

	out, err := runCLI(args("internal", "learn"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("learn failed: %v", err)
	}
	assertContains(t, out.Stdout, "Runtime packet state: blocked")
	assertContains(t, out.Stdout, "Learn role: extract repeated needs")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "memories", "failures.md")), "Missing Stripe key")
}

func TestLearnExtractsDurableSignals(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny chat", "Build a tiny chat MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\ngo test ./... passed.\n\n## Changed Files\n\ncmd/chat.go\n\n## Decisions\n\nKeep storage local-first for the MVP.\n\n## Reusable Patterns\n\nUse table-driven tests for message parsing.\n\n## Blocker\n\nPending.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nAdd persisted chat history.\n\n## Learn Notes\n\n- Decision: Keep CLI output stable.\n- Pattern: Run go test before every runtime packet handoff.\n- Constraint: Do not add external services without credentials.\n- Failure: WebSocket path failed without a local server.\n")

	out, err := runCLI(args("internal", "learn"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("learn failed: %v", err)
	}

	assertContains(t, out.Stdout, "Candidate memories: 7")
	assertContains(t, out.Stdout, "Inserted memories: 7")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "memories", "decisions.md")), "Keep storage local-first")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "memories", "patterns.md")), "table-driven tests")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "memories", "constraints.md")), "external services")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "memories", "failures.md")), "WebSocket path failed")
	assertNotContains(t, readFile(t, filepath.Join(root, ".hyper", "memories", "decisions.md")), "Changed Files")
}

func TestLearnDedupesOverlappingSectionAndLearnSignals(t *testing.T) {
	evidence := "# GOAL-0001 Evidence\n\n## Validation\n\n`./check.sh` passed.\n\n## Decisions\n\nKeep the first slice as a local CLI with file-backed persistence; avoid generated harnesses until repeated need appears.\n\n## Blocker\n\nNone blocking.\n"
	next := "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nHandle empty state.\n\n## Learn Notes\n\n- Decision: Keep the first slice as a local CLI with file-backed persistence.\n- Constraint: Do not create harnesses until repeated evidence shows the project needs one.\n"
	memories := memoriesForDerivedState(goalState{State: "completed", Reason: "done"}, "GOAL-0001", evidence, next)

	overlappingDecisions := 0
	for _, memory := range memories {
		if memory.Kind == "decision" && strings.Contains(memory.Text, "first slice") {
			overlappingDecisions++
		}
	}
	if overlappingDecisions != 1 {
		t.Fatalf("expected one deduped overlapping decision, got %d in %+v", overlappingDecisions, memories)
	}
}

func TestLearnIgnoresHyperRunMetaProgress(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny chat", "Build a tiny chat MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`hyper run` created `GOAL-0001`.\n\n## Changed Files\n\nPending.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nContinue.\n")

	out, err := runCLI(args("internal", "learn"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("learn failed: %v", err)
	}
	assertContains(t, out.Stdout, "Candidate memories: 0")
	assertNotContains(t, readFile(t, filepath.Join(root, ".hyper", "memories", "patterns.md")), "hyper run")
}

func TestGoalStateIgnoresNoIssueBlockerAndFailureNotes(t *testing.T) {
	completed := deriveGoalState("## Validation\n\n`go test ./...` passed.\n\n## Blocker\n\nNone.\n", "## Recommended Next Goal\n\nShip next slice.\n")
	if completed.State != "completed" {
		t.Fatalf("expected no-issue blocker to complete, got %+v", completed)
	}
	completed = deriveGoalState("## Validation\n\nSmoke passed.\n\n## Blocker\n\nNo blocker for this episode. Validation used local MySQL.\n", "## Recommended Next Goal\n\nShip next slice.\n")
	if completed.State != "completed" {
		t.Fatalf("expected no-blocker sentence to complete, got %+v", completed)
	}
	completed = deriveGoalState("## Validation\n\nSmoke passed.\n\n## Blocker\n\nNo blocker for this packet.\n", "## Recommended Next Goal\n\nShip next slice.\n")
	if completed.State != "completed" {
		t.Fatalf("expected no-blocker packet sentence to complete, got %+v", completed)
	}
	completed = deriveGoalState("## Validation\n\nSmoke passed.\n\n## Blocker\n\nNone blocking.\n- Operational note: direct screenshot write returned `EPERM`; saving to `/private/tmp` and copying into the goal folder succeeded.\n", "## Recommended Next Goal\n\nShip next slice.\n")
	if completed.State != "completed" {
		t.Fatalf("expected no-op blocker notes to complete, got %+v", completed)
	}
	completed = deriveGoalState("## Validation\n\n`npm run check` passed.\n\n## Blocker\n\nClear: implementation and validation completed for this packet.\n", "## Recommended Next Goal\n\nShip next slice.\n")
	if completed.State != "completed" {
		t.Fatalf("expected clear completion blocker text to complete, got %+v", completed)
	}
	completed = deriveGoalState("## Validation\n\nBrowser smoke passed.\n\n## Blocker\n\nNo technical blocker. Product stage update is intentionally deferred until the user explicitly accepts the Tiny MVP -> Usable MVP advancement.\n", "## Recommended Next Goal\n\nReview stage advancement.\n")
	if completed.State != "completed" {
		t.Fatalf("expected user-deferred stage note with evidence to complete, got %+v", completed)
	}
	waiting := deriveGoalState("## Blocker\n\nWaiting for user approval before stage advancement.\n", "")
	if waiting.State != "waiting_user" {
		t.Fatalf("expected user decision blocker to wait for user, got %+v", waiting)
	}
	kind, value := parseLearnNote("- Failure: None in this episode.")
	if kind != "" || value != "" {
		t.Fatalf("expected no-op failure learn note to be ignored, got %q %q", kind, value)
	}
	kind, value = parseLearnNote("- Failure: None in this run.")
	if kind != "" || value != "" {
		t.Fatalf("expected no-op failure learn note for this run to be ignored, got %q %q", kind, value)
	}
	kind, value = parseLearnNote("- Failure: None critical for the local-only CLI category.")
	if kind != "" || value != "" {
		t.Fatalf("expected no-critical failure learn note to be ignored, got %q %q", kind, value)
	}
}

func TestValidationMemoryPrefersCommandOverOutputLine(t *testing.T) {
	validation := strings.Join([]string{
		"Command: `./check.sh`",
		"Output:",
		"```text",
		"no items",
		"service-quality smoke passed",
		"```",
	}, "\n")
	if got := firstUsefulValidationMemory(validation); got != "`./check.sh` passed." {
		t.Fatalf("expected command-centered validation memory, got %q", got)
	}
}

func TestStatusDerivesCompletedForNoOpBlocker(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny tasks", "Build a tiny task list MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`npm run build` passed.\n\n## Readiness Evidence\n\nCore UX: Browser smoke verified create and complete flow.\nValidation coverage: `npm run build` passed and primary flow smoke test passed.\n\n## Blocker\n\nNone blocking.\n- Operational note: direct screenshot write from the browser runtime to the workspace returned `EPERM`; saving to `/private/tmp` and copying into the goal folder succeeded.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n")
	state, err := readState(filepath.Join(root, ".hyper", "state.json"))
	if err != nil {
		t.Fatalf("read state failed: %v", err)
	}
	state.Status = "blocked"
	if err := writeJSON(filepath.Join(root, ".hyper", "state.json"), state); err != nil {
		t.Fatalf("write stale state failed: %v", err)
	}

	status, err := runCLI(args("status"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	assertContains(t, status.Stdout, "Status: completed (state.json: blocked)")
	assertContains(t, status.Stdout, "Runtime packet state: completed")
	assertContains(t, status.Stdout, "Next action: hyper repair")
}

func TestCompleteTreatsClearBlockerAsCompleted(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny canvas", "Build a tiny canvas MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\n`npm run check` passed.\nBrowser smoke verified the primary flow.\n\n## Readiness Evidence\n\nCore UX: Browser smoke verified the primary flow.\nValidation coverage: `npm run check` passed.\n\n## Blocker\n\nClear: implementation and validation completed for this packet.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nReview stage advancement.\n\n## Learn Notes\n\n- Failure: None in this run.\n")

	out, err := runCLI(args("complete"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	assertContains(t, out.Stdout, "State: completed")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "state.json")), `"status": "completed"`)
	assertNotContains(t, readFile(t, filepath.Join(root, ".hyper", "memories", "failures.md")), "None in this run")
	assertNotContains(t, readFile(t, filepath.Join(root, ".hyper", "memories", "failures.md")), "implementation and validation completed")
}

func TestGoalStateDerivation(t *testing.T) {
	blocked := deriveGoalState("## Blocker\n\nMissing API key", "")
	if blocked.State != "blocked" {
		t.Fatalf("expected blocked, got %s", blocked.State)
	}
	completed := deriveGoalState("## Validation\n\ngo test passed", "## Recommended Next Goal\n\nShip beta")
	if completed.State != "completed" {
		t.Fatalf("expected completed, got %s", completed.State)
	}
}

func TestStageNormalizationUsesFirstNamedStage(t *testing.T) {
	state := deriveReadinessState(map[string]string{
		"Current Stage": "Tiny MVP moving toward Usable MVP. Do not advance yet.",
		"Product":       "Pickachat is a location-pinned chat web app.",
		"MVP":           "Create a pin and send a message.",
	}, growthState{}, nil)
	if state.Stage != "Tiny MVP" {
		t.Fatalf("expected Tiny MVP, got %s", state.Stage)
	}
	if state.StageGate.CurrentStage != "Tiny MVP" || state.StageGate.NextStage != "Usable MVP" {
		t.Fatalf("expected Tiny MVP gate, got %+v", state.StageGate)
	}
	goal := readinessRecommendedGoal(map[string]string{"Product": "Pickachat is a location-pinned chat web app."}, "Tiny MVP", "persistence")
	assertContains(t, goal, "primary Pickachat flow")

	sustained := deriveReadinessState(map[string]string{
		"Current Stage": "Sustained Service Quality",
		"Product":       "Local Build Relay",
	}, growthState{Candidates: []growthCandidate{{Kind: "validator", Name: "validator-go-test", Status: "active"}}}, []readinessEvidenceRecord{
		readinessEvidenceRecordForAxis("GOAL-0001", "validation_coverage", "`go test ./...` passed and is repeatable."),
		readinessEvidenceRecordForAxis("GOAL-0001", "operations_docs", "Operations and docs: README documents setup, rollback, and smoke command."),
		readinessEvidenceRecordForAxis("GOAL-0001", "maintainability", "Maintainability: Test helper keeps command validation repeatable without hidden local context."),
	})
	if sustained.Stage != "Sustained Service Quality" {
		t.Fatalf("expected Sustained Service Quality, got %s", sustained.Stage)
	}
	if sustained.StageGate.CurrentStage != "Sustained Service Quality" || sustained.StageGate.NextStage != "Sustained Service Quality" {
		t.Fatalf("expected terminal sustained gate, got %+v", sustained.StageGate)
	}
	if sustained.StageGate.Advancement.Candidate {
		t.Fatalf("sustained stage must not create another stage advancement candidate: %+v", sustained.StageGate.Advancement)
	}
	if sustained.NextPressure.Axis == "stage_advancement" {
		t.Fatalf("sustained stage should continue quality work, got %+v", sustained.NextPressure)
	}
	assertContains(t, sustained.NextPressure.RecommendedGoal, "Run active quality checks")
	assertNotContains(t, sustained.NextPressure.RecommendedGoal, "until active validation")
}

func TestReferenceBenchmarkPressureShapesRuntimePacket(t *testing.T) {
	readiness := readinessState{
		Version: 1,
		Stage:   "Beta",
		StageGate: readinessStageGate{
			CurrentStage: "Beta",
			NextStage:    "Service Quality",
			Status:       "not_ready",
			RequiredAxes: []string{"validation_coverage", "security_baseline", "deployment_readiness", "operations_docs", "reference_benchmark"},
		},
		NextPressure: readinessPressure{
			Axis:             "reference_benchmark",
			AxisName:         "Reference benchmark",
			Status:           "missing",
			Reason:           "Reference benchmark is missing for the Beta -> Service Quality gate.",
			RecommendedGoal:  "Compare Tiny CRM against references.",
			WorkBoundary:     "Compare the current result against 3-5 named category references before adding feature breadth; close only the strongest critical below-baseline gap if one is found.",
			ValidationSignal: "Fill Reference Benchmark Evidence with named references, baseline expectations, current comparison, below-baseline gaps, above-baseline strength, and decision.",
		},
	}
	plan := map[string]string{
		"Product":       "Tiny CRM is a local-first sales notes app.",
		"MVP":           "Capture and revisit one customer note.",
		"Current Stage": "Beta",
	}

	goal := readinessRecommendedGoal(plan, "Beta", "reference_benchmark")
	assertContains(t, goal, "3-5 named category references")
	assertContains(t, goal, "define the baseline")
	assertContains(t, goal, "strongest critical below-baseline gap")

	boundary := runtimeWorkBoundary(goal, "Beta", plan, growthState{}, readiness)
	assertContains(t, boundary, "Do not add broad feature work")
	assertContains(t, boundary, "Select 3-5 named references")
	assertContains(t, boundary, "implement only the smallest fix")
	assertContains(t, boundary, "do not advance the stage")

	next := buildNextDoc("GOAL-0009", readiness)
	assertContains(t, next, "durable reference signals")
	assertContains(t, next, "category baseline")
	assertContains(t, next, "comparison-driven constraint")
	assertContains(t, next, "Do not record one-off reference names")
}

func TestServiceQualityStageDefinesOperationalStandard(t *testing.T) {
	done := stageDoneCondition("Service Quality")
	assertContains(t, done, "Required validation, security, deployment, operations, and maintainability evidence")
	assertContains(t, done, "Setup, release, rollback, and recovery paths")
	assertContains(t, done, "Reference benchmark evidence")
	assertContains(t, done, "No critical blocker remains")

	if boundary := stageRuntimeBoundary("Service Quality"); !hasAny(boundary, "security/privacy boundaries", "release and rollback proof", "operator docs") {
		t.Fatalf("expected service quality boundary to name operational acceptance criteria, got %q", boundary)
	} else {
		assertContains(t, boundary, "category-baseline comparison")
	}
	if signal := stageValidationSignal("Service Quality"); !hasAny(signal, "set up", "rolled back", "handed off") {
		t.Fatalf("expected service quality validation signal to describe handoff checks, got %q", signal)
	} else {
		assertContains(t, signal, "compared against category references")
	}

	_, _, axes, evidence := readinessGateDefinition("Service Quality")
	for _, axis := range []string{"validation_coverage", "security_baseline", "deployment_readiness", "operations_docs", "maintainability", "reference_benchmark", "sustained_quality"} {
		found := false
		for _, got := range axes {
			if got == axis {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected service quality gate axis %s in %+v", axis, axes)
		}
	}
	joinedEvidence := strings.Join(evidence, "\n")
	assertContains(t, joinedEvidence, "Required validation")
	assertContains(t, joinedEvidence, "rollback")
	assertContains(t, joinedEvidence, "hidden context")
	assertContains(t, joinedEvidence, "Reference benchmark evidence")
	assertContains(t, joinedEvidence, "Repeated runtime evidence")
}

func TestServiceQualityGateRequiresSustainedGrowthEvidence(t *testing.T) {
	plan := map[string]string{
		"Product":          "Local Build Relay",
		"Current Stage":    "Service Quality",
		"Success Criteria": "Every packet proves the handoff command.",
	}
	evidence := []readinessEvidenceRecord{
		readinessEvidenceRecordForAxis("GOAL-0001", "validation_coverage", "`go test ./...` passed and is repeatable."),
		readinessEvidenceRecordForAxis("GOAL-0001", "security_baseline", "Security baseline: Privacy boundary verified, no cloud sync, no telemetry, and no secrets."),
		readinessEvidenceRecordForAxis("GOAL-0001", "deployment_readiness", "Deployment readiness: Packaged CLI smoke passed outside the development command."),
		readinessEvidenceRecordForAxis("GOAL-0001", "operations_docs", "Operations and docs: README documents setup, rollback, and smoke command."),
		readinessEvidenceRecordForAxis("GOAL-0001", "maintainability", "Maintainability: Test helper keeps command validation repeatable without hidden local context."),
		readinessEvidenceRecordForAxis("GOAL-0001", "reference_benchmark", strings.Join([]string{
			"Category: Local developer handoff CLI.",
			"References: GitHub CLI, Taskfile, Make.",
			"Baseline expectations: documented command, repeatable output, rollback, no hidden credentials.",
			"Current comparison: below baseline = none; meets baseline = command/test/docs/rollback; above baseline = packet evidence loop.",
			"Below-baseline gaps: No critical below-baseline gap.",
			"Above-baseline strength: packet evidence loop.",
			"Decision: Service Quality proof can continue.",
		}, "; ")),
	}

	state := deriveReadinessState(plan, growthState{}, evidence)
	if state.StageGate.Status != "not_ready" {
		t.Fatalf("single service-quality packet must not unlock sustained quality, got %+v", state.StageGate)
	}
	if state.NextPressure.Axis != "sustained_quality" {
		t.Fatalf("expected sustained quality pressure, got %+v", state.NextPressure)
	}
	assertContains(t, strings.Join(state.StageGate.BlockingGaps, "\n"), "Sustained quality")

	fakeActiveEvidence := append([]readinessEvidenceRecord{}, evidence...)
	fakeActiveEvidence = append(fakeActiveEvidence, readinessEvidenceRecordForAxis("GOAL-0002", "sustained_quality", "Sustained quality: Active validator validator-go-test is required and verified before every packet handoff."))
	state = deriveReadinessState(plan, growthState{}, fakeActiveEvidence)
	if state.StageGate.Status != "not_ready" {
		t.Fatalf("text-only active validator evidence must not unlock sustained quality, got %+v", state.StageGate)
	}
	if state.NextPressure.Axis != "sustained_quality" {
		t.Fatalf("expected sustained quality pressure without actual active capability, got %+v", state.NextPressure)
	}

	growth := growthState{Candidates: []growthCandidate{{Kind: "validator", Name: "validator-go-test", Status: "active"}}}
	state = deriveReadinessState(plan, growth, evidence)
	if state.StageGate.Status != "ready" {
		t.Fatalf("active validator should unlock sustained quality gate, got %+v", state.StageGate)
	}
}

func TestServiceQualityPressureFollowsGateOrderOverPlanMentions(t *testing.T) {
	plan := map[string]string{
		"Product":       "Tiny Release Ledger",
		"Current Stage": "Service Quality",
		"MVP":           "Append one release note and one validation result.",
		"Constraints":   "No secrets, no telemetry, deterministic smoke command.",
		"Success Criteria": strings.Join([]string{
			"Validation, security, deployment, docs, rollback, and maintainability are all required before handoff.",
			"Reference comparison should prove the category baseline.",
		}, " "),
	}

	state := deriveReadinessState(plan, growthState{}, nil)
	if state.NextPressure.Axis != "validation_coverage" {
		t.Fatalf("expected service-quality pressure to start at validation coverage, got %+v", state.NextPressure)
	}
	if state.NextPressure.Status != "emerging" {
		t.Fatalf("expected mentioned validation to remain emerging until evidence exists, got %+v", state.NextPressure)
	}
}

func TestTinyMVPPressureFollowsGateOrderOverPlanMentions(t *testing.T) {
	plan := map[string]string{
		"Product":          "Active Guard CLI",
		"Current Stage":    "Tiny MVP",
		"MVP":              "Create one handoff packet and require evidence before the next one.",
		"Success Criteria": "A second run is blocked until the active packet is completed.",
	}

	state := deriveReadinessState(plan, growthState{}, nil)
	if state.NextPressure.Axis != "core_ux" {
		t.Fatalf("expected Tiny MVP pressure to prove the useful flow before validation, got %+v", state.NextPressure)
	}
	if state.NextPressure.Status != "emerging" {
		t.Fatalf("expected mentioned core flow to remain emerging until evidence exists, got %+v", state.NextPressure)
	}
}

func TestServiceQualityPressureWalksRequiredAxesInOrder(t *testing.T) {
	plan := map[string]string{
		"Product":       "Axis Walk CLI",
		"Current Stage": "Service Quality",
		"MVP":           "Create one handoff entry, validate it, and show the latest handoff state.",
		"Constraints":   "No secrets, no telemetry, no network dependency during normal use.",
	}
	evidence := []readinessEvidenceRecord{}
	assertNext := func(want string, growth growthState) {
		t.Helper()
		state := deriveReadinessState(plan, growth, evidence)
		if state.NextPressure.Axis != want {
			t.Fatalf("expected next pressure %s, got %+v", want, state.NextPressure)
		}
	}

	assertNext("validation_coverage", growthState{})
	evidence = append(evidence, readinessEvidenceRecordForAxis("GOAL-0001", "validation_coverage", "Validation coverage: `go test ./...` passed and the handoff smoke command is repeatable."))
	assertNext("security_baseline", growthState{})
	evidence = append(evidence, readinessEvidenceRecordForAxis("GOAL-0002", "security_baseline", "Security baseline: Privacy boundary verified, no cloud sync, no telemetry, no token storage, no secrets, and local-only data handling is explicit."))
	assertNext("deployment_readiness", growthState{})
	evidence = append(evidence, readinessEvidenceRecordForAxis("GOAL-0003", "deployment_readiness", "Deployment readiness: Built the CLI binary and ran the smoke command outside the development command."))
	assertNext("operations_docs", growthState{})
	evidence = append(evidence, readinessEvidenceRecordForAxis("GOAL-0004", "operations_docs", "Operations and docs: README handoff notes cover setup, rollback, recovery, and the smoke command."))
	assertNext("maintainability", growthState{})
	evidence = append(evidence, readinessEvidenceRecordForAxis("GOAL-0005", "maintainability", "Maintainability: Table-driven validation helper keeps command checks repeatable without hidden local context."))
	assertNext("reference_benchmark", growthState{})
	evidence = append(evidence, readinessEvidenceRecordForAxis("GOAL-0006", "reference_benchmark", strings.Join([]string{
		"Category: Local developer handoff CLI.",
		"References: GitHub CLI, Taskfile, Make.",
		"Baseline expectations: documented command, repeatable output, rollback notes, no hidden credentials.",
		"Current comparison: below baseline = none; meets baseline = command/test/docs/rollback; above baseline = packet evidence loop.",
		"Below-baseline gaps: No critical below-baseline gap.",
		"Above-baseline strength: packet evidence loop.",
		"Decision: Service Quality proof can continue.",
	}, "; ")))
	assertNext("sustained_quality", growthState{})

	state := deriveReadinessState(plan, growthState{Candidates: []growthCandidate{{Kind: "validator", Name: "validator-go-test", Status: "active"}}}, evidence)
	if state.NextPressure.Axis != "stage_advancement" || state.StageGate.Status != "ready" {
		t.Fatalf("expected ready stage advancement after active capability, got %+v / %+v", state.NextPressure, state.StageGate)
	}
}

func TestReferenceBenchmarkEvidenceTemplateForBetaAndServiceQuality(t *testing.T) {
	betaEvidence := buildEvidenceDoc("GOAL-0001", "Beta", readinessState{}, growthState{})
	assertContains(t, betaEvidence, "## Reference Benchmark Evidence")
	assertContains(t, betaEvidence, "References: Pending")
	assertContains(t, betaEvidence, "Below-baseline gaps")
	assertContains(t, betaEvidence, "Above-baseline strength")

	serviceEvidence := buildEvidenceDoc("GOAL-0001", "Service Quality", readinessState{}, growthState{})
	assertContains(t, serviceEvidence, "## Reference Benchmark Evidence")

	tinyEvidence := buildEvidenceDoc("GOAL-0001", "Tiny MVP", readinessState{}, growthState{})
	assertNotContains(t, tinyEvidence, "## Reference Benchmark Evidence")

	tasks := buildTasksDoc("GOAL-0001", "Web app", "Service Quality", readinessState{}, growthState{})
	assertContains(t, tasks, "Fill Reference Benchmark Evidence")
}

func TestReferenceBenchmarkTemplateWaitsForPressure(t *testing.T) {
	readiness := readinessState{
		Version: 1,
		Stage:   "Beta",
		Dimensions: []readinessDimension{
			{ID: "security_baseline", Name: "Security baseline", Status: "missing"},
			{ID: "reference_benchmark", Name: "Reference benchmark", Status: "missing"},
		},
		StageGate: readinessStageGate{
			Status:       "not_ready",
			CurrentStage: "Beta",
			NextStage:    "Service Quality",
			RequiredAxes: []string{"validation_coverage", "security_baseline", "deployment_readiness", "operations_docs", "reference_benchmark"},
		},
		NextPressure: readinessPressure{Axis: "security_baseline", AxisName: "Security baseline", Status: "missing"},
	}

	evidence := buildEvidenceDoc("GOAL-0001", "Beta", readiness, growthState{})
	assertContains(t, evidence, "Reference benchmark: Pending.")
	assertNotContains(t, evidence, "## Reference Benchmark Evidence")
	tasks := buildTasksDoc("GOAL-0001", "Local CLI", "Beta", readiness, growthState{})
	assertNotContains(t, tasks, "Fill Reference Benchmark Evidence")
	checklist := doneChecklistDoc("Beta", readiness, growthState{})
	assertNotContains(t, checklist, "Reference Benchmark Evidence lists")

	readiness.NextPressure = readinessPressure{Axis: "reference_benchmark", AxisName: "Reference benchmark", Status: "missing"}
	evidence = buildEvidenceDoc("GOAL-0002", "Beta", readiness, growthState{})
	assertContains(t, evidence, "## Reference Benchmark Evidence")
	tasks = buildTasksDoc("GOAL-0002", "Local CLI", "Beta", readiness, growthState{})
	assertContains(t, tasks, "Fill Reference Benchmark Evidence")
}

func TestReferenceBenchmarkEvidenceNotRepeatedAfterCovered(t *testing.T) {
	readiness := readinessState{
		Version: 1,
		Stage:   "Sustained Service Quality",
		Dimensions: []readinessDimension{
			{ID: "reference_benchmark", Name: "Reference benchmark", Status: "covered"},
			{ID: "sustained_quality", Name: "Sustained quality", Status: "covered"},
		},
		StageGate: readinessStageGate{
			Status:       "ready",
			CurrentStage: "Sustained Service Quality",
			NextStage:    "Sustained Service Quality",
			RequiredAxes: []string{"validation_coverage", "operations_docs", "maintainability", "sustained_quality", "reference_benchmark"},
		},
		NextPressure: readinessPressure{Axis: "sustained_quality", AxisName: "Sustained quality", Status: "ongoing"},
	}

	evidence := buildEvidenceDoc("GOAL-0009", "Sustained Service Quality", readiness, growthState{})
	assertNotContains(t, evidence, "## Reference Benchmark Evidence")
	tasks := buildTasksDoc("GOAL-0009", "Go CLI", "Sustained Service Quality", readiness, growthState{})
	assertNotContains(t, tasks, "Fill Reference Benchmark Evidence")
	checklist := doneChecklistDoc("Sustained Service Quality", readiness, growthState{})
	assertNotContains(t, checklist, "Reference Benchmark Evidence lists")
}

func TestEvidenceTemplateNamesActiveCapabilities(t *testing.T) {
	growth := growthState{
		Candidates: []growthCandidate{
			{
				Kind:   "validator",
				Name:   "validator-check-sh",
				Status: "active",
				Signal: "validation pattern: `./check.sh` passed with output: `release-note add/list/error smoke passed`.",
			},
		},
	}
	evidence := buildEvidenceDoc("GOAL-0010", "Sustained Service Quality", readinessState{}, growth)
	assertContains(t, evidence, "## Active Capability Evidence")
	assertContains(t, evidence, "- validator-check-sh: Pending. Required behavior: validation pattern: `./check.sh` passed")
	assertNotContains(t, evidence, "## Active Capability Evidence\n\nPending.")
}

func TestReferenceBenchmarkEvidenceSectionFeedsReadiness(t *testing.T) {
	root := t.TempDir()
	goalDir := filepath.Join(root, ".hyper", "goals", "GOAL-0001")
	if err := os.MkdirAll(goalDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(goalDir, "evidence.md"), strings.Join([]string{
		"# GOAL-0001 Evidence",
		"",
		"## Reference Benchmark Evidence",
		"",
		"- Category: Static journaling app",
		"- References: Journal A, Journal B, Journal C",
		"- Baseline expectations: daily entry, report, setup, and handoff are understandable",
		"- Current comparison: core entry and report meet baseline; release evidence is above baseline",
		"- Below-baseline gaps: None; no critical user or operator baseline gap remains",
		"- Above-baseline strength: local artifact release evidence and explicit handoff notes",
		"- Decision: Service Quality is allowed from the benchmark perspective",
	}, "\n"))

	records, err := loadReadinessEvidence(root, readinessDimensionDefs())
	if err != nil {
		t.Fatal(err)
	}
	record, ok := readinessEvidenceForAxis(records, "reference_benchmark")
	if !ok {
		t.Fatalf("expected reference benchmark readiness record in %+v", records)
	}
	if record.Status != "covered" {
		t.Fatalf("expected reference benchmark section to be covered, got %+v", record)
	}
}

func TestReferenceBenchmarkNestedReferencesCountAsNamedReferences(t *testing.T) {
	evidence := strings.Join([]string{
		"## Reference Benchmark Evidence",
		"",
		"- Category: Location-based social chat with map markers.",
		"- References: 3-5 named references selected, 5 total: Google Maps; KakaoMap; Snap Map; Pokemon GO; Duolingo.",
		"- Named references: Google Maps, KakaoMap, Snap Map, Pokemon GO, Duolingo.",
		"- Baseline expectations: Pins stay readable at small sizes; the map remains the primary surface; motion adds life without interrupting map use.",
		"- Category baseline: Keep the map readable, keep markers legible at small size, make social presence feel alive.",
		"- Current comparison: below baseline = none; meets baseline = pin readability and timed bubble behavior; above baseline = timed jelly pin chat for social presence.",
		"- No critical below-baseline gap: No critical below-baseline gap and no critical category-baseline gap were found.",
		"- Above-baseline strength: Pickachat has one concrete above-baseline strength: location chat feels alive through jelly mascot pins.",
		"- Decision: Service Quality reference benchmark is covered for the current desktop proof.",
		"- References:",
		"  - Google Maps: map markers must be readable and not obscure the map.",
		"  - KakaoMap: Korean users expect familiar map controls and clear nearby context.",
		"  - Snap Map: social map presence should feel alive instead of static.",
		"  - Pokemon GO: map presence should feel playful and legible while preserving location context.",
		"  - Duolingo: mascot expression should be friendly with a low number of parts.",
	}, "\n")

	record := referenceBenchmarkRecordFromExample(t, evidence)
	if record.Status != "covered" {
		t.Fatalf("expected nested reference benchmark evidence to be covered, got %+v", record)
	}
}

func TestReferenceBenchmarkExampleDocsMatchParser(t *testing.T) {
	body := readFile(t, filepath.Join("..", "..", "docs", "examples", "reference-benchmark.md"))
	covered := markdownCodeBlockAfterHeading(t, body, "Covered Example")
	emerging := markdownCodeBlockAfterHeading(t, body, "Emerging Example")
	blocked := markdownCodeBlockAfterHeading(t, body, "Blocked Example")

	coveredRecord := referenceBenchmarkRecordFromExample(t, covered)
	if coveredRecord.Status != "covered" {
		t.Fatalf("expected covered example to parse as covered, got %+v", coveredRecord)
	}

	emergingRecord := referenceBenchmarkRecordFromExample(t, emerging)
	if emergingRecord.Status != "emerging" {
		t.Fatalf("expected emerging example to parse as emerging, got %+v", emergingRecord)
	}
	if !strings.Contains(emergingRecord.Quality, "reference benchmark needs") {
		t.Fatalf("expected emerging example to report missing benchmark requirements, got %+v", emergingRecord)
	}

	blockedRecords := readinessEvidenceRecordsFromGoalText("GOAL-DOC", "# GOAL-DOC Evidence\n\n"+blocked)
	if record, ok := readinessEvidenceForAxis(blockedRecords, "reference_benchmark"); ok && record.Status == "covered" {
		t.Fatalf("blocked example must not be covered, got %+v", record)
	}
	assertContains(t, blocked, "recovery is below baseline")
	assertContains(t, blocked, "Service Quality is blocked")

	koBody := readFile(t, filepath.Join("..", "..", "docs", "examples", "reference-benchmark_ko.md"))
	assertContains(t, koBody, "## Covered")
	assertContains(t, koBody, "## Status")
}

func TestReadinessEvidenceRequiresAxisLabelAndCoversMySQLPersistence(t *testing.T) {
	defs := readinessDimensionDefs()
	if _, ok := parseReadinessEvidenceLine("GOAL-0001", "Local MySQL proof for browser-created pin returned pin test.", defs); ok {
		t.Fatal("generic validation line should not infer a readiness axis")
	}
	record, ok := parseReadinessEvidenceLine("GOAL-0001", "Data persistence: API smoke persisted a pin/message and MySQL confirmed the browser-created row.", defs)
	if !ok {
		t.Fatal("expected labeled MySQL persistence evidence to parse")
	}
	if record.Axis != "persistence" || record.Status != "covered" {
		t.Fatalf("expected covered persistence evidence, got %+v", record)
	}
}

func TestReadinessEvidenceCoversFileBackedPersistence(t *testing.T) {
	record, ok := parseReadinessEvidenceLine("GOAL-0002", "Data persistence: `.release_notes.json` stored the release note and a separate `go run . list` command re-read it after the add command exited.", readinessDimensionDefs())
	if !ok {
		t.Fatal("expected labeled file persistence evidence to parse")
	}
	if record.Axis != "persistence" || record.Status != "covered" {
		t.Fatalf("expected covered persistence evidence, got %+v", record)
	}

	textFileRecord, ok := parseReadinessEvidenceLine("GOAL-0003", "Data persistence: `notes.txt` stores the added note and `./check.sh` reads it back through a separate list command before export.", readinessDimensionDefs())
	if !ok {
		t.Fatal("expected labeled txt persistence evidence to parse")
	}
	if textFileRecord.Axis != "persistence" || textFileRecord.Status != "covered" {
		t.Fatalf("expected covered txt persistence evidence, got %+v", textFileRecord)
	}
}

func TestReadinessEvidenceCoversRejectedInputErrorHandling(t *testing.T) {
	record, ok := parseReadinessEvidenceLine("GOAL-0001", "Error handling: Empty list returns `no notes`, unsafe input containing a secret-like value is rejected, and the smoke command proves both paths.", readinessDimensionDefs())
	if !ok {
		t.Fatal("expected rejected input error handling evidence to parse")
	}
	if record.Axis != "error_handling" || record.Status != "covered" {
		t.Fatalf("expected covered rejected input error evidence, got %+v", record)
	}
}

func TestReadinessEvidenceCoversPrivacyBoundaryAsSecurityBaseline(t *testing.T) {
	defs := readinessDimensionDefs()
	record, ok := parseReadinessEvidenceLine("GOAL-0001", "Privacy boundary: clipboard content stays local in SQLite, no cloud sync or telemetry, and sensitive text can be deleted locally.", defs)
	if !ok {
		t.Fatal("expected privacy boundary evidence to parse")
	}
	if record.Axis != "security_baseline" || record.Status != "covered" {
		t.Fatalf("expected covered security baseline evidence from privacy boundary, got %+v", record)
	}
}

func TestSustainedQualityEvidenceDoesNotTreatNotActiveAsCovered(t *testing.T) {
	defs := readinessDimensionDefs()
	record, ok := parseReadinessEvidenceLine("GOAL-0002", "Sustained quality: Repeated runtime evidence exists for the same handoff validation pattern, but it is not active required behavior yet.", defs)
	if !ok {
		t.Fatal("expected sustained quality evidence to parse")
	}
	if record.Axis != "sustained_quality" || record.Status != "emerging" {
		t.Fatalf("expected emerging sustained quality evidence, got %+v", record)
	}

	covered, ok := parseReadinessEvidenceLine("GOAL-0004", "Sustained quality: Active validator validator-go-test is required and verified before every packet handoff.", defs)
	if !ok {
		t.Fatal("expected active sustained quality evidence to parse")
	}
	if covered.Axis != "sustained_quality" || covered.Status != "covered" {
		t.Fatalf("expected covered sustained quality evidence, got %+v", covered)
	}
}

func TestGrowthIgnoresNoIssueAndNoChangeSignals(t *testing.T) {
	root := t.TempDir()
	if err := ensureProjectLayout(root); err != nil {
		t.Fatalf("layout failed: %v", err)
	}
	db, err := openDB(root)
	if err != nil {
		t.Fatalf("db open failed: %v", err)
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		t.Fatalf("schema failed: %v", err)
	}
	insertTestMemory(t, db, "pattern", "GOAL-0001 readiness evidence: Deployment readiness: Not changed in this episode; MapLibre bundle-size warning remains expected.")
	insertTestMemory(t, db, "pattern", "GOAL-0002 readiness evidence: Security baseline: No auth, secrets, privileged flows, or third-party write surfaces were added.")
	insertTestMemory(t, db, "failure", "GOAL-0003 learn failure: None in this episode.")
	insertRawTestMemory(t, db, "failure", "GOAL-0004 learn failure: None in this run.", "durable")
	insertRawTestMemory(t, db, "failure", "GOAL-0005 blocked: Clear: implementation and validation completed for this packet.", "durable")
	insertRawTestMemory(t, db, "failure", "GOAL-0006 learn failure: None critical for the local-only CLI category.", "durable")

	state, hyperErr := updateGrowthState(root, db)
	if hyperErr != nil {
		t.Fatalf("growth failed: %v", hyperErr)
	}
	if len(state.Pressures) != 0 {
		t.Fatalf("expected no pressures for no-op signals, got %+v", state.Pressures)
	}
	if len(state.Candidates) != 0 {
		t.Fatalf("expected no candidates for no-op signals, got %+v", state.Candidates)
	}
}

func TestSimilarContextIsCompacted(t *testing.T) {
	longText := strings.Repeat("very long runtime context ", 30) + "tail-marker"
	out := formatSimilarContext([]similarContext{
		{Source: "goal", ID: "GOAL-0001", Kind: "goal", Score: 0.9, Text: longText},
		{Source: "goal", ID: "GOAL-0002", Kind: "goal", Score: 0.8, Text: longText},
		{Source: "goal", ID: "GOAL-0003", Kind: "goal", Score: 0.7, Text: longText},
		{Source: "goal", ID: "GOAL-0004", Kind: "goal", Score: 0.6, Text: longText},
	})
	if strings.Count(out, "\n") != 2 {
		t.Fatalf("expected three compacted context lines, got:\n%s", out)
	}
	assertNotContains(t, out, "tail-marker")
}

func TestRuntimePacketCompactsLongContext(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a tiny notes MVP")
	db, err := openDB(root)
	if err != nil {
		t.Fatalf("db open failed: %v", err)
	}
	defer db.Close()
	if err := ensureSchema(db); err != nil {
		t.Fatalf("schema failed: %v", err)
	}
	long := strings.Repeat("notes prior context ", 80)
	if err := insertRun(db, "RUN-0099", long, "Tiny MVP", "completed", nowISO(), "GOAL-0099", long); err != nil {
		t.Fatalf("insert run failed: %v", err)
	}

	if _, err := runCLI(args("run", "notes"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	goal := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "goal.md"))
	if strings.Count(goal, "notes prior context") > 20 {
		t.Fatalf("expected compact prior context, got:\n%s", goal)
	}
	assertContains(t, goal, "...")
}

func TestUpdateURL(t *testing.T) {
	if !strings.Contains(resolveUpdateURL("github:Example/fork"), "https://github.com/Example/fork/releases/latest/download/hyper-") {
		t.Fatalf("bad github update url")
	}
	request := resolveUpdateRequest("github:Example/fork")
	if !strings.Contains(request.ChecksumURL, "https://github.com/Example/fork/releases/latest/download/checksums.txt") {
		t.Fatalf("bad github checksum url: %s", request.ChecksumURL)
	}
	if !strings.Contains(request.SignatureURL, "https://github.com/Example/fork/releases/latest/download/") || !strings.HasSuffix(request.SignatureURL, ".sigstore.json") {
		t.Fatalf("bad github signature url: %s", request.SignatureURL)
	}
	if !strings.Contains(request.SignatureIdentityRegexp, "https://github.com/Example/fork/.github/workflows/release.yml@refs/tags/v.*") {
		t.Fatalf("bad github signature identity: %s", request.SignatureIdentityRegexp)
	}
	if request.AssetName != updateAssetName() {
		t.Fatalf("bad github asset name: %s", request.AssetName)
	}
	if resolveUpdateURL("https://example.com/hyper") != "https://example.com/hyper" {
		t.Fatalf("explicit URL should pass through")
	}
	t.Setenv("HYPER_RUN_CHECKSUM_URL", "https://example.com/checksums.txt")
	explicit := resolveUpdateRequest("https://example.com/download/hyper-windows-amd64.exe?token=1")
	if explicit.AssetName != "hyper-windows-amd64.exe" {
		t.Fatalf("bad explicit asset name: %s", explicit.AssetName)
	}
	if explicit.ChecksumURL != "https://example.com/checksums.txt" {
		t.Fatalf("bad explicit checksum url: %s", explicit.ChecksumURL)
	}
	if explicit.SignatureURL != "" {
		t.Fatalf("explicit URL should not infer signature URL: %s", explicit.SignatureURL)
	}
	out, err := runCLI(args("update", "https://example.com/hyper"), testRoot(t.TempDir()), fakeUpdater{})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	assertContains(t, out.Stdout, "Installed executable: /tmp/fake-hyper")
	assertContains(t, out.Stdout, "Hyper Run update completed")
}

type testRoot string

func (r testRoot) root() string { return string(r) }

type fakeUpdater struct{}

func (fakeUpdater) update(updateRequest) (updateResult, error) {
	return updateResult{Target: "/tmp/fake-hyper"}, nil
}

func mustRun(t *testing.T, root string, values ...string) {
	t.Helper()
	if _, err := runCLI(args(values...), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("%v failed: %v", values, err)
	}
}

func mustInitWithPlan(t *testing.T, root, product, focus string) {
	t.Helper()
	mustRun(t, root, "init")
	writeFile(t, filepath.Join(root, "plan.md"), testPlan(product, focus))
}

func testPlan(product, focus string) string {
	return "# Product Plan\n\n## Product\n\n" + product + "\n\n## Target Users\n\nSolo builders\n\n## MVP\n\n" + focus + "\n\n## Current Stage\n\nTiny MVP\n\n## Build Style\n\nCLI\n\n## Non-goals\n\nProduction hardening\n\n## Constraints\n\nLocal first\n\n## Success Criteria\n\nOne useful flow works\n\n## Current Focus\n\n" + focus + "\n"
}

func args(values ...string) []string {
	return values
}

func assertContains(t *testing.T, value, expected string) {
	t.Helper()
	if !strings.Contains(value, expected) {
		t.Fatalf("expected %q to contain %q", value, expected)
	}
}

func assertNotContains(t *testing.T, value, unexpected string) {
	t.Helper()
	if strings.Contains(value, unexpected) {
		t.Fatalf("expected %q not to contain %q", value, unexpected)
	}
}

func markdownCodeBlockAfterHeading(t *testing.T, body, heading string) string {
	t.Helper()
	marker := "## " + heading
	sectionStart := strings.Index(body, marker)
	if sectionStart < 0 {
		t.Fatalf("heading %q not found", heading)
	}
	afterHeading := body[sectionStart+len(marker):]
	fenceStart := strings.Index(afterHeading, "```")
	if fenceStart < 0 {
		t.Fatalf("code fence after heading %q not found", heading)
	}
	afterFence := afterHeading[fenceStart+len("```"):]
	firstNewline := strings.Index(afterFence, "\n")
	if firstNewline < 0 {
		t.Fatalf("code fence after heading %q has no body", heading)
	}
	afterFence = afterFence[firstNewline+1:]
	fenceEnd := strings.Index(afterFence, "```")
	if fenceEnd < 0 {
		t.Fatalf("closing code fence after heading %q not found", heading)
	}
	return strings.TrimSpace(afterFence[:fenceEnd])
}

func referenceBenchmarkRecordFromExample(t *testing.T, example string) readinessEvidenceRecord {
	t.Helper()
	records := readinessEvidenceRecordsFromGoalText("GOAL-DOC", "# GOAL-DOC Evidence\n\n"+example)
	record, ok := readinessEvidenceForAxis(records, "reference_benchmark")
	if !ok {
		t.Fatalf("expected reference benchmark record from example:\n%s", example)
	}
	return record
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

func insertTestMemory(t *testing.T, db *sql.DB, kind, text string) {
	t.Helper()
	confidence := 0.8
	if ok, err := insertMemoryIfNew(db, memory{Kind: kind, Text: text, Confidence: confidence, Quality: memoryQuality(kind, text, confidence)}); err != nil {
		t.Fatalf("insert memory failed: %v", err)
	} else if !ok {
		t.Fatalf("expected new memory for %s", text)
	}
}

func insertRawTestMemory(t *testing.T, db *sql.DB, kind, text, quality string) {
	t.Helper()
	_, err := db.Exec(`insert into memories (project_id, kind, text, source_event_ids, confidence, quality, created_at, last_used_at, stale_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?)`, "default", kind, text, nil, 0.8, quality, nowISO(), nil, nil)
	if err != nil {
		t.Fatalf("insert raw memory failed: %v", err)
	}
}

type legacyMemoryFixture struct {
	Kind       string  `json:"kind"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Quality    string  `json:"quality"`
}

func readLegacyMemoryFixture(t *testing.T, name string) []legacyMemoryFixture {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", "migrations", name, "memories.json"))
	if err != nil {
		t.Fatal(err)
	}
	var items []legacyMemoryFixture
	if err := json.Unmarshal(body, &items); err != nil {
		t.Fatal(err)
	}
	return items
}
