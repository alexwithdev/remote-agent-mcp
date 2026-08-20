package tools

import (
	"path/filepath"
	"testing"
)

func TestResolvePathRelative(t *testing.T) {
	root := t.TempDir()
	got, err := ResolvePath(root, "a/b.txt", false)
	if err != nil {
		t.Fatalf("ResolvePath: %v", err)
	}
	want := filepath.Join(root, "a/b.txt")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolvePathAbsoluteWithinRoot(t *testing.T) {
	root := t.TempDir()
	abs := filepath.Join(root, "x.txt")
	got, err := ResolvePath(root, abs, false)
	if err != nil {
		t.Fatalf("ResolvePath: %v", err)
	}
	if got != abs {
		t.Errorf("got %q, want %q", got, abs)
	}
}

func TestResolvePathEscapeRejected(t *testing.T) {
	root := t.TempDir()
	for _, p := range []string{"../x.txt", "../../etc/passwd", filepath.Join(root, "..", "secret")} {
		if _, err := ResolvePath(root, p, false); err == nil {
			t.Errorf("ResolvePath(%q) should fail, got nil error", p)
		}
	}
}

func TestResolvePathEmptyRejected(t *testing.T) {
	if _, err := ResolvePath(t.TempDir(), "", false); err == nil {
		t.Error("empty path should fail")
	}
}

func TestResolvePathAllowAll(t *testing.T) {
	root := t.TempDir()
	abs := "/etc/passwd"
	got, err := ResolvePath(root, abs, true)
	if err != nil {
		t.Fatalf("ResolvePath allowAll: %v", err)
	}
	if got != abs {
		t.Errorf("got %q, want %q", got, abs)
	}
}
