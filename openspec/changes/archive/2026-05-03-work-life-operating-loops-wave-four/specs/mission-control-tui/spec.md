## ADDED Requirements

### Requirement: Mission Control can project operator loops from real existing sources
Mission Control SHALL support an operator loop surface in addition to durable missions, proposals, and live decisions. In the first Wave 4 slice, loop rows SHALL be projected only from real existing sources rather than invented integrations or placeholder data.

#### Scenario: Loop rows are derived only from first-slice real sources
- **WHEN** Mission Control projects operator loops for the current session
- **THEN** it SHALL derive those loops only from real existing sources such as durable missions, pending inquiries, dead-letter or retry backlog, cron-job schedule state, and deterministic follow-up signals from current mission or proposal state
- **AND** the first slice SHALL NOT require fabricated calendar, inbox, or external task-system loop sources

#### Scenario: Loop surface does not replace durable missions
- **WHEN** loop rows are available for the current session
- **THEN** Mission Control SHALL keep durable missions visible as the primary owned work surface
- **AND** loop rows SHALL remain an additive coordination surface rather than a replacement for durable mission records

### Requirement: Mission Control agenda ordering is deterministic
Mission Control SHALL derive an agenda ordering for unresolved loops using deterministic category ordering rather than broad heuristic prioritization.

#### Scenario: Agenda order follows fixed Wave 4 priority
- **WHEN** multiple unresolved loop rows are visible in the same Mission Control session
- **THEN** the agenda SHALL order them by fixed category priority: `waiting_user`, `blocked`, `active`, `scheduled`, `needs_review`, `resolved`
- **AND** rows within the same category SHALL order newer updates first

### Requirement: Scheduled automation loops are cron-job only in the first slice
Wave 4 SHALL keep the first scheduled automation source narrow. Mission Control SHALL surface scheduled automation loops only from cron-job sources until another source has a dedicated adapter.

#### Scenario: Cron job can appear as scheduled automation loop
- **WHEN** a cron job source exists with active, failed, or attention-needing state relevant to the current session
- **THEN** Mission Control MAY project one scheduled automation loop from that cron source

#### Scenario: Workflow runs remain deferred without a dedicated adapter
- **WHEN** workflow-run state exists but no dedicated loop adapter has been introduced for Wave 4
- **THEN** Mission Control SHALL NOT imply workflow-run loops are part of the first scheduled automation slice

### Requirement: Follow-up loops use explicit deterministic predicates
Wave 4 follow-up loops SHALL be generated only from explicit deterministic predicates over existing source facts.

#### Scenario: Follow-up loop is generated from deterministic source fact
- **WHEN** one of the approved first-slice predicates holds, such as an accepted proposal with no active linked execution yet, a recently completed mission still needing review, a failed recurring cron automation, or an unresolved inquiry older than a threshold
- **THEN** Mission Control SHALL project a follow-up loop for that fact
- **AND** the projected loop SHALL be traceable to the underlying source state

#### Scenario: Speculative follow-up loops are not generated
- **WHEN** no approved deterministic predicate matches the current source state
- **THEN** Mission Control SHALL NOT invent a narrative or heuristic follow-up loop

### Requirement: Unsupported external work-life integrations remain explicit non-goals
The first Wave 4 slice SHALL remain honest about unavailable sources. Mission Control SHALL NOT imply support for external work-life operating loops when no real adapter exists.

#### Scenario: Calendar, inbox, and external task integrations stay disabled
- **WHEN** Mission Control renders the Wave 4 loop surface
- **THEN** it SHALL NOT claim calendar events, inbox threads, or third-party external task systems are first-slice loop sources unless a real adapter exists in the application
- **AND** those unsupported domains SHALL remain explicit non-goals for this change
