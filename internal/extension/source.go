package extension

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// manifestFileName is the on-disk manifest filename a pack must use.
const manifestFileName = "extension.yaml"

// Source produces a read-only WorkingCopy of a pack from a location
// (local directory, git URL, etc.). Implementations SHOULD keep Fetch
// side-effect-free outside the returned WorkingCopy.RootDir — inspect
// and install both call Fetch, and inspect guarantees no state remains
// under the user's home dir.
type Source interface {
	Fetch(ctx context.Context) (*WorkingCopy, error)
}

// WorkingCopy is the short-lived handle returned by Source.Fetch. It
// bundles the parsed manifest, the bytes-on-disk root, SHA-256 hashes of
// the manifest and every declared content file, and a Cleanup func that
// removes any temp resources. Local sources return a no-op Cleanup.
type WorkingCopy struct {
	Manifest       *Manifest
	RootDir        string
	ManifestSHA256 string
	FileHashes     map[string]string // keyed by relative path → hex SHA-256
	SourceRef      string            // git commit SHA or local path; recorded in .installed
	Cleanup        func() error
}

// LocalSource reads a pack from an on-disk directory. The directory must
// contain extension.yaml at its root.
type LocalSource struct {
	Dir string
}

// NewLocalSource constructs a LocalSource from a directory path.
func NewLocalSource(dir string) *LocalSource { return &LocalSource{Dir: dir} }

// Fetch reads, parses, and hashes the pack in place. Cleanup is a no-op.
func (s *LocalSource) Fetch(_ context.Context) (*WorkingCopy, error) {
	absRoot, err := filepath.Abs(s.Dir)
	if err != nil {
		return nil, fmt.Errorf("resolve pack dir: %w", err)
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("stat pack dir %q: %w", absRoot, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("pack dir %q is not a directory", absRoot)
	}
	return fetchFromDir(absRoot, absRoot, func() error { return nil })
}

// GitSource clones a git repository into a temp directory and then behaves
// like LocalSource on the clone. The URL may carry a `#<ref>` suffix to
// pin to a specific commit, branch, or tag; when absent, the default
// branch is cloned and the resolved HEAD SHA is recorded.
type GitSource struct {
	URL string
}

// NewGitSource constructs a GitSource from a URL. If the URL has a `#<ref>`
// suffix the fetch pins to that ref.
func NewGitSource(url string) *GitSource { return &GitSource{URL: url} }

// Fetch clones the repo and returns a WorkingCopy with Cleanup that
// removes the temp dir. On error, the temp dir is cleaned up eagerly.
func (s *GitSource) Fetch(ctx context.Context) (*WorkingCopy, error) {
	url, ref := splitGitRef(s.URL)

	tmp, err := os.MkdirTemp("", "lango-extension-")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	if err := cloneAndCheckout(ctx, url, ref, tmp); err != nil {
		_ = os.RemoveAll(tmp)
		return nil, err
	}

	// Record the resolved HEAD SHA for .installed.
	resolvedSHA := resolveHeadSHA(ctx, tmp)
	if ref != "" {
		resolvedSHA = ref + "@" + resolvedSHA
	}

	cleanup := func() error { return os.RemoveAll(tmp) }
	wc, err := fetchFromDir(tmp, tmp, cleanup)
	if err != nil {
		_ = cleanup()
		return nil, err
	}
	wc.SourceRef = resolvedSHA
	return wc, nil
}

// cloneAndCheckout handles the git clone strategy. Branch/tag refs use
// --depth=1 --branch; commit SHAs clone without --branch (shallow clone
// cannot fetch arbitrary SHAs) and then checkout the specific commit.
func cloneAndCheckout(ctx context.Context, url, ref, dst string) error {
	if ref == "" || !looksLikeSHA(ref) {
		// Branch, tag, or default branch: shallow clone with --branch.
		args := []string{"clone", "--depth=1"}
		if ref != "" {
			args = append(args, "--branch", ref)
		}
		args = append(args, url, dst)
		cmd := exec.CommandContext(ctx, "git", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git clone: %s: %w", strings.TrimSpace(string(out)), err)
		}
		return nil
	}

	// Commit SHA: full clone (shallow clone can't fetch arbitrary SHAs),
	// then checkout the exact commit.
	cloneCmd := exec.CommandContext(ctx, "git", "clone", url, dst)
	if out, err := cloneCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone: %s: %w", strings.TrimSpace(string(out)), err)
	}
	checkoutCmd := exec.CommandContext(ctx, "git", "-C", dst, "checkout", ref)
	if out, err := checkoutCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout %s: %s: %w", ref, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// looksLikeSHA returns true if ref is 7–40 lowercase hex characters,
// indicating a commit SHA rather than a branch or tag name.
func looksLikeSHA(ref string) bool {
	if len(ref) < 7 || len(ref) > 40 {
		return false
	}
	for _, c := range ref {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// splitGitRef partitions a URL of the form "repo#ref" into ("repo", "ref").
// An absent ref produces the empty string.
func splitGitRef(urlish string) (string, string) {
	if i := strings.LastIndex(urlish, "#"); i >= 0 {
		return urlish[:i], urlish[i+1:]
	}
	return urlish, ""
}

// resolveHeadSHA reports the current HEAD commit of a checkout. On
// error, returns the empty string — the caller still has the URL.
func resolveHeadSHA(ctx context.Context, repoDir string) string {
	cmd := exec.CommandContext(ctx, "git", "-C", repoDir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// fetchFromDir is the shared body that reads, validates, and hashes a
// pack rooted at dir. sourceRef is the value recorded in .installed.
func fetchFromDir(dir, sourceRef string, cleanup func() error) (*WorkingCopy, error) {
	manifestPath := filepath.Join(dir, manifestFileName)
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest %s: %w", manifestPath, err)
	}
	m, err := ParseManifest(strings.NewReader(string(manifestBytes)))
	if err != nil {
		return nil, err
	}
	manifestSum := sha256.Sum256(manifestBytes)
	hashes := map[string]string{}

	hashFile := func(rel string) error {
		abs, err := ResolvePath(dir, rel)
		if err != nil {
			return fmt.Errorf("path safety for %q: %w", rel, err)
		}
		data, err := os.ReadFile(abs)
		if err != nil {
			return fmt.Errorf("read %s: %w", rel, err)
		}
		sum := sha256.Sum256(data)
		hashes[filepath.ToSlash(filepath.Clean(rel))] = hex.EncodeToString(sum[:])
		return nil
	}
	for _, s := range m.Contents.Skills {
		absPath, err := ResolvePath(dir, s.Path)
		if err != nil {
			return nil, err
		}
		info, statErr := os.Stat(absPath)
		if statErr != nil {
			return nil, fmt.Errorf("stat skill %q: %w", s.Path, statErr)
		}

		// Determine the directory to hash: for directories use it directly,
		// for SKILL.md promote to the parent directory, otherwise hash the
		// single file only.
		var hashDirAbs string
		if info.IsDir() {
			hashDirAbs = absPath
		} else {
			if err := hashFile(s.Path); err != nil {
				return nil, err
			}
			if strings.HasSuffix(filepath.Base(s.Path), "SKILL.md") {
				parentRel := filepath.Dir(s.Path)
				parentAbs, pErr := ResolvePath(dir, parentRel)
				if pErr != nil {
					return nil, fmt.Errorf("hash skill dir %q: %w", parentRel, pErr)
				}
				hashDirAbs = parentAbs
			}
		}

		// Walk and hash every file in the skill directory so tamper
		// detection covers all content that copySkillsToStore/copyPackFiles copy.
		if hashDirAbs != "" {
			resolvedDir, err := filepath.EvalSymlinks(dir)
			if err != nil {
				return nil, fmt.Errorf("resolve pack root: %w", err)
			}
			if walkErr := filepath.Walk(hashDirAbs, func(path string, fi os.FileInfo, err error) error {
				if err != nil || fi.IsDir() {
					return err
				}
				rel, relErr := filepath.Rel(resolvedDir, path)
				if relErr != nil {
					return relErr
				}
				normalized := filepath.ToSlash(filepath.Clean(rel))
				if _, exists := hashes[normalized]; exists {
					return nil
				}
				return hashFile(normalized)
			}); walkErr != nil {
				return nil, fmt.Errorf("hash skill dir %q: %w", s.Path, walkErr)
			}
		}
	}
	for _, p := range m.Contents.Prompts {
		if err := hashFile(p.Path); err != nil {
			return nil, err
		}
	}

	return &WorkingCopy{
		Manifest:       m,
		RootDir:        dir,
		ManifestSHA256: hex.EncodeToString(manifestSum[:]),
		FileHashes:     hashes,
		SourceRef:      sourceRef,
		Cleanup:        cleanup,
	}, nil
}
