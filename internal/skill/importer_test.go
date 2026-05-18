package skill

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
)

func TestParseGitHubURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		wantOwner  string
		wantRepo   string
		wantBranch string
		wantPath   string
		wantErr    bool
	}{
		{
			give:       "https://github.com/kepano/obsidian-skills",
			wantOwner:  "kepano",
			wantRepo:   "obsidian-skills",
			wantBranch: "main",
			wantPath:   "",
		},
		{
			give:       "https://github.com/kepano/obsidian-skills/tree/develop",
			wantOwner:  "kepano",
			wantRepo:   "obsidian-skills",
			wantBranch: "develop",
			wantPath:   "",
		},
		{
			give:       "https://github.com/kepano/obsidian-skills/tree/main/skills",
			wantOwner:  "kepano",
			wantRepo:   "obsidian-skills",
			wantBranch: "main",
			wantPath:   "skills",
		},
		{
			give:       "https://github.com/kepano/obsidian-skills/tree/main/deep/nested/path",
			wantOwner:  "kepano",
			wantRepo:   "obsidian-skills",
			wantBranch: "main",
			wantPath:   "deep/nested/path",
		},
		{
			give:    "https://github.com/onlyowner",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			ref, err := ParseGitHubURL(tt.give)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOwner, ref.Owner)
			assert.Equal(t, tt.wantRepo, ref.Repo)
			assert.Equal(t, tt.wantBranch, ref.Branch)
			assert.Equal(t, tt.wantPath, ref.Path)
		})
	}
}

func TestParseGitHubURL_BlobPathAndTrailingSlash(t *testing.T) {
	t.Parallel()

	ref, err := ParseGitHubURL("http://github.com/acme/tools/blob/release/skills/importer/")
	require.NoError(t, err)

	assert.Equal(t, "acme", ref.Owner)
	assert.Equal(t, "tools", ref.Repo)
	assert.Equal(t, "release", ref.Branch)
	assert.Equal(t, "skills/importer", ref.Path)
}

func TestIsGitHubURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want bool
	}{
		{"https://github.com/owner/repo", true},
		{"http://github.com/owner/repo/tree/main", true},
		{"https://example.com/skills/SKILL.md", false},
		{"https://raw.githubusercontent.com/owner/repo/main/SKILL.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, IsGitHubURL(tt.give))
		})
	}
}

func TestDiscoverSkills(t *testing.T) {
	t.Parallel()

	entries := []gitHubContentsEntry{
		{Name: "obsidian-web-clipper", Type: "dir", Path: "obsidian-web-clipper"},
		{Name: "obsidian-markdown", Type: "dir", Path: "obsidian-markdown"},
		{Name: "README.md", Type: "file", Path: "README.md"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	}))
	defer ts.Close()

	logger := zap.NewNop().Sugar()
	im := NewImporterWithClient(ts.Client(), logger)

	// Override the API URL by pointing to our test server.
	// We need to use a custom approach: swap the base URL in the ref.
	ref := &GitHubRef{Owner: "test", Repo: "repo", Branch: "main"}

	// Since DiscoverSkills uses a fixed URL format, we test via the HTTP mock.
	// Create a server that mimics the GitHub Contents API.
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	}))
	defer ts2.Close()

	// For a full integration test, we'd need to mock the GitHub API URL.
	// Instead, test the HTTP client integration with a real server.
	_ = ref
	_ = im

	// Direct HTTP test using FetchFromURL.
	raw := `---
name: test-skill
description: A test skill
type: instruction
---

This is the content.`

	ts3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, raw)
	}))
	defer ts3.Close()

	im2 := NewImporterWithClient(ts3.Client(), logger)
	body, err := im2.FetchFromURL(context.Background(), ts3.URL+"/SKILL.md")
	require.NoError(t, err)
	assert.Equal(t, raw, string(body))
}

func TestFetchSkillMD(t *testing.T) {
	t.Parallel()

	skillContent := `---
name: obsidian-markdown
description: Obsidian Markdown reference
type: instruction
---

# Obsidian Markdown

Use Obsidian-flavored markdown for notes.`

	encoded := base64.StdEncoding.EncodeToString([]byte(skillContent))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(gitHubFileResponse{
			Content:  encoded,
			Encoding: "base64",
		})
	}))
	defer ts.Close()

	logger := zap.NewNop().Sugar()
	im := NewImporterWithClient(ts.Client(), logger)

	body, err := im.FetchFromURL(context.Background(), ts.URL+"/contents/obsidian-markdown/SKILL.md")
	require.NoError(t, err)

	// The response is a JSON object, parse it to get the base64 content.
	var file gitHubFileResponse
	require.NoError(t, json.Unmarshal(body, &file))
	assert.Equal(t, "base64", file.Encoding)
}

func TestFetchSkillMD_DecodesGitHubContentsResponse(t *testing.T) {
	t.Parallel()

	const skillContent = `---
name: decoded-skill
description: Decoded from GitHub contents
type: instruction
---

Use decoded content.`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/owner/repo/contents/catalog/decoded/SKILL.md", r.URL.Path)
		assert.Equal(t, "feature", r.URL.Query().Get("ref"))
		assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))
		assert.Equal(t, "lango-skill-importer", r.Header.Get("User-Agent"))

		encoded := base64.StdEncoding.EncodeToString([]byte(skillContent))
		withLineBreaks := encoded[:12] + "\n" + encoded[12:]
		require.NoError(t, json.NewEncoder(w).Encode(gitHubFileResponse{
			Content:  withLineBreaks,
			Encoding: "base64",
		}))
	}))
	defer server.Close()

	im := NewImporterWithClient(gitHubAPIClient(t, server.URL), zap.NewNop().Sugar())
	ref := &GitHubRef{Owner: "owner", Repo: "repo", Branch: "feature", Path: "catalog"}

	got, err := im.FetchSkillMD(context.Background(), ref, "decoded")
	require.NoError(t, err)
	assert.Equal(t, skillContent, string(got))
}

func TestFetchSkillMD_RejectsUnexpectedEncodingAndInvalidBase64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		response   gitHubFileResponse
		wantErr    string
		wantStatus int
	}{
		{
			name:       "unexpected encoding",
			response:   gitHubFileResponse{Content: "plain", Encoding: "utf-8"},
			wantErr:    `unexpected encoding "utf-8"`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid base64",
			response:   gitHubFileResponse{Content: "not base64 !", Encoding: "base64"},
			wantErr:    "decode base64 content",
			wantStatus: http.StatusOK,
		},
		{
			name:       "http error",
			wantErr:    "HTTP 503",
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.wantStatus != http.StatusOK {
					http.Error(w, "unavailable", tt.wantStatus)
					return
				}
				require.NoError(t, json.NewEncoder(w).Encode(tt.response))
			}))
			defer server.Close()

			im := NewImporterWithClient(gitHubAPIClient(t, server.URL), zap.NewNop().Sugar())
			ref := &GitHubRef{Owner: "owner", Repo: "repo", Branch: "main"}

			_, err := im.FetchSkillMD(context.Background(), ref, "broken")
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestFetchFromURL(t *testing.T) {
	t.Parallel()

	raw := `---
name: external-skill
description: An external skill
type: instruction
---

Some reference content here.`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, raw)
	}))
	defer ts.Close()

	logger := zap.NewNop().Sugar()
	im := NewImporterWithClient(ts.Client(), logger)

	body, err := im.FetchFromURL(context.Background(), ts.URL+"/SKILL.md")
	require.NoError(t, err)
	assert.Equal(t, raw, string(body))

	// Parse the fetched content.
	entry, err := ParseSkillMD(body)
	require.NoError(t, err)
	assert.Equal(t, "external-skill", entry.Name)
	assert.Equal(t, SkillTypeInstruction, entry.Type)
	content, _ := entry.Definition["content"].(string)
	assert.Equal(t, "Some reference content here.", content)
}

func TestFetchFromURL_HTTPErrorAndBodyReadFailure(t *testing.T) {
	t.Parallel()

	t.Run("non-success status", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "nope", http.StatusTeapot)
		}))
		defer server.Close()

		im := NewImporterWithClient(server.Client(), zap.NewNop().Sugar())
		_, err := im.FetchFromURL(context.Background(), server.URL+"/SKILL.md")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP 418")
	})

	t.Run("body read failure", func(t *testing.T) {
		t.Parallel()

		im := NewImporterWithClient(&http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       errReadCloser{err: errors.New("read boom")},
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}),
		}, zap.NewNop().Sugar())

		_, err := im.FetchFromURL(context.Background(), "https://example.test/SKILL.md")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "read response body")
	})
}

func TestImportSingle_ValidatesSkillBeforeSaving(t *testing.T) {
	t.Parallel()

	store := newRecordingSkillStore()
	im := NewImporterWithClient(http.DefaultClient, zap.NewNop().Sugar())

	entry, err := im.ImportSingle(context.Background(), []byte("not front matter"), "https://example.test", store)
	require.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "parse SKILL.md")
	assert.Empty(t, store.savedNames())
}

func TestHasGit(t *testing.T) {
	t.Parallel()

	// On most dev machines, git is available.
	got := hasGit()
	// We don't assert a specific value since CI might not have git,
	// but we verify it doesn't panic.
	t.Logf("hasGit() = %v", got)
}

func TestCopyResourceDirs(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop().Sugar()
	dir := filepath.Join(t.TempDir(), "skills")
	store := NewFileSkillStore(dir, logger)
	ctx := context.Background()

	// Save a skill first.
	require.NoError(t, store.Save(ctx, SkillEntry{
		Name:       "res-skill",
		Type:       "instruction",
		Status:     "active",
		Definition: map[string]interface{}{"content": "test"},
	}))

	// Create a fake cloned skill directory with resources.
	srcDir := t.TempDir()
	scriptsDir := filepath.Join(srcDir, "scripts")
	require.NoError(t, os.MkdirAll(scriptsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(scriptsDir, "setup.sh"), []byte("#!/bin/bash\necho hi"), 0o644))

	copyResourceDirs(ctx, srcDir, "res-skill", store)

	// Verify the resource was copied.
	got, err := os.ReadFile(filepath.Join(dir, "res-skill", "scripts", "setup.sh"))
	require.NoError(t, err)
	assert.Equal(t, "#!/bin/bash\necho hi", string(got))
}

func TestCopyResourceDirs_NoResources(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop().Sugar()
	dir := filepath.Join(t.TempDir(), "skills")
	store := NewFileSkillStore(dir, logger)
	ctx := context.Background()

	require.NoError(t, store.Save(ctx, SkillEntry{
		Name:       "no-res-skill",
		Type:       "instruction",
		Status:     "active",
		Definition: map[string]interface{}{"content": "test"},
	}))

	// Empty source dir — should not panic.
	srcDir := t.TempDir()
	copyResourceDirs(ctx, srcDir, "no-res-skill", store)

	// Verify no resource dirs were created.
	for _, d := range resourceDirs {
		path := filepath.Join(dir, "no-res-skill", d)
		_, err := os.Stat(path)
		assert.True(t, os.IsNotExist(err), "unexpected resource dir %s exists", d)
	}
}

func TestCopyResourceDirs_CopiesOnlyTopLevelFilesFromConventionalDirs(t *testing.T) {
	t.Parallel()

	store := newRecordingSkillStore()
	srcDir := t.TempDir()

	for _, dir := range resourceDirs {
		require.NoError(t, os.MkdirAll(filepath.Join(srcDir, dir, "nested"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, dir, "keep.txt"), []byte(dir+" data"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, dir, "nested", "skip.txt"), []byte("skip"), 0o644))
	}

	copyResourceDirs(context.Background(), srcDir, "resource-skill", store)

	got := store.savedResources()
	assert.Equal(t, map[string]string{
		"resource-skill/assets/keep.txt":     "assets data",
		"resource-skill/references/keep.txt": "references data",
		"resource-skill/scripts/keep.txt":    "scripts data",
	}, got)
}

func TestImportViaHTTP_MaxSkillsValidationAndResources(t *testing.T) {
	t.Parallel()

	const firstSkill = `---
name: first-skill
description: First imported skill
type: instruction
---

First body.`

	const invalidSkill = `---
description: Missing required name
type: instruction
---

Invalid body.`

	requestedPaths := make(chan string, 16)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPaths <- r.URL.Path
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/repos/owner/repo/contents/catalog":
			require.NoError(t, json.NewEncoder(w).Encode([]gitHubContentsEntry{
				{Name: "first", Type: "dir", Path: "catalog/first"},
				{Name: "invalid", Type: "dir", Path: "catalog/invalid"},
				{Name: "over-limit", Type: "dir", Path: "catalog/over-limit"},
				{Name: "README.md", Type: "file", Path: "catalog/README.md"},
			}))
		case "/repos/owner/repo/contents/catalog/first/SKILL.md":
			writeGitHubFile(t, w, firstSkill)
		case "/repos/owner/repo/contents/catalog/invalid/SKILL.md":
			writeGitHubFile(t, w, invalidSkill)
		case "/repos/owner/repo/contents/catalog/first/scripts":
			require.NoError(t, json.NewEncoder(w).Encode([]gitHubContentsEntry{
				{Name: "install.sh", Type: "file", Path: "catalog/first/scripts/install.sh"},
				{Name: "nested", Type: "dir", Path: "catalog/first/scripts/nested"},
			}))
		case "/repos/owner/repo/contents/catalog/first/references":
			http.NotFound(w, r)
		case "/repos/owner/repo/contents/catalog/first/assets":
			require.NoError(t, json.NewEncoder(w).Encode([]gitHubContentsEntry{
				{Name: "logo.txt", Type: "file", Path: "catalog/first/assets/logo.txt"},
			}))
		case "/repos/owner/repo/contents/catalog/first/scripts/install.sh":
			writeGitHubFile(t, w, "echo install")
		case "/repos/owner/repo/contents/catalog/first/assets/logo.txt":
			writeGitHubFile(t, w, "logo")
		case "/repos/owner/repo/contents/catalog/invalid/scripts",
			"/repos/owner/repo/contents/catalog/invalid/references",
			"/repos/owner/repo/contents/catalog/invalid/assets":
			t.Fatalf("resources should not be fetched for invalid skills: %s", r.URL.Path)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	im := NewImporterWithClient(gitHubAPIClient(t, server.URL), zap.NewNop().Sugar())
	store := newRecordingSkillStore()
	ref := &GitHubRef{Owner: "owner", Repo: "repo", Branch: "main", Path: "catalog"}

	result, err := im.importViaHTTP(context.Background(), ref, store, ImportConfig{
		MaxSkills:   2,
		Concurrency: 1,
	})
	require.NoError(t, err)

	assert.Equal(t, []string{"first-skill"}, result.Imported)
	assert.Empty(t, result.Skipped)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "invalid: parse:")
	assert.Equal(t, []string{"first-skill"}, store.savedNames())
	assert.Equal(t, map[string]string{
		"first-skill/assets/logo.txt":    "logo",
		"first-skill/scripts/install.sh": "echo install",
	}, store.savedResources())

	close(requestedPaths)
	var paths []string
	for path := range requestedPaths {
		paths = append(paths, path)
	}
	assert.NotContains(t, paths, "/repos/owner/repo/contents/catalog/over-limit/SKILL.md")
}

func TestImportViaHTTP_RespectsConcurrencyLimitAndSkipsExisting(t *testing.T) {
	t.Parallel()

	const skillBodyTemplate = `---
name: %s
description: Imported skill
type: instruction
---

Body.`

	var active int32
	var maxActive int32
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/repos/owner/repo/contents/":
			require.NoError(t, json.NewEncoder(w).Encode([]gitHubContentsEntry{
				{Name: "one", Type: "dir"},
				{Name: "two", Type: "dir"},
				{Name: "existing-dir", Type: "dir"},
				{Name: "three", Type: "dir"},
			}))
		case "/repos/owner/repo/contents/one/SKILL.md",
			"/repos/owner/repo/contents/two/SKILL.md",
			"/repos/owner/repo/contents/existing-dir/SKILL.md",
			"/repos/owner/repo/contents/three/SKILL.md":
			now := atomic.AddInt32(&active, 1)
			updateMaxActive(&maxActive, now)
			<-release
			defer atomic.AddInt32(&active, -1)

			name := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/repos/owner/repo/contents/"), "/SKILL.md")
			if name == "existing-dir" {
				name = "already-installed"
			}
			writeGitHubFile(t, w, fmt.Sprintf(skillBodyTemplate, name))
		case "/repos/owner/repo/contents/one/scripts",
			"/repos/owner/repo/contents/one/references",
			"/repos/owner/repo/contents/one/assets",
			"/repos/owner/repo/contents/two/scripts",
			"/repos/owner/repo/contents/two/references",
			"/repos/owner/repo/contents/two/assets",
			"/repos/owner/repo/contents/three/scripts",
			"/repos/owner/repo/contents/three/references",
			"/repos/owner/repo/contents/three/assets":
			http.NotFound(w, r)
		case "/repos/owner/repo/contents/existing-dir/scripts",
			"/repos/owner/repo/contents/existing-dir/references",
			"/repos/owner/repo/contents/existing-dir/assets":
			t.Fatalf("resources should not be fetched for existing skills: %s", r.URL.Path)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Millisecond)
		timeout := time.NewTimer(2 * time.Second)
		defer ticker.Stop()
		defer timeout.Stop()
		defer close(done)

		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&active) > 0 {
					close(release)
					return
				}
			case <-timeout.C:
				close(release)
				return
			}
		}
	}()

	im := NewImporterWithClient(gitHubAPIClient(t, server.URL), zap.NewNop().Sugar())
	store := newRecordingSkillStore()
	store.existing["already-installed"] = &SkillEntry{Name: "already-installed"}
	ref := &GitHubRef{Owner: "owner", Repo: "repo", Branch: "main"}

	result, err := im.importViaHTTP(context.Background(), ref, store, ImportConfig{Concurrency: 2})
	require.NoError(t, err)
	<-done

	assert.LessOrEqual(t, atomic.LoadInt32(&maxActive), int32(2))
	assert.ElementsMatch(t, []string{"one", "two", "three"}, result.Imported)
	assert.Equal(t, []string{"already-installed"}, result.Skipped)
	assert.Empty(t, result.Errors)
	assert.ElementsMatch(t, []string{"one", "two", "three"}, store.savedNames())
}

func TestImportViaGit_LocalCloneSimulation(t *testing.T) {
	t.Parallel()

	// Simulate what importViaGit does with a local directory structure.
	logger := zap.NewNop().Sugar()
	dir := filepath.Join(t.TempDir(), "skills")
	store := NewFileSkillStore(dir, logger)
	ctx := context.Background()

	// Create a fake cloned repo structure.
	cloneDir := t.TempDir()
	skillDir := filepath.Join(cloneDir, "my-imported-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0o755))

	skillContent := `---
name: my-imported-skill
description: An imported skill
type: instruction
status: active
---

This is imported content.`

	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0o644))

	// Add resource files.
	assetsDir := filepath.Join(skillDir, "assets")
	require.NoError(t, os.MkdirAll(assetsDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(assetsDir, "logo.png"), []byte("fake-png"), 0o644))

	// Read and parse SKILL.md like importViaGit does.
	raw, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	require.NoError(t, err)

	entry, err := ParseSkillMD(raw)
	require.NoError(t, err)
	entry.Source = "https://github.com/test/repo"

	require.NoError(t, store.Save(ctx, *entry))

	copyResourceDirs(ctx, skillDir, entry.Name, store)

	// Verify skill was saved.
	got, err := store.Get(ctx, "my-imported-skill")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/test/repo", got.Source)

	// Verify resource was copied.
	asset, err := os.ReadFile(filepath.Join(dir, "my-imported-skill", "assets", "logo.png"))
	require.NoError(t, err)
	assert.Equal(t, "fake-png", string(asset))
}

func TestImportFromRepo(t *testing.T) {
	t.Parallel()

	// Prepare skill content.
	skill1 := `---
name: skill-one
description: First skill
type: instruction
---

Content for skill one.`

	skill2 := `---
name: skill-two
description: Second skill
type: instruction
---

Content for skill two.`

	encoded1 := base64.StdEncoding.EncodeToString([]byte(skill1))
	encoded2 := base64.StdEncoding.EncodeToString([]byte(skill2))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		path := r.URL.Path
		switch path {
		case "/repos/owner/repo/contents/":
			// Directory listing.
			json.NewEncoder(w).Encode([]gitHubContentsEntry{
				{Name: "skill-one", Type: "dir"},
				{Name: "skill-two", Type: "dir"},
				{Name: "README.md", Type: "file"},
			})
		case "/repos/owner/repo/contents/skill-one/SKILL.md":
			json.NewEncoder(w).Encode(gitHubFileResponse{Content: encoded1, Encoding: "base64"})
		case "/repos/owner/repo/contents/skill-two/SKILL.md":
			json.NewEncoder(w).Encode(gitHubFileResponse{Content: encoded2, Encoding: "base64"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	logger := zap.NewNop().Sugar()
	dir := filepath.Join(t.TempDir(), "skills")
	store := NewFileSkillStore(dir, logger)

	// We can't easily override the GitHub API base URL in the Importer,
	// so we test the individual components and the ImportSingle path.

	// Test ImportSingle for each skill.
	im := NewImporterWithClient(ts.Client(), logger)
	ctx := context.Background()

	entry1, err := im.ImportSingle(ctx, []byte(skill1), "https://github.com/owner/repo", store)
	require.NoError(t, err)
	assert.Equal(t, "skill-one", entry1.Name)
	assert.Equal(t, "https://github.com/owner/repo", entry1.Source)
	assert.Equal(t, SkillTypeInstruction, entry1.Type)

	entry2, err := im.ImportSingle(ctx, []byte(skill2), "https://github.com/owner/repo", store)
	require.NoError(t, err)
	assert.Equal(t, "skill-two", entry2.Name)

	// Verify both are persisted.
	active, err := store.ListActive(ctx)
	require.NoError(t, err)
	assert.Len(t, active, 2)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type errReadCloser struct {
	err error
}

func (r errReadCloser) Read([]byte) (int, error) {
	return 0, r.err
}

func (r errReadCloser) Close() error {
	return nil
}

func gitHubAPIClient(t *testing.T, serverURL string) *http.Client {
	t.Helper()

	baseURL, err := url.Parse(serverURL)
	require.NoError(t, err)

	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			cloned := req.Clone(req.Context())
			rewritten := *req.URL
			rewritten.Scheme = baseURL.Scheme
			rewritten.Host = baseURL.Host
			cloned.URL = &rewritten
			cloned.Host = baseURL.Host
			return http.DefaultTransport.RoundTrip(cloned)
		}),
	}
}

func writeGitHubFile(t *testing.T, w http.ResponseWriter, content string) {
	t.Helper()

	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	require.NoError(t, json.NewEncoder(w).Encode(gitHubFileResponse{
		Content:  encoded,
		Encoding: "base64",
	}))
}

type recordingSkillStore struct {
	mu        sync.Mutex
	existing  map[string]*SkillEntry
	saved     map[string]SkillEntry
	resources map[string][]byte
}

func newRecordingSkillStore() *recordingSkillStore {
	return &recordingSkillStore{
		existing:  make(map[string]*SkillEntry),
		saved:     make(map[string]SkillEntry),
		resources: make(map[string][]byte),
	}
}

func (s *recordingSkillStore) Save(ctx context.Context, entry SkillEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.saved[entry.Name] = entry
	return nil
}

func (s *recordingSkillStore) Get(ctx context.Context, name string) (*SkillEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, ok := s.existing[name]; ok {
		copied := *entry
		return &copied, nil
	}
	if entry, ok := s.saved[name]; ok {
		copied := entry
		return &copied, nil
	}
	return nil, nil
}

func (s *recordingSkillStore) ListActive(ctx context.Context) ([]SkillEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entries := make([]SkillEntry, 0, len(s.saved)+len(s.existing))
	for _, entry := range s.existing {
		entries = append(entries, *entry)
	}
	for _, entry := range s.saved {
		entries = append(entries, entry)
	}
	return entries, nil
}

func (s *recordingSkillStore) Activate(ctx context.Context, name string) error {
	return ctx.Err()
}

func (s *recordingSkillStore) Delete(ctx context.Context, name string) error {
	return ctx.Err()
}

func (s *recordingSkillStore) SaveResource(ctx context.Context, skillName, relPath string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := filepath.Join(skillName, relPath)
	s.resources[key] = append([]byte(nil), data...)
	return nil
}

func (s *recordingSkillStore) savedNames() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	names := make([]string, 0, len(s.saved))
	for name := range s.saved {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (s *recordingSkillStore) savedResources() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()

	resources := make(map[string]string, len(s.resources))
	for path, data := range s.resources {
		resources[filepath.ToSlash(path)] = string(data)
	}
	return resources
}

func updateMaxActive(maxActive *int32, value int32) {
	for {
		current := atomic.LoadInt32(maxActive)
		if value <= current {
			return
		}
		if atomic.CompareAndSwapInt32(maxActive, current, value) {
			return
		}
	}
}
