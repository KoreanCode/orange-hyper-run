# Project Instructions

<!-- hyper-run:start -->
## Hyper Run

When the user writes `$hyper`, `$hyper run`, `$hyper-run`, `$hyper advance`, `$hyper doctor`, `hyper run`, or asks Hyper Run to continue the project, treat it as a project workflow command inside the current Codex session.

Use `.agents/skills/hyper/SKILL.md` as the thin Codex Desktop router. Keep product judgment, execution state, learning, and generated project knowledge in `plan.md`, `.hyper/`, and the `hyper` CLI rather than in static skill text.

Method: Hyper Run is an evidence-first project growth protocol. Execution logs create pressure, pressure creates candidates, and repeated proof promotes project-specific structure.

Protocol: Hyper Run runtime packets are agent-agnostic. Codex Desktop is one consumer, but the packet can be read by CLI agents and other coding assistants.

Principles: No structure before pressure. No stage advancement without evidence. No harness before repeated need. No memory without reusable signal.

Learn role: Learn is not a summary. It extracts what the project repeatedly needed, failed at, or proved from `evidence.md` and `next.md` so future runtime packets can change work boundaries, validation signals, stop conditions, readiness pressure, and capability candidates.

Required workflow:

1. Run `hyper run [focus]` only when a new runtime packet is needed.
2. Read the generated runtime packet path from the CLI output, or read `.hyper/state.json` and use `current_goal_path`.
3. Read `.hyper/goals/<GOAL-ID>/goal.md` and `.hyper/goals/<GOAL-ID>/tasks.md`.
4. Implement the smallest coherent step that satisfies the current episode.
5. Run the safest available validation or record why validation is blocked.
6. Update `.hyper/goals/<GOAL-ID>/evidence.md` with validation output, readiness evidence, active capability evidence, pressure signals, changed files, decisions, reusable patterns, and blockers.
7. Write `.hyper/goals/<GOAL-ID>/next.md` with the next recommended runtime episode and Learn Notes.
8. Run `hyper complete` so Learn, Growth, and Readiness refresh from the completed packet.
9. Do not start another `hyper run` until evidence, next notes, and `hyper complete` are done.

Use `hyper init` only for project setup. Do not pass the project objective to `hyper init`; put product context in `plan.md` and use `hyper run [focus]` for the current execution focus.

Use `hyper advance` only when `hyper status` says the stage gate is ready and the user accepts the stage change.

Use `hyper doctor` when install, PATH, project state, SQLite, or Codex routing looks wrong.
<!-- hyper-run:end -->
