---
name: hyper
description: Thin Codex Desktop router for Hyper Run. Use when the user says $hyper, $hyper run, $hyper init, $hyper status, $hyper status --short, $hyper migrate, $hyper advance, $hyper doctor, $hyper resume, hyper run, or asks Hyper Run to continue the current project.
---

# Hyper Router

This skill is only a Codex Desktop compatibility shim. Do not move product strategy, learning, validation policy, generated harnesses, or project-specific execution knowledge into this file.

Source of truth:
- `plan.md` for the human-owned product brief.
- `.hyper/` for goals, evidence, logs, memory, generated candidates, and runtime state.
- The `hyper` CLI for creating or resuming runtime packets.

Method:
- Evidence-first project growth protocol: execution logs create pressure, pressure creates candidates, and repeated proof promotes project-specific structure.
- Agent-agnostic runtime packet protocol for Codex, CLI agents, and other coding assistants.
- Growth order: Execution -> Evidence -> Pressure Ledger -> Candidate -> Structure when proven.

Principles:
- No structure before pressure.
- No stage advancement without evidence.
- No harness before repeated need.
- No memory without reusable signal.

Learn role:
- Learn is not a summary of the last run.
- Learn extracts what the project repeatedly needed, failed at, or proved from `evidence.md` and `next.md`.
- Future runtime packets use those signals to change work boundaries, validation signals, stop conditions, readiness pressure, and capability candidates.

Command mapping:
- `$hyper init`: run `hyper init` in the current project root. Ask the user to review `plan.md` before deep implementation.
- `$hyper run [focus]`: run `hyper run [focus]`; if `plan.md` has `Target Stage`, plain `hyper run` uses it as the guarded auto target until that target stage's readiness proof is complete. Read the generated runtime packet, implement it in the current Codex session, update `evidence.md`, and write `next.md`.
- `$hyper run --auto --until <stage> [focus]`: run `hyper run --auto --until <stage> [focus]` as an explicit target override, then continue packet by packet using `.hyper/next-packet.md` until the target stage proof is complete or a guard stops progress.
- `$hyper complete`: run `hyper complete` after evidence and next notes are written so project readiness is refreshed.
- `$hyper status`: run `hyper status` and use the dashboard to decide whether to complete, repair, advance, migrate, or start the next packet.
- `$hyper status --short`: run `hyper status --short` when the user wants only the current stage, gate, proof, and next action.
- `$hyper migrate`: run `hyper migrate` after CLI updates or when growth state/candidates look stale; then check `hyper status --short`.
- `$hyper advance`: run `hyper advance` only after `hyper status` shows the stage gate is ready and either `.hyper/next-packet.md` is continuing an active auto target or the user accepts the stage change.
- `$hyper doctor`: run `hyper doctor` and use the diagnostics to fix install, PATH, project state, or routing issues.
- `$hyper resume`: run `hyper resume`, read the active runtime packet path, and continue the same evidence and next-step rules.
- `hyper run [focus]`: treat this the same as `$hyper run [focus]` when the user is speaking inside Codex Desktop.

Execution rules:
1. Run a CLI command only when a new or resumed runtime packet is needed; if `plan.md` has `Target Stage`, plain `hyper run` uses it as the guarded auto target until that target stage's readiness proof is complete.
2. Read the generated runtime packet in `goal.md` and the checklist in `tasks.md` before editing project files.
3. Keep implementation scoped to the current runtime episode.
4. Run the safest available validation, or record why validation is blocked.
5. Update the active runtime packet's `evidence.md` with changed files, validation output, readiness evidence, active capability evidence, pressure signals, decisions, reusable patterns, and blockers.
6. Write the active runtime packet's `next.md` with the next recommended runtime episode and Learn Notes.
7. Run `hyper complete`; if the finish gate fails, fix the same packet using `review.md` before continuing.
8. In auto mode, read `.hyper/next-packet.md`, obey its Guard and Progress Guard, and continue only through the planned next command: `run` continues, `advance` requires Stage Advancement Review authorization or user acceptance, `complete-current` fixes review.md/evidence.md/next.md in the same packet, and `stop` reports the stop reason and waits.
9. Do not start another `hyper run` before evidence, next notes, and `hyper complete` are done.
10. Use `hyper advance` only when readiness says the gate is ready and either the planned auto target authorizes continuation or the user accepts the stage change.
