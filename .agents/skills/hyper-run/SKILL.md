---
name: hyper-run
description: Command-style entry point for running the next Hyper Run runtime packet. Use when the user says $hyper-run, $hyper run, hyper run, or asks to execute Hyper Run in the current repository.
---

# Hyper Run

Use this skill as the direct run entry point. For `$hyper run`, the router skill at `.agents/skills/hyper/SKILL.md` may trigger first; both paths must lead to the same runtime flow.

Behavior:
- Treat `$hyper-run`, `$hyper run`, and `hyper run` as Codex-native workflow entry points for the current repository.
- Keep the growth order explicit: Execution -> Evidence -> Pressure Ledger -> Candidate -> Structure when proven.
- Run `hyper run [focus]` when a new runtime packet is needed; `plan.md` `Target Stage` makes plain `hyper run` use guarded auto continuation until that target stage's readiness proof is complete and keeps the continuation command as plain `hyper run`.
- Run `hyper run --auto --until <stage> [focus]` when the user wants to override the plan target.
- Read the generated runtime packet at `.hyper/goals/<GOAL-ID>/goal.md` and `tasks.md` before implementation.
- Implement the work directly in the current Codex session.
- Run the safest available validation or record why validation is blocked.
- Update `evidence.md` with validation output, readiness evidence, active capability evidence, pressure signals, changed files, decisions, reusable patterns, and blockers.
- Write `next.md` with the next recommended runtime episode and Learn Notes.
- Run `hyper complete` internally as the agent finish gate after evidence and next notes are written; if it fails, fix the same packet using `review.md`.
- In auto mode, read `.hyper/next-packet.md`, obey its Guard and Progress Guard, and continue through the planned command until a guard stops progress.
- If `.hyper/next-packet.md` says `Action: run`, execute only its `Command` and continue the next packet.
- If it says `Action: advance`, continue only when the Stage Advancement Review authorizes the active auto target or the user accepts the stage change.
- If it says `Action: complete-current`, stay in the same packet and fix evidence.md, next.md, and review.md findings.
- If it says `Action: stop`, report the reason shown in `.hyper/next-packet.md`; this may be target proof complete, blocked, waiting for user input, or another stop condition.
- Do not start another `hyper run` until evidence, next notes, and the agent finish gate are done.
