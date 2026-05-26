# Reference Benchmark Evidence Examples

Reference Benchmark Evidence is used when a project is moving from Beta to Service Quality, or when it is already in Service Quality.

It is not a generic score. It answers one product question:

> Is this service at least at the category baseline, and does it have one clear strength?

Hyper Run treats the evidence as covered only when it includes:

- Category
- 3-5 named references
- Baseline expectations
- Current comparison using below/meets/above baseline
- Below-baseline gaps with no critical user or operator gap remaining
- Above-baseline strength
- Decision

## Covered Example

```md
## Reference Benchmark Evidence

- Category: Developer CLI / project-growth runtime
- References: namba-ai, pi.dev, Claude Code
- Baseline expectations: install is clear; one command creates useful work context; status and recovery are understandable
- Current comparison: setup meets baseline; status output meets baseline; project-local evidence pressure is above baseline
- Below-baseline gaps: None; no critical user or operator baseline gap remains
- Above-baseline strength: project-local evidence, readiness pressure, and finish-gate review are stronger than a plain prompt wrapper
- Decision: Service Quality is allowed from the benchmark perspective; next pressure should reduce auto-continuation friction
```

Why this is covered:

- It names a category.
- It lists 3 references.
- It states the category baseline.
- It compares the current product with below/meets/above-baseline language.
- It says no critical baseline gap remains.
- It names one above-baseline strength.
- It makes a decision and turns the comparison into next pressure.

## Emerging Example

```md
## Reference Benchmark Evidence

- Category: Developer CLI
- References: similar tools
- Baseline expectations: should be easy to use
- Current comparison: good enough
- Below-baseline gaps: Pending
- Above-baseline strength: evidence loop
- Decision: continue
```

Why this is emerging:

- References are not named.
- The baseline is too vague.
- Current comparison does not say below/meets/above baseline.
- Below-baseline gaps are still pending.
- The decision does not create clear next pressure.

## Blocked Example

```md
## Reference Benchmark Evidence

- Category: Local-first note app
- References: Apple Notes, Notion, Obsidian
- Baseline expectations: a user can create, find, edit, and recover notes from a documented path
- Current comparison: create and edit meet baseline; recovery is below baseline; setup meets baseline
- Below-baseline gaps: recovery path is below baseline because deleted or corrupted notes cannot be restored
- Above-baseline strength: local setup is simpler than the references
- Decision: Service Quality is blocked until recovery reaches the category baseline
```

Why this blocks Service Quality:

- The evidence is useful, but it names a critical below-baseline gap.
- Hyper Run should turn that gap into the next runtime pressure instead of advancing the stage.

## Status Output Examples

When the benchmark is required but missing:

```text
Hyper Run Status
Project: Tiny CRM
Stage: Beta
Gate: Beta -> Service Quality (not_ready)
Proof: functional covered, surface covered, operational covered, benchmark missing
Next: hyper run "Compare Tiny CRM against 3-5 category references and close any core below-baseline gap."
Benchmark: missing - Reference comparison has not proven category baseline and differentiating strength.
Gap: Reference benchmark: Reference comparison has not proven category baseline and differentiating strength.
Guard: Do not advance Beta until blocking readiness gaps are closed.
```

When the benchmark is covered:

```text
Hyper Run Status
Project: Tiny CRM
Stage: Beta
Gate: Beta -> Service Quality (ready)
Proof: functional covered, surface covered, operational covered, benchmark covered
Next: hyper advance
Benchmark: covered - GOAL-0008 readiness evidence: Category: Developer CLI; References: namba-ai, pi.dev, Claude Code; ...
Gap: none; stage advancement is ready
Guard: accept the stage change before running `hyper advance`
```
