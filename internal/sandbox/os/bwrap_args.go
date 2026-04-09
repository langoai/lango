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
//   - DenyPaths (dir)     → --tmpfs <p>                (empty tmpfs over the tree)
//   - DenyPaths (file)    → --ro-bind /dev/null <p>    (read yields EOF, write EACCES)
//
// PR 5c added file-level deny via --ro-bind /dev/null <file>, closing the
// directory-only limitation of PR 3/4. compileBwrapArgs still returns an error
// if a deny path is missing or refers to a non-regular, non-directory file
// (device nodes, sockets, fifos — uncommon, likely user config mistake).
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
// Mount order is load-bearing: bubblewrap processes options left-to-right,
// and the later `--ro-bind / /` would shadow any earlier `/proc`, `/dev`,
// or `/run` mounts (the root bind replaces whatever is at the target,
// undoing the specialised mounts). Follow the standard bwrap wrapper
// pattern — root bind first, then `--proc /proc`, `--dev /dev`,
// `--tmpfs /run` layered on top — so the sandboxed child sees a fresh
// procfs (new PID namespace), a filtered /dev, and an empty /run.
func compileBwrapArgs(policy Policy) ([]string, error) {
	args := []string{
		"--die-with-parent",
		"--unshare-pid",
		"--unshare-ipc",
		"--unshare-uts",
		"--unshare-cgroup-try",
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

	// Overlay standard special mounts AFTER the root/read bind so they are
	// not shadowed by the later root mount.
	args = append(args,
		"--proc", "/proc",
		"--dev", "/dev",
		"--tmpfs", "/run",
	)

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
		mode := fi.Mode()
		switch {
		case mode.IsDir():
			// Mount an empty tmpfs over the directory so read/write both
			// yield empty + EACCES for the sandboxed child.
			args = append(args, "--tmpfs", clean)
		case mode.IsRegular():
			// Bind /dev/null read-only over the file so reads yield EOF
			// and writes fail with EACCES while preserving the parent
			// directory structure. This closes the file-level deny gap
			// that PR 3/4 left open.
			args = append(args, "--ro-bind", "/dev/null", clean)
		default:
			return nil, fmt.Errorf("bwrap deny path %q: unsupported file mode %s (not a regular file or directory)", p, mode)
		}
	}

	switch policy.Network {
	case NetworkDeny, NetworkUnixOnly:
		args = append(args, "--unshare-net")
	case NetworkAllow:
		// host network — no flag
	}

	return args, nil
}
