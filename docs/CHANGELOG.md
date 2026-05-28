# Changelog

## Unreleased

- Refresh README and supporting docs for the `v0.6.1` behavior.
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
