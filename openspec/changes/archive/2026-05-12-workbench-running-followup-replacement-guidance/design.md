## Overview

This is a contract-clarification change for the in-flight follow-up loop.

## Design Decisions

### Running follow-up loops support both free typing and starter replacement

Once a follow-up draft exists during streaming, the operator can:

- keep typing
- use editing keys
- press `1/2/3` to replace it with a starter prompt
- press `Enter` to interrupt-and-run it

The docs and spec now describe that full interaction set.
