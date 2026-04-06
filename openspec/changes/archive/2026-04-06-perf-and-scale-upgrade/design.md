## Context

Phase 1-5 완료 후 cockpit TUI의 성능 특성을 분석한 결과, render() O(n) 루프와 per-entry style 할당이 1000+ entries에서 5-15ms/render 지연을 발생시키고, transcript/task/grant의 무한 성장이 장기 세션에서 메모리 문제를 유발함을 확인.

## Goals / Non-Goals

**Goals:**
- Entry block memoization으로 per-render style 할당 비용 제거
- Module-level style pre-allocation으로 NewStyle() 호출 제거
- Transcript entry cap (2000)으로 메모리 및 render 비용 bounded
- Terminal task cap (500)으로 BackgroundManager 메모리 bounded
- Grant lazy cleanup으로 GrantStore 메모리 bounded
- Approval diff styled line cache로 스크롤 시 re-style 제거

**Non-Goals:**
- Viewport virtualization (Bubble Tea viewport 모델 자체 변경 필요 — scope 밖)
- EventBus Unsubscribe 추가 (실질적으로 문제 아님)
- MetricsCollector/ChannelTracker 정리 (이미 bounded)

## Decisions

### D1: cachedBlock field로 entry-level memoization
- transcriptItem에 cachedBlock string 추가
- render()에서 cachedBlock != ""이면 캐시 사용
- width 변경, tool finalize, thinking finalize 시 무효화
- **Why**: strings.Join + viewport.SetContent 비용은 잔존하지만, 가장 비싼 부분(style 생성 + markdown render)을 제거

### D2: Module-level style pre-allocation (sidebar.go 패턴)
- lipgloss.Style은 immutable value type — base style 공유 시 복사만 발생 (alloc 없음)
- 색상 등 가변 속성은 per-call chain 유지
- **Why**: 검증된 패턴 (sidebar.go가 이미 사용 중)

### D3: Terminal task FIFO cap (삭제가 아닌 밀어내기)
- maxTerminalTasks = 500, CompletedAt 기준 oldest eviction
- Status/Result/retry는 cap 내에서 정상 동작
- **Why**: TTL 삭제는 Status/Result/retry를 깨뜨림 — cap은 동일한 메모리 bound를 제공하면서 UX 영향 최소화

### D4: GrantStore List()에서 lazy cleanup
- cleanExpiredLocked() 내부 헬퍼 추출 (lock 불요)
- List() RLock → Lock 승격 후 cleanup + listing
- **Why**: RLock과 Lock 교차 호출은 deadlock — Lock 승격이 필요

### D5: Diff line cache — scrollOffset을 key에서 제외
- cache key: content + width + splitMode
- scroll은 cached lines slice windowing만
- **Why**: 가장 비싼 상호작용(스크롤)에서 100% cache hit

### D6: Transcript trimming — in-flight entries 보존 + 새 backing array
- trim 범위에서 active tool/thinking entries를 수집하여 tombstone 뒤에 보존
- make()로 새 slice 생성하여 old backing array를 GC 대상으로
- tombstone count 누적 (tombstone 유무에 따라 boundary 조정)
- **Why**: finalizeToolResult/finalizeThinking이 항상 active entry를 찾을 수 있도록 보장

## Risks / Trade-offs

- [Risk] Terminal task cap 500 초과 시 oldest task not found → Mitigation: 실질적으로 500 완료 task 이후에만 발생, docs에 반영
- [Risk] GrantStore List() Lock 승격으로 read concurrency 감소 → Mitigation: List()는 UI tick (2초)에서만 호출, 병목 아님
- [Risk] cachedBlock이 메모리 증가 유발 → Mitigation: entry trimming이 2000 cap으로 제한, 새 backing array로 해제
