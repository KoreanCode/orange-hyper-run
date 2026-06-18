# Known Limitations

Hyper Run is usable, but some parts are still early. This document is intentionally explicit so users can judge where the tool is stable today.

## Stable Enough For Daily Testing

- Creating project-local Hyper Run state with `hyper init`.
- Generating runtime packets with `hyper run`.
- Closing packets through the agent finish gate (`hyper complete` for agent/manual recovery).
- Guarded auto continuation from `plan.md` `Target Stage` or `hyper run --auto --until <stage>`.
- Explicit stage advancement with `hyper advance`.
- Reading status and readiness with `hyper status`.
- Diagnosing common install/project issues with `hyper doctor`.
- Repairing simple state mismatches with `hyper repair`.
- Updating project state after CLI upgrades with `hyper migrate`.
- Checksum-verified install/update from GitHub releases, with optional cosign signature verification.
- macOS, Linux, and Windows x64 CI builds and tests for the native CLI.

## Still Experimental

- Growth pressure classification is heuristic. It is useful, but not a formal semantic model.
- Capability candidates use an explicit threshold-based activation policy. The policy is intentionally conservative and does not yet support team-specific custom thresholds.
- Auto mode is packet-by-packet continuation, not an unattended autonomous background runner. `Target Stage` makes plain `hyper run` choose the target and stops only after that target stage's readiness proof is complete; finish gates still apply, stage advancement only continues after the Stage Advancement Review shows ready proof and no blocking gaps, blocked/waiting packets require a deliberate focused follow-up, and `.hyper/next-packet.md` includes a Progress Guard for no-progress loops.
- Reference benchmark evidence is structured and checked, but the quality of the comparison still depends on the evidence written by the agent or developer.
- Package boundaries are still mostly inside `internal/app`; stage vocabulary has been split into `internal/stage`, and larger policy/storage/runtime splits should remain pressure-driven.
- Existing `.hyper/` projects may need `hyper migrate` after larger growth/readiness changes.

## Security And Supply Chain

- Release builds publish checksums.
- The macOS/Linux installer, Windows PowerShell installer, and `hyper update` verify checksums for GitHub release downloads.
- Release builds publish cosign keyless signature bundles. Installers and `hyper update` verify them when `cosign` is available, and can require them with `HYPER_RUN_VERIFY_SIGNATURE=required`.
- Signature verification depends on the local `cosign` executable; checksum verification remains the default baseline.

## Agent Behavior

- Hyper Run creates runtime packets; it does not force an AI agent to behave correctly.
- Codex Desktop routing is a thin compatibility layer. The source of truth remains `plan.md`, `.hyper/`, and the native CLI.
- Good evidence still depends on the agent updating `evidence.md` and `next.md` honestly; manual developers can still do the same for recovery/debugging.
