## Overview

Use the Cobra command writer as the only output sink for `lango agent trace metrics`.

## Decisions

- preserve existing table and JSON payload shapes
- seed a real `turntrace.EntStore` in tests rather than introducing a new fake store seam
- use the existing `storage.NewFacade(..., WithEntClient(...))` path so runtime wiring stays representative

## Risks

- none beyond test setup complexity, mitigated by isolated temp stores per test
