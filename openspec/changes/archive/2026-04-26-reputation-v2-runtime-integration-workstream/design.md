## Design Summary

This workstream is documentation-only. It does not change Go behavior.

Key design points:

- describe Reputation V2 as a landed canonical contract rather than a planned redesign
- document the separated runtime signals: composite trust, earned trust, durable negative units, and temporary safety signals
- document canonical trust-entry states: `bootstrap`, `established`, `review`, and `temporarily_unsafe`
- document how runtime consumers use that contract:
  - firewall admission blocks `review` and `temporarily_unsafe`
  - handshake auto-approval only applies to returning peers in `established`
  - economy pricing and risk use bootstrap-effective trust for first-time peers and earned trust for returning peers
  - post-pay routing opens only for `established` returning peers
  - team reputation bridges record operational incidents without treating them as automatic durable damage
- narrow the track-level remaining work to owner-root-aware policy adoption, broader dispute-to-reputation feeds, and richer operator trust surfaces
