package main

import (
	"database/sql"
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
	assertContains(t, out.Stdout, "Status: initialized")
	assertContains(t, out.Stdout, "$hyper run")
	assertContains(t, out.Stdout, "Fill in plan.md")
	assertContains(t, readFile(t, filepath.Join(root, "plan.md")), "# Product Plan")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "state.json")), "initialized")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "logs", "project.jsonl")), "project_initialized")
	assertContains(t, readFile(t, filepath.Join(root, "AGENTS.md")), "$hyper run")
	assertContains(t, readFile(t, filepath.Join(root, ".agents", "skills", "hyper", "SKILL.md")), "name: hyper")
	assertContains(t, readFile(t, filepath.Join(root, ".agents", "skills", "hyper", "SKILL.md")), "compatibility shim")
	assertContains(t, readFile(t, filepath.Join(root, ".agents", "skills", "hyper-run", "SKILL.md")), "name: hyper-run")
	assertContains(t, readFile(t, filepath.Join(root, ".agents", "skills", "hyper-run", "SKILL.md")), "hyper run")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "codex-desktop.md")), "$hyper run")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "commands", "hyper-run.md")), "Required flow")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), `"version": 1`)
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), `"pressure_ledger"`)
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), `"No structure before pressure."`)
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "readiness", "state.json")), `"version": 1`)
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
	assertContains(t, goal, "## Stage Runtime Behavior")
	assertContains(t, goal, "## Active Capabilities")
	assertContains(t, goal, "Next readiness pressure")
	assertContains(t, goal, "Capture readiness evidence")
	assertNotContains(t, goal, "## Scope")
	assertNotContains(t, goal, "## Non-goals")
	evidence := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"))
	assertContains(t, evidence, "## Readiness Evidence")
	assertContains(t, evidence, "Core UX: Pending.")
	assertContains(t, evidence, "## Active Capability Evidence")
	assertContains(t, evidence, "## Decisions")
	assertContains(t, evidence, "## Reusable Patterns")
	next := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"))
	assertContains(t, next, "## Learn Notes")
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
	assertContains(t, out.Stdout, "Readiness gate: Tiny MVP -> Usable MVP (ready)")
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

	status, err := runCLI(args("status"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	assertContains(t, status.Stdout, "Readiness gate: Tiny MVP -> Usable MVP (ready)")
	assertContains(t, status.Stdout, "Stage advancement:")
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

func TestAutoLearnFeedsNextGoalContext(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny CRM", "Build a tiny CRM MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\nCustomer records persisted in SQLite. go test passed.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nAdd persisted customer records.\n")

	out, err := runCLI(args("run", "Add persisted customer records"), testRoot(root), fakeUpdater{})
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	assertContains(t, out.Stdout, "Auto learn: completed, inserted 1")
	assertContains(t, out.Stdout, "Similar context: ")
	assertContains(t, strings.ToLower(readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0002", "goal.md"))), "customer records persisted")
	assertContains(t, readFile(t, filepath.Join(root, ".hyper", "logs", "RUN-0001.jsonl")), "auto_learn_completed")
}

func TestGrowthStateChangesNextRuntimePacket(t *testing.T) {
	root := t.TempDir()
	mustInitWithPlan(t, root, "Tiny notes", "Build a local-first notes MVP")
	mustRun(t, root, "run")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\ngo test ./... passed.\n\n## Changed Files\n\ncmd/notes.go\n\n## Decisions\n\nKeep local-first storage.\n\n## Reusable Patterns\n\nRun go test before every runtime packet handoff.\n\n## Blocker\n\nPending.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nAdd note editing polish.\n\n## Learn Notes\n\n- Pattern: Run go test before every runtime packet handoff.\n- Constraint: Do not add external services without credentials.\n")

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
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\ngo test ./... passed.\n\n## Changed Files\n\ncmd/app.go\n\n## Decisions\n\nPending.\n\n## Reusable Patterns\n\nRun go test before every runtime packet handoff.\n\n## Blocker\n\nPending.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nAdd CLI persistence.\n\n## Learn Notes\n\n- Pattern: Run go test before every runtime packet handoff.\n")

	if _, err := runCLI(args("run", "Add CLI persistence"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0002", "evidence.md"), "# GOAL-0002 Evidence\n\n## Validation\n\ngo test ./... passed.\n\n## Changed Files\n\ncmd/storage.go\n\n## Decisions\n\nPending.\n\n## Reusable Patterns\n\nRun go test before every runtime packet handoff.\n\n## Blocker\n\nPending.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0002", "next.md"), "# GOAL-0002 Next\n\n## Recommended Next Goal\n\nPolish CLI output.\n\n## Learn Notes\n\n- Pattern: Run go test before every runtime packet handoff.\n")

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
	assertNotContains(t, readFile(t, filepath.Join(root, ".hyper", "growth", "state.json")), "Required active validator validator-go-test")
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
	assertContains(t, readiness, `"axis": "persistence"`)
	goal := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "goal.md"))
	assertContains(t, goal, "Current gate: Usable MVP -> Beta")
	assertContains(t, goal, "Next readiness pressure: Data persistence")
	assertContains(t, goal, "Make the primary Tiny CRM flow persist real user data")
	assertContains(t, goal, "Capture readiness evidence for Data persistence")
}

func TestReadinessEvidenceProgressesSelectedAxis(t *testing.T) {
	root := t.TempDir()
	mustRun(t, root, "init")
	writeFile(t, filepath.Join(root, "plan.md"), "# Product Plan\n\n## Product\n\nTiny CRM\n\n## Target Users\n\nSolo sellers\n\n## MVP\n\nAdd and revisit customer notes.\n\n## Current Stage\n\nUsable MVP\n\n## Build Style\n\nWeb app\n\n## Non-goals\n\nTeam collaboration\n\n## Constraints\n\nLocal first\n\n## Success Criteria\n\nPrimary customer notes flow works without manual data edits.\n\n## Current Focus\n\nImprove customer notes.\n")

	if _, err := runCLI(args("run"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "evidence.md"), "# GOAL-0001 Evidence\n\n## Validation\n\nBrowser smoke passed.\n\n## Readiness Evidence\n\nData persistence: Customer notes persist across reload using local storage.\n\n## Changed Files\n\nsrc/App.tsx\n\n## Decisions\n\nKeep storage local-first.\n\n## Reusable Patterns\n\nPending.\n\n## Blocker\n\nPending.\n")
	writeFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0001", "next.md"), "# GOAL-0001 Next\n\n## Recommended Next Goal\n\nHandle empty and failure states.\n\n## Learn Notes\n\n- Pattern: Record readiness evidence with an axis label.\n")

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

	if _, err := runCLI(args("run"), testRoot(root), fakeUpdater{}); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	readiness := readFile(t, filepath.Join(root, ".hyper", "readiness", "state.json"))
	assertContains(t, readiness, `"candidate": true`)
	assertContains(t, readiness, `"plan_change": "Current Stage -> Usable MVP"`)
	goal := readFile(t, filepath.Join(root, ".hyper", "goals", "GOAL-0002", "goal.md"))
	assertContains(t, goal, "Stage advancement candidate")
	assertContains(t, goal, "Recommend updating plan.md Current Stage to Usable MVP")
	assertContains(t, goal, "Do not auto-edit plan.md")
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
	kind, value := parseLearnNote("- Failure: None in this episode.")
	if kind != "" || value != "" {
		t.Fatalf("expected no-op failure learn note to be ignored, got %q %q", kind, value)
	}
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
	if resolveUpdateURL("https://example.com/hyper") != "https://example.com/hyper" {
		t.Fatalf("explicit URL should pass through")
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

func (fakeUpdater) update(string) (updateResult, error) {
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
