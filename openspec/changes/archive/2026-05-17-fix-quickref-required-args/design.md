# Design

This is a documentation truthfulness fix. Runtime command behavior is already
correct and dedicated CLI pages already show the accepted forms.

The regression guard should stay narrow:

- Assert README provenance quick-reference snippets include required operands.
- Assert CLI index provenance quick-reference snippets include required operands.
- Assert README and CLI index P2P reputation snippets include `--peer-did <did>`.

The guard intentionally avoids parsing every command table; it only protects the
stale high-value snippets found in this change.
