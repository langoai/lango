package os

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

const seatbeltTemplate = `(version 1)
(deny default)

;; Allow process execution
(allow process-exec)
{{- if .AllowFork}}
(allow process-fork)
{{- end}}

;; Allow read access
{{- if .ReadOnlyGlobal}}
(allow file-read*)
{{- else}}
{{- range .ReadPaths}}
(allow file-read* (subpath "{{.}}"))
{{- end}}
{{- end}}

;; Allow write access to specific paths
{{- range .WritePaths}}
(allow file-write* (subpath "{{.}}"))
{{- end}}

;; Deny write access to specific paths (overrides allow)
{{- range .DenyPaths}}
(deny file-write* (subpath "{{.}}"))
{{- end}}

;; Network policy
{{- if eq .NetworkMode "deny"}}
(deny network*)
{{- else if eq .NetworkMode "unix-only"}}
(deny network*)
(allow network* (local unix))
{{- else if eq .NetworkMode "allow"}}
(allow network*)
{{- end}}
{{- if and (eq .NetworkMode "deny") (gt (len .AllowedIPs) 0)}}
;; Allowed outbound IPs
{{- range .AllowedIPs}}
(allow network-outbound (remote ip "{{.}}:*"))
{{- end}}
{{- end}}

;; Standard permits
(allow sysctl-read)
(allow mach-lookup)
(allow ipc-posix-shm-read-data)
(allow ipc-posix-shm-write-data)
(allow signal (target self))
`

type seatbeltData struct {
	ReadOnlyGlobal bool
	ReadPaths      []string
	WritePaths     []string
	DenyPaths      []string
	NetworkMode    string
	AllowedIPs     []string
	AllowFork      bool
}

// GenerateSeatbeltProfile creates a macOS Seatbelt .sb profile string from a Policy.
// All paths are resolved to absolute paths and validated against injection.
func GenerateSeatbeltProfile(policy Policy) (string, error) {
	data := seatbeltData{
		ReadOnlyGlobal: policy.Filesystem.ReadOnlyGlobal,
		NetworkMode:    string(policy.Network),
		AllowFork:      policy.Process.AllowFork,
	}

	// Resolve and validate paths.
	for _, p := range policy.Filesystem.ReadPaths {
		clean, err := sanitizePath(p)
		if err != nil {
			return "", fmt.Errorf("read path %q: %w", p, err)
		}
		data.ReadPaths = append(data.ReadPaths, clean)
	}
	for _, p := range policy.Filesystem.WritePaths {
		clean, err := sanitizePath(p)
		if err != nil {
			return "", fmt.Errorf("write path %q: %w", p, err)
		}
		data.WritePaths = append(data.WritePaths, clean)
	}
	for _, p := range policy.Filesystem.DenyPaths {
		clean, err := sanitizePath(p)
		if err != nil {
			return "", fmt.Errorf("deny path %q: %w", p, err)
		}
		data.DenyPaths = append(data.DenyPaths, clean)
	}
	for _, ip := range policy.AllowedNetworkIPs {
		if err := validateIP(ip); err != nil {
			return "", fmt.Errorf("allowed IP %q: %w", ip, err)
		}
		data.AllowedIPs = append(data.AllowedIPs, ip)
	}

	tmpl, err := template.New("seatbelt").Parse(seatbeltTemplate)
	if err != nil {
		return "", fmt.Errorf("parse seatbelt template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute seatbelt template: %w", err)
	}
	return buf.String(), nil
}

// sanitizePath resolves a path to absolute and validates it against injection characters.
func sanitizePath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("resolve absolute: %w", err)
	}
	// Reject paths containing characters that could break the S-expression.
	for _, c := range abs {
		if c == '"' || c == '(' || c == ')' || c == ';' || c == '\n' || c == '\r' {
			return "", fmt.Errorf("%w: path contains invalid character %q", ErrInvalidPolicy, string(c))
		}
	}
	return abs, nil
}

// validateIP validates an IP address string for Seatbelt profile injection safety.
func validateIP(ip string) error {
	// Allow only alphanumeric, dots, colons (IPv6), and star for port wildcard.
	for _, c := range ip {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') && c != '.' && c != ':' {
			return fmt.Errorf("%w: IP contains invalid character %q", ErrInvalidPolicy, string(c))
		}
	}
	if strings.TrimSpace(ip) == "" {
		return fmt.Errorf("%w: empty IP", ErrInvalidPolicy)
	}
	return nil
}
