# Changelog

## Unreleased

- Verify GitHub release checksums during `hyper update`.
- Add a Windows PowerShell installer with SHA256 checksum verification.

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
