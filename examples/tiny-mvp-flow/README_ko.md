# Tiny MVP Flow 예제

이 예제는 Hyper Run의 Golden Path를 보여줍니다. 작은 product plan이 하나의 runtime packet이 되고, agent가 packet을 완료하고, evidence를 남기고, Learn signal이 다음 context가 되는 흐름입니다.

예제 제품은 local-first browser task list인 `Pocket Tasks`입니다. 애플리케이션 소스는 일부러 넣지 않았습니다. 이 폴더는 Hyper Run의 service loop를 설명하는 산출물에 집중합니다.

## 명령 흐름

```bash
hyper init
# plan.md를 채웁니다
hyper run "Build the smallest local task list MVP"
# Codex Desktop이 GOAL-0001을 실행하고 evidence.md / next.md를 업데이트합니다
hyper complete
hyper run "Add persistence polish after the core flow works"
```

`hyper complete`는 packet을 닫고 Learn, Growth, Readiness를 갱신합니다. 두 번째 `hyper run`은 이 갱신된 project state에서 시작해야 합니다. `hyper internal learn`은 learning을 수동으로 확인하거나 디버깅할 때만 사용합니다.

Codex Desktop에서는 다음처럼 사용합니다.

```text
$hyper run Build the smallest local task list MVP
```

## 읽을 파일

```text
plan.md
.hyper/goals/GOAL-0001/
  goal.md
  tasks.md
  evidence.md
  next.md
.hyper/capabilities/
  candidates/
    validator/
  active/
    validator/
  retired/
    validator/
.hyper/growth/
  state.json
.hyper/readiness/
  state.json
.hyper/memories/
  decisions.md
  patterns.md
  constraints.md
  failures.md
```

## 이 예제가 보여주는 것

- `plan.md`는 사람이 관리하는 가벼운 product brief로 남습니다.
- `goal.md`는 permanent spec이 아니라 runtime packet입니다.
- `tasks.md`는 한 episode를 위한 execution checklist입니다.
- `evidence.md`는 validation, axis-slot readiness evidence, active capability evidence, changed files, decisions, reusable patterns, blockers를 기록합니다.
- `next.md`는 다음 runtime episode를 추천하고 structured Learn Notes를 남깁니다.
- `.hyper/memories/`는 이후 packet이 가져올 durable signal을 저장합니다.
- `.hyper/growth/state.json`은 다음 packet의 boundary와 validation behavior를 바꾸는 pressure를 저장합니다.
- `.hyper/readiness/state.json`은 MVP 작업이 service quality로 계속 이동하도록 stage-gate readiness pressure를 저장합니다.
- `.hyper/capabilities/`는 repeated pressure가 candidate, promotable structure, active structure, retired structure가 될 때 lifecycle metadata를 저장합니다.
- `.hyper/capabilities/active/validator/` 아래 active validator는 다음 runtime packet에서 required validation behavior가 됩니다.

## Golden Path 결과

`GOAL-0001`과 `hyper complete` 이후 프로젝트에는 작동하는 local MVP flow 하나와 다음 작업으로 이어갈 evidence가 생깁니다. 다음 `hyper run`은 같은 결정을 다시 발견하지 않아야 합니다. Tiny MVP에서는 localStorage가 storage 선택이고, browser smoke test가 validation pattern이며, credentials가 생기기 전까지 external service가 scope 밖이라는 것을 알고 시작해야 합니다.
