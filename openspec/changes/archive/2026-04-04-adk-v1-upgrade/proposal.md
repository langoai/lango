## Why

google.golang.org/adk v1.0.0이 릴리스되었다. lango는 v0.6.0을 사용 중이며, GA 버전으로 업그레이드하여 안정성과 향후 호환성을 확보해야 한다. v1.0.0은 Go API 수준에서 완전 하위 호환이므로, 코드 변경 없이 의존성만 갱신하면 된다.

## What Changes

- `go.mod`에서 `google.golang.org/adk` 버전을 `v0.6.0` → `v1.0.0`으로 범프
- `go.sum` 전이 의존성 자동 갱신 (`a2a-go v0.3.10`, `go-sdk v1.4.1`, `grpc v1.79.3` 등)
- `internal/adk/mcp_spike_test.go`에서 `mcptoolset.ConfirmationProvider` → `tool.ConfirmationProvider` 타입 참조 변경 (v1.0.0에서 타입이 `tool` 패키지로 이동)
- 프로덕션 코드 변경 없음 — 모든 ADK 인터페이스(`session.Service`, `model.LLM`, `agent.Agent`, `tool.Tool`, `runner.Runner`)가 소스 레벨에서 동일하거나 additive 변경만 존재

## Capabilities

### New Capabilities

(없음 — 이 변경은 의존성 업그레이드이며, 새로운 기능을 도입하지 않는다)

### Modified Capabilities

- `adk-architecture`: ADK 의존성 버전이 v0.6.0에서 v1.0.0으로 변경됨. 인터페이스 계약은 동일하나, 사용 가능한 ADK 기능 surface가 확장됨 (RunOption, AutoCreateSession, HITL tool confirmation, workflow agents 등)

## Impact

- **의존성**: `google.golang.org/adk v0.6.0` → `v1.0.0`, 전이 의존성 약 5개 마이너 범프
- **코드**: spike test 1파일만 수정 (`internal/adk/mcp_spike_test.go`), 프로덕션 코드 0줄 변경
- **테스트**: 전체 테스트 스위트 통과 확인 완료 (`go test ./...` all pass)
- **빌드**: `go build ./...` + `go vet ./...` 통과 확인 완료
- **리스크**: LOW — 소스 호환이 모듈 캐시 diff로 검증됨
