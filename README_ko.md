![Hyper Run banner](assets/readme/banner.png)

<p align="right">
  <a href="./README.md"><kbd>English</kbd></a>
  <a href="./README_ko.md"><kbd>한국어</kbd></a>
</p>

# Hyper Run

Hyper Run은 하네스 없이 시작하는 프로젝트 성장 런타임입니다. 사람이 관리하는 `plan.md`에서 출발해 다음 실행 에피소드를 위한 runtime packet을 만들고, 실행 기록을 `.hyper/`에 남기며, 완료되거나 막힌 작업에서 재사용 가능한 컨텍스트를 학습합니다.

프로젝트 명령 모델은 다음과 같습니다.

```bash
hyper init
hyper run [focus]
hyper complete
hyper status
hyper version
hyper update [source]
```

Codex Desktop에서는 같은 개념을 명령 관례처럼 사용할 수 있습니다.

```text
$hyper
$hyper init
$hyper run
```

`$hyper run`은 현재 workspace에서 CLI를 실행하고, 생성된 runtime packet을 읽고, 현재 episode를 구현하고, evidence를 업데이트하고, 다음 추천 runtime episode를 작성한다는 뜻입니다.

`hyper init`은 `AGENTS.md`, `.agents/skills/hyper/SKILL.md`, `.agents/skills/hyper-run/SKILL.md`, `.hyper/codex-desktop.md`, `.hyper/commands/hyper-run.md`에 로컬 Codex Desktop 실행 규칙을 기록합니다. 그래서 `$hyper run` 흐름은 README 설명이 아니라 프로젝트 안의 규칙으로 남습니다.

`hyper` skill은 의도적으로 얇게 둡니다. Codex Desktop이 `$hyper run`을 잡아 native CLI와 `.hyper/` 상태로 다시 연결하는 역할만 합니다. 제품 판단, 학습, 검증 증거, 생성된 하네스는 static skill 안이 아니라 `plan.md`, `.hyper/`, CLI에 남습니다.

제품 경계와 예제 흐름은 다음 문서를 보면 됩니다.

- [서비스 정의](docs/SERVICE_DEFINITION_ko.md)
- [Tiny MVP Flow 예제](examples/tiny-mvp-flow/README_ko.md)

## 요구사항

- release binary로 설치하면 별도 runtime 의존성 없음
- source에서 빌드할 때는 Go 1.21 이상
- Hyper Run이 `plan.md`와 `.hyper/`를 만들 수 있는 프로젝트 디렉터리

## Native Binary 설치

최신 native binary를 `~/.local/bin/hyper`에 설치합니다.

```bash
curl -fsSL https://raw.githubusercontent.com/KoreanCode/orange-hyper-run/main/install.sh | sh
```

수동 설치:

```bash
mkdir -p ~/.local/bin
curl -fsSL https://github.com/KoreanCode/orange-hyper-run/releases/latest/download/hyper-darwin-arm64 -o ~/.local/bin/hyper
chmod +x ~/.local/bin/hyper
```

Intel macOS는 `hyper-darwin-amd64`, Linux x64는 `hyper-linux-amd64`, Linux ARM64는 `hyper-linux-arm64`를 사용합니다. `~/.local/bin`이 `PATH`에 들어 있어야 합니다.

설치된 binary를 확인합니다.

```bash
hyper version
```

그다음 원하는 프로젝트 안에서 Hyper Run을 실행합니다.

```bash
cd my-project
hyper init
# plan.md를 채워주세요
hyper run "가장 작은 사용 가능한 MVP를 만들어줘"
# evidence.md와 next.md를 업데이트한 뒤
hyper complete
```

## Source에서 설치

source 설치를 원하면 다음처럼 설치합니다.

```bash
go install github.com/KoreanCode/orange-hyper-run/cmd/hyper@latest
```

## 업데이트

현재 native executable을 최신 GitHub release로 업데이트합니다.

```bash
hyper update
```

`hyper update`는 먼저 현재 실행 중인 executable을 교체합니다. 해당 경로에 쓸 수 없으면 최신 binary를 `~/.local/bin/hyper`에 설치하고, 그 디렉터리가 `PATH`에 없으면 경고합니다.

fork에서 업데이트하려면:

```bash
hyper update github:OWNER/orange-hyper-run
```

직접 binary URL도 줄 수 있습니다.

```bash
hyper update https://example.com/hyper-darwin-arm64
```

user-local 설치 경로를 강제하려면:

```bash
HYPER_INSTALL_PATH="$HOME/.local/bin/hyper" hyper update
```

Release binary는 `v*` tag를 push하면 GitHub Actions release workflow가 빌드합니다.

## 로컬 개발 설치

이 repository에서:

```bash
go test ./...
go build -o dist/hyper ./cmd/hyper
```

다른 프로젝트 디렉터리에서 local binary를 실행합니다.

```bash
cd ../my-project
../orange-hyper-run/dist/hyper init
# plan.md를 채워주세요
../orange-hyper-run/dist/hyper run "가장 작은 사용 가능한 MVP를 만들어줘"
../orange-hyper-run/dist/hyper complete
```

## 프로젝트 설정

대상 프로젝트에서 `hyper init`을 한 번 실행합니다. 로컬 `.hyper/` 런타임 설정을 설치하고, `plan.md`가 없으면 빈 draft를 만듭니다.

Hyper Run은 사용자가 프로젝트 루트의 `plan.md`를 검토한 뒤 가장 잘 작동합니다.

```markdown
# Product Plan

## Product

무엇을 만들고 있나요?

## Target Users

누구를 위한 제품인가요?

## MVP

가장 작은 완성형 제품은 무엇인가요?

## Current Stage

Tiny MVP

## Build Style

Web app

## Non-goals

지금 만들지 않을 것은 무엇인가요?

## Constraints

기술, 제품, 시간, UX 제약은 무엇인가요?

## Success Criteria

이번 stage가 끝났다는 기준은 무엇인가요?

## Current Focus

다음 run에서 무엇을 진행해야 하나요?
```

`plan.md`가 없으면 `hyper run`은 실행을 멈추고 먼저 프로젝트를 초기화하라고 안내합니다. `plan.md`가 아직 비어 있는데 README나 docs에 기획 문맥이 있으면 Hyper Run은 사용자가 검토할 수 있도록 `.hyper/plan-candidates.md`를 씁니다.

## `hyper init`이 만드는 것

```text
AGENTS.md
.agents/
  skills/
    hyper/
      SKILL.md
    hyper-run/
      SKILL.md
.hyper/
  codex-desktop.md
  commands/
    hyper-run.md
  capabilities/
    candidates/
      harness/
      skill/
      validator/
    active/
      harness/
      skill/
      validator/
    retired/
      harness/
      skill/
      validator/
  growth/
    state.json
  readiness/
    state.json
  hyper.sqlite
  state.json
  logs/
  memories/
    decisions.md
    patterns.md
    failures.md
    constraints.md
  validators/
    generated/
  skills/
    generated/
  harnesses/
    generated/
plan.md
```

## `hyper run`이 만드는 것

```text
.hyper/
  logs/
  goals/
    GOAL-0001/
      goal.md
      tasks.md
      evidence.md
      review.md
      next.md
```

각 run은 Codex Desktop이나 다른 실행 agent에게 넘길 수 있는 runtime packet을 만듭니다. 이것은 오래 유지되는 SPEC이 아니라 `plan.md`, 로그, evidence, memory, 현재 프로젝트 상태에서 생성되는 다음 실행 episode입니다. 이전 active packet의 evidence가 아직 pending이면 새 `hyper run`은 차단됩니다. 기본 handoff는 prompt 기반이며 다음과 같은 payload를 포함합니다.

```text
Read .hyper/goals/GOAL-0001/goal.md as a runtime packet and complete it checkpoint by checkpoint.
```

`evidence.md`에는 axis-slot 형태의 `Readiness Evidence`와 `Active Capability Evidence`가 포함되어 stage-gate 진행 증거와 active validator 통과 여부를 일반 validation output과 분리해 기록할 수 있습니다.

## 학습 루프

runtime packet이 완료되었거나 막혔다면 `evidence.md`와 `next.md`를 업데이트한 뒤 다음을 실행합니다.

```bash
hyper complete
```

`hyper complete`는 active packet을 닫고, durable signal을 학습하고, Growth와 Readiness를 새 evidence 기준으로 갱신하고, `.hyper/state.json`을 업데이트합니다.

Learn은 일반 요약이 아닙니다. 다음 작업에 영향을 줘야 하는 지속성 있는 신호만 추출합니다.

- 앞으로도 유지되어야 하는 결정
- 재사용 가능한 구현 또는 검증 패턴
- 반복하면 안 되는 blocker와 failure
- 이후 runtime packet이 지켜야 하는 constraint

가장 강한 Learn 신호는 다음 섹션에서 나옵니다.

```text
evidence.md
  Validation
  Readiness Evidence
  Decisions
  Reusable Patterns
  Blocker

next.md
  Recommended Next Goal
  Learn Notes
```

명시적인 구조화 신호는 `Learn Notes`에 적습니다.

```text
- Decision: MVP가 사용 가능해질 때까지 auth는 local-first로 유지한다.
- Pattern: 기존 Playwright smoke test로 핵심 사용자 흐름을 검증한다.
- Constraint: credentials 없이는 유료 서비스를 추가하지 않는다.
- Failure: 이전 API 경로는 필수 token이 없어 실패했다.
```

다음 `hyper run`도 새 runtime packet을 만들기 전에 이전 active runtime packet 상태를 확인합니다. 수동 학습은 디버깅용으로 남아 있습니다.

```bash
hyper internal learn
```

Hyper Run은 재사용 가능한 memory를 SQLite와 `.hyper/memories/`에 저장합니다. 이후 `hyper run`은 유사한 이전 컨텍스트를 찾아 새 runtime packet의 `Continue From`에 포함합니다. Learn은 changed files나 긴 notes를 그대로 memory로 보지 않고, durable decision, pattern, failure, constraint만 누적합니다.

## Service Readiness Model

Growth가 업데이트된 뒤 Hyper Run은 `.hyper/readiness/state.json`을 씁니다. 이 state는 tiny MVP 작업을 service-level quality까지 이어주는 다리입니다. 다음 readiness 축을 추적합니다.

- product completeness
- core UX
- data persistence
- error handling
- validation coverage
- security baseline
- deployment readiness
- operations and docs
- maintainability

다음 runtime packet에는 `Stage Gate` 섹션이 들어갑니다. Hyper Run은 현재 stage에서 다음 gate를 정하고, 가장 약한 missing readiness axis를 찾아 다음 episode의 concrete pressure로 바꿉니다. 이 pressure는 `Current Episode`, `Work Boundary`, `Validation Signals`, `Stop When`에 영향을 줍니다.

Readiness 진행도는 axis label이 있는 evidence로 기록합니다. 새 packet은 모든 readiness axis 슬롯을 포함합니다.

```text
## Readiness Evidence

Core UX: 브라우저에서 주요 add/edit flow가 작동한다.
Data persistence: SQLite 또는 localStorage로 사용자 기록이 reload 뒤에도 유지된다.
Validation coverage: 주요 flow가 반복 가능한 smoke test로 검증된다.
```

`hyper complete`와 `hyper status`에서 axis-labeled readiness evidence는 해당 axis를 `covered`로 올리고, stage gate blocking gap에서 제거하며, 같은 pressure가 다시 선택되지 않게 합니다.

Evidence는 axis에 맞게 충분히 구체적이어야 합니다. 예를 들어 `Validation coverage: tested`는 emerging evidence에 머물고, `Validation coverage: \`go test ./...\` passed and is repeatable`처럼 command와 통과 증거가 있어야 covered가 됩니다. UX evidence는 smoke pass, browser verification, screenshot 같은 증거가 필요합니다. Deployment evidence는 build, release, hosted URL, CI, deploy proof가 필요합니다.

Stage gate가 ready가 되어도 Hyper Run은 `plan.md`를 자동 수정하지 않습니다. 다음 runtime packet에 stage advancement candidate를 표시하고, 사용자가 수락하거나 거절할 수 있도록 정확한 `Current Stage` 변경을 권고합니다.

Beta와 Service Quality stage에서는 `.hyper/validators/generated/`와 `.hyper/capabilities/candidates/validator/` 아래 quiet validator candidate도 생성합니다. 이 파일들은 active validator로 승격되기 전까지 required behavior가 아닙니다.

목표는 무거운 프로세스를 추가하는 것이 아닙니다. 사용자는 계속 `hyper run`을 실행하고, 프로젝트가 자기 안에서 service quality의 의미를 점진적으로 학습합니다.

## Project Growth Engine

Learn이 실행된 뒤 Hyper Run은 `.hyper/growth/state.json`을 업데이트합니다. 이것은 사용자에게 보여주기 위한 planning report가 아니라, 다음 runtime packet의 컴파일 방식을 바꾸는 project-local growth state입니다.

Growth pressure는 반복되거나 지속성 있는 Learn signal에서 나옵니다.

- stable decision과 constraint는 `Work Boundary`에 영향을 줍니다.
- repeated validation pattern은 `Validation Signals`에 영향을 줍니다.
- recurring failure는 `Stop When`에 영향을 줍니다.
- repeated pattern은 `.hyper/validators/generated/` 또는 `.hyper/skills/generated/` 아래 quiet candidate를 만들 수 있습니다.
- `.hyper/capabilities/active/validator/` 아래 active validator는 다음 runtime packet의 `Validation Signals`에 required behavior로 들어갑니다.
- harness candidate는 여러 repeated pressure가 harness threshold를 넘은 뒤에만 생성됩니다.

Phase 1 안정화는 exact text match에만 의존하지 않고 유사 Learn signal을 clustering하며, 단순 progress noise를 제외하고, pressure를 `stable_decision`, `repeated_validation`, `implementation_pattern`, `recurring_constraint`, `recurring_failure`로 분류합니다.

Phase 2는 capability lifecycle을 추가합니다.

```text
observed -> repeated -> promotable -> active -> retired
```

Lifecycle metadata는 `.hyper/capabilities/` 아래 기록됩니다. 발견 가능성을 위해 `.hyper/validators/generated/`, `.hyper/skills/generated/`, `.hyper/harnesses/generated/`에도 legacy candidate 파일을 계속 쓰지만, 활성 여부는 lifecycle status가 결정합니다. Candidate 또는 promotable validator는 `.hyper/capabilities/active/validator/` 아래 active validator가 되기 전까지 required behavior가 되지 않습니다.

사용자는 계속 `hyper run`을 실행합니다. 프로젝트는 다음 packet이 더 구체적으로 컴파일되는 방식으로 성장합니다.

## 유용한 명령

```bash
hyper init
hyper run "고객 저장 기능 추가"
hyper complete
hyper status
hyper resume
hyper update
hyper version
hyper internal learn
```

## 현재 상태

현재는 native Go MVP 구현입니다. TUI 없이 작은 표면적을 유지합니다: 한 번의 초기화, 반복되는 run, 로컬 파일, SQLite 로그, prompt 기반 실행 handoff, 가벼운 학습.

## 라이선스

MIT License입니다. 자세한 내용은 [LICENSE](LICENSE)를 참고하세요.
