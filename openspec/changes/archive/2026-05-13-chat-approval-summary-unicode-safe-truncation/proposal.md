## Why

Approval transcript event summaries are currently truncated with byte slicing. That can split multibyte characters and produce invalid or mangled output for Korean, emoji, or other non-ASCII summaries.

## What Changes

- Switch approval transcript summary truncation to Unicode-safe display truncation.
- Add a regression covering non-ASCII summary text.
- Record the truncation safety contract in OpenSpec.

## Impact

- Prevents malformed approval transcript text for multibyte summaries.
- Improves rendering safety without changing the visible contract for ASCII summaries.
