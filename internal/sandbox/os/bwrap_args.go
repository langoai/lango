package os

import (
	"fmt"
	"os"
)

// compileBwrapArgs converts a Policy into bubblewrap (bwrap) CLI arguments.
//
// The returned slice contains the bwrap flags only — callers must prepend the
// bwrap binary path and append "--" followed by the original argv. Example:
//
//	flags, _ := compileBwrapArgs(policy)
//	cmd.Args = append(append([]string{bwrapPath}, flags...), append([]string{"--"}, originalArgs...)...)
//
// This function is platform-agnostic on purpose so the argv compiler can be
// unit-tested on macOS without invoking bwrap.
//
// Filesystem mapping (default-deny model):
//   - ReadOnlyGlobal=true → --ro-bind / /
//   - ReadPaths           → --ro-bind <p> <p>          (only when ReadOnlyGlobal=false)
//   - WritePaths          → --bind <p> <p>             (overlaid on the read-only root)
//   - DenyPaths           → --tmpfs <p>                (directory only — see note below)
//
// PR 3 supports directory-level deny only because bwrap's --tmpfs cannot be
// mounted on top of a regular file. File-level deny (e.g. --ro-bind /dev/null
// <file>) is planned for PR 4. compileBwrapArgs returns an error if a deny
// path is missing or refers to a non-directory.
//
// Network mapping:
//   - NetworkDeny / NetworkUnixOnly → --unshare-net
//   - NetworkAllow                  → no flag (host network)
//
// Note: NetworkUnixOnly is treated identically to NetworkDeny here because
// bwrap has no AF_UNIX-only filter. AF_UNIX sockets that are reachable via
// the bound filesystem still work because the socket file is shared via the
// --bind / --ro-bind mounts.
//
// Process namespaces (always enabled):
//   - --die-with-parent       parent exit kills the child
//   - --unshare-pid           PID namespace isolation
//   - --unshare-ipc           SysV IPC namespace isolation
//   - --unshare-uts           hostname namespace isolation
//   - --unshare-cgroup-try    cgroup namespace, best-effort (kernel >= 4.6)
//
// Standard mounts (always present):
//   - --proc /proc, --dev /dev, --tmpfs /run
func compileBwrapArgs(policy Policy) ([]string, error) {
	args := []string{
		"--die-with-parent",
		"--unshare-pid",
		"--unshare-ipc",
		"--unshare-uts",
		"--unshare-cgroup-try",
		"--proc", "/proc",
		"--dev", "/dev",
		"--tmpfs", "/run",
	}

	if policy.Filesystem.ReadOnlyGlobal {
		args = append(args, "--ro-bind", "/", "/")
	} else {
		for _, p := range policy.Filesystem.ReadPaths {
			clean, err := sanitizePath(p)
			if err != nil {
				return nil, fmt.Errorf("read path %q: %w", p, err)
			}
			args = append(args, "--ro-bind", clean, clean)
		}
	}

	for _, p := range policy.Filesystem.WritePaths {
		clean, err := sanitizePath(p)
		if err != nil {
			return nil, fmt.Errorf("write path %q: %w", p, err)
		}
		args = append(args, "--bind", clean, clean)
	}

	for _, p := range policy.Filesystem.DenyPaths {
		clean, err := sanitizePath(p)
		if err != nil {
			return nil, fmt.Errorf("deny path %q: %w", p, err)
		}
		fi, err := os.Stat(clean)
		if err != nil {
			return nil, fmt.Errorf("bwrap deny path %q: %w", p, err)
		}
		if !fi.IsDir() {
			return nil, fmt.Errorf("bwrap deny path %q must be a directory; file-level deny not yet supported", p)
		}
		args = append(args, "--tmpfs", clean)
	}

	switch policy.Network {
	case NetworkDeny, NetworkUnixOnly:
		args = append(args, "--unshare-net")
	case NetworkAllow:
		// host network — no flag
	}

	return args, nil
}
