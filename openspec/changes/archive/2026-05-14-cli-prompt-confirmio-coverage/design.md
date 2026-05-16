## Overview

This is a pure prompt-package regression expansion. The helper already accepts injected streams, so the missing work is only test coverage.

## Decisions

### Cover all primary response branches

The new tests explicitly assert:
- `yes` returns approval
- `no` returns denial
- empty reader returns an error

## Non-Goals

- No prompt API changes
- No runtime behavior changes
