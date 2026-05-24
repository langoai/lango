## Overview

The health command already encapsulates its HTTP work in a small block, so a single HTTP-client seam is enough to make timeout behavior deterministic in tests.

## Decision

- Add `newHealthHTTPClientFn` for timeout control under test
- Keep success output behavior unchanged
- Assert that failure paths do not emit the success payload

## Consequences

- Top-level health command behavior is covered on both success and failure branches
