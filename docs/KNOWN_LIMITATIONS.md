# Known Limitations

Hyper Run is usable, but some parts are still early. This document is intentionally explicit so users can judge where the tool is stable today.

## Stable Enough For Daily Testing

- Creating project-local Hyper Run state with `hyper init`.
- Generating runtime packets with `hyper run`.
- Closing packets with `hyper complete`.
- Reading status and readiness with `hyper status`.
- Diagnosing common install/project issues with `hyper doctor`.
- Repairing simple state mismatches with `hyper repair`.
- macOS, Linux, and Windows x64 CI builds and tests for the native CLI.

## Still Experimental

- Growth pressure classification is heuristic. It is useful, but not a formal semantic model.
- Capability candidates are generated from repeated evidence, but activation policy is still conservative.
- Package boundaries are still mostly inside `cmd/hyper`; this is acceptable for the current CLI size, but should be split later.
- Existing `.hyper/` projects may need `hyper migrate` after larger growth/readiness changes.
- Windows release binaries are built and CI-tested, but there is no PowerShell installer script yet.

## Security And Supply Chain

- Release builds publish checksums.
- The installer verifies checksums for GitHub release downloads.
- Release binaries are not signed yet. Cosign signing is a future step.
- `hyper update` should gain checksum verification next so it matches `install.sh`.

## Agent Behavior

- Hyper Run creates runtime packets; it does not force an AI agent to behave correctly.
- Codex Desktop routing is a thin compatibility layer. The source of truth remains `plan.md`, `.hyper/`, and the native CLI.
- Good evidence still depends on the agent or developer updating `evidence.md` and `next.md` honestly.
