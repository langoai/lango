## Why

The README internal CLI inventory currently duplicates the `a2a/` row. That makes the public package tree noisier than the real repository structure and weakens trust in the inventory as a source of truth.

## What Changes

- remove the duplicate `a2a/` row from the README internal tree
- strengthen the existing A2A/agent inventory guard so the README must contain exactly one `a2a/` row
- sync the main docs-only and test-coverage specs with that uniqueness contract

## Impact

- cleaner public inventory documentation
- executable protection against accidental duplicate package rows
- better README truthfulness for the shipped CLI tree
