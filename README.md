![Hyper Run banner](assets/readme/banner.png)

<p align="right">
  <a href="./README.md"><kbd>English</kbd></a>
  <a href="./README_ko.md"><kbd>한국어</kbd></a>
</p>

# Hyper Run

Hyper Run is an evidence-first project growth protocol. Execution logs create pressure, pressure creates candidates, and repeated proof promotes project-specific structure.

You write a simple `plan.md`. Hyper Run turns it into the next small work packet, stores progress under `.hyper/`, and uses completed evidence to make the next packet more specific.

It is agent-agnostic: Codex Desktop, CLI agents, Cursor-style agents, or other coding assistants can consume the same runtime packet. The basic loop is still one command:

```bash
hyper run
```

## Why Use It?

AI coding sessions often lose project context:

- the next task becomes too broad
- previous decisions are forgotten
- validation evidence is scattered
- small MVP work does not naturally grow into service quality

Hyper Run keeps that context inside the project. It is not a task splitter or a full project manager. It creates the next focused runtime packet, learns from the result, and lets repeated evidence decide when the project needs stronger structure.

## Core Ideas

Hyper Run has a few internal concepts, but they are simple:

| Concept | Simple meaning |
| --- | --- |
| `plan.md` | The human-written product brief. It says what the product is, who it is for, and what stage it is in. |
| Runtime packet | The next small work bundle generated from `plan.md` and project history. Usually this is `.hyper/goals/GOAL-0001/goal.md`. |
| Evidence | Proof that the work was done and checked. This goes in `evidence.md`. |
| Proof Contract | The runtime packet's short proof boundary: functional proof, surface proof, and operational proof. |
| Learn | The step that extracts what the project repeatedly needed, failed at, or proved. It is not a generic summary. |
| Pressure Ledger | The project ledger of unresolved or repeated pressure. For example, if every run needs the same validation, Hyper Run can suggest a validator candidate. |
| Readiness | Stage contracts that check whether the project is ready to move from Tiny MVP to Usable MVP, Beta, and Service Quality. |
| Capability candidate | A suggested validator, skill, or harness. It is only a candidate until enough repeated evidence proves it should be active. |

The key idea is **harness-less growth**. A project does not need a harness on day one. It starts with `plan.md`, runs small packets, records evidence, and only creates stronger structure when the project repeatedly proves it needs that structure.

```text
Execution -> Evidence -> Pressure Ledger -> Capability candidate -> Structure when proven
```

This is the main difference from a harness-first workflow. A harness usually starts with a predefined workflow. Hyper Run starts with execution evidence and lets the project earn validators, skills, agents, or harnesses only when repeated pressure makes them useful.

Each runtime packet also asks for proof in three plain areas:

- Functional proof: the smallest useful behavior works.
- Surface proof: if a user-facing screen changed, a real user can understand the screen, take the primary action, and see the result or recovery state.
- Operational proof: the safest available build, test, smoke, setup, or handoff path is repeatable, or the blocker is documented.

Surface proof is not a design harness. Screenshot, browser, responsive, accessibility, Figma, or design-system checks only become stronger candidates after repeated evidence shows the project needs them.

## Principles

Hyper Run follows four product rules:

- No structure before pressure.
- No stage advancement without evidence.
- No harness before repeated need.
- No memory without reusable signal.

These rules matter because Hyper Run should not create process for its own sake. Structure appears only when the project keeps proving it needs that structure.

## Pressure Ledger

The Pressure Ledger lives in `.hyper/growth/state.json`. It tracks repeated validation needs, recurring failures, reusable implementation patterns, constraints, and readiness gaps.

The ledger does not immediately force new behavior. It moves through a lifecycle:

```text
observed -> repeated -> promotable -> active -> retired
```

Before a threshold is reached, generated validators, skills, agents, or harnesses remain candidates. After repeated proof, they can become active project-specific structure.

## Stage Contracts

Stages are not just labels. Each stage changes what `goal.md` asks Codex or another coding agent to prove.

| Stage | Contract |
| --- | --- |
| Tiny MVP | Existence proof: prove one useful flow exists with the smallest reversible product slice. |
| Usable MVP | Usability proof: make the primary flow usable end-to-end for a real user. |
| Beta | Repeatability proof: prove reliability around realistic data, failures, validation, docs, and release readiness. |
| Service Quality | Operability proof: treat security, deployment, operations, rollback, and repeatable validation as required product behavior. |

`hyper run` keeps working while the project has unresolved growth pressure. When pressure repeats, Hyper Run creates candidates. When evidence keeps confirming the same need, those candidates can become active project structure.

## Basic Flow

```bash
hyper init
# edit plan.md once

hyper run "Build the smallest usable MVP"
# implement the generated packet
# update evidence.md and next.md

hyper complete
hyper status --short
hyper advance   # only when the stage gate is ready and you accept the stage change
hyper doctor
hyper run "Next improvement"
```

## Execution Flow

```mermaid
flowchart TD
  A["hyper init<br/>create plan.md and .hyper/"] --> B["Edit plan.md<br/>product brief and current stage"]
  B --> C["hyper run [focus]<br/>create runtime packet"]
  C --> D["Implement current GOAL<br/>read goal.md and tasks.md"]
  D --> E["Update evidence.md and next.md"]
  E --> F["hyper complete<br/>run Finish Gate"]
  F --> G{"Finish Gate passed?"}
  G -- "No" --> H["Write review.md findings<br/>stay in the same GOAL"]
  H --> D
  G -- "Yes" --> I["Learn<br/>refresh memory, growth, readiness"]
  I --> J["Next Packet Planner<br/>write .hyper/next-packet.md"]
  J --> K{"Stage gate ready?"}
  K -- "Yes, user accepts" --> L["hyper advance<br/>update plan.md stage"]
  K -- "No or not accepted" --> M["hyper status --short<br/>review next action"]
  L --> N{"Auto until target reached?"}
  M --> N
  N -- "No" --> C
  N -- "Yes" --> O["Stop and review<br/>choose the next service target"]
```

`hyper complete` runs a finish gate before learning. If validation, readiness evidence, active capability evidence, or `next.md` is not good enough yet, it writes findings to the current packet's `review.md` and keeps you in the same packet.

For longer Codex Desktop sessions, start with an auto target:

```bash
hyper run --auto --until service-quality "Keep upgrading this service"
```

Auto mode does not skip evidence or silently advance stages. It keeps the next packet command planned in `.hyper/next-packet.md`; stage changes still require explicit acceptance with `hyper advance`.

In Codex Desktop you can use the same idea as a project command:

```text
$hyper init
$hyper run
```

`$hyper run` means Codex should run the native `hyper` CLI, read the generated `.hyper/goals/.../goal.md`, implement it, update evidence, and prepare the next recommendation.

## Install

### macOS / Linux

Install the latest native binary:

```bash
curl -fsSL https://raw.githubusercontent.com/KoreanCode/orange-hyper-run/main/install.sh | sh
```

For GitHub release installs, the installer downloads `checksums.txt` and verifies the binary with SHA256 before moving it into place.

Check it:

```bash
hyper version
```

Manual macOS install:

Apple Silicon:

```bash
mkdir -p ~/.local/bin
curl -fsSL https://github.com/KoreanCode/orange-hyper-run/releases/latest/download/hyper-darwin-arm64 -o ~/.local/bin/hyper
chmod +x ~/.local/bin/hyper
hyper version
```

Intel Mac:

```bash
mkdir -p ~/.local/bin
curl -fsSL https://github.com/KoreanCode/orange-hyper-run/releases/latest/download/hyper-darwin-amd64 -o ~/.local/bin/hyper
chmod +x ~/.local/bin/hyper
hyper version
```

### Windows

Install the latest Windows x64 binary with PowerShell:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/KoreanCode/orange-hyper-run/main/install.ps1 | iex"
```

The PowerShell installer downloads `checksums.txt` and verifies the binary with SHA256 before moving it into place.

If the installer warns that `~\.local\bin` is not on `PATH`, add it:

```powershell
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\.local\bin", "User")
```

Open a new terminal, then check:

```powershell
hyper version
```

Other release binaries:

- `hyper-darwin-amd64` for Intel macOS
- `hyper-linux-amd64` for Linux x64
- `hyper-linux-arm64` for Linux ARM64
- `hyper-windows-amd64.exe` for Windows x64

Make sure `~/.local/bin` is on your `PATH`.

## Install From Source

```bash
go install github.com/KoreanCode/orange-hyper-run/cmd/hyper@latest
```

## Update

```bash
hyper update
```

This downloads the latest GitHub release. If Hyper Run cannot replace the current executable, it installs to `~/.local/bin/hyper`.
For GitHub release updates, Hyper Run downloads `checksums.txt` and verifies the binary before replacing the executable.

To update from a fork:

```bash
hyper update github:OWNER/orange-hyper-run
```

## Project Setup

Run this once inside your project:

```bash
hyper init
```

It creates:

- `plan.md`
- `.hyper/`
- Codex Desktop routing files such as `AGENTS.md` and `.agents/skills/...`

Then fill in `plan.md` in plain language:

```markdown
# Product Plan

## Product

What are we building?

## Target Users

Who is it for?

## MVP

What is the smallest useful version?

## Current Stage

Tiny MVP

## Build Style

Web app

## Non-goals

What should not be built yet?

## Constraints

Technical or product constraints.

## Success Criteria

How do we know this stage is done?

## Current Focus

What should the next run improve?
```

If `plan.md` is sparse, Hyper Run may create `.hyper/plan-candidates.md` from README or docs so you can copy useful product context into `plan.md`.

## What `hyper run` Does

`hyper run` creates a new runtime packet:

```text
.hyper/goals/GOAL-0001/
  goal.md
  tasks.md
  evidence.md
  review.md
  next.md
```

The important files are:

- `goal.md`: what to build now
- `tasks.md`: checkpoints for this run
- `evidence.md`: proof of what changed and what was validated
- `next.md`: what should happen next

Hyper Run blocks a new `hyper run` if the previous packet still has pending evidence. Finish the current packet with `hyper complete` first.

## What `hyper complete` Does

After implementation, update `evidence.md` and `next.md`, then run:

```bash
hyper complete
```

This closes the current packet and updates project memory:

- decisions to keep
- reusable patterns
- failures or blockers
- constraints
- readiness progress

`hyper complete` also prints the next recommended action. If the gate is ready, it will tell you to run `hyper advance`. Otherwise it will point to the next smallest `hyper run` focus. The next `hyper run` uses the learned information to change the work boundary, validation signals, stop conditions, readiness pressure, and capability candidates.

## Readiness In Simple Terms

Hyper Run tries to grow the project stage by stage:

```text
Tiny MVP -> Usable MVP -> Beta -> Service Quality
```

It checks whether the project has evidence for things like:

- product clarity
- core UX
- persistence
- error handling
- validation
- security
- deployment
- docs
- maintainability

You record this in `evidence.md`:

```text
## Readiness Evidence

Core UX: Browser smoke test passed for create and complete flow.
Validation coverage: `go test ./...` passed and is repeatable.
Data persistence: Records survive reload using SQLite.
```

When enough evidence exists, `hyper status` shows the next stage is ready. Hyper Run still does not change the stage silently. If you accept the recommendation, run:

```bash
hyper advance
```

That updates `plan.md` from the current stage to the next stage, refreshes readiness, and then the next `hyper run` starts using the new stage behavior.

## Commands

```bash
hyper init                  # install Hyper Run files in this project
hyper run [focus]           # create the next runtime packet
hyper run --auto --until service-quality [focus]
hyper complete              # run the finish gate, close the packet, and learn
hyper advance               # apply an accepted stage change when the gate is ready
hyper status                # show current stage, gaps, and readiness
hyper status --short        # show only stage, gate, proof, and next action
hyper doctor                # diagnose install, PATH, project state, and Codex routing
hyper repair                # reconcile state.json when packet evidence and state disagree
hyper migrate               # refresh growth/readiness state after Hyper Run upgrades
hyper resume                # print the current handoff again
hyper update                # update the native binary
hyper version               # show version and binary path
hyper internal learn        # debug/manual learning command
```

## Local Development

From this repository:

```bash
go test ./...
go vet ./...
go build -o dist/hyper ./cmd/hyper
```

Then test it in another project:

```bash
cd ../my-project
../orange-hyper-run/dist/hyper init
../orange-hyper-run/dist/hyper run "Build the smallest usable MVP"
../orange-hyper-run/dist/hyper complete
```

## More Detail

- [Service Definition](docs/SERVICE_DEFINITION.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Tiny MVP Flow Example](examples/tiny-mvp-flow/README.md)
- [Before / After Demo](examples/before-after-demo/README.md)
- [Roadmap](docs/ROADMAP.md)
- [Changelog](docs/CHANGELOG.md)
- [Known Limitations](docs/KNOWN_LIMITATIONS.md)

## License

MIT License. See [LICENSE](LICENSE).
