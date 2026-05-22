# Before / After Demo

This is a short practical demo script for explaining Hyper Run in about one minute. It is not a polished GIF yet; it is the source scenario for a future recording.

## Problem

Without project-local runtime memory, a coding agent can lose context between sessions:

```text
Session 1:
- Build a tiny MVP.
- Decide to keep storage local-first.
- Run `npm run build`.
- Notice deployment is not ready.

Session 2:
- The agent starts from a broad prompt again.
- It may forget the local-first decision.
- It may skip the same validation.
- It may try to add a harness or broad architecture too early.
```

## With Hyper Run

Initialize once:

```bash
hyper init
```

Fill in `plan.md`, then run:

```bash
hyper run "Build the smallest usable MVP"
```

Hyper Run creates:

```text
.hyper/goals/GOAL-0001/
  goal.md
  tasks.md
  evidence.md
  next.md
```

After the agent implements the packet, it records evidence:

```text
## Validation

`npm run build` passed.

## Decisions

Keep storage local-first.

## Readiness Evidence

Core UX: Browser smoke passed for create and complete flow.
Validation coverage: `npm run build` passed and is repeatable.
```

Close the packet:

```bash
hyper complete
```

Then inspect:

```bash
hyper status
```

Example result:

```text
Action:
  Next action: hyper run "Review readiness evidence and update plan.md Current Stage to Usable MVP if the evidence is accepted."
  Why now: Tiny MVP gate is ready.
  Do not do yet: Do not treat candidates as active harnesses or validators before promotion.
```

## What Changed

Before Hyper Run, the next AI session depends on a broad prompt and memory in the chat.

After Hyper Run, the next AI session reads project-local state:

- `plan.md` for product intent
- `.hyper/goals/.../evidence.md` for proof
- `.hyper/goals/.../next.md` for the recommended next episode
- `.hyper/growth/state.json` for repeated pressure
- `.hyper/readiness/state.json` for stage readiness

The agent still implements the code, but the project now carries its own context.

## Recording Checklist

- Show a sparse project with only README and source files.
- Run `hyper init`.
- Fill a small `plan.md`.
- Run `hyper run`.
- Show `goal.md`.
- Simulate or perform one small implementation.
- Fill `evidence.md` and `next.md`.
- Run `hyper complete`.
- Run `hyper status`.
- Highlight the `Action` section and the readiness gate.
