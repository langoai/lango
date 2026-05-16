## Why

The README internal CLI inventory still truncates the memory family at `agent`, even though the actual command surface is `lango memory agent <name>`.

## What Changes

- update the README internal tree to include the `agent <name>` placeholder
- tighten the existing memory inventory guard so the README must keep that placeholder
- sync the main docs-only and test-coverage specs with the clearer contract

## Impact

- more truthful README inventory wording
- clearer operator guidance for the per-agent memory command
- stronger regression protection against placeholder loss
