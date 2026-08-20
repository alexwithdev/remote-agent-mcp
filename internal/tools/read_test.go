package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "f.txt")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTemp: %v", err)
	}
	return p
}

func TestReadFileFull(t *testing.T) {
	p := writeTemp(t, "line1\nline2\nline3\n")
	res, err := ReadFile(p, 1, 2000)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if res.Content != "line1\nline2\nline3" {
		t.Errorf("Content = %q", res.Content)
	}
	if res.TotalLines != 3 {
		t.Errorf("TotalLines = %d, want 3", res.TotalLines)
	}
	if res.TotalBytes != 18 {
		t.Errorf("TotalBytes = %d, want 18", res.TotalBytes)
	}
	if res.Truncated {
		t.Error("Truncated should be false")
	}
}

func TestReadFileOffsetLimit(t *testing.T) {
	p := writeTemp(t, "1\n2\n3\n4\n5\n")
	res, err := ReadFile(p, 2, 2)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if res.Content != "2\n3" {
		t.Errorf("Content = %q, want \"2\\n3\"", res.Content)
	}
	if res.TotalLines != 5 {
		t.Errorf("TotalLines = %d, want 5", res.TotalLines)
	}
}

func TestReadFileTruncation(t *testing.T) {
	// 2500 lines: exceeds the 2000 line limit.
	var b strings.Builder
	for i := 0; i < 2500; i++ {
		b.WriteString("x\n")
	}
	p := writeTemp(t, b.String())
	res, err := ReadFile(p, 1, 2000)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !res.Truncated {
		t.Error("expected truncated to be true with 2500 lines")
	}
	if res.TotalLines != 2500 {
		t.Errorf("TotalLines = %d, want 2500", res.TotalLines)
	}
}

func TestReadFilePastEOF(t *testing.T) {
	p := writeTemp(t, "a\nb\n")
	res, err := ReadFile(p, 100, 10)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if res.Content != "" {
		t.Errorf("Content = %q, want empty past EOF", res.Content)
	}
}

func TestReadFileMissing(t *testing.T) {
	if _, err := ReadFile(filepath.Join(t.TempDir(), "nope"), 1, 10); err == nil {
		t.Error("missing file should error")
	}
}

func TestSplitLines(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"a", 1},
		{"a\nb", 2},
		{"a\nb\n", 2},
		{"a\nb\nc\n", 3},
	}
	for _, c := range cases {
		if got := len(splitLines(c.in)); got != c.want {
			t.Errorf("splitLines(%q) len = %d, want %d", c.in, got, c.want)
		}
	}
}
