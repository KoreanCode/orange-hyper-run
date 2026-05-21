package main

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
	return "<!-- hyper-run:start -->\n## Hyper Run\n\nWhen the user writes `$hyper`, `$hyper run`, `$hyper-run`, `hyper run`, or asks Hyper Run to continue the project, treat it as a project workflow command inside the current Codex session.\n\nUse `.agents/skills/hyper/SKILL.md` as the thin Codex Desktop router. Keep product judgment, execution state, learning, and generated project knowledge in `plan.md`, `.hyper/`, and the `hyper` CLI rather than in static skill text.\n\nLearn role: Learn is not a summary. It extracts durable decisions, reusable implementation patterns, blockers/failures, and constraints from `evidence.md` and `next.md` so future runtime packets can continue with better context.\n\nRequired workflow:\n\n1. Run `hyper run [focus]` only when a new runtime packet is needed.\n2. Read the generated runtime packet path from the CLI output, or read `.hyper/state.json` and use `current_goal_path`.\n3. Read `.hyper/goals/<GOAL-ID>/goal.md` and `.hyper/goals/<GOAL-ID>/tasks.md`.\n4. Implement the smallest coherent step that satisfies the current episode.\n5. Run the safest available validation or record why validation is blocked.\n6. Update `.hyper/goals/<GOAL-ID>/evidence.md` with validation output, readiness evidence, active capability evidence, changed files, decisions, reusable patterns, and blockers.\n7. Write `.hyper/goals/<GOAL-ID>/next.md` with the next recommended runtime episode and Learn Notes.\n8. Run `hyper complete` so Learn, Growth, and Readiness refresh from the completed packet.\n9. Do not start another `hyper run` until evidence, next notes, and `hyper complete` are done.\n\nUse `hyper init` only for project setup. Do not pass the project objective to `hyper init`; put product context in `plan.md` and use `hyper run [focus]` for the current execution focus.\n<!-- hyper-run:end -->\n"
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
		"description: Thin Codex Desktop router for Hyper Run. Use when the user says $hyper, $hyper run, $hyper init, $hyper resume, hyper run, or asks Hyper Run to continue the current project.",
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
		"Learn role:",
		"- Learn is not a summary of the last run.",
		"- Learn extracts durable decisions, reusable implementation patterns, blockers/failures, and constraints from `evidence.md` and `next.md`.",
		"- Future runtime packets retrieve those signals through `.hyper/memories/` and similar-context search.",
		"",
		"Command mapping:",
		"- `$hyper init`: run `hyper init` in the current project root. Ask the user to review `plan.md` before deep implementation.",
		"- `$hyper run [focus]`: run `hyper run [focus]`, read the generated runtime packet, implement it in the current Codex session, update `evidence.md`, and write `next.md`.",
		"- `$hyper complete`: run `hyper complete` after evidence and next notes are written so project readiness is refreshed.",
		"- `$hyper resume`: run `hyper resume`, read the active runtime packet path, and continue the same evidence and next-step rules.",
		"- `hyper run [focus]`: treat this the same as `$hyper run [focus]` when the user is speaking inside Codex Desktop.",
		"",
		"Execution rules:",
		"1. Run a CLI command only when a new or resumed runtime packet is needed.",
		"2. Read the generated runtime packet in `goal.md` and the checklist in `tasks.md` before editing project files.",
		"3. Keep implementation scoped to the current runtime episode.",
		"4. Run the safest available validation, or record why validation is blocked.",
		"5. Update the active runtime packet's `evidence.md` with changed files, validation output, readiness evidence, active capability evidence, decisions, reusable patterns, and blockers.",
		"6. Write the active runtime packet's `next.md` with the next recommended runtime episode and Learn Notes.",
		"7. Run `hyper complete` to close the current packet and refresh Learn, Growth, and Readiness.",
		"8. Do not start another `hyper run` before evidence, next notes, and `hyper complete` are done.",
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
		"- Run `hyper run [focus]` when a new runtime packet is needed.",
		"- Read the generated runtime packet at `.hyper/goals/<GOAL-ID>/goal.md` and `tasks.md` before implementation.",
		"- Implement the work directly in the current Codex session.",
		"- Run the safest available validation or record why validation is blocked.",
		"- Update `evidence.md` with validation output, readiness evidence, active capability evidence, changed files, decisions, reusable patterns, and blockers.",
		"- Write `next.md` with the next recommended runtime episode and Learn Notes.",
		"- Run `hyper complete` after evidence and next notes are written.",
		"- Do not start another `hyper run` until evidence, next notes, and `hyper complete` are done.",
		"",
	}, "\n")
}

func codexDesktopGuide() string {
	return "# Hyper Run for Codex Desktop\n\nUse these local project rules when the user writes a Hyper Run command in Codex Desktop.\n\n`$hyper` is a thin router skill. It exists only so Codex Desktop can catch `$hyper run`; product strategy and execution memory stay in `plan.md`, `.hyper/`, and the `hyper` CLI.\n\n## Learn Role\n\nLearn is not a summary. Learn extracts durable decisions, reusable implementation patterns, blockers/failures, and constraints from `evidence.md` and `next.md`, then stores those signals in `.hyper/memories/` for future runtime packets.\n\n## $hyper init\n\n1. Run `hyper init` in the project root when the project is not initialized.\n2. Ask the user to fill in or review `plan.md` before deep implementation.\n3. Do not put the project objective after `hyper init`; use `plan.md` and `hyper run [focus]` instead.\n4. Do not overwrite an existing active runtime packet.\n\n## $hyper run\n\n1. Run `hyper run [focus]` in the project root.\n2. Read the runtime packet path from stdout, or read `.hyper/state.json` and use `current_goal_path`.\n3. Read the generated `goal.md` runtime packet and `tasks.md` checklist.\n4. Work checkpoint by checkpoint toward the current episode.\n5. Run the smallest safe validation available.\n6. Update `evidence.md` with validation output, readiness evidence, active capability evidence, changed files, decisions, reusable patterns, and blockers.\n7. Update `next.md` with the next recommended runtime episode and Learn Notes.\n8. Run `hyper complete` so Learn, Growth, and Readiness refresh from the completed packet.\n9. Stop early for destructive actions, missing credentials, unclear product scope, or repeated validation failure.\n10. Do not start another `hyper run` before evidence, next notes, and `hyper complete` are done.\n\n## $hyper resume\n\n1. Run `hyper resume`.\n2. Read the active runtime packet path from the handoff.\n3. Continue the same execution rules as `$hyper run`.\n"
}

func hyperRunCommandGuide() string {
	return "# $hyper run\n\nMeaning: create the next Hyper Run runtime packet, execute the current episode, record evidence, capture Learn signals, and leave a useful next step.\n\nRequired flow:\n\n1. Execute `hyper run [focus]`.\n2. Open the generated runtime packet under `.hyper/goals/<GOAL-ID>/`.\n3. Implement the smallest coherent step that satisfies the current episode in `goal.md`.\n4. Mark real progress in `evidence.md`, including validation, readiness evidence, active capability evidence, decisions, reusable patterns, and blockers.\n5. Write the next recommended runtime episode and Learn Notes in `next.md`.\n6. Execute `hyper complete` to close the packet and refresh Learn, Growth, and Readiness.\n\nCompletion requires implementation evidence, Learn signals where applicable, a next recommendation, and a completed Hyper packet.\n"
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
	return appendText(filepath.Join(root, rel), "- "+mem.Text+"\n")
}
