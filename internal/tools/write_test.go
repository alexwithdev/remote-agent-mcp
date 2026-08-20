package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileCreatesParentDirs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b", "c")
	p := filepath.Join(dir, "f.txt")
	res, err := WriteFile(p, "hello")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if !res.Created {
		t.Error("Created should be true for a new file")
	}
	if res.BytesWritten != 5 {
		t.Errorf("BytesWritten = %d, want 5", res.BytesWritten)
	}
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("content = %q, want hello", string(got))
	}
}

func TestWriteFileOverwritePreservesMode(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.txt")
	if err := os.WriteFile(p, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	res, err := WriteFile(p, "new-content")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if res.Created {
		t.Error("Created should be false when overwriting")
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v, want 0600 preserved", info.Mode().Perm())
	}
}

func TestWriteFileEmptyContent(t *testing.T) {
	p := filepath.Join(t.TempDir(), "empty.txt")
	res, err := WriteFile(p, "")
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if res.BytesWritten != 0 {
		t.Errorf("BytesWritten = %d, want 0", res.BytesWritten)
	}
}
