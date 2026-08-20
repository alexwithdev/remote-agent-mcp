package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEditFileSingle(t *testing.T) {
	p := writeTemp(t, "hello world\n")
	res, err := EditFile(p, "world", "there", false)
	if err != nil {
		t.Fatalf("EditFile: %v", err)
	}
	if res.Occurrences != 1 {
		t.Errorf("Occurrences = %d, want 1", res.Occurrences)
	}
	got, _ := os.ReadFile(p)
	if string(got) != "hello there\n" {
		t.Errorf("file = %q", string(got))
	}
}

func TestEditFileReplaceAll(t *testing.T) {
	p := writeTemp(t, "foo foo foo\n")
	res, err := EditFile(p, "foo", "bar", true)
	if err != nil {
		t.Fatalf("EditFile: %v", err)
	}
	if res.Occurrences != 3 {
		t.Errorf("Occurrences = %d, want 3", res.Occurrences)
	}
	got, _ := os.ReadFile(p)
	if string(got) != "bar bar bar\n" {
		t.Errorf("file = %q", string(got))
	}
}

func TestEditFileNotFound(t *testing.T) {
	p := writeTemp(t, "nothing here\n")
	if _, err := EditFile(p, "absent", "x", false); err == nil {
		t.Error("missing oldString should error")
	}
}

func TestEditFileMultipleWithoutReplaceAll(t *testing.T) {
	p := writeTemp(t, "dup dup\n")
	if _, err := EditFile(p, "dup", "x", false); err == nil {
		t.Error("multiple matches without replaceAll should error")
	}
}

func TestEditFileEmptyOldString(t *testing.T) {
	p := writeTemp(t, "x\n")
	if _, err := EditFile(p, "", "x", false); err == nil {
		t.Error("empty oldString should error")
	}
}

func TestEditFilePreservesContent(t *testing.T) {
	p := writeTemp(t, "aaa bbb ccc\n")
	res, err := EditFile(p, "bbb", "BBB", false)
	if err != nil {
		t.Fatalf("EditFile: %v", err)
	}
	if res.Path == "" {
		t.Error("Path should be set")
	}
	got, _ := os.ReadFile(p)
	if string(got) != "aaa BBB ccc\n" {
		t.Errorf("file = %q", string(got))
	}
}

func TestEditFileMissingFile(t *testing.T) {
	if _, err := EditFile(filepath.Join(t.TempDir(), "nope"), "a", "b", false); err == nil {
		t.Error("missing file should error")
	}
}
