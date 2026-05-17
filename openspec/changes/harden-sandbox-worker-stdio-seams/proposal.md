# Harden Sandbox Worker Stdio Seams

## Why

`internal/sandbox.RunWorker` still binds the public worker wrapper directly to `os.Stdin` and `os.Stdout`, even though `RunWorkerWithIO` is already injectable. The worker stdout stream is a JSON protocol channel, so regression tests should be able to exercise the public wrapper without replacing process-global standard streams.

## What Changes

- Add unexported package-level stdio seams for the public sandbox worker wrapper.
- Keep `RunWorkerWithIO` as the protocol implementation and leave the worker JSON protocol unchanged.
- Add focused regression coverage proving `RunWorker` uses the injected stdin/stdout seams.

## Impact

- No CLI behavior change.
- No sandbox protocol change.
- Test-only seam overrides must remain serialized because the seams are package globals.
