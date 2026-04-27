## Design Summary

This workstream keeps the existing dead-letter operator surfaces:

- `lango status dead-letter-summary`
- cockpit `dead-letters` page summary strip

and extends them additively with grouped actor-family summaries.

Rules:

- aggregate actor families from each backlog row's current `latest_manual_replay_actor`
- expose grouped CLI buckets as `by_actor_family`
- render the CLI grouped buckets in a `By actor family` table section
- render the cockpit grouped buckets in a compact `actor families:` strip line
- use the initial built-in taxonomy:
  - `operator`
  - `system`
  - `service`
  - `unknown`
- keep matching case-insensitive with `unknown` fallback
- preserve raw top latest manual replay actors alongside the grouped view

The slice does not add grouped dispatch families, configurable actor taxonomy, or trend/time-window summary backends. Those remain follow-on work.
