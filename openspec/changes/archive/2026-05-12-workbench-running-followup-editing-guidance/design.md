## Overview

This is a contract-clarification change layered on top of the existing running-state follow-up behavior.

## Design Decisions

### Running-state loop includes edit intent, not just interrupt intent

Once a follow-up draft exists during streaming, the operator can:

- keep typing
- use editing keys
- press `Enter` to interrupt and run the staged follow-up

The docs and spec now state that explicitly so the product contract matches the actual interaction model.
