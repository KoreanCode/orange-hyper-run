# Changelog

## Unreleased

- Add a Service Quality Self Review gate that requires plan alignment, core loop quality, product satisfaction, no drift, validation match, and an explicit pass/fail verdict before packet completion.
- Keep Service Quality packets open for repair when the Self Review verdict is `fail`.
- Add Product satisfaction as a readiness axis for Beta, Service Quality, and Sustained Service Quality gates.
- Add a no-drift runtime guard to packet work boundaries and stop conditions so agents record blockers instead of silently widening product direction.
- Let `plan.md` define `Target Stage`, making plain `hyper run` default to guarded auto continuation toward that target.
- Keep plan-target continuation commands as plain `hyper run`, while preserving `--auto --until` as an explicit override.
- Keep stored auto targets synchronized when `plan.md` `Target Stage` changes or is removed.
- Add explicit stage advancement review output for status, next-packet planning, and `hyper advance`.
- Make `hyper doctor` warn when `.hyper/next-packet.md` is missing required guard, continuation, or stage advancement review sections.
- Tighten Reference Benchmark Evidence so below-baseline gaps only pass when explicitly non-critical, deferred, out of scope, or a non-goal.
- Add Codex Desktop continuation guidance to `.hyper/next-packet.md` so auto mode clearly says when to run, repair, advance, or stop.
- Include next-packet continuation instructions in auto-mode `hyper run` and `hyper resume` Codex Desktop payloads.

## v0.6.3 - 2026-05-29

- Fix plan parsing so explicit `Current Stage` headings win over roadmap headings such as `0단계: 화면 검증`.
- Show a status/doctor refresh warning when `state.json` stores an old stage that differs from `plan.md`.
- Refresh the stored stage during `hyper migrate` and keep `.hyper/next-packet.md` aligned with the corrected stage.

## v0.6.2 - 2026-05-28

- Improve `hyper doctor` and `hyper status --short` action guidance for stale project state.
- Refresh README and supporting docs for the `v0.6.x` behavior.
- Add maintainer release checklists in English and Korean.
- Add update-after-install and troubleshooting flows.
- Expand before/after demo and reference benchmark examples.

## v0.6.1 - 2026-05-27

- Require the finish gate to pass before another runtime packet can start.
- Refresh and validate `.hyper/next-packet.md` during `hyper complete`, `hyper migrate`, `hyper advance`, and `hyper doctor`.
- Tighten readiness evidence matching, active capability evidence, repeated validation grouping, command-pattern classification, and failure-pressure handling.
- Improve first-run plan parsing, service-quality packet guidance, sustained quality flow, short status output, and Windows path display.
- Prevent `hyper repair` from bypassing a failed finish gate.

## v0.6.0 - 2026-05-26

- Add Service Quality reference benchmark evidence and status output.
- Require deployment, security, docs, operations, and benchmark proof when those readiness axes are the current pressure.
- Simplify README onboarding language and clarify the product loop.
- Fix staticcheck coverage around benchmark readiness helpers.

## v0.5.6 - 2026-05-26

- Fix readiness reconciliation after project state changes.

## v0.5.5 - 2026-05-26

- Add Learn quality-gate filtering for weak/noisy memory signals.
- Refresh legacy memory quality during `hyper migrate` with fixture coverage.
- Sign release assets with cosign keyless bundles and optionally verify them during install/update.

## v0.5.4 - 2026-05-26

- Verify GitHub release checksums during `hyper update`.
- Add a Windows PowerShell installer with SHA256 checksum verification.
- Add trusted install/update verification tests.

## v0.5.3 - 2026-05-26

- Add finish-gate review and auto continuation planning.
- Add `.hyper/next-packet.md` as the planned next command handoff.
- Improve short status and completion guidance.

## v0.5.2 - 2026-05-22

- Add the explicit stage advancement workflow with `hyper advance`.
- Recommend stage changes without applying them silently.

## v0.5.1 - 2026-05-22

- Split the CLI entrypoint from the application runtime package.
- Add Proof Contract sections for functional, surface, and operational proof.
- Add Surface Proof Evidence templates, readiness extraction, growth pressure learning, and proof gap status output.
- Improve readiness inference from real project runs.
- Read Korean product plan aliases in status output.

## v0.5.0 - 2026-05-22

- Add PR/push CI for tests, vet, staticcheck, and govulncheck.
- Add checksum verification to `install.sh` for GitHub release installs.
- Add macOS and Windows install guidance to the README.
- Add roadmap, known limitations, and a before/after demo guide.
- Normalize plan import candidate paths for stable Windows CI output.

## v0.4.1

- Improve growth status display.
- Shorten command-based candidate names in `hyper status`.
- Hide passive/noisy growth signals from status summaries.
- Preserve plan-covered product readiness when weak runtime evidence exists.

## v0.4.0

- Define the evidence-first growth protocol more clearly.
- Add pressure ledger, readiness, and capability candidate behavior.
- Add release automation for native binaries and checksums.

## Earlier Releases

- Add native Go CLI.
- Add `hyper init`, `hyper run`, `hyper complete`, `hyper status`, `hyper doctor`, `hyper update`, and Codex Desktop routing files.
