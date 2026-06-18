# 변경 기록

## Unreleased

## v0.6.8 - 2026-06-18

- 일반 사용자의 흐름을 `hyper run` 중심으로 유지하도록 README, status, doctor, run-blocking, repair, advance, migrate, resume, 생성된 Codex routing 출력에서 `hyper complete`를 agent finish gate 및 수동 복구 명령으로 설명하게 했습니다.
- 생성 runtime handoff를 갱신했습니다. `goal.md`/`tasks.md`에는 finish gate를 계속 명시하되, user-facing payload, status, doctor는 agent가 evidence/next notes를 마무리하고 gate를 내부적으로 실행하도록 안내합니다.
- `hyper migrate`가 project-installed `$hyper-run` skill routing을 최신 agent finish-gate wording에 맞게 갱신하게 했습니다.
- local secret, certificate, log, coverage, 임시 build output, editor/cache file이 Git에 들어가지 않도록 `.gitignore`를 보강했습니다.

## v0.6.7 - 2026-06-16

- Decision Hierarchy, Autonomous Work Plan, Autonomous Safety Policy, Capability Expansion Policy, Research Evidence Policy, Loop Progress Policy, Product Satisfaction Policy를 runtime packet에 생성하고, 이에 맞는 task/evidence row를 함께 생성하게 했습니다.
- local source propagation 경로를 더 안전하게 문서화했습니다. locally built `hyper`를 먼저 검증한 뒤, 사용자가 해당 checkout을 active PATH executable로 바꾸려는 경우에만 의도적으로 `~/.local/bin/hyper`에 설치하도록 안내합니다.
- runtime template smoke test와 AI-assisted 작업의 승인 경계를 release checklist에 추가했습니다. PATH install, tag, push, release publish는 명시 승인 전에는 진행하지 않습니다.
- `plan.md`에 `Target Stage`가 없을 때 `hyper status`와 `hyper doctor`가 single-packet mode를 더 분명하게 설명하고, Service Quality / Sustained Service Quality 단계에 맞춘 안내를 보여줍니다.
- 다음 packet reason이 packet의 `next.md` 추천을 그대로 따르는 경우와 surface-proof gap을 먼저 닫기 위해 추천을 override하는 경우를 구분하게 했습니다.

## v0.6.6 - 2026-06-08

- generic 또는 Sustained Service Quality 후속 작업을 계획할 때 완료된 packet의 `next.md` Recommended Next Goal을 `.hyper/next-packet.md`에 반영합니다.
- stage advancement, target 완료, 같은 packet 보강은 packet next-goal 추천보다 계속 우선합니다.
- 마지막 packet에 browser, screenshot, surface proof gap이 남아 있으면 visual/accessibility surface-proof 후속 작업을 우선 추천합니다.
- `hyper status`와 `hyper doctor`도 같은 next-goal-aware 계획을 검증해 오래된 generic next-packet handoff를 감지하게 했습니다.

## v0.6.5 - 2026-06-05

- 서비스 품질까지 계속 개발하려는 긴 focus를 쓰면서 `plan.md`에 `Target Stage`가 없으면, plain `hyper run`이 자동 continuation이 아니라 single packet만 만든다는 점을 안내합니다.
- 구현과 command validation은 통과했지만 Browser URL policy 때문에 surface proof만 막힌 경우, 전체 packet을 blocked로 닫지 않고 focused follow-up이 필요한 completed packet으로 처리합니다.
- 일반 품질 작업을 이어가기 전에 허용된 browser proof나 반복 가능한 fallback surface check로 다음 packet을 계획하게 했습니다.
- SQLite status count를 깨끗하게 읽지 못할 때 `.hyper/logs`와 `.hyper/goals` 기준으로 fallback해서 `Runs recorded: 0` 같은 오해를 줄였습니다.

## v0.6.4 - 2026-06-05

- canonical stage vocabulary, target alias, stage ordering을 `internal/stage`로 분리해서 plan parsing, auto target, readiness, status가 같은 package boundary를 공유하게 했습니다.
- `service-quality`, `sustained-service-quality` 같은 slug-style stage 값을 Current Stage 정규화에서도 일관되게 받게 했습니다.
- govulncheck가 패치된 standard library 기준으로 실행되도록 CI와 release build를 Go `1.26.4`로 고정했습니다.
- Service Quality Self Review gate를 추가했습니다. packet 완료 전에 plan alignment, core loop quality, product satisfaction, no drift, validation match, pass/fail verdict를 요구합니다.
- Self Review verdict가 `fail`이면 Service Quality packet을 닫지 않고 같은 packet에서 보강하게 합니다.
- 실패한 Self Review의 구체적인 field와 verdict 내용을 finish-gate finding, next-packet correction plan, `hyper resume`에 포함하게 했습니다.
- Beta와 Service Quality finish gate에서 이미 충족된 경우가 아니면 Reference Benchmark Evidence를 요구하게 했습니다.
- 현재 readiness pressure가 reference benchmark인 경우에도 Reference Benchmark Evidence를 요구하고, 중복된 generic readiness finding은 만들지 않게 했습니다.
- Service Quality packet 완료 전에 Self Review, Reference Benchmark Evidence, active validator proof를 모두 만족해야 하는 테스트를 추가했습니다.
- Product satisfaction을 Beta, Service Quality, Sustained Service Quality gate의 readiness axis로 추가했습니다.
- agent가 제품 방향을 조용히 넓히지 않고 blocker로 기록하도록 runtime packet의 work boundary와 stop condition에 no-drift guard를 추가했습니다.
- candidate, promotable candidate, active required behavior가 구분되도록 threshold 기반 capability activation policy를 growth state, capability file, status output에 명시하게 했습니다.
- `plan.md`에 `Target Stage`를 정의하면 plain `hyper run`이 그 목표까지 guarded auto continuation으로 동작하게 했습니다.
- `Target Stage`는 해당 target stage의 readiness proof가 완료된 뒤에만 완료로 보게 했습니다. 그래서 `Target Stage: Service Quality`는 stage 진입 직후 멈추지 않고 Service Quality packet 안에서도 계속 진행합니다.
- plan target에서 온 continuation 명령은 plain `hyper run`으로 유지하고, `--auto --until`은 명시적인 override로 남겼습니다.
- 명시적인 `--until`을 runtime target source로 기록해 CLI 출력, `state.json`, 생성된 `goal.md`가 실수로 `plan.md` target을 설명하지 않게 했습니다.
- 이후 `hyper run --auto`가 이전 command-line target을 이어갈 때도 명시적인 `--until` source를 유지하게 했습니다.
- plain `hyper run`은 `plan.md` target으로 돌아가고, 생성된 `--auto --until` 명령은 명시 override를 유지한다는 규칙을 문서와 테스트로 고정했습니다.
- 활성 명시 `--until` override가 `plan.md` target과 다를 때 `hyper status`가 둘을 같이 보여주게 했습니다.
- `plan.md`의 `Target Stage`가 바뀌거나 제거되면 저장된 auto target도 함께 맞추게 했습니다.
- status, next-packet planning, `hyper advance`에 명시적인 stage advancement review 출력을 추가했습니다.
- active auto target에서는 Stage Advancement Review가 ready proof와 blocking gap 없음 상태를 보여준 뒤 `hyper advance`까지 이어갈 수 있게 했습니다.
- `plan.md Target Stage`가 잘못된 값이면 `hyper run`, `hyper doctor`와 동일하게 `hyper advance`도 stage 변경을 막게 했습니다.
- active auto target에서 stage gate가 ready일 때 불필요한 filler runtime packet을 만들지 않고, `hyper run`이 검토된 `hyper advance`로 다시 안내하게 했습니다.
- target-proof-complete, gate-ready advancement처럼 runtime packet을 만들지 않는 auto 판단도 project log와 SQLite event에 기록하게 했습니다.
- finish gate가 실패하면 `.hyper/next-packet.md`를 `complete-current`로 갱신하고, 오래된 실패 packet 상태도 migration으로 복구하게 했습니다.
- finish gate 실패 handoff를 쓸 때도 최신 `plan.md Target Stage` 변경 또는 제거를 반영하게 했습니다.
- `plan.md Target Stage`가 잘못된 값이면 `hyper complete`도 completion handoff를 쓰지 않고 먼저 막게 했습니다.
- `plan.md Target Stage`가 잘못된 값이면 `hyper status`가 migrate 대신 plan 수정으로 안내하게 했습니다.
- `plan.md Target Stage`가 잘못된 값이면 `hyper migrate`, `hyper repair`, `hyper resume`도 stale auto-continuation 상태를 쓰거나 보여주지 않게 막았습니다.
- 알 수 없는 stage 이름이 Tiny MVP로 조용히 fallback되지 않도록 `plan.md Current Stage`도 init, run, status, doctor, complete, advance, migrate, repair, resume 전체에서 검증하게 했습니다.
- 기존 `plan.md` stage field가 잘못된 경우 `hyper init`이 `.hyper/` routing state를 쓰기 전에 막히게 해서 실패한 초기화가 부작용을 남기지 않게 했습니다.
- CLI 업데이트 이후 다시 `hyper init`을 할 필요가 없도록 `hyper migrate`가 Codex Desktop routing file, generated command guide, 빠진 Hyper Run directory를 갱신하게 했습니다.
- active runtime packet이 있어도 `plan.md` stage field가 잘못되어 있으면 `hyper status`가 plan 수정을 우선 안내하게 했습니다. plan이 잘못된 상태에서는 completion과 continuation이 막히기 때문입니다.
- `hyper complete`, runtime packet을 만들지 않는 auto-run stop, `hyper advance`, `hyper repair` 출력에 next-packet planned action과 continuation guard를 함께 보여주게 했습니다.
- 같은 packet 재작업을 바로 시작할 수 있도록 `hyper status`와 `hyper status --short`에 현재 finish-gate review finding을 표시하게 했습니다.
- finish gate 실패 때문에 `hyper run`이 막힐 때도 현재 review finding을 함께 보여줘 loop가 같은 packet 재작업으로 향하게 했습니다.
- finish-gate 실패 에러에도 `Planned action: complete-current`와 continuation guard를 포함해 `.hyper/next-packet.md`를 열기 전에도 같은 packet 재작업이 필요하다는 점을 알 수 있게 했습니다.
- `not ready`, `insufficient`, `incomplete`, `not service-quality` 같은 Self Review verdict를 `ready` 단어가 들어 있다는 이유로 통과시키지 않고 finish-gate 실패로 처리하게 했습니다.
- Reference Benchmark decision이 Service Quality 진행을 명시적으로 허용해야 통과하게 했습니다. blocked, not ready, only after 같은 결정은 finish-gate 실패로 남습니다.
- Service Quality가 진행 가능하다고 함께 말하는 `not blocked`, `unblocked` benchmark/readiness 문구는 긍정 표현으로 처리해 blocker 관련 false failure가 나지 않게 했습니다.
- `complete-current` next-packet plan과 `hyper resume` 출력에 현재 `review.md` finding을 직접 표시하게 했습니다.
- `review.md`에 finish-gate evidence, next-note, finding hash를 기록하고, 같은 finding이 반복되면 auto continuation이 멈춰야 한다는 반복 실패 경고를 표시하게 했습니다.
- finish-gate 실패부터 같은 packet 보강, stage advancement, 다음 runtime packet 생성까지 이어지는 correction loop end-to-end 테스트를 추가했습니다.
- plain `hyper run`이 `--auto --until` 없이도 plan target 기준으로 Tiny MVP부터 Service Quality까지 stage를 올리고, Service Quality에서 Sustained Service Quality까지 이어간 뒤 목표에서 멈추는 multi-stage 테스트를 추가했습니다.
- finish gate가 실패한 상태에서는 `hyper status --short`가 새 작업 대신 `review.md` 보강과 같은 packet 재완료를 안내하게 했습니다.
- finish gate가 실패한 상태에서는 `hyper doctor`의 next action도 `review.md` 보강을 우선 안내하게 했습니다.
- `hyper status`와 `hyper status --short`에 다음 next-packet action을 표시하게 했습니다.
- `hyper status`와 `hyper status --short`에 `.hyper/next-packet.md` handoff 경로를 표시하게 했습니다.
- 일반 active packet이 아직 완료되지 않은 상태에서는 `.hyper/next-packet.md`가 이미 있는 것처럼 보이지 않게 했습니다.
- `hyper repair` 이후 갱신된 planned action과 `.hyper/next-packet.md` 경로를 출력하게 했습니다.
- `hyper migrate` 이후에도 갱신된 planned action과 next action을 출력하게 했습니다.
- 아직 끝나지 않은 active packet에서 `hyper migrate`가 evidence/next notes 작성 전 `hyper complete`를 바로 안내하지 않게 했습니다.
- `.hyper/next-packet.md`에 필요한 guard, continuation, stage advancement review section이 빠지면 `hyper doctor`가 경고하게 했습니다.
- auto target 변경 이후 `.hyper/next-packet.md`에 오래된 guard, continuation, advancement review 문구가 남아 있으면 `hyper doctor`가 경고하게 했습니다.
- `hyper doctor`가 `.hyper/next-packet.md`의 mode, reason, readiness gate, readiness pressure metadata 최신성도 검증하게 했습니다.
- 첫 runtime packet이 아직 없을 때는 `hyper doctor`가 `.hyper/next-packet.md`를 요구하지 않게 했습니다.
- Reference Benchmark Evidence의 below-baseline gap은 non-critical, deferred, out of scope, non-goal이 명시된 경우에만 통과하게 더 엄격하게 했습니다.
- `.hyper/next-packet.md`에 Codex Desktop continuation 안내를 추가해 auto mode가 run, 같은 packet 보강, advance, stop 중 무엇을 해야 하는지 더 명확하게 했습니다.
- `.hyper/next-packet.md`와 `hyper doctor`에 auto-mode Progress Guard를 추가해 진전 없는 반복 명령이나 반복 fix finding이 valid progress처럼 이어지지 않고 멈추게 했습니다.
- auto mode의 `hyper run`과 `hyper resume` Codex Desktop payload에 next-packet continuation 안내를 포함했습니다.
- Codex router skill에서 `run`, `advance`, `complete-current`, `stop` next-packet action별 행동을 더 명확하게 했습니다.
- `complete-current` handoff가 `hyper repair` 명령과 혼동되지 않도록 review/evidence/next notes를 고치는 흐름으로 명확히 했습니다.
- `blocked`와 `waiting_user` packet을 auto continuation의 terminal stop 상태로 처리해 blocker나 사용자 결정을 기다리게 하고, 다음 `hyper run`으로 이어가지 않게 했습니다.
- terminal blocked/waiting stop 이후 plan target을 쓰는 plain `hyper run`이 새 packet을 만들지 않게 하고, 명시적인 follow-up focus가 있을 때만 다음 packet을 시작하게 했습니다.

## v0.6.3 - 2026-05-29

- 명시적인 `현재 단계` heading이 `0단계: 화면 검증` 같은 roadmap heading보다 우선되도록 plan parsing을 수정했습니다.
- `state.json`에 저장된 오래된 stage가 `plan.md`와 다르면 `hyper status`와 `hyper doctor`가 refresh 경고를 보여줍니다.
- `hyper migrate`가 저장된 stage를 갱신하고 `.hyper/next-packet.md`를 수정된 stage와 맞춥니다.

## v0.6.2 - 2026-05-28

- 오래된 프로젝트 상태를 다룰 때 `hyper doctor`와 `hyper status --short`의 action guidance를 개선했습니다.
- `v0.6.x` 동작 기준으로 README와 보조 문서를 갱신했습니다.
- 영문/한글 maintainer release checklist를 추가했습니다.
- 업데이트 이후 확인 흐름과 문제 해결 흐름을 추가했습니다.
- before/after 데모와 reference benchmark 예시를 보강했습니다.

## v0.6.1 - 2026-05-27

- finish gate가 통과되기 전에는 다음 runtime packet을 시작할 수 없게 했습니다.
- `hyper complete`, `hyper migrate`, `hyper advance`, `hyper doctor`에서 `.hyper/next-packet.md`를 갱신하고 검증합니다.
- readiness evidence matching, active capability evidence, repeated validation grouping, command-pattern classification, failure-pressure 처리를 더 엄격하게 만들었습니다.
- first-run plan parsing, service-quality packet guidance, sustained quality flow, short status output, Windows path display를 개선했습니다.
- `hyper repair`가 실패한 finish gate를 우회하지 못하게 했습니다.

## v0.6.0 - 2026-05-26

- Service Quality reference benchmark evidence와 status 출력을 추가했습니다.
- 현재 pressure가 해당 축일 때 deployment, security, docs, operations, benchmark proof를 요구하도록 했습니다.
- README onboarding 문장을 더 쉽게 정리하고 제품 loop를 명확하게 설명했습니다.
- benchmark readiness helper 주변 staticcheck 문제를 수정했습니다.

## v0.5.6 - 2026-05-26

- 프로젝트 상태 변경 이후 readiness reconciliation을 수정했습니다.

## v0.5.5 - 2026-05-26

- 약하거나 noisy한 memory signal을 걸러내는 Learn quality gate를 추가했습니다.
- `hyper migrate`에서 legacy memory quality를 현재 규칙으로 보정하고 fixture 테스트를 추가했습니다.
- release asset에 cosign keyless signature bundle을 만들고 install/update에서 선택적으로 검증합니다.

## v0.5.4 - 2026-05-26

- `hyper update`가 GitHub release checksum을 검증하도록 개선했습니다.
- SHA256 checksum 검증을 포함한 Windows PowerShell installer를 추가했습니다.
- trusted install/update 검증 테스트를 추가했습니다.

## v0.5.3 - 2026-05-26

- finish-gate review와 auto continuation planning을 추가했습니다.
- 다음 명령 handoff로 `.hyper/next-packet.md`를 추가했습니다.
- short status와 completion guidance를 개선했습니다.

## v0.5.2 - 2026-05-22

- `hyper advance`를 사용하는 명시적인 stage advancement workflow를 추가했습니다.
- stage 변경을 자동 적용하지 않고 추천만 하도록 했습니다.

## v0.5.1 - 2026-05-22

- CLI entrypoint와 application runtime package를 분리했습니다.
- functional, surface, operational proof를 담는 Proof Contract 섹션을 추가했습니다.
- Surface Proof Evidence 템플릿, readiness 추출, growth pressure 학습, proof gap status 출력을 추가했습니다.
- 실제 프로젝트 실행 evidence에서 readiness를 더 잘 추론하도록 개선했습니다.
- status 출력에서 한국어 product plan alias를 읽도록 개선했습니다.

## v0.5.0 - 2026-05-22

- PR/push마다 test, vet, staticcheck, govulncheck가 도는 CI를 추가합니다.
- GitHub release 설치 경로에 대해 `install.sh` checksum 검증을 추가합니다.
- README에 macOS와 Windows 설치 방법을 추가합니다.
- roadmap, known limitations, before/after 데모 문서를 추가합니다.
- Windows CI에서도 plan import candidate 경로가 안정적으로 출력되도록 정규화합니다.

## v0.4.1

- growth status 표시를 개선했습니다.
- `hyper status`에서 command 기반 candidate 이름을 짧게 표시합니다.
- passive/noisy growth signal을 status 요약에서 숨깁니다.
- plan으로 이미 충족된 product readiness가 약한 runtime evidence로 downgrade되지 않게 했습니다.

## v0.4.0

- evidence-first growth protocol을 더 명확히 정의했습니다.
- pressure ledger, readiness, capability candidate 동작을 추가했습니다.
- native binary와 checksums 릴리즈 자동화를 추가했습니다.

## 이전 릴리즈

- native Go CLI를 추가했습니다.
- `hyper init`, `hyper run`, `hyper complete`, `hyper status`, `hyper doctor`, `hyper update`, Codex Desktop routing file을 추가했습니다.
