# Proposal: Sandbox Status Output Format

## Summary

Make `lango sandbox status` scriptable by adding an `--output table|json|plain` flag while preserving the existing human-readable default output.

## Problem

Most production-facing CLI inspection commands in Lango expose machine-readable output. `lango sandbox status` currently prints only a fixed human-readable report, which makes wrappers, CI checks, and operator tooling parse fragile text.

The sandbox status command is a high-value diagnostic because it reports whether local tool execution is actually isolated, whether fail-closed is active, which backend was selected, and why a backend is unavailable.

## Goals

- Add explicit `--output table|json|plain` support to `lango sandbox status`.
- Keep the current default table output unchanged for interactive users.
- Provide structured JSON covering configuration, active isolation, platform capabilities, backend availability, Linux warning state, and optional recent decisions.
- Keep Recent Sandbox Decisions graceful: unavailable audit storage must not fail status rendering.

## Non-Goals

- Do not change sandbox enforcement behavior.
- Do not add JSON output to `lango sandbox test`.
- Do not change P2P container sandbox commands.
