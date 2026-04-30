# Design

## Boundary

`cmd/lango` owns process assembly and is the only production package that may import `internal/app`. Packages under `internal/cli/**` must not import `internal/app`.

## Hook Registry

`BuildHookRegistry` moves from `internal/app` to exported `internal/toolchain.BuildHookRegistry`. Runtime bootstrap and CLI snapshot mode both call the same toolchain function.

## Cockpit Status

Cockpit status page uses `func() []types.FeatureStatus` as its provider contract. The unused concrete app status field on `cockpit.Deps` is removed.

## Status Dead-Letter Loader

`internal/cli/status` exports a narrow dead-letter bridge API and accepts a `DeadLetterBridgeLoader`. `cmd/lango` owns the lazy app-backed loader in `cmd/lango/dead_letter_status.go`.

## Enforcement

`internal/archtest` fails when production imports of `github.com/langoai/lango/internal/app` appear outside the allowed `cmd/lango` entrypoint files. It also fails when packages under `internal/cli/**` import `github.com/langoai/lango/internal/app`.
