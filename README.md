![Hyper Run banner](assets/readme/banner.png)

<p align="right">
  <a href="./README.md"><kbd>English</kbd></a>
  <a href="./README_ko.md"><kbd>한국어</kbd></a>
</p>

# Hyper Run

Hyper Run is a harness-less project growth runtime. It starts from a human-owned `plan.md`, creates a concrete runtime packet for the next episode, records execution in `.hyper/`, and learns reusable context from completed or blocked work.

The project command model is:

```bash
hyper init
hyper run [focus]
hyper version
hyper update [source]
```

In Codex Desktop, use the same idea as a command convention:

```text
$hyper
$hyper init
$hyper run
```

`$hyper run` means: run the CLI in the current workspace, read the generated runtime packet, execute the current episode, update evidence, and write the next recommended runtime episode.

`hyper init` writes local Codex Desktop rules into `AGENTS.md`, `.agents/skills/hyper/SKILL.md`, `.agents/skills/hyper-run/SKILL.md`, `.hyper/codex-desktop.md`, and `.hyper/commands/hyper-run.md` so the `$hyper run` workflow is part of the project, not just this README.

The `hyper` skill is intentionally thin. It only lets Codex Desktop catch `$hyper run` and route it back to the native CLI plus `.hyper/` state. Product strategy, learning, validation evidence, and generated harnesses stay outside the static skill.

For the product boundary and example workflow, see:

- [Service Definition](docs/SERVICE_DEFINITION.md)
- [Tiny MVP Flow Example](examples/tiny-mvp-flow/README.md)

## Requirements

- No runtime dependency when installed from a release binary
- Go 1.21 or newer when building from source
- A project directory where Hyper Run can create `plan.md` and `.hyper/`

## Install Native Binary

Install the latest native binary to `~/.local/bin/hyper`:

```bash
curl -fsSL https://raw.githubusercontent.com/KoreanCode/orange-hyper-run/main/install.sh | sh
```

Manual install:

```bash
mkdir -p ~/.local/bin
curl -fsSL https://github.com/KoreanCode/orange-hyper-run/releases/latest/download/hyper-darwin-arm64 -o ~/.local/bin/hyper
chmod +x ~/.local/bin/hyper
```

Use `hyper-darwin-amd64` for Intel macOS, `hyper-linux-amd64` for Linux x64, and `hyper-linux-arm64` for Linux ARM64. Make sure `~/.local/bin` is on your `PATH`.

Check the installed binary:

```bash
hyper version
```

Then run Hyper Run inside any target project:

```bash
cd my-project
hyper init
# Fill in plan.md
hyper run "Build the smallest usable MVP"
```

## Install From Source

If you prefer source installation:

```bash
go install github.com/KoreanCode/orange-hyper-run/cmd/hyper@latest
```

## Update

Update the current native executable from the latest GitHub release:

```bash
hyper update
```

`hyper update` first tries to replace the currently running executable. If that path is not writable, it installs the latest binary to `~/.local/bin/hyper` and warns if that directory is not on `PATH`.

To update from a fork:

```bash
hyper update github:OWNER/orange-hyper-run
```

You can also pass a direct binary URL:

```bash
hyper update https://example.com/hyper-darwin-arm64
```

To force a user-local install path:

```bash
HYPER_INSTALL_PATH="$HOME/.local/bin/hyper" hyper update
```

Release binaries are built by the GitHub Actions release workflow when a `v*` tag is pushed.

## Local Development Install

From this repository:

```bash
go test ./...
go build -o dist/hyper ./cmd/hyper
```

Then run the local binary from another project directory:

```bash
cd ../my-project
../orange-hyper-run/dist/hyper init
# Fill in plan.md
../orange-hyper-run/dist/hyper run "Build the smallest usable MVP"
```

## Project Setup

Run `hyper init` once in the target project. It installs the local `.hyper/` runtime settings and creates a blank draft `plan.md` if one does not already exist.

Hyper Run works best after the user reviews `plan.md` at the project root:

```markdown
# Product Plan

## Product

What are we building?

## Target Users

Who is it for?

## MVP

What is the smallest coherent product?

## Current Stage

Tiny MVP

## Build Style

Web app

## Non-goals

What should Hyper Run avoid for now?

## Constraints

Technical, product, time, or UX constraints.

## Success Criteria

How do we know this stage is done?

## Current Focus

What should the next run advance?
```

If `plan.md` does not exist, `hyper run` stops and asks you to initialize the project first.

## What `hyper init` Creates

```text
AGENTS.md
.agents/
  skills/
    hyper/
      SKILL.md
    hyper-run/
      SKILL.md
.hyper/
  codex-desktop.md
  commands/
    hyper-run.md
  capabilities/
    candidates/
      harness/
      skill/
      validator/
    active/
      harness/
      skill/
      validator/
    retired/
      harness/
      skill/
      validator/
  growth/
    state.json
  readiness/
    state.json
  hyper.sqlite
  state.json
  logs/
  memories/
    decisions.md
    patterns.md
    failures.md
    constraints.md
  validators/
    generated/
  skills/
    generated/
  harnesses/
    generated/
plan.md
```

## What `hyper run` Creates

```text
.hyper/
  logs/
  goals/
    GOAL-0001/
      goal.md
      tasks.md
      evidence.md
      review.md
      next.md
```

Each run creates a runtime packet that can be handed to Codex Desktop or another execution agent. It is not a long-lived SPEC; it is the next execution episode derived from `plan.md`, logs, evidence, memory, and current project state. The default handoff is prompt-based and includes:

```text
Read .hyper/goals/GOAL-0001/goal.md as a runtime packet and complete it checkpoint by checkpoint.
```

`evidence.md` includes `Readiness Evidence` and `Active Capability Evidence` so stage-gate progress and required active validators can be confirmed separately from general validation output.

## Learning Loop

After a runtime packet is completed or blocked, update its `evidence.md` and `next.md`.

Learn is not a generic summary. It extracts only durable signals that should influence future work:

- decisions that should remain true
- reusable implementation or validation patterns
- blockers and failures to avoid repeating
- constraints that future runtime packets must respect

The strongest Learn signals come from these sections:

```text
evidence.md
  Validation
  Readiness Evidence
  Decisions
  Reusable Patterns
  Blocker

next.md
  Recommended Next Goal
  Learn Notes
```

Use `Learn Notes` for explicit structured signals:

```text
- Decision: Keep auth local-first until the MVP is usable.
- Pattern: Validate the main user flow with the existing Playwright smoke test.
- Constraint: Do not add paid services without credentials.
- Failure: The previous API path failed because the required token was missing.
```

The next `hyper run` automatically learns from the previous active runtime packet before creating the next one. Manual learning remains available for debugging:

```bash
hyper internal learn
```

Hyper Run stores reusable memories in SQLite and `.hyper/memories/`. Future `hyper run` calls retrieve similar prior context and include it in the next runtime packet under `Continue From`. Learn does not treat changed files or long notes as memory unless they contain a durable decision, pattern, failure, or constraint.

## Service Readiness Model

After Growth is updated, Hyper Run writes `.hyper/readiness/state.json`. This state is the bridge from tiny MVP work to service-level quality. It tracks project readiness across these axes:

- product completeness
- core UX
- data persistence
- error handling
- validation coverage
- security baseline
- deployment readiness
- operations and docs
- maintainability

The next runtime packet includes a `Stage Gate` section. Hyper Run uses the current stage to choose the next gate, finds the weakest missing readiness axis, and turns that into a concrete pressure for the next episode. This pressure affects `Current Episode`, `Work Boundary`, `Validation Signals`, and `Stop When`.

Record readiness progress with axis-labeled evidence:

```text
## Readiness Evidence

Core UX: The primary add/edit flow works in the browser.
Data persistence: User records survive reload using SQLite or localStorage.
Validation coverage: The primary flow is covered by a repeatable smoke test.
```

On the next `hyper run`, axis-labeled readiness evidence moves that axis to `covered`, removes it from the stage gate blocking gaps, and prevents the same pressure from being selected again.

Evidence must be specific enough for the axis. For example, `Validation coverage: tested` is only emerging evidence, while `Validation coverage: \`go test ./...\` passed and is repeatable` is covered evidence. UX evidence should mention a smoke pass, browser verification, or screenshot. Deployment evidence should include a build, release, hosted URL, CI, or deploy proof.

When a stage gate becomes ready, Hyper Run does not edit `plan.md` automatically. It emits a stage advancement candidate in the next runtime packet and recommends the exact `Current Stage` change for the user to accept or reject.

Beta and Service Quality stages also create quiet validator candidates under `.hyper/validators/generated/` and `.hyper/capabilities/candidates/validator/`. These are not required behavior until they are promoted to active validators.

The goal is not to add a heavy process. The user still runs `hyper run`; the project gradually learns what service quality means for itself.

## Project Growth Engine

After Learn runs, Hyper Run updates `.hyper/growth/state.json`. This is not a user-facing planning report. It is the project-local growth state that changes how the next runtime packet is compiled.

Growth pressure comes from repeated or durable Learn signals:

- stable decisions and constraints affect `Work Boundary`
- repeated validation patterns affect `Validation Signals`
- recurring failures affect `Stop When`
- repeated patterns can create quiet candidates under `.hyper/validators/generated/` or `.hyper/skills/generated/`
- active validators under `.hyper/capabilities/active/validator/` become required behavior in the next runtime packet's `Validation Signals`
- a harness candidate is generated only after multiple repeated pressures cross the harness threshold

Phase 1 stabilization keeps pressure cleaner by clustering similar Learn signals instead of requiring exact text matches, ignoring noisy progress-only signals, and classifying pressure as `stable_decision`, `repeated_validation`, `implementation_pattern`, `recurring_constraint`, or `recurring_failure`.

Phase 2 adds a capability lifecycle:

```text
observed -> repeated -> promotable -> active -> retired
```

Lifecycle metadata is written under `.hyper/capabilities/`. Legacy candidate files under `.hyper/validators/generated/`, `.hyper/skills/generated/`, and `.hyper/harnesses/generated/` are still written for discoverability, but activation depends on the lifecycle status. Candidate and promotable validators do not become required behavior until an active validator exists under `.hyper/capabilities/active/validator/`.

The user still runs `hyper run`; the project grows by making the next packet more specific.

## Useful Commands

```bash
hyper init
hyper run "Add customer persistence"
hyper status
hyper resume
hyper update
hyper version
hyper internal learn
```

## Current Status

This is a native Go MVP implementation. It intentionally avoids a TUI and keeps the surface area small: one user-facing command, local files, SQLite logs, prompt-based execution handoff, and lightweight learning.

## License

MIT License. See [LICENSE](LICENSE).
