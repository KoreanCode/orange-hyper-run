# 알려진 한계

Hyper Run은 사용할 수 있는 상태지만, 일부는 아직 초기 단계입니다. 사용자가 어디까지 안정적인지 판단할 수 있도록 명확히 적습니다.

## 일상 테스트에 충분히 안정적인 부분

- `hyper init`으로 프로젝트 로컬 Hyper Run 상태 생성
- `hyper run`으로 runtime packet 생성
- `hyper complete`로 packet 종료
- `hyper run --auto --until <stage>`를 통한 guarded auto continuation
- `hyper advance`를 통한 명시적인 stage advancement
- `hyper status`로 상태와 readiness 확인
- `hyper doctor`로 설치/프로젝트 문제 진단
- `hyper repair`로 단순 state 불일치 복구
- `hyper migrate`로 CLI 업그레이드 이후 프로젝트 상태 갱신
- GitHub release에서 checksum 검증을 거치는 install/update와 선택적 cosign signature 검증
- macOS, Linux, Windows x64에서 native CLI CI 빌드와 테스트

## 아직 실험적인 부분

- Growth pressure 분류는 heuristic입니다. 유용하지만 정식 semantic model은 아닙니다.
- Capability candidate는 반복 evidence에서 생성되지만, activation policy는 아직 보수적입니다.
- Auto mode는 packet 단위 continuation이지, 완전히 방치해도 되는 autonomous background runner는 아닙니다.
- Reference benchmark evidence는 구조화되어 있고 검증되지만, 비교의 품질은 여전히 agent나 개발자가 남긴 evidence 품질에 의존합니다.
- 패키지 경계는 아직 대부분 `internal/app` 안에 있습니다. 현재 CLI 크기에서는 괜찮지만, 이후 분리가 필요합니다.
- 기존 `.hyper/` 프로젝트는 큰 growth/readiness 변경 뒤 `hyper migrate`가 필요할 수 있습니다.

## 보안과 공급망

- release build는 checksums를 제공합니다.
- macOS/Linux installer, Windows PowerShell installer, `hyper update`는 GitHub release download에 대해 checksum을 검증합니다.
- release build는 cosign keyless signature bundle을 제공합니다. `cosign`이 있으면 installer와 `hyper update`가 signature를 검증하고, `HYPER_RUN_VERIFY_SIGNATURE=required`로 필수화할 수 있습니다.
- signature 검증은 로컬 `cosign` executable에 의존합니다. checksum 검증은 기본 baseline으로 유지됩니다.

## 에이전트 동작

- Hyper Run은 runtime packet을 만듭니다. AI agent의 행동을 강제로 보장하지는 않습니다.
- Codex Desktop routing은 얇은 compatibility layer입니다. truth source는 `plan.md`, `.hyper/`, native CLI입니다.
- 좋은 evidence는 여전히 agent나 개발자가 `evidence.md`와 `next.md`를 성실하게 업데이트해야 합니다.
