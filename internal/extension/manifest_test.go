package extension

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validManifest = `
schema: lango.extension/v1
name: python-dev
version: 0.1.0
description: Python dev pack
author: langoai
license: Apache-2.0
homepage: https://example.com
contents:
  skills:
    - name: pytest-refactor
      path: skills/pytest-refactor/SKILL.md
  modes:
    - name: python-review
      systemHint: Focus on Python idioms.
      tools: ["@python"]
      skills: [pytest-refactor]
  prompts:
    - path: prompts/python.md
      section: python
`

func TestParseManifest_HappyPath(t *testing.T) {
	t.Parallel()

	m, err := ParseManifest(strings.NewReader(validManifest))
	require.NoError(t, err)
	assert.Equal(t, SchemaV1, m.Schema)
	assert.Equal(t, "python-dev", m.Name)
	assert.Equal(t, "0.1.0", m.Version)
	assert.Equal(t, "Apache-2.0", m.License)
	require.Len(t, m.Contents.Skills, 1)
	assert.Equal(t, "pytest-refactor", m.Contents.Skills[0].Name)
	require.Len(t, m.Contents.Modes, 1)
	require.Len(t, m.Contents.Prompts, 1)
	assert.Equal(t, "python", m.Contents.Prompts[0].Section)
}

func TestParseManifest_UnknownContentsKeyRejected(t *testing.T) {
	t.Parallel()

	bad := strings.Replace(validManifest, "contents:\n  skills:", "contents:\n  tools:\n    - name: x\n  skills:", 1)
	_, err := ParseManifest(strings.NewReader(bad))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown field")
}

func TestParseManifest_UnknownTopLevelKeyRejected(t *testing.T) {
	t.Parallel()

	bad := validManifest + "\nsignature: abc\n"
	_, err := ParseManifest(strings.NewReader(bad))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown field")
}

func TestParseManifest_FutureSchemaRejected(t *testing.T) {
	t.Parallel()

	bad := strings.Replace(validManifest, "lango.extension/v1", "lango.extension/v2", 1)
	_, err := ParseManifest(strings.NewReader(bad))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSchemaMismatch)
	assert.Contains(t, err.Error(), "upgrade lango")
}

func TestParseManifest_MissingSchemaRejected(t *testing.T) {
	t.Parallel()

	bad := strings.Replace(validManifest, "schema: lango.extension/v1\n", "", 1)
	_, err := ParseManifest(strings.NewReader(bad))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing schema")
}

func TestValidate_InvalidName(t *testing.T) {
	t.Parallel()

	cases := []string{"Python_Dev", "UPPER", "x", "-leading", "trailing-", "has spaces"}
	for _, name := range cases {
		m := &Manifest{
			Schema: SchemaV1, Name: name, Version: "0.1.0", Description: "x",
		}
		err := m.Validate()
		assert.Error(t, err, "name %q should be rejected", name)
	}
}

func TestValidate_InvalidVersion(t *testing.T) {
	t.Parallel()

	cases := []string{"0.1", "1", "v1.0.0", "1.0.0.0", "latest"}
	for _, v := range cases {
		m := &Manifest{
			Schema: SchemaV1, Name: "foo", Version: v, Description: "x",
		}
		err := m.Validate()
		assert.Error(t, err, "version %q should be rejected", v)
	}
}

func TestValidateContentPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		path  string
		valid bool
	}{
		{"empty", "", false},
		{"absolute unix", "/etc/passwd", false},
		{"absolute windows", "\\Windows", false},
		{"parent traversal", "../outside.md", false},
		{"nested parent traversal", "skills/../../outside.md", false},
		{"clean nested", "skills/foo/SKILL.md", true},
		{"dot prefix", "./SKILL.md", true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContentPath(tt.path)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestResolvePath_ContainmentEnforced(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "skills"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "skills", "SKILL.md"), []byte("x"), 0o644))

	// Happy path.
	resolved, err := ResolvePath(root, "skills/SKILL.md")
	require.NoError(t, err)
	evalRoot, _ := filepath.EvalSymlinks(root)
	assert.True(t, strings.HasPrefix(resolved, evalRoot), "resolved=%s evalRoot=%s", resolved, evalRoot)
}

func TestResolvePath_SymlinkEscapeRejected(t *testing.T) {
	t.Parallel()

	outside := t.TempDir()
	root := t.TempDir()
	target := filepath.Join(outside, "secret.md")
	require.NoError(t, os.WriteFile(target, []byte("secret"), 0o644))
	link := filepath.Join(root, "escape.md")
	require.NoError(t, os.Symlink(target, link))

	_, err := ResolvePath(root, "escape.md")
	require.Error(t, err, "symlink escape must be rejected")
	assert.Contains(t, err.Error(), "escapes pack root")
}

func TestValidate_DuplicateSkillName(t *testing.T) {
	t.Parallel()

	m := &Manifest{
		Schema:      SchemaV1,
		Name:        "pack",
		Version:     "0.1.0",
		Description: "x",
		Contents: Contents{
			Skills: []SkillRef{
				{Name: "foo", Path: "a/SKILL.md"},
				{Name: "foo", Path: "b/SKILL.md"},
			},
		},
	}
	err := m.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate skill name")
}

func TestValidate_DuplicateModeName(t *testing.T) {
	t.Parallel()

	m := &Manifest{
		Schema: SchemaV1, Name: "pack", Version: "0.1.0", Description: "x",
		Contents: Contents{
			Modes: []ModeRef{{Name: "alpha"}, {Name: "alpha"}},
		},
	}
	err := m.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate mode name")
}

func TestValidate_InvalidPromptPath(t *testing.T) {
	t.Parallel()

	m := &Manifest{
		Schema: SchemaV1, Name: "pack", Version: "0.1.0", Description: "x",
		Contents: Contents{
			Prompts: []PromptRef{{Path: "../evil.md"}},
		},
	}
	err := m.Validate()
	require.Error(t, err)
	var e *os.PathError
	// not a PathError — just sanity that it's a validation error.
	assert.False(t, errors.As(err, &e))
	assert.Contains(t, err.Error(), "parent-directory")
}
