---
name: hyper
description: Thin Codex Desktop router for Hyper Run. Use when the user says $hyper, $hyper run, $hyper init, $hyper doctor, $hyper resume, hyper run, or asks Hyper Run to continue the current project.
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
- `$hyper run [focus]`: run `hyper run [focus]`, read the generated runtime packet, implement it in the current Codex session, update `evidence.md`, and write `next.md`.
- `$hyper complete`: run `hyper complete` after evidence and next notes are written so project readiness is refreshed.
- `$hyper doctor`: run `hyper doctor` and use the diagnostics to fix install, PATH, project state, or routing issues.
- `$hyper resume`: run `hyper resume`, read the active runtime packet path, and continue the same evidence and next-step rules.
- `hyper run [focus]`: treat this the same as `$hyper run [focus]` when the user is speaking inside Codex Desktop.

Execution rules:
1. Run a CLI command only when a new or resumed runtime packet is needed.
2. Read the generated runtime packet in `goal.md` and the checklist in `tasks.md` before editing project files.
3. Keep implementation scoped to the current runtime episode.
4. Run the safest available validation, or record why validation is blocked.
5. Update the active runtime packet's `evidence.md` with changed files, validation output, readiness evidence, active capability evidence, pressure signals, decisions, reusable patterns, and blockers.
6. Write the active runtime packet's `next.md` with the next recommended runtime episode and Learn Notes.
7. Run `hyper complete` to close the current packet and refresh Learn, Growth, and Readiness.
8. Do not start another `hyper run` before evidence, next notes, and `hyper complete` are done.
