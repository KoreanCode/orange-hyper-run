![Hyper Run banner](assets/readme/banner.png)

<p align="right">
  <a href="./README.md"><kbd>English</kbd></a>
  <a href="./README_ko.md"><kbd>한국어</kbd></a>
</p>

# Hyper Run

Hyper Run은 evidence-first 프로젝트 성장 프로토콜입니다. 실행 로그가 pressure를 만들고, pressure가 candidate를 만들고, 반복된 proof가 프로젝트 전용 structure를 승격시킵니다.

프로젝트 루트에 `plan.md`를 적어두면, Hyper Run이 다음에 할 작은 작업 단위를 만들고, 진행 기록을 `.hyper/`에 저장하고, 완료 evidence를 바탕으로 다음 작업을 더 구체적으로 만듭니다.

Codex Desktop, CLI 에이전트, Cursor 스타일 에이전트, 다른 코딩 assistant가 같은 runtime packet을 읽을 수 있는 agent-agnostic 구조입니다. 기본 흐름은 여전히 한 명령입니다.

```bash
hyper run
```

## 왜 필요한가요?

AI 코딩을 오래 이어가다 보면 이런 문제가 생깁니다.

- 다음 작업이 너무 넓어짐
- 이전 결정이 잊힘
- 검증 evidence가 흩어짐
- 작은 MVP 작업이 서비스 품질까지 자연스럽게 이어지지 않음

Hyper Run은 이 문맥을 프로젝트 안에 남깁니다. 단순 작업 쪼개기 도구나 복잡한 프로젝트 관리 도구가 아니라, 다음 작업 packet을 만들고 결과에서 배우며, 반복 evidence가 쌓일 때만 더 강한 구조를 만들게 하는 런타임입니다.

## 핵심 개념

Hyper Run 안에는 몇 가지 내부 개념이 있지만, 어렵게 볼 필요는 없습니다.

| 개념 | 쉽게 말하면 |
| --- | --- |
| `plan.md` | 사람이 적는 제품 설명서입니다. 무엇을 만들고, 누구를 위한 것이고, 현재 어느 단계인지 적습니다. |
| Runtime packet | 다음에 할 작은 작업 묶음입니다. 보통 `.hyper/goals/GOAL-0001/goal.md`로 만들어집니다. |
| Evidence | 작업을 했고 검증했다는 증거입니다. `evidence.md`에 남깁니다. |
| Learn | 프로젝트가 반복해서 필요로 했거나, 실패했거나, 증명한 것을 뽑아내는 단계입니다. 단순 요약이 아닙니다. |
| Pressure Ledger | 프로젝트가 아직 풀지 못했거나 반복해서 겪은 pressure를 기록하는 장부입니다. 예를 들어 매번 같은 검증이 필요하면 validator 후보를 만들 수 있습니다. |
| Readiness | Tiny MVP에서 Usable MVP, Beta, Service Quality로 넘어갈 준비가 됐는지 보는 단계별 성장 계약입니다. |
| Capability candidate | validator, skill, harness 후보입니다. 반복 evidence가 충분하기 전까지는 바로 강제되지 않습니다. |

핵심은 **하네스 없이 성장하는 구조**입니다. 처음부터 하네스를 만들지 않고, `plan.md`에서 시작해 작은 packet을 실행하고, evidence를 쌓고, 정말 필요하다는 반복 신호가 생겼을 때만 더 강한 구조를 만듭니다.

```text
Execution -> Evidence -> Pressure Ledger -> Capability candidate -> Structure when proven
```

이 점이 하네스 우선 방식과 가장 다릅니다. 보통 하네스는 미리 정해진 workflow에서 시작합니다. Hyper Run은 실행 evidence에서 시작하고, 반복 pressure가 쌓였을 때만 validator, skill, agent, harness 같은 구조를 후보로 만들고 활성화합니다.

## 원칙

Hyper Run은 네 가지 제품 원칙을 따릅니다.

- No structure before pressure.
- No stage advancement without evidence.
- No harness before repeated need.
- No memory without reusable signal.

중요한 이유는 간단합니다. Hyper Run은 절차를 만들기 위해 절차를 만들면 안 됩니다. 프로젝트가 계속 필요하다고 증명할 때만 구조가 생깁니다.

## Pressure Ledger

Pressure Ledger는 `.hyper/growth/state.json`에 저장됩니다. 반복되는 validation 필요, recurring failure, 재사용 가능한 구현 패턴, constraint, readiness gap을 기록합니다.

ledger는 바로 새 행동을 강제하지 않습니다. lifecycle을 거칩니다.

```text
observed -> repeated -> promotable -> active -> retired
```

threshold 전에는 생성된 validator, skill, agent, harness가 candidate로만 남습니다. 반복 proof가 충분하면 active project-specific structure가 됩니다.

## Stage 계약

Stage는 단순한 라벨이 아닙니다. 각 stage는 `goal.md`가 Codex나 다른 코딩 에이전트에게 무엇을 증명하라고 요구하는지를 바꿉니다.

| Stage | 계약 |
| --- | --- |
| Tiny MVP | Existence proof: 가장 작고 되돌릴 수 있는 제품 조각으로 유용한 flow 하나가 존재함을 증명합니다. |
| Usable MVP | Usability proof: 실제 사용자가 primary flow를 처음부터 끝까지 사용할 수 있게 만듭니다. |
| Beta | Repeatability proof: 현실적인 데이터, 실패 처리, 검증, 문서, release readiness 주변의 신뢰성을 증명합니다. |
| Service Quality | Operability proof: 보안, 배포, 운영, rollback, 반복 가능한 검증을 제품 동작의 일부로 다룹니다. |

`hyper run`은 프로젝트에 아직 풀리지 않은 growth pressure가 있을 때 계속 다음 작업을 만듭니다. pressure가 반복되면 후보를 만들고, 같은 필요가 계속 증명되면 그 후보가 active structure가 됩니다.

## 기본 흐름

```bash
hyper init
# plan.md를 한 번 채웁니다

hyper run "가장 작은 사용 가능한 MVP를 만들어줘"
# 생성된 packet을 구현합니다
# evidence.md와 next.md를 업데이트합니다

hyper complete
hyper status
hyper doctor
hyper run "다음 개선 작업"
```

Codex Desktop에서는 프로젝트 명령처럼 사용할 수 있습니다.

```text
$hyper init
$hyper run
```

`$hyper run`은 Codex가 native `hyper` CLI를 실행하고, 생성된 `.hyper/goals/.../goal.md`를 읽고, 구현한 뒤, evidence와 다음 추천 작업까지 남긴다는 뜻입니다.

## 설치

### macOS / Linux

최신 native binary를 설치합니다.

```bash
curl -fsSL https://raw.githubusercontent.com/KoreanCode/orange-hyper-run/main/install.sh | sh
```

GitHub release에서 설치할 때 installer는 `checksums.txt`를 함께 내려받고, binary를 옮기기 전에 SHA256 checksum을 검증합니다.

설치 확인:

```bash
hyper version
```

macOS 수동 설치:

Apple Silicon:

```bash
mkdir -p ~/.local/bin
curl -fsSL https://github.com/KoreanCode/orange-hyper-run/releases/latest/download/hyper-darwin-arm64 -o ~/.local/bin/hyper
chmod +x ~/.local/bin/hyper
hyper version
```

Intel Mac:

```bash
mkdir -p ~/.local/bin
curl -fsSL https://github.com/KoreanCode/orange-hyper-run/releases/latest/download/hyper-darwin-amd64 -o ~/.local/bin/hyper
chmod +x ~/.local/bin/hyper
hyper version
```

### Windows

PowerShell에서 Windows x64 binary를 내려받습니다.

```powershell
New-Item -ItemType Directory -Force "$env:USERPROFILE\.local\bin" | Out-Null
Invoke-WebRequest -Uri "https://github.com/KoreanCode/orange-hyper-run/releases/latest/download/hyper-windows-amd64.exe" -OutFile "$env:USERPROFILE\.local\bin\hyper.exe"
& "$env:USERPROFILE\.local\bin\hyper.exe" version
```

사용자 `PATH`에 추가합니다.

```powershell
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\.local\bin", "User")
```

새 터미널을 열고 확인합니다.

```powershell
hyper version
```

Windows binary는 CI에서 빌드/테스트하지만, 아직 PowerShell installer script는 없습니다.

다른 release binary:

- `hyper-darwin-amd64`: Intel macOS
- `hyper-linux-amd64`: Linux x64
- `hyper-linux-arm64`: Linux ARM64
- `hyper-windows-amd64.exe`: Windows x64

`~/.local/bin`이 `PATH`에 들어 있어야 합니다.

## Source에서 설치

```bash
go install github.com/KoreanCode/orange-hyper-run/cmd/hyper@latest
```

## 업데이트

```bash
hyper update
```

최신 GitHub release를 내려받습니다. 현재 실행 파일을 교체할 수 없으면 `~/.local/bin/hyper`에 설치합니다.

fork에서 업데이트하려면:

```bash
hyper update github:OWNER/orange-hyper-run
```

## 프로젝트 설정

대상 프로젝트 안에서 한 번 실행합니다.

```bash
hyper init
```

생성되는 것:

- `plan.md`
- `.hyper/`
- Codex Desktop 라우팅 파일인 `AGENTS.md`, `.agents/skills/...`

그다음 `plan.md`를 평범한 문장으로 채웁니다.

```markdown
# Product Plan

## Product

무엇을 만들고 있나요?

## Target Users

누구를 위한 제품인가요?

## MVP

가장 작은 유용한 버전은 무엇인가요?

## Current Stage

Tiny MVP

## Build Style

Web app

## Non-goals

아직 만들지 않을 것은 무엇인가요?

## Constraints

기술 또는 제품 제약은 무엇인가요?

## Success Criteria

이번 단계가 끝났다는 기준은 무엇인가요?

## Current Focus

다음 run에서 무엇을 개선해야 하나요?
```

`plan.md`가 너무 비어 있으면 Hyper Run이 README나 docs를 읽고 `.hyper/plan-candidates.md`를 만들 수 있습니다. 거기서 쓸 만한 제품 문맥을 `plan.md`로 옮기면 됩니다.

## `hyper run`이 하는 일

`hyper run`은 새 runtime packet을 만듭니다.

```text
.hyper/goals/GOAL-0001/
  goal.md
  tasks.md
  evidence.md
  review.md
  next.md
```

중요한 파일은 다음입니다.

- `goal.md`: 지금 만들 작업
- `tasks.md`: 이번 run의 체크포인트
- `evidence.md`: 무엇을 바꿨고 어떻게 검증했는지
- `next.md`: 다음에 무엇을 해야 하는지

이전 packet의 evidence가 아직 pending이면 새 `hyper run`은 막힙니다. 먼저 `hyper complete`로 현재 packet을 닫아야 합니다.

## `hyper complete`가 하는 일

구현이 끝나면 `evidence.md`와 `next.md`를 업데이트한 뒤 실행합니다.

```bash
hyper complete
```

이 명령은 현재 packet을 닫고 프로젝트 memory를 업데이트합니다.

- 유지해야 할 결정
- 재사용할 패턴
- 실패나 blocker
- 지켜야 할 제약
- readiness 진행 상태

다음 `hyper run`은 이 정보를 사용합니다.
구체적으로 다음 work boundary, validation signal, stop condition, readiness pressure, capability candidate에 영향을 줍니다.

## Readiness를 쉽게 말하면

Hyper Run은 프로젝트가 단계별로 커지도록 돕습니다.

```text
Tiny MVP -> Usable MVP -> Beta -> Service Quality
```

다음 항목에 대한 evidence가 있는지 봅니다.

- 제품 정의
- 핵심 UX
- 데이터 저장
- 에러 처리
- 검증
- 보안
- 배포
- 문서
- 유지보수성

이 내용은 `evidence.md`에 이렇게 적습니다.

```text
## Readiness Evidence

Core UX: 브라우저 smoke test에서 생성과 완료 flow가 통과했다.
Validation coverage: `go test ./...`가 통과했고 반복 실행 가능하다.
Data persistence: SQLite로 저장한 records가 reload 뒤에도 유지된다.
```

evidence가 충분하면 `hyper status`에서 다음 stage로 올릴 준비가 됐는지 보여줍니다. Hyper Run은 stage 변경을 추천하지만, `plan.md`를 자동으로 수정하지는 않습니다.

## 명령어

```bash
hyper init                  # 프로젝트에 Hyper Run 파일 설치
hyper run [focus]           # 다음 runtime packet 생성
hyper complete              # 현재 packet을 닫고 학습
hyper status                # 현재 stage, gap, readiness 확인
hyper doctor                # 설치, PATH, 프로젝트 상태, Codex 라우팅 진단
hyper repair                # packet evidence와 state.json이 어긋날 때 상태 복구
hyper migrate               # Hyper Run 업그레이드 뒤 growth/readiness 상태 갱신
hyper resume                # 현재 handoff 다시 출력
hyper update                # native binary 업데이트
hyper version               # 버전과 binary 경로 확인
hyper internal learn        # 디버그/수동 학습 명령
```

## 로컬 개발

이 repository에서:

```bash
go test ./...
go vet ./...
go build -o dist/hyper ./cmd/hyper
```

다른 프로젝트에서 테스트합니다.

```bash
cd ../my-project
../orange-hyper-run/dist/hyper init
../orange-hyper-run/dist/hyper run "가장 작은 사용 가능한 MVP를 만들어줘"
../orange-hyper-run/dist/hyper complete
```

## 더 자세한 문서

- [서비스 정의](docs/SERVICE_DEFINITION_ko.md)
- [Tiny MVP Flow 예제](examples/tiny-mvp-flow/README_ko.md)
- [Before / After 데모](examples/before-after-demo/README_ko.md)
- [로드맵](docs/ROADMAP_ko.md)
- [변경 기록](docs/CHANGELOG_ko.md)
- [알려진 한계](docs/KNOWN_LIMITATIONS_ko.md)

## 라이선스

MIT License입니다. 자세한 내용은 [LICENSE](LICENSE)를 참고하세요.
