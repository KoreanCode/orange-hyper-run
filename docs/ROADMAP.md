# Roadmap

This roadmap describes the current product direction. Hyper Run should stay small at the entrypoint and become stronger only when project evidence proves a need.

## Near Term

- Keep `hyper run` as the main user command.
- Make `hyper status`, `hyper doctor`, `hyper repair`, and `hyper migrate` reliable enough for daily use.
- Improve installer and release trust with checksum verification and CI checks.
- Tighten Learn quality so only durable project signals become memory.
- Make runtime packets easier for coding agents to execute without extra explanation.

## Next

- Add migration coverage for older `.hyper/` project states.
- Improve `hyper update` with checksum verification, matching the installer.
- Add clearer examples for web app, CLI, and desktop app projects.
- Add a release checklist for maintainers.
- Start splitting packages only after the domain boundaries stabilize.

## Later

- Consider cosign signatures for release binaries.
- Add richer project demos, including terminal recordings or short GIFs.
- Add optional agent/capability activation policies for teams that want stricter gates.
- Evaluate package extraction into focused internal packages such as state, storage, growth, readiness, and packet generation.

## Non-Goals For Now

- Do not turn Hyper Run into a full project management app.
- Do not require a harness before the project has repeated evidence.
- Do not make static skills the source of project truth.
- Do not auto-advance project stages without explicit user review.
