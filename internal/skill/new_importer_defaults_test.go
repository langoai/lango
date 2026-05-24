package skill

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewImporterDefaults(t *testing.T) {
	t.Parallel()

	im := NewImporter(zap.NewNop().Sugar())

	require.NotNil(t, im)
	require.NotNil(t, im.client)
	assert.Equal(t, 30*time.Second, im.client.Timeout)
	assert.NotNil(t, im.logger)
}

func TestImportFromRepoGitFallbackUsesHTTPWithoutLiveNetwork(t *testing.T) {
	const skillBody = `---
name: runChatValidModeReachesAppBuilderBeforeTui2-http-fallback
description: Imported after deterministic git failure
type: instruction
---

Fallback content.`

	installFakeGit(t, "fail", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/repos/owner/repo/contents/catalog":
			require.Equal(t, "feature", r.URL.Query().Get("ref"))
			require.NoError(t, json.NewEncoder(w).Encode([]gitHubContentsEntry{
				{Name: "fallback", Type: "dir", Path: "catalog/fallback"},
			}))
		case "/repos/owner/repo/contents/catalog/fallback/SKILL.md":
			writeGitHubFile(t, w, skillBody)
		case "/repos/owner/repo/contents/catalog/fallback/scripts",
			"/repos/owner/repo/contents/catalog/fallback/references",
			"/repos/owner/repo/contents/catalog/fallback/assets":
			http.NotFound(w, r)
		default:
			t.Fatalf("unexpected path: %s", r.URL.String())
		}
	}))
	defer server.Close()

	im := NewImporterWithClient(gitHubAPIClient(t, server.URL), zap.NewNop().Sugar())
	store := newRecordingSkillStore()
	ref := &GitHubRef{Owner: "owner", Repo: "repo", Branch: "feature", Path: "catalog"}

	result, err := im.ImportFromRepo(context.Background(), ref, store, ImportConfig{Concurrency: 1})
	require.NoError(t, err)

	assert.Equal(t, []string{"runChatValidModeReachesAppBuilderBeforeTui2-http-fallback"}, result.Imported)
	assert.Empty(t, result.Skipped)
	assert.Empty(t, result.Errors)
	assert.Equal(t, []string{"runChatValidModeReachesAppBuilderBeforeTui2-http-fallback"}, store.savedNames())
}

func TestImportViaGitCopiesFixtureAndHandlesBranches(t *testing.T) {
	fixture := t.TempDir()
	writeSkillFixture(t, filepath.Join(fixture, "packs", "valid"), "runChatValidModeReachesAppBuilderBeforeTui2-git-valid", "Valid git import.")
	writeSkillFixture(t, filepath.Join(fixture, "packs", "existing"), "runChatValidModeReachesAppBuilderBeforeTui2-existing", "Existing git import.")
	writeRawSkillFixture(t, filepath.Join(fixture, "packs", "invalid"), `---
description: Missing name
type: instruction
---

Invalid.`)
	require.NoError(t, os.WriteFile(filepath.Join(fixture, "packs", "README.md"), []byte("not a skill"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(fixture, "packs", "valid", "scripts", "nested"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(fixture, "packs", "valid", "scripts", "setup.sh"), []byte("echo setup"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(fixture, "packs", "valid", "scripts", "nested", "skip.sh"), []byte("skip"), 0o644))

	logPath := installFakeGit(t, "copy-fixture", fixture)

	im := NewImporterWithClient(http.DefaultClient, zap.NewNop().Sugar())
	store := newRecordingSkillStore()
	store.existing["runChatValidModeReachesAppBuilderBeforeTui2-existing"] = &SkillEntry{Name: "runChatValidModeReachesAppBuilderBeforeTui2-existing"}
	ref := &GitHubRef{Owner: "owner", Repo: "repo", Branch: "runChatValidModeReachesAppBuilderBeforeTui2-branch", Path: "packs"}

	result, err := im.importViaGit(context.Background(), ref, store, ImportConfig{})
	require.NoError(t, err)
	assertGitCloneArgs(t, logPath, "--depth=1 --branch runChatValidModeReachesAppBuilderBeforeTui2-branch https://github.com/owner/repo.git")

	assert.Equal(t, []string{"runChatValidModeReachesAppBuilderBeforeTui2-git-valid"}, result.Imported)
	assert.Equal(t, []string{"runChatValidModeReachesAppBuilderBeforeTui2-existing"}, result.Skipped)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "invalid: parse:")
	assert.Equal(t, []string{"runChatValidModeReachesAppBuilderBeforeTui2-git-valid"}, store.savedNames())
	assert.Equal(t, map[string]string{
		"runChatValidModeReachesAppBuilderBeforeTui2-git-valid/scripts/setup.sh": "echo setup",
	}, store.savedResources())
}

func TestImportViaGitReadClonedDirFailure(t *testing.T) {
	fixture := t.TempDir()
	installFakeGit(t, "copy-fixture", fixture)

	im := NewImporterWithClient(http.DefaultClient, zap.NewNop().Sugar())
	ref := &GitHubRef{Owner: "owner", Repo: "repo", Branch: "main", Path: "missing"}

	result, err := im.importViaGit(context.Background(), ref, newRecordingSkillStore(), ImportConfig{})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "read cloned dir")
}

func TestImportSingleWithResourcesHTTPBranches(t *testing.T) {
	const skillBody = `---
name: runChatValidModeReachesAppBuilderBeforeTui2-single-http
description: Single HTTP import
type: instruction
---

Single content.`

	t.Setenv("PATH", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/repos/owner/repo/contents/root/single/SKILL.md":
			writeGitHubFile(t, w, skillBody)
		case "/repos/owner/repo/contents/root/single/scripts":
			require.NoError(t, json.NewEncoder(w).Encode([]gitHubContentsEntry{
				{Name: "run.sh", Type: "file", Path: "root/single/scripts/run.sh"},
				{Name: "ignored-dir", Type: "dir", Path: "root/single/scripts/ignored-dir"},
			}))
		case "/repos/owner/repo/contents/root/single/references":
			fmt.Fprint(w, "not json")
		case "/repos/owner/repo/contents/root/single/assets":
			require.NoError(t, json.NewEncoder(w).Encode([]gitHubContentsEntry{
				{Name: "bad.bin", Type: "file", Path: "root/single/assets/bad.bin"},
			}))
		case "/repos/owner/repo/contents/root/single/scripts/run.sh":
			writeGitHubFile(t, w, "echo run")
		case "/repos/owner/repo/contents/root/single/assets/bad.bin":
			require.NoError(t, json.NewEncoder(w).Encode(gitHubFileResponse{
				Content:  base64.StdEncoding.EncodeToString([]byte("bad")),
				Encoding: "utf-8",
			}))
		default:
			t.Fatalf("unexpected path: %s", r.URL.String())
		}
	}))
	defer server.Close()

	im := NewImporterWithClient(gitHubAPIClient(t, server.URL), zap.NewNop().Sugar())
	store := newRecordingSkillStore()
	ref := &GitHubRef{Owner: "owner", Repo: "repo", Branch: "main", Path: "root"}

	entry, err := im.ImportSingleWithResources(context.Background(), ref, "single", store)
	require.NoError(t, err)

	assert.Equal(t, "runChatValidModeReachesAppBuilderBeforeTui2-single-http", entry.Name)
	assert.Equal(t, "https://github.com/owner/repo", entry.Source)
	assert.Equal(t, []string{"runChatValidModeReachesAppBuilderBeforeTui2-single-http"}, store.savedNames())
	assert.Equal(t, map[string]string{
		"runChatValidModeReachesAppBuilderBeforeTui2-single-http/scripts/run.sh": "echo run",
	}, store.savedResources())
}

func TestImportSingleViaGitSuccessAndFallbackFailure(t *testing.T) {
	t.Run("success from cloned fixture", func(t *testing.T) {
		fixture := t.TempDir()
		writeSkillFixture(t, filepath.Join(fixture, "skills", "single"), "runChatValidModeReachesAppBuilderBeforeTui2-single-git", "Single git import.")
		require.NoError(t, os.MkdirAll(filepath.Join(fixture, "skills", "single", "assets"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(fixture, "skills", "single", "assets", "note.txt"), []byte("asset"), 0o644))
		logPath := installFakeGit(t, "copy-fixture", fixture)

		im := NewImporterWithClient(http.DefaultClient, zap.NewNop().Sugar())
		store := newRecordingSkillStore()
		ref := &GitHubRef{Owner: "owner", Repo: "repo", Branch: "runChatValidModeReachesAppBuilderBeforeTui2-single-branch", Path: "skills"}

		entry, err := im.ImportSingleWithResources(context.Background(), ref, "single", store)
		require.NoError(t, err)
		assertGitCloneArgs(t, logPath, "--depth=1 --branch runChatValidModeReachesAppBuilderBeforeTui2-single-branch https://github.com/owner/repo.git")

		assert.Equal(t, "runChatValidModeReachesAppBuilderBeforeTui2-single-git", entry.Name)
		assert.Equal(t, []string{"runChatValidModeReachesAppBuilderBeforeTui2-single-git"}, store.savedNames())
		assert.Equal(t, map[string]string{
			"runChatValidModeReachesAppBuilderBeforeTui2-single-git/assets/note.txt": "asset",
		}, store.savedResources())
	})

	t.Run("clone failure falls back to HTTP error", func(t *testing.T) {
		installFakeGit(t, "fail", "")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "missing", http.StatusNotFound)
		}))
		defer server.Close()

		im := NewImporterWithClient(gitHubAPIClient(t, server.URL), zap.NewNop().Sugar())
		store := newRecordingSkillStore()
		ref := &GitHubRef{Owner: "owner", Repo: "repo", Branch: "main"}

		entry, err := im.ImportSingleWithResources(context.Background(), ref, "missing", store)
		require.Error(t, err)
		assert.Nil(t, entry)
		assert.Contains(t, err.Error(), `fetch SKILL.md for "missing"`)
		assert.Contains(t, err.Error(), "HTTP 404")
		assert.Empty(t, store.savedNames())
	})
}

func TestCloneRepoGitFailureIncludesOutput(t *testing.T) {
	installFakeGit(t, "fail", "")

	im := NewImporterWithClient(http.DefaultClient, zap.NewNop().Sugar())
	cloneDir, err := im.cloneRepo(context.Background(), &GitHubRef{
		Owner:  "owner",
		Repo:   "repo",
		Branch: "main",
	})

	require.Error(t, err)
	assert.Empty(t, cloneDir)
	assert.Contains(t, err.Error(), "runChatValidModeReachesAppBuilderBeforeTui2 fake git failure")
}

func installFakeGit(t *testing.T, mode, fixture string) string {
	t.Helper()

	binDir := t.TempDir()
	scriptPath := filepath.Join(binDir, "git")
	logPath := filepath.Join(binDir, "git.args")
	script := fmt.Sprintf(`#!/bin/sh
set -eu
mode=%q
fixture=%q
log=%q
printf '%%s\n' "$*" >> "$log"
if [ "$mode" = "fail" ]; then
  echo "runChatValidModeReachesAppBuilderBeforeTui2 fake git failure" >&2
  exit 42
fi
dest="${@: -1}"
mkdir -p "$dest"
cp -R "$fixture"/. "$dest"/
`, mode, fixture, logPath)
	// macOS /bin/sh does not support ${@: -1}; use a POSIX loop instead.
	script = strings.ReplaceAll(script, `dest="${@: -1}"`, `dest=""
for arg in "$@"; do
  dest="$arg"
done`)
	require.NoError(t, os.WriteFile(scriptPath, []byte(script), 0o755))
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return logPath
}

func assertGitCloneArgs(t *testing.T, logPath, wantFragment string) {
	t.Helper()

	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), wantFragment)
}

func writeSkillFixture(t *testing.T, dir, name, body string) {
	t.Helper()

	writeRawSkillFixture(t, dir, fmt.Sprintf(`---
name: %s
description: Import fixture
type: instruction
---

%s`, name, body))
}

func writeRawSkillFixture(t *testing.T, dir, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644))
}
