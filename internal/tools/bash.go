package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultTimeout = 30 * time.Second
	maxTimeout     = 5 * time.Minute
	maxOutputBytes = 50 * 1024
)

type bashArgs struct {
	Command string `json:"command" jsonschema:"Shell command to execute"`
	Cwd     string `json:"cwd,omitempty" jsonschema:"Working directory, relative to root (default: root)"`
	Timeout int    `json:"timeout,omitempty" jsonschema:"Timeout in seconds (default 30, max 300)"`
}

// BashResult is the structured output of the bash tool.
type BashResult struct {
	Stdout          string `json:"stdout" jsonschema:"Standard output"`
	Stderr          string `json:"stderr" jsonschema:"Standard error"`
	ExitCode        int    `json:"exitCode" jsonschema:"Process exit code"`
	StdoutTruncated bool   `json:"stdoutTruncated,omitempty" jsonschema:"True when stdout was truncated"`
	StderrTruncated bool   `json:"stderrTruncated,omitempty" jsonschema:"True when stderr was truncated"`
}

// dangerousPatterns match shell commands that warrant human confirmation.
// They are applied case-insensitively.
var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\brm\s+.*-[a-z]*[rf][a-z]*\b`), // rm -r / -f / -rf
	regexp.MustCompile(`\brm\s+.*(--recursive|--force)`),
	regexp.MustCompile(`\bdd\b`),
	regexp.MustCompile(`\bmkfs\b`),
	regexp.MustCompile(`\bfdisk\b`),
	regexp.MustCompile(`\bshutdown\b`),
	regexp.MustCompile(`\breboot\b`),
	regexp.MustCompile(`\bhalt\b`),
	regexp.MustCompile(`\bpoweroff\b`),
	regexp.MustCompile(`\bsudo\b`),
	regexp.MustCompile(`\bchmod\b`),
	regexp.MustCompile(`\bchown\b`),
	regexp.MustCompile(`\s>\s*[^>&|;]`), // overwrite redirection
}

// IsDangerousCommand reports whether cmd matches any dangerous pattern.
func IsDangerousCommand(cmd string) bool {
	lower := strings.ToLower(cmd)
	for _, re := range dangerousPatterns {
		if re.MatchString(lower) {
			return true
		}
	}
	return false
}

// Truncate truncates s to at most maxBytes, trimming to a UTF-8 rune boundary.
// It reports whether truncation occurred.
func Truncate(s string, maxBytes int) (string, bool) {
	if len(s) <= maxBytes {
		return s, false
	}
	cut := s[:maxBytes]
	for len(cut) > 0 && !utf8.ValidString(cut) {
		cut = cut[:len(cut)-1]
	}
	return cut, true
}

// ResolveShell returns the preferred shell if it is executable, otherwise
// falls back to /bin/sh. The final value is returned unchanged if neither is
// found, so the caller surfaces a clear error at execution time.
func ResolveShell(preferred string) string {
	if preferred == "" {
		preferred = "/bin/bash"
	}
	if _, err := exec.LookPath(preferred); err == nil {
		return preferred
	}
	if _, err := exec.LookPath("/bin/sh"); err == nil {
		return "/bin/sh"
	}
	return preferred
}

// RunCommand executes command via the given shell with -c and a timeout,
// capturing truncated stdout/stderr and the exit code.
func RunCommand(ctx context.Context, shell, cwd, command string, timeout time.Duration) (BashResult, error) {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if timeout > maxTimeout {
		timeout = maxTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, shell, "-c", command)
	cmd.Dir = cwd
	// WaitDelay bounds how long Wait waits for orphaned subprocesses to close
	// the I/O pipes after the context is canceled. Without it, a shell that
	// forks a child (e.g. dash's `sh -c "sleep 30"`) keeps the pipes open until
	// the child exits, so a timed-out command can hang for the child's full
	// lifetime instead of returning promptly.
	cmd.WaitDelay = 1 * time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return BashResult{}, fmt.Errorf("command timed out after %s", timeout)
	}
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return BashResult{}, fmt.Errorf("running command: %w", err)
		}
	}
	out, outTrunc := Truncate(stdout.String(), maxOutputBytes)
	errOut, errTrunc := Truncate(stderr.String(), maxOutputBytes)
	return BashResult{
		Stdout:          out,
		Stderr:          errOut,
		ExitCode:        exitCode,
		StdoutTruncated: outTrunc,
		StderrTruncated: errTrunc,
	}, nil
}

func (t *ToolSet) registerBash(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "bash",
		Description: "Run a shell command in the working directory (tests, installs, lint, etc.).",
	}, wrapTool(t.logger, "bash", t.bash))
}

func (t *ToolSet) bash(ctx context.Context, req *mcp.CallToolRequest, args bashArgs) (*mcp.CallToolResult, BashResult, error) {
	if IsDangerousCommand(args.Command) {
		if err := Confirm(ctx, req.Session, fmt.Sprintf("Allow this potentially dangerous command?\n\n%s", args.Command), t.opts.AllowUnconfirmed); err != nil {
			return nil, BashResult{}, err
		}
	}

	cwd := t.opts.Root
	if args.Cwd != "" {
		var err error
		cwd, err = ResolvePath(t.opts.Root, args.Cwd, t.opts.AllowAll)
		if err != nil {
			return nil, BashResult{}, err
		}
	}

	res, err := RunCommand(ctx, t.opts.Shell, cwd, args.Command, time.Duration(args.Timeout)*time.Second)
	if err != nil {
		return nil, BashResult{}, err
	}
	return nil, res, nil
}
