## MODIFIED Requirements

### Requirement: Platform capability detection
The system SHALL detect available OS sandbox primitives via `Probe()` returning `PlatformCapabilities` with `HasSeatbelt`, `SeatbeltReason`, `HasLandlock`, `LandlockABI`, `LandlockReason`, `HasSeccomp`, `SeccompReason`, `Platform`, `KernelVersion`. Probe functions SHALL NOT use concrete type-casts on isolator instances. On Linux, `probeLandlockKernel` and `probeSeccompKernel` SHALL use real syscalls via `golang.org/x/sys/unix` (not stub returns).

The `HasSeccomp` field doc comment SHALL state explicitly that a `true` value indicates only that the kernel exposes the seccomp prctl interface and does NOT prove that BPF filters are installable. The qualified description SHALL appear in `SeccompReason`.

#### Scenario: macOS probe detects seatbelt
- **WHEN** `Probe()` is called on macOS with sandbox-exec available
- **THEN** `HasSeatbelt` SHALL be true

#### Scenario: Linux probe without concrete cast
- **WHEN** `Probe()` is called on Linux
- **THEN** `probePlatform()` uses standalone probe functions without constructing isolator instances

#### Scenario: Reason fields populated
- **WHEN** `Probe()` is called on any platform
- **THEN** reason fields explain the probe result (e.g., `"sandbox-exec found"`, `"Landlock ABI 3"`, `"seccomp interface present (PR_GET_SECCOMP=N)"`, `"not on Linux"`)

#### Scenario: Linux Landlock probe uses real syscall
- **WHEN** `probeLandlockKernel()` runs on a Linux kernel ≥ 5.13 that supports Landlock
- **THEN** it SHALL return `(true, abi, "Landlock ABI N")` where `abi > 0`, having called `unix.Syscall(SYS_LANDLOCK_CREATE_RULESET, 0, 0, LANDLOCK_CREATE_RULESET_VERSION)`

#### Scenario: Linux Landlock probe handles ENOSYS
- **WHEN** `probeLandlockKernel()` runs on a Linux kernel that does not support Landlock
- **THEN** it SHALL return `(false, 0, "Landlock not supported by this kernel (requires 5.13+)")` after the syscall returns ENOSYS

#### Scenario: Linux seccomp probe captures presence only
- **WHEN** `probeSeccompKernel()` runs and `unix.PrctlRetInt(PR_GET_SECCOMP, ...)` returns successfully
- **THEN** it SHALL return `(true, reason)` where `reason` contains `"seccomp interface present"` and the substring `"BPF filter capability not directly verified"`

#### Scenario: Linux seccomp probe augments with /proc/self/status
- **WHEN** `/proc/self/status` is readable and contains a `Seccomp:` line
- **THEN** the seccomp reason SHALL include `"/proc/self/status Seccomp=<value>"`

#### Scenario: HasSeccomp doc comment carries the caveat
- **WHEN** a developer reads the `PlatformCapabilities.HasSeccomp` field doc comment
- **THEN** it SHALL state that `true` means only that the kernel exposes the prctl interface and does NOT prove BPF filter capability
