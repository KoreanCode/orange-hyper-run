# 변경 기록

## Unreleased

- Service Quality Self Review gate를 추가했습니다. packet 완료 전에 plan alignment, core loop quality, product satisfaction, no drift, validation match, pass/fail verdict를 요구합니다.
- Self Review verdict가 `fail`이면 Service Quality packet을 닫지 않고 같은 packet에서 재작업하게 합니다.
- Product satisfaction을 Beta, Service Quality, Sustained Service Quality gate의 readiness axis로 추가했습니다.
- agent가 제품 방향을 조용히 넓히지 않고 blocker로 기록하도록 runtime packet의 work boundary와 stop condition에 no-drift guard를 추가했습니다.
- `plan.md`에 `Target Stage`를 정의하면 plain `hyper run`이 그 목표까지 guarded auto continuation으로 동작하게 했습니다.
- plan target에서 온 continuation 명령은 plain `hyper run`으로 유지하고, `--auto --until`은 명시적인 override로 남겼습니다.
- `.hyper/next-packet.md`에 Codex Desktop continuation 안내를 추가해 auto mode가 run, repair, advance, stop 중 무엇을 해야 하는지 더 명확하게 했습니다.
- auto mode의 `hyper run`과 `hyper resume` Codex Desktop payload에 next-packet continuation 안내를 포함했습니다.

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
