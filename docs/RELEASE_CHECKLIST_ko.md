# 릴리즈 체크리스트

새 Hyper Run 릴리즈를 배포할 때 이 체크리스트를 사용합니다.

AI가 릴리즈 준비를 보조하는 경우에도 maintainer가 명시적으로 승인하기 전에는 active `hyper` 실행 파일 교체, tag 생성, tag push, GitHub release 발행을 진행하지 않습니다.

## 1. 릴리즈 브랜치 준비

- 브랜치가 최신 `main` 기준인지 확인합니다.
- `docs/CHANGELOG.md`와 `docs/CHANGELOG_ko.md`에 새 버전 섹션이 있는지 확인합니다.
- README의 설치/업데이트 안내가 release asset과 맞는지 확인합니다.
- runtime packet 또는 generator template 변경이 있다면, 사용자가 `hyper update`로 받게 될 생성 산출물(`goal.md`, `tasks.md`, `evidence.md`)이 changelog에 명시되어 있는지 확인합니다.
- 로컬 테스트 프로젝트나 임시 binary가 stage되지 않았는지 확인합니다.

## 2. 로컬 검증 실행

```bash
go test -count=1 ./...
go vet ./...
staticcheck ./...
govulncheck ./...
git diff --check
```

sandbox 환경에서 Go cache에 쓸 수 없다면 writable cache 경로를 지정합니다.

```bash
GOCACHE=/private/tmp/hyper-go-cache go test -count=1 ./...
```

runtime packet 또는 generator template 변경이 있다면, tag를 만들기 전에 local binary를 빌드하고 disposable project에서 packet generation을 검증합니다.

```bash
GOCACHE=/private/tmp/hyper-go-cache go build -o /private/tmp/hyper-local ./cmd/hyper
/private/tmp/hyper-local version
```

disposable project 검증에서는 새로 생성된 packet이 다음 파일에 새 user-visible section을 포함하는지 확인해야 합니다.

- `.hyper/goals/<GOAL-ID>/goal.md`
- `.hyper/goals/<GOAL-ID>/tasks.md`
- `.hyper/goals/<GOAL-ID>/evidence.md`

자율 runtime template 릴리즈라면 배포 전에 다음 생성 section을 확인합니다.

- `Decision Hierarchy`
- `Autonomous Work Plan`
- `Autonomous Safety Policy`
- `Capability Expansion Policy`
- `Research Evidence Policy`
- `Loop Progress Policy`
- `Product Satisfaction Policy`

## 3. PR 병합

- 브랜치를 push합니다.
- PR을 엽니다.
- Linux, macOS, Windows CI를 기다립니다.
- CI가 통과한 뒤 merge합니다.

## 4. 릴리즈 태그 생성

```bash
git switch main
git pull --ff-only origin main
git tag -a vX.Y.Z -m "vX.Y.Z"
git push origin vX.Y.Z
```

## 5. GitHub Actions 확인

두 workflow가 모두 통과해야 합니다.

- CI
- Release

Release workflow는 다음 asset을 올려야 합니다.

- `hyper-darwin-arm64`
- `hyper-darwin-amd64`
- `hyper-linux-arm64`
- `hyper-linux-amd64`
- `hyper-windows-amd64.exe`
- `checksums.txt`
- 각 binary와 `checksums.txt`에 대한 `.sigstore.json` bundle

## 6. 설치와 업데이트 smoke test

macOS 또는 Linux:

```bash
hyper update
hyper version
hyper doctor
```

기존 프로젝트에서는 binary 업데이트 뒤 project state도 갱신합니다.

```bash
hyper migrate
hyper doctor
hyper status --short
```

Windows PowerShell:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/KoreanCode/orange-hyper-run/main/install.ps1 | iex"
hyper version
hyper doctor
```

## 7. Checksum과 선택적 signature 검증

기본 사용자 경로에서는 checksum이 자동 검증되어야 합니다.

`cosign`이 설치되어 있으면 install/update가 sigstore bundle도 검증해야 합니다. signature 검증을 필수로 하려면:

```bash
HYPER_RUN_VERIFY_SIGNATURE=required hyper update
```

PowerShell:

```powershell
$env:HYPER_RUN_VERIFY_SIGNATURE="required"
hyper update
```

## 8. 릴리즈 노트 확인

릴리즈를 알리기 전에 release page에서 다음을 확인합니다.

- 예상한 tag
- 모든 asset
- changelog와 맞는 generated note 또는 manual note
- 이전 asset을 실수로 덮어쓰지 않았는지

## 9. 문제가 있으면

- tag는 push됐지만 release가 실패했다면 Release workflow를 먼저 확인합니다.
- asset이 빠졌다면 실패 원인을 파악한 뒤 Release workflow rerun 또는 필요한 조치를 합니다.
- 사용자가 아직 받지 않은 명확히 잘못된 tag라면 GitHub release와 tag를 삭제한 뒤 수정된 tag를 다시 냅니다.
- 사용자가 이미 받았을 가능성이 있으면 patch release를 새로 냅니다.
