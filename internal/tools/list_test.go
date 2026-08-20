package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListDir(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "a.txt"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(root, ".hidden"), []byte("h"), 0o644)
	os.Mkdir(filepath.Join(root, "sub"), 0o755)

	entries, err := ListDir(root)
	if err != nil {
		t.Fatalf("ListDir: %v", err)
	}

	names := map[string]string{}
	for _, e := range entries {
		names[e.Name] = e.Type
	}
	if names["a.txt"] != "file" {
		t.Errorf("a.txt type = %q, want file", names["a.txt"])
	}
	if names[".hidden"] != "file" {
		t.Error("hidden file .hidden should be listed")
	}
	if names["sub"] != "dir" {
		t.Errorf("sub type = %q, want dir", names["sub"])
	}
	if len(entries) != 3 {
		t.Errorf("entries = %d, want 3", len(entries))
	}
}

func TestListDirMissing(t *testing.T) {
	if _, err := ListDir(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Error("missing directory should error")
	}
}

func TestListDirEmpty(t *testing.T) {
	entries, err := ListDir(t.TempDir())
	if err != nil {
		t.Fatalf("ListDir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("entries = %d, want 0", len(entries))
	}
}
