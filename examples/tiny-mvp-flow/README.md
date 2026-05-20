# Tiny MVP Flow Example

This example shows the Golden Path for Hyper Run: a tiny product plan becomes one runtime packet, the agent completes the packet, evidence is written, and Learn signals become future context.

The example product is `Pocket Tasks`, a local-first browser task list. The application source is intentionally omitted. This folder focuses on the Hyper Run artifacts that explain the service loop.

## Command Flow

```bash
hyper init
# Fill in plan.md
hyper run "Build the smallest local task list MVP"
# Codex Desktop executes GOAL-0001 and updates evidence.md / next.md
hyper run "Add persistence polish after the core flow works"
```

The second `hyper run` automatically learns from the previous active runtime packet before creating the next one. `hyper internal learn` is available only when you want to inspect or debug learning manually.

In Codex Desktop, the equivalent entrypoint is:

```text
$hyper run Build the smallest local task list MVP
```

## Files To Read

```text
plan.md
.hyper/goals/GOAL-0001/
  goal.md
  tasks.md
  evidence.md
  next.md
.hyper/capabilities/
  candidates/
    validator/
  active/
    validator/
  retired/
    validator/
.hyper/growth/
  state.json
.hyper/readiness/
  state.json
.hyper/memories/
  decisions.md
  patterns.md
  constraints.md
  failures.md
```

## What This Demonstrates

- `plan.md` stays human-owned and lightweight.
- `goal.md` is a runtime packet, not a permanent spec.
- `tasks.md` defines the execution checklist for one episode.
- `evidence.md` records validation, readiness evidence, active capability evidence, changed files, decisions, reusable patterns, and blockers.
- `next.md` recommends the next runtime episode and includes structured Learn Notes.
- `.hyper/memories/` stores durable signals that future packets can retrieve.
- `.hyper/growth/state.json` stores pressure that changes the next packet's boundary and validation behavior.
- `.hyper/readiness/state.json` stores stage-gate readiness pressure so MVP work keeps moving toward service quality.
- `.hyper/capabilities/` stores lifecycle metadata when repeated pressure becomes a candidate, promotable structure, active structure, or retired structure.
- Active validators under `.hyper/capabilities/active/validator/` become required validation behavior in the next runtime packet.

## Golden Path Outcome

After `GOAL-0001`, the project has one working local MVP flow and enough evidence to continue. The next `hyper run` should not rediscover the same choices. It should know that localStorage is the accepted Tiny MVP storage choice, browser smoke testing is the validation pattern, and external services are out of scope until credentials exist.
