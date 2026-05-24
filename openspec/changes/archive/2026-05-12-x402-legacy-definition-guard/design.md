## Overview

This is a small test-hardening change for x402 source hygiene.

## Design Decisions

### Distinguish references from definitions

The repository-wide scan catches call sites, while the package-local scan now catches the legacy definition itself. Keeping both makes the “removed” contract explicit from two angles without introducing any external tooling.
