package skills

import (
	"io/fs"
	"testing"
)

func TestDefaultFSContainsEmbeddedSkillFiles(t *testing.T) {
	fsys, err := DefaultFS()
	if err != nil {
		t.Fatalf("DefaultFS returned error: %v", err)
	}
	if fsys == nil {
		t.Fatal("DefaultFS returned nil filesystem")
	}

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		t.Fatalf("read embedded root: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("embedded root has no entries")
	}

	data, err := fs.ReadFile(fsys, ".placeholder/SKILL.md")
	if err != nil {
		t.Fatalf("read embedded SKILL.md: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("embedded SKILL.md is empty")
	}
	if _, err := fs.Stat(fsys, ".placeholder"); err != nil {
		t.Fatalf("stat embedded placeholder directory: %v", err)
	}
}
