package tools

import (
	"context"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestIsDangerousCommand(t *testing.T) {
	dangerous := []string{
		"rm -rf /tmp/x",
		"rm -fr /var/log",
		"sudo rm -r /etc",
		"dd if=/dev/zero of=/dev/sda",
		"mkfs.ext4 /dev/sdb1",
		"fdisk /dev/sda",
		"shutdown -h now",
		"reboot",
		"chmod 777 /",
		"chown root /etc/passwd",
		"echo hi > /etc/hosts",
	}
	for _, c := range dangerous {
		if !IsDangerousCommand(c) {
			t.Errorf("IsDangerousCommand(%q) = false, want true", c)
		}
	}

	safe := []string{
		"ls -la",
		"cat /etc/hosts",
		"echo hello",
		"grep error app.log",
		"pwd",
		"go test ./...",
	}
	for _, c := range safe {
		if IsDangerousCommand(c) {
			t.Errorf("IsDangerousCommand(%q) = true, want false", c)
		}
	}
}

func TestTruncate(t *testing.T) {
	if got, trunc := Truncate("short", 100); trunc || got != "short" {
		t.Errorf("got %q, trunc %v", got, trunc)
	}
	got, trunc := Truncate("abcdefghij", 5)
	if !trunc || got != "abcde" {
		t.Errorf("got %q, trunc %v", got, trunc)
	}
}

func TestTruncateUTF8Boundary(t *testing.T) {
	// "héllo" is 6 bytes; truncating at 4 would split "é" (2 bytes).
	s := "héllo"
	got, trunc := Truncate(s, 4)
	if !trunc {
		t.Fatal("expected truncation")
	}
	if !strings.HasPrefix(s, got) {
		t.Errorf("got %q is not a prefix of %q", got, s)
	}
	// The result must be valid UTF-8.
	if !utf8.ValidString(got) {
		t.Errorf("got %q is not valid UTF-8", got)
	}
}

func TestRunCommand(t *testing.T) {
	res, err := RunCommand(context.Background(), "/bin/sh", t.TempDir(), "echo hello", time.Second)
	if err != nil {
		t.Fatalf("RunCommand: %v", err)
	}
	if res.Stdout != "hello\n" {
		t.Errorf("Stdout = %q, want \"hello\\n\"", res.Stdout)
	}
	if res.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", res.ExitCode)
	}
}

func TestRunCommandExitCode(t *testing.T) {
	res, err := RunCommand(context.Background(), "/bin/sh", t.TempDir(), "exit 7", time.Second)
	if err != nil {
		t.Fatalf("RunCommand: %v", err)
	}
	if res.ExitCode != 7 {
		t.Errorf("ExitCode = %d, want 7", res.ExitCode)
	}
}

func TestRunCommandOutputTruncation(t *testing.T) {
	// Generate more than 50KB of output.
	res, err := RunCommand(context.Background(), "/bin/sh", t.TempDir(), "yes x | head -c 100000", time.Second)
	if err != nil {
		t.Fatalf("RunCommand: %v", err)
	}
	if !res.StdoutTruncated {
		t.Error("expected stdout to be truncated")
	}
	if len(res.Stdout) != maxOutputBytes {
		t.Errorf("Stdout len = %d, want %d", len(res.Stdout), maxOutputBytes)
	}
}

func TestRunCommandTimeout(t *testing.T) {
	start := time.Now()
	_, err := RunCommand(context.Background(), "/bin/sh", t.TempDir(), "sleep 30", 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if time.Since(start) > 5*time.Second {
		t.Error("timeout did not take effect promptly")
	}
}

func TestResolveShellFallback(t *testing.T) {
	got := ResolveShell("/bin/bash")
	if got == "" {
		t.Fatal("ResolveShell returned empty")
	}
	// On Alpine (no bash) it should fall back to /bin/sh; elsewhere bash itself.
	if got != "/bin/bash" && got != "/bin/sh" {
		t.Errorf("ResolveShell = %q, want /bin/bash or /bin/sh", got)
	}
}
