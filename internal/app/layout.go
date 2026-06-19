package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ensureProjectLayout(root string) *hyperError {
	for _, rel := range []string{
		".agents/skills/hyper",
		".agents/skills/hyper-run",
		".hyper",
		".hyper/commands",
		".hyper/capabilities/candidates",
		".hyper/capabilities/candidates/harness",
		".hyper/capabilities/candidates/skill",
		".hyper/capabilities/candidates/validator",
		".hyper/capabilities/active",
		".hyper/capabilities/active/harness",
		".hyper/capabilities/active/skill",
		".hyper/capabilities/active/validator",
		".hyper/capabilities/retired",
		".hyper/capabilities/retired/harness",
		".hyper/capabilities/retired/skill",
		".hyper/capabilities/retired/validator",
		".hyper/growth",
		".hyper/readiness",
		".hyper/logs",
		".hyper/goals",
		".hyper/memories",
		".hyper/verified-evidence",
		".hyper/skills/generated",
		".hyper/agents/candidates",
		".hyper/agents/active",
		".hyper/agents/retired",
		".hyper/agent_trials",
		".hyper/harnesses/generated",
		".hyper/validators/generated",
	} {
		if err := os.MkdirAll(filepath.Join(root, rel), 0755); err != nil {
			return ioError(err)
		}
	}
	return nil
}

func ensureCodexDesktopRules(root string) *hyperError {
	if err := ensureAgentsGuide(root); err != nil {
		return err
	}
	if err := ensureGeneratedFile(filepath.Join(root, ".agents", "skills", "hyper", "SKILL.md"), hyperRouterSkillGuide()); err != nil {
		return err
	}
	if err := ensureGeneratedFile(filepath.Join(root, ".agents", "skills", "hyper-run", "SKILL.md"), hyperRunSkillGuide()); err != nil {
		return err
	}
	if err := ensureGeneratedFile(filepath.Join(root, hyperDir, "codex-desktop.md"), codexDesktopGuide()); err != nil {
		return err
	}
	return ensureGeneratedFile(filepath.Join(root, hyperDir, "commands", "hyper-run.md"), hyperRunCommandGuide())
}

func ensureAgentsGuide(root string) *hyperError {
	path := filepath.Join(root, "AGENTS.md")
	section := agentsHyperRunSection()
	if !exists(path) {
		return writeText(path, "# Project Instructions\n\n"+section)
	}
	body := readIfExists(path)
	if strings.Contains(body, "<!-- hyper-run:start -->") {
		return replaceHyperRunSection(path, body, section)
	}
	prefix := "\n\n"
	if strings.HasSuffix(body, "\n") {
		prefix = "\n"
	}
	return appendText(path, prefix+section)
}

func agentsHyperRunSection() string {
	return strings.Join([]string{
		"<!-- hyper-run:start -->",
		"## Hyper Run",
		"",
		"When the user writes `$hyper`, `$hyper run`, `$hyper-run`, `$hyper status`, `$hyper status --short`, `$hyper verify`, `$hyper migrate`, `$hyper advance`, `$hyper doctor`, `hyper run`, or asks Hyper Run to continue the project, treat it as a project workflow command inside the current Codex session.",
		"",
		"Use `.agents/skills/hyper/SKILL.md` as the thin Codex Desktop router. Keep product judgment, execution state, learning, and generated project knowledge in `plan.md`, `.hyper/`, and the `hyper` CLI rather than in static skill text.",
		"",
		"Method: Hyper Run is an evidence-first project growth protocol. Execution logs create pressure, pressure creates candidates, and repeated proof promotes project-specific structure.",
		"",
		"Protocol: Hyper Run runtime packets are agent-agnostic. Codex Desktop is one consumer, but the packet can be read by CLI agents and other coding assistants.",
		"",
		"Principles: No structure before pressure. No stage advancement without evidence. No harness before repeated need. No memory without reusable signal.",
		"",
		"Learn role: Learn is not a summary. It extracts what the project repeatedly needed, failed at, or proved from `evidence.md` and `next.md` so future runtime packets can change work boundaries, validation signals, stop conditions, readiness pressure, and capability candidates.",
		"",
		"Required workflow:",
		"",
		"1. Run `hyper run [focus]` only when a new runtime packet is needed; if `plan.md` has `Target Stage`, plain `hyper run` uses it as the guarded auto target until that stage's readiness proof is complete.",
		"2. Read the generated runtime packet path from the CLI output, or read `.hyper/state.json` and use `current_goal_path`.",
		"3. Read `.hyper/goals/<GOAL-ID>/goal.md` and `.hyper/goals/<GOAL-ID>/tasks.md`.",
		"4. Implement the smallest coherent step that satisfies the current episode.",
		"5. Run the safest available validation or record why validation is blocked; prefer `hyper verify -- <command>` for repeatable command proof.",
		"6. Update `.hyper/goals/<GOAL-ID>/evidence.md` with validation output or Verified Evidence IDs, readiness evidence, active capability evidence, pressure signals, changed files, decisions, reusable patterns, and blockers.",
		"7. Write `.hyper/goals/<GOAL-ID>/next.md` with the next recommended runtime episode and Learn Notes.",
		"8. Run the agent finish gate with `hyper complete`; if it fails, fix the same packet using `review.md` before continuing.",
		"9. In auto mode, read `.hyper/next-packet.md`, obey its Guard and Progress Guard, and continue only through the planned next command: `run` continues, `advance` requires Stage Advancement Review authorization or user acceptance, `complete-current` fixes review.md/evidence.md/next.md in the same packet, and `stop` reports the stop reason and waits.",
		"10. Do not start another `hyper run` until evidence, next notes, and the agent finish gate are done.",
		"",
		"Use `hyper init` only for project setup. Do not pass the project objective to `hyper init`; put product context in `plan.md` and use `hyper run [focus]` for the current execution focus.",
		"",
		"Use `hyper status --short` when the user wants the current stage, gate, proof, and next action without the full pressure ledger.",
		"",
		"Use `hyper migrate` when project state, growth rules, or generated candidates need to be refreshed after a CLI update.",
		"",
		"Use `hyper run --auto --until <stage> [focus]` only to override the `plan.md` target. Auto mode still requires finish-gate evidence and target-stage readiness proof; if `.hyper/next-packet.md` says `advance`, continue only after the Stage Advancement Review shows ready proof and no blocking gaps.",
		"",
		"Use `hyper advance` only when `hyper status` says the stage gate is ready and either `.hyper/next-packet.md` is continuing an active auto target or the user accepts the stage change.",
		"",
		"Use `hyper doctor` when install, PATH, project state, SQLite, or Codex routing looks wrong.",
		"<!-- hyper-run:end -->",
		"",
	}, "\n")
}

func replaceHyperRunSection(path, current, section string) *hyperError {
	startMarker := "<!-- hyper-run:start -->"
	endMarker := "<!-- hyper-run:end -->"
	start := strings.Index(current, startMarker)
	end := strings.Index(current, endMarker)
	if start == -1 || end == -1 || end < start {
		return nil
	}
	end += len(endMarker)
	next := current[:start] + strings.TrimRight(section, "\n") + current[end:]
	if !strings.HasSuffix(next, "\n") {
		next += "\n"
	}
	if next == current {
		return nil
	}
	return writeText(path, next)
}

func ensureGeneratedFile(path, body string) *hyperError {
	if readIfExists(path) == body {
		return nil
	}
	return writeText(path, body)
}

func hyperRouterSkillGuide() string {
	return strings.Join([]string{
		"---",
		"name: hyper",
		"description: Thin Codex Desktop router for Hyper Run. Use when the user says $hyper, $hyper run, $hyper init, $hyper status, $hyper status --short, $hyper verify, $hyper migrate, $hyper advance, $hyper doctor, $hyper resume, hyper run, or asks Hyper Run to continue the current project.",
		"---",
		"",
		"# Hyper Router",
		"",
		"This skill is only a Codex Desktop compatibility shim. Do not move product strategy, learning, validation policy, generated harnesses, or project-specific execution knowledge into this file.",
		"",
		"Source of truth:",
		"- `plan.md` for the human-owned product brief.",
		"- `.hyper/` for goals, evidence, logs, memory, generated candidates, and runtime state.",
		"- The `hyper` CLI for creating or resuming runtime packets.",
		"",
		"Method:",
		"- " + growthRuntimeDefinition,
		"- " + runtimeProtocolDefinition,
		"- Growth order: " + growthLoopDefinition,
		"",
		"Principles:",
		"- No structure before pressure.",
		"- No stage advancement without evidence.",
		"- No harness before repeated need.",
		"- No memory without reusable signal.",
		"",
		"Learn role:",
		"- Learn is not a summary of the last run.",
		"- Learn extracts what the project repeatedly needed, failed at, or proved from `evidence.md` and `next.md`.",
		"- Future runtime packets use those signals to change work boundaries, validation signals, stop conditions, readiness pressure, and capability candidates.",
		"",
		"Command mapping:",
		"- `$hyper init`: run `hyper init` in the current project root. Ask the user to review `plan.md` before deep implementation.",
		"- `$hyper run [focus]`: run `hyper run [focus]`; if `plan.md` has `Target Stage`, plain `hyper run` uses it as the guarded auto target until that target stage's readiness proof is complete. Read the generated runtime packet, implement it in the current Codex session, update `evidence.md`, and write `next.md`.",
		"- `$hyper run --auto --until <stage> [focus]`: run `hyper run --auto --until <stage> [focus]` as an explicit target override, then continue packet by packet using `.hyper/next-packet.md` until the target stage proof is complete or a guard stops progress.",
		"- `$hyper complete`: advanced/recovery command. Run it only as the agent finish gate after evidence and next notes are written so project readiness is refreshed.",
		"- `$hyper verify -- <command>`: run repeatable validation through the CLI so exit code, log hashes, commit SHA, worktree status hash, goal ID, and run ID are recorded under `.hyper/verified-evidence/`.",
		"- `$hyper status`: run `hyper status` and use the dashboard to decide whether the agent should finish the packet, repair, advance, migrate, or start the next packet.",
		"- `$hyper status --short`: run `hyper status --short` when the user wants only the current stage, gate, proof, and next action.",
		"- `$hyper migrate`: run `hyper migrate` after CLI updates or when growth state/candidates look stale; then check `hyper status --short`.",
		"- `$hyper advance`: run `hyper advance` only after `hyper status` shows the stage gate is ready and either `.hyper/next-packet.md` is continuing an active auto target or the user accepts the stage change.",
		"- `$hyper doctor`: run `hyper doctor` and use the diagnostics to fix install, PATH, project state, or routing issues.",
		"- `$hyper resume`: run `hyper resume`, read the active runtime packet path, and continue the same evidence and next-step rules.",
		"- `hyper run [focus]`: treat this the same as `$hyper run [focus]` when the user is speaking inside Codex Desktop.",
		"",
		"Execution rules:",
		"1. Run a CLI command only when a new or resumed runtime packet is needed; if `plan.md` has `Target Stage`, plain `hyper run` uses it as the guarded auto target until that target stage's readiness proof is complete.",
		"2. Read the generated runtime packet in `goal.md` and the checklist in `tasks.md` before editing project files.",
		"3. Keep implementation scoped to the current runtime episode.",
		"4. Run the safest available validation, or record why validation is blocked. Prefer `hyper verify -- <command>` when a real command can prove the behavior.",
		"5. Update the active runtime packet's `evidence.md` with changed files, validation output or Verified Evidence IDs, readiness evidence, active capability evidence, pressure signals, decisions, reusable patterns, and blockers.",
		"6. Write the active runtime packet's `next.md` with the next recommended runtime episode and Learn Notes.",
		"7. Run the agent finish gate with `hyper complete`; if it fails, fix the same packet using `review.md` before continuing.",
		"8. In auto mode, read `.hyper/next-packet.md`, obey its Guard and Progress Guard, and continue only through the planned next command: `run` continues, `advance` requires Stage Advancement Review authorization or user acceptance, `complete-current` fixes review.md/evidence.md/next.md in the same packet, and `stop` reports the stop reason and waits.",
		"9. Do not start another `hyper run` before evidence, next notes, and the agent finish gate are done.",
		"10. Use `hyper advance` only when readiness says the gate is ready and either the planned auto target authorizes continuation or the user accepts the stage change.",
		"",
	}, "\n")
}

func hyperRunSkillGuide() string {
	return strings.Join([]string{
		"---",
		"name: hyper-run",
		"description: Command-style entry point for running the next Hyper Run runtime packet. Use when the user says $hyper-run, $hyper run, hyper run, or asks to execute Hyper Run in the current repository.",
		"---",
		"",
		"# Hyper Run",
		"",
		"Use this skill as the direct run entry point. For `$hyper run`, the router skill at `.agents/skills/hyper/SKILL.md` may trigger first; both paths must lead to the same runtime flow.",
		"",
		"Behavior:",
		"- Treat `$hyper-run`, `$hyper run`, and `hyper run` as Codex-native workflow entry points for the current repository.",
		"- Keep the growth order explicit: " + growthLoopDefinition,
		"- Run `hyper run [focus]` when a new runtime packet is needed; `plan.md` `Target Stage` makes plain `hyper run` use guarded auto continuation until that target stage's readiness proof is complete and keeps the continuation command as plain `hyper run`.",
		"- Run `hyper run --auto --until <stage> [focus]` when the user wants to override the plan target.",
		"- Read the generated runtime packet at `.hyper/goals/<GOAL-ID>/goal.md` and `tasks.md` before implementation.",
		"- Implement the work directly in the current Codex session.",
		"- Run the safest available validation or record why validation is blocked; prefer `hyper verify -- <command>` for repeatable command proof.",
		"- Update `evidence.md` with validation output or Verified Evidence IDs, readiness evidence, active capability evidence, pressure signals, changed files, decisions, reusable patterns, and blockers.",
		"- Write `next.md` with the next recommended runtime episode and Learn Notes.",
		"- Run `hyper complete` internally as the agent finish gate after evidence and next notes are written; if it fails, fix the same packet using `review.md`.",
		"- In auto mode, read `.hyper/next-packet.md`, obey its Guard and Progress Guard, and continue through the planned command until a guard stops progress.",
		"- If `.hyper/next-packet.md` says `Action: run`, execute only its `Command` and continue the next packet.",
		"- If it says `Action: advance`, continue only when the Stage Advancement Review authorizes the active auto target or the user accepts the stage change.",
		"- If it says `Action: complete-current`, stay in the same packet and fix evidence.md, next.md, and review.md findings.",
		"- If it says `Action: stop`, report the reason shown in `.hyper/next-packet.md`; this may be target proof complete, blocked, waiting for user input, or another stop condition.",
		"- Do not start another `hyper run` until evidence, next notes, and the agent finish gate are done.",
		"",
	}, "\n")
}

func codexDesktopGuide() string {
	return strings.Join([]string{
		"# Hyper Run for Codex Desktop",
		"",
		"Use these local project rules when the user writes a Hyper Run command in Codex Desktop.",
		"",
		"`$hyper` is a thin router skill. It exists only so Codex Desktop can catch `$hyper run`; product strategy and execution memory stay in `plan.md`, `.hyper/`, and the `hyper` CLI.",
		"",
		"## Method",
		"",
		growthRuntimeDefinition,
		"",
		"Runtime protocol: " + runtimeProtocolDefinition,
		"",
		"Growth order: " + growthLoopDefinition,
		"",
		"Principles:",
		"",
		"- No structure before pressure.",
		"- No stage advancement without evidence.",
		"- No harness before repeated need.",
		"- No memory without reusable signal.",
		"",
		"## Learn Role",
		"",
		"Learn is not a summary. Learn extracts what the project repeatedly needed, failed at, or proved from `evidence.md` and `next.md`. Future runtime packets use those signals to change work boundaries, validation signals, stop conditions, readiness pressure, and capability candidates.",
		"",
		"## $hyper init",
		"",
		"1. Run `hyper init` in the project root when the project is not initialized.",
		"2. Ask the user to fill in or review `plan.md` before deep implementation.",
		"3. Do not put the project objective after `hyper init`; use `plan.md` and `hyper run [focus]` instead.",
		"4. Do not overwrite an existing active runtime packet.",
		"",
		"## $hyper run",
		"",
		"1. Run `hyper run [focus]` in the project root. If `plan.md` has `Target Stage`, plain `hyper run` will use it as the guarded auto target until that stage's readiness proof is complete.",
		"2. Read the runtime packet path from stdout, or read `.hyper/state.json` and use `current_goal_path`.",
		"3. Read the generated `goal.md` runtime packet and `tasks.md` checklist.",
		"4. Work checkpoint by checkpoint toward the current episode.",
		"5. Run the smallest safe validation available; prefer `hyper verify -- <command>` for repeatable command proof.",
		"6. Update `evidence.md` with validation output or Verified Evidence IDs, readiness evidence, active capability evidence, pressure signals, changed files, decisions, reusable patterns, and blockers.",
		"7. Update `next.md` with the next recommended runtime episode and Learn Notes.",
		"8. Run the agent finish gate with `hyper complete`; if it fails, fix the same packet using `review.md` before continuing.",
		"9. In auto mode, read `.hyper/next-packet.md`, obey its Guard and Progress Guard, and continue only through the planned next command: `run` continues, `advance` requires Stage Advancement Review authorization or user acceptance, `complete-current` fixes review.md/evidence.md/next.md in the same packet, and `stop` reports the stop reason and waits.",
		"10. Stop early for destructive actions, missing credentials, unclear product scope, or repeated validation failure.",
		"11. Do not start another `hyper run` before evidence, next notes, and the agent finish gate are done.",
		"",
		"## $hyper doctor",
		"",
		"1. Run `hyper doctor`.",
		"2. Use the diagnostics to resolve install, PATH, project state, SQLite, or Codex routing issues before starting another packet.",
		"",
		"## $hyper advance",
		"",
		"1. Run `hyper status` first and confirm the stage gate is ready.",
		"2. Run `hyper advance` only when the user accepts the stage change, or when `.hyper/next-packet.md` is continuing an active auto target after a ready Stage Advancement Review.",
		"3. After advancement, use the refreshed `hyper status` output to choose the next `hyper run` focus.",
		"",
		"## $hyper resume",
		"",
		"1. Run `hyper resume`.",
		"2. Read the active runtime packet path from the handoff.",
		"3. Continue the same execution rules as `$hyper run`.",
		"",
	}, "\n")
}

func hyperRunCommandGuide() string {
	return "# $hyper run\n\nMeaning: create the next Hyper Run runtime packet, execute the current episode, record evidence, capture Learn signals, and let repeated pressure become future structure. If `plan.md` has `Target Stage`, plain `hyper run` uses it as the guarded auto target until that target stage's readiness proof is complete and remains the continuation command in `.hyper/next-packet.md` until a guard stops progress.\n\nGrowth order: " + growthLoopDefinition + "\n\nPrinciples: " + growthPrinciplesLine() + "\n\nRequired flow:\n\n1. Execute `hyper run [focus]`.\n2. Open the generated runtime packet under `.hyper/goals/<GOAL-ID>/`.\n3. Implement the smallest coherent step that satisfies the current episode in `goal.md`.\n4. Prefer `hyper verify -- <command>` for repeatable command proof, then mark real progress in `evidence.md`, including validation or Verified Evidence IDs, readiness evidence, active capability evidence, pressure signals, decisions, reusable patterns, and blockers.\n5. Write the next recommended runtime episode and Learn Notes in `next.md`.\n6. Run `hyper complete` internally as the agent finish gate to close the packet and refresh Learn, Growth, and Readiness.\n7. Read `.hyper/next-packet.md`, obey its Guard and Progress Guard, and continue only through the planned command: if it says `run`, continue only through that command; if it says `advance`, continue only when the active auto target authorizes it after the Stage Advancement Review, otherwise wait for user review; if it says `complete-current`, fix review.md/evidence.md/next.md in the same packet and rerun the agent finish gate; if it says `stop`, report the stop reason from `.hyper/next-packet.md` and wait.\n\nCompletion requires implementation evidence, Learn signals where applicable, a next recommendation, and a completed Hyper packet.\n"
}

func ensureMemoryFiles(root string) *hyperError {
	for rel, title := range map[string]string{
		".hyper/memories/decisions.md":   "Decisions",
		".hyper/memories/patterns.md":    "Patterns",
		".hyper/memories/failures.md":    "Failures",
		".hyper/memories/constraints.md": "Constraints",
	} {
		path := filepath.Join(root, rel)
		if !exists(path) {
			if err := writeText(path, fmt.Sprintf("# %s\n\n", title)); err != nil {
				return err
			}
		}
	}
	return nil
}

func appendMemoryMarkdown(root string, mem memory) *hyperError {
	rel := ""
	switch mem.Kind {
	case "decision":
		rel = ".hyper/memories/decisions.md"
	case "pattern":
		rel = ".hyper/memories/patterns.md"
	case "failure":
		rel = ".hyper/memories/failures.md"
	case "constraint":
		rel = ".hyper/memories/constraints.md"
	default:
		return nil
	}
	quality := firstNonBlank(mem.Quality, "weak")
	return appendText(filepath.Join(root, rel), "- ["+quality+"] "+mem.Text+"\n")
}
