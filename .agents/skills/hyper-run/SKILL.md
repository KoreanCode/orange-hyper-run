---
name: hyper-run
description: Command-style entry point for running the next Hyper Run runtime packet. Use when the user says $hyper-run, $hyper run, hyper run, or asks to execute Hyper Run in the current repository.
---

# Hyper Run

Use this skill as the direct run entry point. For `$hyper run`, the router skill at `.agents/skills/hyper/SKILL.md` may trigger first; both paths must lead to the same runtime flow.

Behavior:
- Treat `$hyper-run`, `$hyper run`, and `hyper run` as Codex-native workflow entry points for the current repository.
- Run `hyper run [focus]` when a new runtime packet is needed.
- Read the generated runtime packet at `.hyper/goals/<GOAL-ID>/goal.md` and `tasks.md` before implementation.
- Implement the work directly in the current Codex session.
- Run the safest available validation or record why validation is blocked.
- Update `evidence.md` with validation output, readiness evidence, active capability evidence, changed files, decisions, reusable patterns, and blockers.
- Write `next.md` with the next recommended runtime episode and Learn Notes.
- Run `hyper complete` after evidence and next notes are written.
- Do not start another `hyper run` until evidence, next notes, and `hyper complete` are done.
