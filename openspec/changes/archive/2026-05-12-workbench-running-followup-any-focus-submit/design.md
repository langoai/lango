## Overview

This is a small correctness fix for the running follow-up loop.

## Design Decisions

### Treat staged follow-up drafts like armed starter prompts for Enter routing

If a running-state follow-up draft exists, `Enter` from the other empty-workbench lanes now routes into the composer submit path the same way seeded starter prompts already do. That closes the last obvious non-composer submission gap in the workbench quick-start loop.
