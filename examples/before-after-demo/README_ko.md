# Before / After 데모

Hyper Run을 1분 안에 설명하기 위한 실사용 데모 스크립트입니다. 아직 완성된 GIF는 아니고, 이후 녹화의 기준 시나리오입니다.

## 문제

프로젝트 로컬 runtime memory가 없으면 코딩 에이전트는 세션 사이에 문맥을 잃기 쉽습니다.

```text
Session 1:
- 작은 MVP를 만든다.
- storage는 local-first로 유지하기로 결정한다.
- `npm run build`를 실행한다.
- deployment는 아직 준비되지 않았음을 확인한다.

Session 2:
- 에이전트가 다시 넓은 prompt에서 시작한다.
- local-first 결정을 잊을 수 있다.
- 같은 validation을 건너뛸 수 있다.
- 너무 이른 harness나 큰 구조를 만들려고 할 수 있다.
```

## Hyper Run을 쓰면

한 번 초기화합니다.

```bash
hyper init
```

`plan.md`를 채운 뒤 실행합니다.

```bash
hyper run "가장 작은 사용 가능한 MVP를 만들어줘"
```

Hyper Run은 다음 파일을 만듭니다.

```text
.hyper/goals/GOAL-0001/
  goal.md
  tasks.md
  evidence.md
  next.md
```

에이전트가 packet을 구현한 뒤 evidence를 남깁니다.

```text
## Validation

`npm run build` passed.

## Decisions

Keep storage local-first.

## Readiness Evidence

Core UX: Browser smoke passed for create and complete flow.
Validation coverage: `npm run build` passed and is repeatable.
```

packet을 닫습니다.

```bash
hyper complete
```

상태를 봅니다.

```bash
hyper status
```

예시 결과:

```text
Action:
  Next action: hyper advance
  Why now: Tiny MVP gate is ready.
  Do not do yet: Do not run `hyper advance` unless the user accepts the stage advancement.
```

## 무엇이 달라졌나

Hyper Run 전에는 다음 AI 세션이 넓은 prompt와 채팅 기억에 의존합니다.

Hyper Run 후에는 다음 AI 세션이 프로젝트 안의 상태를 읽습니다.

- `plan.md`: 제품 의도
- `.hyper/goals/.../evidence.md`: 증명된 내용
- `.hyper/goals/.../next.md`: 다음 추천 episode
- `.hyper/next-packet.md`: 다음 실행 명령 계획
- `.hyper/growth/state.json`: 반복 pressure
- `.hyper/readiness/state.json`: stage readiness

코드는 여전히 에이전트가 구현하지만, 이제 프로젝트가 자기 문맥을 직접 보관합니다.

## 터미널 데모 스크립트

이 transcript를 터미널 녹화의 기준으로 사용할 수 있습니다.

### Before

```text
$ codex "가장 작은 task app MVP를 만들어줘"
...
`npm run build` 통과.
Decision: storage는 local-first로 유지.
Deployment는 아직 준비되지 않음.

$ codex "task app 이어서 진행해줘"
...
Agent가 storage를 다시 물어본다.
Agent가 `npm run build`가 기대 validation인지 모른다.
MVP flow가 안정되기 전에 넓은 architecture를 제안한다.
```

### After

```bash
hyper init
# plan.md에 product, users, MVP, stage, constraints를 적습니다
hyper run "가장 작은 task app MVP를 만들어줘"
```

생성된 packet을 보여줍니다.

```text
.hyper/goals/GOAL-0001/
  goal.md
  tasks.md
  evidence.md
  review.md
  next.md
```

구현 후:

```bash
hyper complete
hyper status --short
```

기대 status 형태:

```text
Project: Pocket Tasks
Stage: Tiny MVP
Gate: Tiny MVP -> Usable MVP (ready)
Proof: functional covered, surface covered, operational covered
Next: hyper advance
Guard: accept the stage change before running `hyper advance`
```

다음 세션은 채팅 기억이 아니라 프로젝트 파일에서 이어받습니다.

```bash
hyper resume
hyper status --short
```

핵심 차이는 Hyper Run이 앱을 대신 만든다는 것이 아닙니다. 다음 agent가 제품 단계, 이미 결정한 내용, 검증 명령, 다음 작업 경계를 다시 발견하지 않아도 된다는 점입니다.

## 녹화 체크리스트

- README와 source file만 있는 작은 프로젝트를 보여줍니다.
- `hyper init`을 실행합니다.
- 작은 `plan.md`를 채웁니다.
- `hyper run`을 실행합니다.
- `goal.md`를 보여줍니다.
- 작은 구현을 시뮬레이션하거나 실제 수행합니다.
- `evidence.md`와 `next.md`를 채웁니다.
- `hyper complete`를 실행합니다.
- `hyper status --short`를 실행합니다.
- `Action` 섹션과 readiness gate를 강조합니다.
- 새 터미널 또는 agent session에서 `hyper resume`을 실행해 프로젝트가 handoff를 가지고 있음을 보여줍니다.
