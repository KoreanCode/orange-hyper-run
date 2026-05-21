# Hyper Run 서비스 정의

Hyper Run은 사람이 관리하는 `plan.md`를 반복 가능한 runtime packet, evidence, durable learning으로 바꿔서 작은 MVP가 무거운 하네스 없이 큰 프로젝트로 성장할 수 있게 하는 프로젝트 성장 런타임입니다.

## 대상 사용자

Hyper Run은 Codex Desktop 또는 CLI agent로 여러 세션에 걸쳐 프로젝트를 만들고, 방향성, 실행 상태, 검증 증거, 재사용 가능한 학습을 프로젝트 안에 남기고 싶은 builder를 위한 도구입니다.

주요 사용자는 generic task manager, static spec system, 완전 자율 플랫폼을 원하는 사람이 아닙니다. 가벼운 plan에서 시작해 다음 coherent step을 실행하고, 배운 것을 보존하며, 프로젝트가 충분히 성장했을 때만 generated harness를 만들고 싶은 사람입니다.

## 문제

Agent 작업은 작은 task에서는 잘 작동하지만 큰 프로젝트로 이어질 때 자주 약해집니다. 세션마다 context가 사라지고, 매 요청을 새 task처럼 다루거나, 제품 형태가 안정되기 전에 static spec에 너무 많은 비용을 씁니다.

Harness-first 방식은 반복 workflow가 안정된 뒤에는 도움이 됩니다. 하지만 tiny MVP 단계에서는 무겁습니다. 작은 MVP에는 빠른 실행이 필요하고, 성장하는 프로젝트에는 완전한 harness보다 먼저 memory, evidence, 반복 가능한 run boundary가 필요합니다.

## 제품 약속

Hyper Run이 보장하려는 loop는 작고 반복 가능합니다.

1. 사람이 관리하는 `plan.md`에서 시작합니다.
2. 다음 실행 episode를 위한 runtime packet 하나를 생성합니다.
3. Codex Desktop 또는 다른 agent가 현재 repo 안에서 그 packet을 실행합니다.
4. validation evidence, changed files, decisions, reusable patterns, blockers, next steps를 기록합니다.
5. 다음 run에 영향을 줘야 하는 durable signal만 학습합니다.
6. 다음 runtime packet을 만들 때 유사한 이전 context를 가져옵니다.
7. product, UX, persistence, validation, security, deployment, operations, maintainability 축으로 service readiness를 측정합니다.
8. 프로젝트가 증거를 통해 필요성을 보였을 때만 generated skill, agent, validator, harness로 성장합니다.

## Hyper Run인 것

- `plan.md`, `.hyper/`, `hyper` CLI를 중심으로 한 로컬 프로젝트 런타임
- Codex Desktop에서 `$hyper run`으로 사용할 수 있는 command convention
- 다음 coherent execution episode를 만드는 runtime packet generator
- 파일과 SQLite 기반의 로컬 evidence/learning layer
- tiny MVP 작업이 usable, beta, service-quality stage로 이어지도록 하는 service-readiness gate
- 프로젝트가 충분히 성장한 뒤 필요한 harness를 만들 수 있게 하는 harness-less starting point

## Hyper Run이 아닌 것

- Codex Desktop 또는 coding agent의 대체재가 아닙니다.
- Cloud agent platform이 아닙니다.
- Static SPEC manager가 아닙니다.
- Project management app이 아닙니다.
- Test framework가 아닙니다.
- 필수 TUI가 아닙니다.
- 모든 run이 사용자 판단, credentials, validation 없이 완료된다는 약속이 아닙니다.

## Harness와의 관계

Hyper Run은 프로젝트 생애주기에서 harness보다 위쪽에 있습니다.

Harness는 반복 workflow, 안정된 validation path, 명확한 execution boundary가 생긴 뒤 유용합니다. Hyper Run은 그 이전 단계에 존재합니다. 프로젝트가 harness 없이 실행되고, evidence를 모으고, 어떤 harness가 필요한지 발견하게 합니다.

의도한 흐름은 다음과 같습니다.

```text
plan.md
  -> runtime packets
  -> evidence and Learn signals
  -> repeated patterns
  -> generated validators, skills, agents, or harnesses when needed
```

## Run Contract

하나의 `hyper run`은 정확히 하나의 runtime packet을 만듭니다. 그 packet은 실행 agent가 다음을 완료했을 때 끝난 것으로 봅니다.

- `plan.md`, `goal.md`, `tasks.md`를 읽었습니다.
- packet의 `Stage Gate`와 selected readiness pressure를 확인했습니다.
- 현재 episode를 위한 가장 작은 coherent step을 구현했습니다.
- 실행 가능한 가장 안전한 validation을 돌렸거나, validation이 막힌 이유를 기록했습니다.
- `evidence.md`에 validation output, readiness evidence, active capability evidence, changed files, decisions, reusable patterns, blockers를 기록했습니다.
- `next.md`에 다음 추천 runtime episode와 structured Learn Notes를 기록했습니다.
- `hyper complete`를 실행해 Learn, Growth, Readiness가 완료된 packet 기준으로 갱신되었습니다.
- destructive action, missing credentials, unclear product scope, repeated validation failure 앞에서 멈췄습니다.

`hyper run` 하나를 무한 background loop로 보면 안 됩니다. 무한한 것은 unchecked command 하나가 아니라, 반복되는 프로젝트 성장 loop입니다. 이전 active packet의 evidence가 아직 pending이면 새 `hyper run`은 차단됩니다.

## Learn 역할

Learn은 요약 시스템이 아닙니다.

Learn은 완료되었거나 막힌 runtime packet에서 durable signal을 추출합니다.

- `decision`: 이후 run이 존중해야 하는 제품 또는 기술 결정
- `pattern`: 재사용 가능한 구현 또는 검증 방식
- `constraint`: 이후 run이 어기면 안 되는 경계
- `failure`: 반복하면 안 되는 blocker 또는 실패한 접근

Learn은 ordinary progress notes, changed-file lists, temporary observations를 그대로 저장하지 않습니다. durable decision, pattern, constraint, failure가 들어 있을 때만 memory로 봅니다.

## Service Readiness 역할

Readiness는 "service quality까지 계속 간다"는 말을 구체적인 실행 판단으로 바꾸는 부분입니다.

Hyper Run은 `.hyper/readiness/state.json`에 readiness axis, current stage gate, blocking gap, next selected pressure를 씁니다. 다음 runtime packet은 이 state를 사용해 지금 무엇을 진행해야 하는지, 어떤 evidence가 필요한지, 언제 stage advancement를 주장하면 안 되는지 판단합니다.

Readiness evidence는 누적 진행도로 반영됩니다. 새 evidence 파일에는 모든 readiness axis 슬롯이 들어갑니다. `evidence.md`에 `Data persistence: Customer records survive reload` 같은 axis-labeled line이 있으면, `hyper complete`와 `hyper status`는 해당 axis를 `covered`로 올리고, 관련 gate gap을 제거하고, 해결된 pressure를 반복하지 않고 다음으로 약한 pressure를 선택합니다.

Readiness evidence에는 기본 품질 기준도 있습니다. `Validation coverage: tested`처럼 모호한 label은 covered가 아니라 emerging evidence로 봅니다. Covered evidence는 axis에 맞는 proof shape이 있어야 합니다. Validation은 command 또는 smoke test, UX는 browser 또는 screenshot proof, persistence는 reload/restart/storage proof, deployment는 hosted/build/release proof, operations는 docs/runbook/rollback proof가 필요합니다.

현재 gate의 required axis가 모두 covered가 되면 Hyper Run은 stage advancement candidate를 만듭니다. 정확한 `plan.md` `Current Stage` 변경을 권고하지만 자동 적용하지 않습니다. Stage 이동은 사람이 검토하되, 프로젝트 상태는 명확히 드러나게 합니다.

Beta와 Service Quality stage에서는 repeatable smoke, security, deployment, operations check를 위한 quiet validator candidate를 만들 수 있습니다. 이들은 repeated evidence로 active validator가 되기 전까지 candidate로 남습니다.

기본 readiness path는 다음과 같습니다.

```text
Tiny MVP -> Usable MVP -> Beta -> Service Quality
```

Readiness는 Learn이나 Growth를 대체하지 않습니다. Learn은 durable signal을 추출합니다. Growth는 repeated signal을 behavior와 capability로 바꿉니다. Readiness는 프로젝트가 실제 service에 가까워지고 있는지 묻고, 가장 약한 missing axis를 다음 run의 압력으로 선택합니다.

## 서비스 경계

Hyper Run의 user-facing layer는 작게 유지해야 합니다.

- `hyper init`: project-local runtime files를 초기화합니다.
- `hyper run [focus]`: 다음 runtime packet을 만듭니다.
- `hyper complete`: active packet을 닫고 Learn, Growth, Readiness를 갱신합니다.
- `hyper resume`: active packet을 재개합니다.
- `hyper status`: 현재 runtime state를 보여줍니다.
- `hyper update`: native binary를 업데이트합니다.

새 기능은 곧바로 top-level command가 되기보다, 먼저 `.hyper/` 아래 generated project knowledge로 존재하는 편이 좋습니다.

## 성공 기준

Hyper Run은 사용자가 다음을 할 수 있을 때 성공입니다.

- `hyper init`, `plan.md`, `hyper run`, `hyper complete`만으로 프로젝트를 시작합니다.
- Harness를 먼저 설계하지 않고 tiny MVP를 만듭니다.
- 여러 세션에 걸쳐 같은 프로젝트를 이어가며 유용한 context를 회수합니다.
- 무엇이 바뀌었고, 무엇이 통과했고, 무엇이 실패했고, 다음에 무엇을 해야 하는지 evidence로 확인합니다.
- Learn signal이 충분히 쌓였을 때만 project-specific validator, skill, agent, harness를 만듭니다.

## 현재 MVP 경계

현재 제품은 다음에 집중해야 합니다.

- Native CLI install/update
- Project-local initialization
- Runtime packet generation
- Codex Desktop `$hyper` routing
- Evidence and next-step templates
- Durable Learn extraction
- Project Growth Engine pressure state
- Service Readiness state와 Stage Gate runtime compilation
- 유사 signal clustering과 noisy signal filtering
- repeated pressure에서 active structure로 이어지는 capability lifecycle
- Growth-informed runtime packet compilation
- Threshold 뒤에서 조용히 생성되는 validator, skill, harness candidate
- Similar-context retrieval
- 명확한 Golden Path example

그 밖의 기능은 product surface area가 되기 전에 run contract를 기준으로 평가해야 합니다.
