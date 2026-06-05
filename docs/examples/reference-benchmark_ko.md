# Reference Benchmark Evidence 예시

Reference Benchmark Evidence는 프로젝트가 Beta에서 Service Quality로 넘어가거나, 이미 Service Quality 단계에 있을 때 사용합니다.

일반 점수표가 아닙니다. 한 가지 제품 질문에 답합니다.

> 이 서비스는 category baseline 이상인가? 그리고 명확한 강점이 하나 이상 있는가?

Hyper Run은 아래 항목이 모두 있을 때만 covered로 봅니다.

- Category
- 3-5개의 named reference
- Baseline expectations
- below/meets/above baseline으로 표현된 Current comparison
- critical user/operator gap이 없다는 Below-baseline gaps
- Above-baseline strength
- Decision

## Covered 예시

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

covered인 이유:

- category가 있습니다.
- reference 3개가 이름으로 적혀 있습니다.
- category baseline이 있습니다.
- 현재 제품을 below/meets/above baseline 기준으로 비교합니다.
- critical baseline gap이 없다고 명시합니다.
- above-baseline strength가 있습니다.
- decision이 있고, 비교 결과를 다음 pressure로 바꿉니다.

Below-baseline line에 below-baseline 영역이 언급된다면, 그 영역이 현재 service boundary에서 왜 critical하지 않은지 명시해야 합니다. 예를 들어 deferred, out of scope, explicit non-goal이어야 합니다.

## Emerging 예시

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

emerging인 이유:

- reference 이름이 없습니다.
- baseline이 너무 모호합니다.
- current comparison이 below/meets/above baseline으로 표현되지 않았습니다.
- below-baseline gaps가 pending입니다.
- decision이 Service Quality 진행을 명시적으로 허용하지 않습니다.

## Blocked 예시

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

Service Quality가 blocked인 이유:

- 증거 자체는 유용합니다.
- 하지만 critical below-baseline gap이 남아 있습니다.
- decision이 추가 작업 전에는 Service Quality가 blocked라고 판단합니다.
- Hyper Run은 stage를 올리지 않고, 이 gap을 다음 runtime pressure로 바꿔야 합니다.

## Category Template

아래 예시는 시작점입니다. 실제 프로젝트의 evidence에 맞게 reference와 비교 내용을 바꿔야 합니다.

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

## Status 출력 예시

benchmark가 필요한데 missing인 경우:

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

benchmark가 covered인 경우:

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
