## ADDED Requirements

### Requirement: Tracing exporter writer seam regressions stay executable
Tracing exporter writer routing regressions SHALL be enforced by executable
tests instead of relying only on manual review.

#### Scenario: Tracing stdout writer seam is covered
- **WHEN** the stdout trace exporter is initialized
- **THEN** an executable test SHALL verify that flushed spans use the injected
  writer seam
