# Hyper Run Service Definition

Hyper Run is a harness-less project growth runtime that turns a human-owned `plan.md` into repeated execution packets, evidence, and durable learning so a small MVP can grow into a larger project without starting from a heavy harness.

## Target User

Hyper Run is for builders who use Codex Desktop or CLI agents to develop projects over many sessions and want a simple way to keep direction, execution state, validation evidence, and reusable learning attached to the project.

The primary user is not looking for a generic task manager, a static spec system, or a full autonomous platform. They want to start with a lightweight plan, run the next coherent step, preserve what was learned, and let the project accumulate enough structure to justify generated harnesses later.

## Problem

Agent work often fails to scale from small tasks to large projects because each session loses context, treats every request as a fresh task, or over-invests in static specs before the product shape is known.

Harness-first systems can help once the project has stable workflows, but they are too heavy at the beginning. A tiny MVP needs quick execution, not a large fixed process. A growing project needs memory, evidence, and repeatable run boundaries before it needs a complete harness.

## Product Promise

Hyper Run promises a small, repeatable project loop:

1. Start from a human-owned `plan.md`.
2. Generate one runtime packet for the next execution episode.
3. Let Codex Desktop or another agent execute that packet inside the current repo.
4. Record validation evidence, changed files, decisions, reusable patterns, blockers, and next steps.
5. Learn only durable signals that should influence future runs.
6. Retrieve similar prior context when the next runtime packet is created.
7. Measure service readiness across product, UX, persistence, validation, security, deployment, operations, maintainability, benchmark fit, and product satisfaction.
8. Grow toward generated skills, agents, validators, or harnesses only when the project has earned that structure.
9. Keep the next command explicit in `.hyper/next-packet.md` so auto continuation remains packet-by-packet and reviewable.

## What Hyper Run Is

- A local project runtime around `plan.md`, `.hyper/`, and the `hyper` CLI.
- A command convention for Codex Desktop through `$hyper run`.
- A runtime packet generator for the next coherent execution episode.
- A local evidence and learning layer backed by files and SQLite.
- A service-readiness gate that keeps tiny MVP work moving toward usable, beta, and service-quality stages.
- A finish gate that blocks the next packet until evidence and next-step notes are good enough or a real blocker is recorded.
- A Self Review gate for Service Quality and Sustained Service Quality packets, so working code is not treated as enough when product satisfaction, core loop quality, no-drift, or plan alignment is still weak.
- A harness-less starting point that can later create project-specific harnesses when the evidence shows they are useful.

## What Hyper Run Is Not

- Not a replacement for Codex Desktop or the coding agent.
- Not a cloud agent platform.
- Not a static SPEC manager.
- Not a project management app.
- Not a test framework.
- Not a required TUI.
- Not a promise that every run will complete without user judgment, credentials, or validation.

## Relationship To Harnesses

Hyper Run sits above a harness in the project lifecycle.

A harness is useful after the project has repeated workflows, stable validation paths, and clear execution boundaries. Hyper Run exists before that point. It lets the project run harness-less, gather evidence, and discover what harnesses are worth generating.

The intended path is:

```text
plan.md
  -> runtime packets
  -> evidence and Learn signals
  -> repeated patterns
  -> generated validators, skills, agents, or harnesses when needed
```

## Run Contract

One CLI invocation of `hyper run` creates at most one runtime packet. If `plan.md` has `Target Stage`, plain `hyper run` enters guarded auto mode and writes the planned continuation command after each completed packet. When that target comes from `plan.md`, the continuation command stays plain `hyper run`; `--auto --until` is reserved for explicit command-line overrides. The packet is complete when the executing agent has:

- Read `plan.md`, `goal.md`, and `tasks.md`.
- Checked the packet's `Stage Gate` and selected readiness pressure.
- Implemented the smallest coherent step for the current episode.
- Run the safest available validation or recorded why validation is blocked.
- Updated `evidence.md` with validation output, readiness evidence, active capability evidence, changed files, decisions, reusable patterns, and blockers.
- Updated `next.md` with the next recommended runtime episode and structured Learn Notes.
- Run `hyper complete` so Learn, Growth, and Readiness refresh from the completed packet.
- Stopped before destructive actions, missing credentials, unclear product scope, or repeated validation failure.

`hyper run` should not be treated as an unchecked background loop. The long-running part is packet-by-packet continuation: create one packet, execute it, check evidence, learn, then follow `.hyper/next-packet.md` only if the guard allows it. That file includes both the next command and Codex Desktop continuation guidance. A new `hyper run` is blocked while the previous active packet still has pending evidence.

## Learn Role

Learn is not a summary system.

Learn extracts durable signals from completed or blocked runtime packets:

- `decision`: a product or technical choice future runs should respect
- `pattern`: a reusable implementation or validation approach
- `constraint`: a boundary future runs must not violate
- `failure`: a blocker or failed approach future runs should avoid repeating

Learn should ignore ordinary progress notes, changed-file lists, and temporary observations unless they contain a durable decision, pattern, constraint, or failure.

## Service Readiness Role

Readiness is the part of Hyper Run that makes "keep going until service quality" concrete.

Hyper Run writes `.hyper/readiness/state.json` with readiness axes, the current stage gate, blocking gaps, and the next selected pressure. The next runtime packet uses that state to decide what should be advanced now, what evidence is required, and when not to claim stage advancement.

Readiness evidence is progressive. New evidence files include slots for every readiness axis. When `evidence.md` contains axis-labeled lines such as `Data persistence: Customer records survive reload`, `hyper complete` and `hyper status` mark that axis as `covered`, remove the matching gate gap, and select the next weakest pressure instead of repeating the solved one.

Readiness evidence also has a basic quality bar. A vague label such as `Validation coverage: tested` is treated as emerging evidence, not covered evidence. Covered evidence should include the proof shape expected by the axis: commands or smoke tests for validation, browser or screenshot proof for UX, reload/restart/storage proof for persistence, hosted/build/release/artifact proof for deployment, and docs/runbook/rollback proof for operations.

Product satisfaction is a readiness axis, not just a free-form opinion. It should prove that the result still fits the target user, the core loop feels coherent, visible or operational details are acceptable, and the packet did not drift away from `plan.md`.

Runtime packets include a Proof Contract with three evidence boundaries: Functional Proof, Surface Proof, and Operational Proof. Surface Proof is required only when the packet changes a user-facing screen or flow. It should connect screenshots or browser smoke evidence to the affected surface, primary user action, checked states, viewport, and remaining surface gaps.

Surface Proof is intentionally evidence-first. Repeated surface evidence can become visual-smoke, responsive-check, accessibility, Figma handoff, or design-system candidates, but those candidates are not required behavior until repeated proof promotes them.

When all required axes for the current gate are covered, Hyper Run creates a stage advancement candidate. It recommends the exact `plan.md` `Current Stage` change but does not apply it automatically. This keeps stage movement human-reviewed while still making the project state explicit.

Beta and Service Quality stages can generate quiet validator candidates for repeatable smoke, security, deployment, and operations checks. They remain candidates until repeated evidence promotes them to active validators.

The default readiness path is:

```text
Tiny MVP -> Usable MVP -> Beta -> Service Quality
```

## Service Quality Standard

Service Quality does not mean "perfect production." It means the project has enough evidence that a real operator or tester can run, validate, recover, compare, and continue the service without relying on hidden context from the current agent session.

In Hyper Run, Service Quality is not only operational readiness. It also requires reference benchmark evidence: the result should meet the basic expectations of its category and have at least one concrete strength compared with similar products, tools, apps, or workflows.

Hyper Run should treat a project as Service Quality only when these criteria are covered:

| Area | Required Proof |
| --- | --- |
| Product boundary | The primary value loop is complete enough to test, and non-goals or deferred surfaces are explicit. |
| Validation | Required commands, smoke checks, or manual checks are repeatable from documented steps. |
| UX and surface | Critical user-facing flows have current browser, screenshot, or equivalent surface proof for the touched states. |
| Data and persistence | Data creation, readback, deletion, fallback, or migration behavior is proven for the current architecture. |
| Security and privacy | Secrets, permissions, local or remote data handling, content boundaries, and misuse risks are documented and checked. |
| Deployment or release | A hosted URL, packaged artifact, native build, CLI binary, container, or equivalent release path runs outside the development path. |
| Operations | Setup, environment, smoke path, rollback, recovery, stop conditions, and handoff notes are documented. |
| Maintainability | The next operator can identify the main code paths, known risks, active validators, and highest-friction follow-up. |
| Reference benchmark | 3-5 comparable references define the category baseline; no core expectation is below baseline, and at least one strength is above baseline. |
| Product satisfaction | The result is acceptable for the target user, core loop, visual or copy quality, and no-drift direction check. |

Reference Benchmark Evidence should not be a generic scorecard. It should turn outside comparison into the next execution pressure:

```md
## Reference Benchmark Evidence

- Category: Developer CLI / project-growth runtime
- References: Tool A, Tool B, Tool C
- Baseline expectations: install is clear; one command creates useful work context; status and recovery are understandable
- Current comparison: setup meets baseline; evidence loop is above baseline; auto continuation is below baseline
- Below-baseline gaps: auto continuation and stage advance clarity
- Above-baseline strength: project-local evidence and readiness pressure
- Decision: Service Quality is blocked until auto continuation and stage advance reach the category baseline
```

Hyper Run treats this evidence as covered only when the benchmark has a category, 3-5 named references, baseline expectations, a current below/meets/above-baseline comparison, no critical below-baseline gap, one above-baseline strength, and a decision. `hyper status` shows the benchmark line whenever it is required by the current gate.

Service Quality is blocked when any of these are true:

- The only proof is "it works on my machine" without commands, artifact, URL, screenshot, or smoke output.
- A critical credential, data, security, deployment, or rollback step is unknown.
- The project cannot be restarted, installed, packaged, served, or run from the documented path.
- The next operator must infer hidden agent decisions that are not in `plan.md`, `.hyper/`, docs, or code.
- The next recommended packet is broad feature work while validation, security, deployment, operations, or maintainability gaps remain.
- Reference Benchmark Evidence has a below-baseline gap in a core user or operator expectation.
- The project has no concrete above-baseline strength compared with its references.

Once Service Quality is reached, the next stage is not "add everything." It is Sustained Service Quality: monitoring repeated failures, promoting proven validators, reducing operational friction, and only then expanding product breadth.

Readiness does not replace Learn or Growth. Learn extracts durable signals. Growth turns repeated signals into behavior and capabilities. Readiness asks whether the project is becoming a usable service and selects the weakest missing axis for the next run.

## Service Boundary

Hyper Run should stay small at the user-facing layer:

- `hyper init` initializes project-local runtime files.
- `hyper run [focus]` creates the next runtime packet and uses `plan.md` `Target Stage` as the default auto target when present.
- `hyper complete` closes the active packet and refreshes Learn, Growth, and Readiness.
- `hyper resume` resumes the active packet.
- `hyper status` shows current runtime state.
- `hyper update` updates the native binary.

New capabilities should usually become generated project knowledge under `.hyper/` before they become permanent top-level commands.

## Success Criteria

Hyper Run is successful when a user can:

- Start a project with only `hyper init`, `plan.md`, `hyper run`, and `hyper complete`.
- Build a tiny MVP without designing a harness first.
- Continue the same project across many sessions with useful context retrieval.
- See clear evidence of what changed, what passed, what failed, and what should happen next.
- Accumulate enough Learn signals to create project-specific validators, skills, agents, or harnesses only when those structures are justified.

## Current MVP Boundary

The current product should stay focused on:

- Native CLI installation and update.
- Project-local initialization.
- Runtime packet generation.
- Codex Desktop `$hyper` routing.
- Evidence and next-step templates.
- Finish-gate review with `review.md`.
- Auto continuation planning through `.hyper/next-packet.md`.
- Explicit stage advancement with `hyper advance`.
- Durable Learn extraction.
- Project Growth Engine pressure state.
- Service Readiness state and Stage Gate runtime compilation.
- Similar signal clustering and noisy signal filtering.
- Capability lifecycle from repeated pressure to active structure.
- Growth-informed runtime packet compilation.
- Quiet validator, skill, and harness candidates behind thresholds.
- Similar-context retrieval.
- Reference benchmark evidence for Service Quality gates.
- Checksum-verified install/update and optional cosign signature verification.
- A clear Golden Path example.

Everything else should be evaluated through the run contract before becoming product surface area.
