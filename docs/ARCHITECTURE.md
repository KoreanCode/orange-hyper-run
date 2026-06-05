# Architecture

Hyper Run now keeps the executable entrypoint separate from product runtime logic.

## Package Layout

| Path | Responsibility |
| --- | --- |
| `cmd/hyper` | Thin native executable entrypoint. It only passes process arguments to the app package and exits with the returned code. |
| `internal/app` | Current Hyper Run application runtime: command routing, project state, runtime packets, finish gate, Learn, Growth, Readiness, SQLite storage, repair, update, next-packet planning, and doctor checks. |
| `internal/buildinfo` | Release build metadata injected by the release workflow and displayed by `hyper version`. |
| `internal/stage` | Canonical stage vocabulary, target-stage aliases, and stage ordering used by plan parsing, auto targets, readiness, and status output. |

## Direction

This is the first package boundary. It removes product logic from `cmd/hyper` without changing behavior.

Most domain behavior still lives in `internal/app`, but stage vocabulary is now separated because target-stage behavior is used across plan parsing, readiness, auto continuation, and status checks. Further splits should stay this small and pressure-driven.

Future package splits should happen only around proven pressure:

- `internal/storage` when SQLite schema and queries become harder to evolve safely.
- `internal/runtime` when runtime packet generation needs independent tests and adapters.
- `internal/learn`, `internal/growth`, and `internal/readiness` when those loops need separate policy evolution.
- `internal/project` when plan, layout, repair, and migration behavior need clearer ownership.

## Rule

`cmd/hyper` should stay small. New product behavior belongs under `internal/` first, then moves into a narrower package only when repeated maintenance pressure proves the boundary.
