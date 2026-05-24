## 1. Implementation

- [x] 1.1 Add a reader-based stdin helper for passphrase pipe reads
- [x] 1.2 Add an internal acquisition helper that accepts injected stdin/stderr
- [x] 1.3 Keep `Acquire(...)` behavior unchanged by delegating to the helper

## 2. Tests

- [x] 2.1 Remove `os.Stdin` swapping from stdin-pipe tests
- [x] 2.2 Add keyring warning coverage through an injected stderr writer

## 3. Downstream

- [x] 3.1 Update OpenSpec passphrase-acquisition spec with the stream-seam contract
- [x] 3.2 Add change proposal
- [x] 3.3 Add change design
- [x] 3.4 Add delta spec
