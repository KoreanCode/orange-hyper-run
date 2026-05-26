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
- decision이 다음 pressure로 이어지지 않습니다.

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
- Hyper Run은 stage를 올리지 않고, 이 gap을 다음 runtime pressure로 바꿔야 합니다.

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
