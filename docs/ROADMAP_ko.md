# 로드맵

이 로드맵은 현재 제품 방향을 설명합니다. Hyper Run은 진입점은 작게 유지하고, 프로젝트 evidence가 필요를 증명할 때만 구조를 키우는 방향을 유지합니다.

## 가까운 단계

- `hyper run`을 주 사용자 명령으로 유지합니다.
- `hyper status`, `hyper doctor`, `hyper repair`, `hyper migrate`를 일상적으로 믿고 쓸 수 있게 만듭니다.
- checksum 검증 다음 단계로 signing과 더 명확한 검증 문서를 통해 릴리즈 신뢰도를 높입니다.
- Learn 품질을 더 엄격히 해서 durable한 프로젝트 신호만 memory가 되게 합니다.
- 코딩 에이전트가 추가 설명 없이 실행할 수 있도록 runtime packet을 더 명확하게 만듭니다.

## 다음 단계

- 오래된 `.hyper/` 프로젝트 상태에 대한 migration 커버리지를 늘립니다.
- web app, CLI, desktop app 예제를 더 명확하게 추가합니다.
- maintainer용 release checklist를 추가합니다.
- 도메인 경계가 안정된 뒤 패키지 분리를 시작합니다.

## 이후 단계

- release binary에 cosign 서명을 검토합니다.
- 터미널 녹화나 짧은 GIF 같은 실사용 데모를 추가합니다.
- 더 엄격한 gate를 원하는 팀을 위해 agent/capability activation policy를 선택적으로 제공합니다.
- state, storage, growth, readiness, packet generation 같은 내부 패키지 분리를 검토합니다.

## 지금 하지 않는 것

- Hyper Run을 완전한 프로젝트 관리 앱으로 만들지 않습니다.
- 반복 evidence 없이 harness를 먼저 요구하지 않습니다.
- static skill을 프로젝트 truth source로 만들지 않습니다.
- 사용자 확인 없이 stage를 자동으로 올리지 않습니다.
