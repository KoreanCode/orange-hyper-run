# 아키텍처

Hyper Run은 이제 실행 진입점과 제품 런타임 로직을 분리합니다.

## 패키지 구조

| 경로 | 책임 |
| --- | --- |
| `cmd/hyper` | 얇은 native executable entrypoint입니다. process argument를 app package로 넘기고 반환된 exit code로 종료합니다. |
| `internal/app` | 현재 Hyper Run application runtime입니다. command routing, project state, runtime packet, finish gate, Learn, Growth, Readiness, SQLite storage, repair, update, next-packet planning, doctor check를 담당합니다. |
| `internal/buildinfo` | release workflow가 주입하고 `hyper version`이 표시하는 build metadata를 담당합니다. |

## 방향

이 변경은 첫 번째 package boundary입니다. 동작을 바꾸지 않고 `cmd/hyper`에서 제품 로직을 제거합니다.

`v0.6.1` 기준으로 대부분의 domain behavior는 아직 `internal/app` 안에 있습니다. 다음 안정적인 package boundary는 반복적인 유지보수 pressure가 증명될 때 나누는 방향입니다.

앞으로의 세부 package split은 실제 pressure가 증명될 때 진행합니다.

- SQLite schema와 query 변경이 어려워지면 `internal/storage`.
- runtime packet 생성과 adapter가 독립 테스트를 필요로 하면 `internal/runtime`.
- Learn, Growth, Readiness 정책이 따로 진화해야 하면 `internal/learn`, `internal/growth`, `internal/readiness`.
- plan, layout, repair, migration의 소유권이 더 필요해지면 `internal/project`.

## 규칙

`cmd/hyper`는 작게 유지합니다. 새로운 제품 동작은 먼저 `internal/` 아래에 두고, 반복적인 유지보수 pressure가 증명될 때 더 좁은 package로 분리합니다.
