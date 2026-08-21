package logging

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewStderrOnly(t *testing.T) {
	var buf bytes.Buffer
	logger, closer, err := New(Config{
		Level:  "info",
		Stderr: &buf,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer closer.Close()

	logger.Info("hello")
	logger.Debug("should be filtered")

	out := buf.String()
	if !strings.Contains(out, "hello") {
		t.Errorf("stderr output missing info message: %q", out)
	}
	if strings.Contains(out, "should be filtered") {
		t.Errorf("stderr output should not contain debug message at info level: %q", out)
	}
}

func TestNewLevelFiltering(t *testing.T) {
	cases := []struct {
		level    string
		emit     func(*slog.Logger)
		expected string
	}{
		{"debug", func(l *slog.Logger) { l.Debug("dbg") }, "dbg"},
		{"info", func(l *slog.Logger) { l.Debug("dbg"); l.Info("inf") }, "inf"},
		{"warn", func(l *slog.Logger) { l.Info("inf"); l.Warn("wrn") }, "wrn"},
		{"error", func(l *slog.Logger) { l.Warn("wrn"); l.Error("err") }, "err"},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		logger, closer, err := New(Config{Level: c.level, Stderr: &buf})
		if err != nil {
			t.Fatalf("New(%q): %v", c.level, err)
		}
		c.emit(logger)
		closer.Close()
		out := buf.String()
		if !strings.Contains(out, c.expected) {
			t.Errorf("level %q: output %q missing %q", c.level, out, c.expected)
		}
	}
}

func TestNewInvalidLevel(t *testing.T) {
	_, _, err := New(Config{Level: "bogus", Stderr: &bytes.Buffer{}})
	if err == nil {
		t.Fatal("New with invalid level should error")
	}
}

func TestNewDualWrite(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "app.log")
	var buf bytes.Buffer
	logger, closer, err := New(Config{
		Level:   "info",
		File:    logFile,
		Stderr:  &buf,
		MaxSize: 10,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	logger.Info("dual write message")
	closer.Close()

	if !strings.Contains(buf.String(), "dual write message") {
		t.Errorf("stderr missing message: %q", buf.String())
	}
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("reading log file: %v", err)
	}
	if !strings.Contains(string(data), "dual write message") {
		t.Errorf("log file missing message: %q", string(data))
	}
}
