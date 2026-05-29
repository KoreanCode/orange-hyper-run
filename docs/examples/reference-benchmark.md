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

If the below-baseline line mentions a below-baseline area, it must explicitly say why that area is not critical for the current service boundary, for example because it is deferred, out of scope, or an explicit non-goal.

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

## Category Templates

Use these as starting points. Replace the references and comparisons with evidence from the actual project.

### Web App

```md
## Reference Benchmark Evidence

- Category: Small authenticated web app
- References: Linear, Notion, Trello
- Baseline expectations: a user can sign in, complete the primary workflow, recover from empty/error states, and understand what changed after an action
- Current comparison: sign-in and primary workflow meet baseline; empty/error states meet baseline; keyboard-first speed is below Linear but not critical for this product stage; project-local recovery notes are above baseline
- Below-baseline gaps: None critical; advanced keyboard navigation is deferred and documented as a non-goal
- Above-baseline strength: the app has clearer project-local validation and rollback notes than a typical small MVP
- Decision: Service Quality is allowed from the benchmark perspective; next pressure should improve keyboard shortcuts only after operational checks stay green
```

### CLI Tool

```md
## Reference Benchmark Evidence

- Category: Developer CLI
- References: GitHub CLI, Vercel CLI, Railway CLI
- Baseline expectations: install is clear, help output explains common commands, errors tell the user what to do next, and update behavior is documented
- Current comparison: install meets baseline; help output meets baseline; update checksum verification is above baseline; interactive onboarding is below Vercel CLI but not critical
- Below-baseline gaps: None critical; interactive onboarding is deferred because command help and doctor output cover the current operator path
- Above-baseline strength: checksum-verified update plus project-local doctor checks are stronger than a plain download script
- Decision: Service Quality is allowed from the benchmark perspective; next pressure should reduce first-run wording friction
```

### Local-First App

```md
## Reference Benchmark Evidence

- Category: Local-first notes or task app
- References: Apple Notes, Obsidian, Todoist
- Baseline expectations: a user can create, edit, list, search, and recover local data from documented steps
- Current comparison: create/edit/list meet baseline; search meets baseline for the MVP scope; recovery and backup meet baseline through documented export; collaboration is below baseline for Todoist but out of scope
- Below-baseline gaps: None critical; collaboration is an explicit non-goal for the current service boundary
- Above-baseline strength: local data ownership and documented export/recovery are clearer than many small web MVPs
- Decision: Service Quality is allowed from the benchmark perspective; next pressure should monitor data migration evidence before adding sync
```

### Design-Heavy App

```md
## Reference Benchmark Evidence

- Category: Design-heavy consumer web app
- References: Airbnb, Duolingo, Arc
- Baseline expectations: the first screen communicates the product, primary action is obvious, responsive layout works, and visual style supports the product domain
- Current comparison: primary action and responsive layout meet baseline; visual polish meets baseline; motion depth is below Duolingo but not critical; domain-specific visual identity is above baseline for the MVP
- Below-baseline gaps: None critical; advanced animation polish is deferred until performance and accessibility checks stay green
- Above-baseline strength: the product-specific visual system is clearer than a generic template UI
- Decision: Service Quality is allowed from the benchmark perspective; next pressure should add accessibility and performance evidence for the visual system
```

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
