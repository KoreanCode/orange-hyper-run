# Changelog

## Unreleased

## v0.6.5 - 2026-06-05

- Warn when a long-running service-quality focus is used without `plan.md` `Target Stage`, so users know plain `hyper run` will create a single packet instead of continuing automatically.
- Treat Browser URL policy surface-proof gaps as completed packets that need a focused follow-up when implementation and command validation already passed.
- Plan the next packet toward allowed browser proof or a repeatable fallback surface check before general quality work resumes.
- Fall back to `.hyper/logs` and `.hyper/goals` counts when SQLite status counts cannot be read cleanly, avoiding misleading `Runs recorded: 0` output.

## v0.6.4 - 2026-06-05

- Split canonical stage vocabulary, target aliases, and stage ordering into `internal/stage` so plan parsing, auto targets, readiness, and status share one package boundary.
- Accept slug-style stage values such as `service-quality` and `sustained-service-quality` consistently when normalizing Current Stage values.
- Pin CI and release builds to Go `1.26.4` so govulncheck runs against the patched standard library.
- Add a Service Quality Self Review gate that requires plan alignment, core loop quality, product satisfaction, no drift, validation match, and an explicit pass/fail verdict before packet completion.
- Keep Service Quality packets open for correction when the Self Review verdict is `fail`.
- Include concrete failed Self Review fields and verdict text in finish-gate findings, next-packet correction plans, and `hyper resume`.
- Require Reference Benchmark Evidence in Beta and Service Quality finish gates unless the axis was already covered.
- Keep Reference Benchmark Evidence required even when it is the current readiness pressure, avoiding duplicate generic readiness findings.
- Add coverage for Service Quality packets that must satisfy Self Review, Reference Benchmark Evidence, and active validator proof before completion.
- Add Product satisfaction as a readiness axis for Beta, Service Quality, and Sustained Service Quality gates.
- Add a no-drift runtime guard to packet work boundaries and stop conditions so agents record blockers instead of silently widening product direction.
- Record an explicit threshold-based capability activation policy in growth state, capability files, and status output so candidates, promotable candidates, and active required behavior are distinguishable.
- Let `plan.md` define `Target Stage`, making plain `hyper run` default to guarded auto continuation toward that target.
- Treat `Target Stage` as complete only when that target stage's readiness proof is complete, so `Target Stage: Service Quality` continues into Service Quality packets instead of stopping immediately after entering the stage.
- Keep plan-target continuation commands as plain `hyper run`, while preserving `--auto --until` as an explicit override.
- Record explicit `--until` as the runtime target source so CLI output, `state.json`, and generated `goal.md` do not accidentally describe the `plan.md` target.
- Preserve the explicit `--until` source when a later `hyper run --auto` continues a previously selected command-line target.
- Document and test that plain `hyper run` returns to the `plan.md` target, while generated `--auto --until` commands keep an explicit override.
- Show the `plan.md` target in `hyper status` when an active explicit `--until` override points somewhere else.
- Keep stored auto targets synchronized when `plan.md` `Target Stage` changes or is removed.
- Add explicit stage advancement review output for status, next-packet planning, and `hyper advance`.
- Let active auto targets continue through `hyper advance` after the Stage Advancement Review shows ready proof and no blocking gaps.
- Block `hyper advance` when `plan.md Target Stage` is invalid, matching the `hyper run` and `hyper doctor` plan-target validation.
- Avoid creating filler runtime packets when an active auto target hits a ready stage gate; `hyper run` now points back to the reviewed `hyper advance`.
- Record no-packet auto decisions such as target-proof-complete and gate-ready advancement in project logs and SQLite events.
- Refresh `.hyper/next-packet.md` to `complete-current` when the finish gate fails, including during migration of older failed-packet states.
- Make finish-gate failure handoffs follow the latest `plan.md Target Stage` change or removal before writing `.hyper/next-packet.md`.
- Block `hyper complete` from writing completion handoffs when `plan.md Target Stage` is invalid.
- Make `hyper status` point users to fix invalid `plan.md Target Stage` instead of suggesting migration.
- Block `hyper migrate`, `hyper repair`, and `hyper resume` from writing or showing stale auto-continuation state when `plan.md Target Stage` is invalid.
- Validate `plan.md Current Stage` across init, run, status, doctor, complete, advance, migrate, repair, and resume so unknown stage names cannot silently fall back to Tiny MVP.
- Validate existing `plan.md` stage fields before `hyper init` writes `.hyper/` routing state, keeping failed initialization side-effect free.
- Refresh Codex Desktop routing files, generated command guides, and missing Hyper Run directories during `hyper migrate`, so CLI updates do not require rerunning `hyper init`.
- Make `hyper status` prioritize invalid `plan.md` stage fields even when a runtime packet is active, because completion and continuation are blocked until the plan is fixed.
- Show the next-packet planned action and continuation guard in `hyper complete`, no-packet auto-run stops, `hyper advance`, and `hyper repair` output.
- Show current finish-gate review findings in `hyper status` and `hyper status --short`, so same-packet correction can start without opening multiple files first.
- Include current review findings when `hyper run` is blocked by a failed finish gate, keeping the loop pointed at same-packet correction.
- Include `Planned action: complete-current` and the continuation guard in finish-gate failure errors, so agents do not need to open `.hyper/next-packet.md` before knowing they must fix the same packet.
- Treat Self Review verdicts such as `not ready`, `insufficient`, `incomplete`, or `not service-quality` as finish-gate failures instead of accepting them because they contain words like `ready`.
- Require Reference Benchmark decisions to explicitly allow Service Quality to proceed; decisions that say the service is blocked, not ready, or allowed only after more work remain finish-gate failures.
- Accept clear `not blocked` or `unblocked` benchmark/readiness wording when it also says Service Quality can proceed, avoiding false failures on positive blocker language.
- Surface current `review.md` findings inside `complete-current` next-packet plans and `hyper resume` output.
- Record finish-gate evidence, next-note, and findings hashes in `review.md`, then surface repeated same-finding failures so auto continuation stops when fixes are not addressing the gate.
- Add end-to-end coverage for the same-packet correction loop from finish-gate failure through stage advancement and the next runtime packet.
- Add multi-stage plan-target coverage proving plain `hyper run` can advance Tiny MVP through Service Quality, carry Service Quality into Sustained Service Quality, and stop at the target without `--auto --until`.
- Point `hyper status --short` at `review.md` when the finish gate failed so the same packet is fixed instead of starting new work.
- Point `hyper doctor` next actions at `review.md` correction when the finish gate failed.
- Show the planned next-packet action in `hyper status` and `hyper status --short`.
- Show the `.hyper/next-packet.md` handoff path in `hyper status` and `hyper status --short`.
- Avoid implying `.hyper/next-packet.md` exists while a normal active packet is still pending completion.
- Show the refreshed planned action and `.hyper/next-packet.md` path after `hyper repair`.
- Show planned and next actions after `hyper migrate` refreshes project state.
- Keep `hyper migrate` from telling an unfinished active packet to run `hyper complete` before evidence and next notes are updated.
- Make `hyper doctor` warn when `.hyper/next-packet.md` is missing required guard, continuation, or stage advancement review sections.
- Make `hyper doctor` warn when `.hyper/next-packet.md` keeps stale guard, continuation, or advancement review text after the auto target changes.
- Make `hyper doctor` verify `.hyper/next-packet.md` mode, reason, readiness gate, and readiness pressure metadata.
- Avoid requiring `.hyper/next-packet.md` in `hyper doctor` before the first runtime packet exists.
- Tighten Reference Benchmark Evidence so below-baseline gaps only pass when explicitly non-critical, deferred, out of scope, or a non-goal.
- Add Codex Desktop continuation guidance to `.hyper/next-packet.md` so auto mode clearly says when to run, fix the current packet, advance, or stop.
- Add an auto-mode Progress Guard to `.hyper/next-packet.md` and `hyper doctor` so repeated no-progress commands or repeated fix findings stop the continuation loop instead of looking like valid progress.
- Include next-packet continuation instructions in auto-mode `hyper run` and `hyper resume` Codex Desktop payloads.
- Clarify Codex router skills so `run`, `advance`, `complete-current`, and `stop` next-packet actions have explicit behavior.
- Clarify `complete-current` handoffs so agents fix review/evidence/next notes instead of confusing it with the `hyper repair` command.
- Treat `blocked` and `waiting_user` packets as terminal stop states for auto continuation, with next-packet guards that wait for the blocker or user decision instead of starting another run.
- Keep a plain `hyper run` using the plan target from creating a new packet after a terminal blocked/waiting stop unless the user gives an explicit follow-up focus.

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
