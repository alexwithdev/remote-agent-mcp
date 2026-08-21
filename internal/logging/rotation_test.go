package logging

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRotation verifies that writing past the size limit rotates the log file,
// retains a bounded number of backups, and compresses them.
func TestRotation(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "app.log")

	// MaxSize is in MB; use 1MB so a single large write triggers rotation.
	logger, closer, err := New(Config{
		Level:      "debug",
		File:       logFile,
		Stderr:     &bytes.Buffer{},
		MaxSize:    1,
		MaxBackups: 2,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Write enough data to force several rotations.
	big := strings.Repeat("x", 256*1024) // 256KB per message
	for i := 0; i < 20; i++ {
		logger.Info("bulk", "chunk", i, "data", big)
	}
	closer.Close()

	// The current log file must exist.
	if _, err := os.Stat(logFile); err != nil {
		t.Fatalf("current log file missing: %v", err)
	}

	// lumberjack compresses and prunes backups asynchronously in a background
	// goroutine, so poll until the mill settles.
	var backups []string
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("ReadDir: %v", err)
		}
		backups = backups[:0]
		for _, e := range entries {
			if e.Name() != "app.log" {
				backups = append(backups, e.Name())
			}
		}
		if len(backups) > 0 && len(backups) <= 2 {
			allCompressed := true
			for _, b := range backups {
				if !strings.HasSuffix(b, ".gz") {
					allCompressed = false
					break
				}
			}
			if allCompressed {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	if len(backups) == 0 {
		t.Fatal("expected at least one rotated backup file")
	}
	for _, b := range backups {
		if !strings.HasSuffix(b, ".gz") {
			t.Errorf("backup %q should be gzip-compressed (.gz)", b)
		}
	}
	// MaxBackups=2 bounds the number of retained backups.
	if len(backups) > 2 {
		t.Errorf("retained %d backups, want at most 2", len(backups))
	}
}
