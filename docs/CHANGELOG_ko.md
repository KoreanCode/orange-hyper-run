# 변경 기록

## Unreleased

- `hyper update`가 GitHub release checksum을 검증하도록 개선했습니다.
- SHA256 checksum 검증을 포함한 Windows PowerShell installer를 추가했습니다.

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
