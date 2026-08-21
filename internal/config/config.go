// Package config loads runtime configuration from environment variables and
// command-line flags. Flags take precedence over environment variables, which
// take precedence over defaults.
package config

import (
	"flag"
	"os"
	"strconv"
)

const defaultAddr = "0.0.0.0:9090"

// Config holds all runtime configuration for the MCP server.
type Config struct {
	// Addr is the listen address for the Streamable HTTP server.
	Addr string
	// Root is the root directory that file tools are restricted to.
	Root string
	// Shell is the shell used by the bash tool (default /bin/bash, with
	// automatic fallback to /bin/sh when absent).
	Shell string
	// Token, when non-empty, requires a matching bearer token on every request.
	Token string
	// AllowAll permits full filesystem access outside Root.
	AllowAll bool
	// AllowUnconfirmed permits destructive actions without elicitation
	// confirmation (used when the client does not support elicitation).
	AllowUnconfirmed bool
	// LogFile is the path to the log file. When empty, logs go only to stderr.
	LogFile string
	// LogLevel is the minimum log level to emit (debug/info/warn/error).
	LogLevel string
	// LogMaxSize is the maximum size of a single log file in megabytes before
	// rotation.
	LogMaxSize int
	// LogMaxBackups is the maximum number of rotated log files to retain.
	LogMaxBackups int
}

// Load reads configuration from environment variables and flags, returning the
// merged result.
func Load(args []string) (*Config, error) {
	cfg := &Config{
		Addr:          defaultAddr,
		Root:          defaultRoot(),
		Shell:         "/bin/bash",
		LogLevel:      "info",
		LogMaxSize:    10,
		LogMaxBackups: 5,
	}
	applyEnv(cfg)

	fs := flag.NewFlagSet("remote-agent-mcp", flag.ContinueOnError)
	fs.StringVar(&cfg.Addr, "addr", cfg.Addr, "listen address (env MCP_ADDR)")
	fs.StringVar(&cfg.Root, "root", cfg.Root, "root directory for file access (env MCP_ROOT)")
	fs.StringVar(&cfg.Shell, "shell", cfg.Shell, "shell for the bash tool (env MCP_SHELL)")
	fs.StringVar(&cfg.Token, "token", cfg.Token, "optional bearer token (env MCP_TOKEN)")
	fs.BoolVar(&cfg.AllowAll, "allow-all", cfg.AllowAll, "allow full filesystem access (env MCP_ALLOW_ALL)")
	fs.BoolVar(&cfg.AllowUnconfirmed, "allow-unconfirmed", cfg.AllowUnconfirmed, "allow destructive actions without confirmation (env MCP_ALLOW_UNCONFIRMED)")
	fs.StringVar(&cfg.LogFile, "log-file", cfg.LogFile, "log file path; empty means stderr only (env MCP_LOG_FILE)")
	fs.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "log level: debug, info, warn, error (env MCP_LOG_LEVEL)")
	fs.IntVar(&cfg.LogMaxSize, "log-max-size", cfg.LogMaxSize, "max log file size in MB before rotation (env MCP_LOG_MAX_SIZE)")
	fs.IntVar(&cfg.LogMaxBackups, "log-max-backups", cfg.LogMaxBackups, "max rotated log files to retain (env MCP_LOG_MAX_BACKUPS)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return cfg, nil
}

func defaultRoot() string {
	d, err := os.Getwd()
	if err != nil {
		return "."
	}
	return d
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("MCP_ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := os.Getenv("MCP_ROOT"); v != "" {
		cfg.Root = v
	}
	if v := os.Getenv("MCP_SHELL"); v != "" {
		cfg.Shell = v
	}
	if v := os.Getenv("MCP_TOKEN"); v != "" {
		cfg.Token = v
	}
	if v := os.Getenv("MCP_ALLOW_ALL"); v != "" {
		cfg.AllowAll = parseBool(v)
	}
	if v := os.Getenv("MCP_ALLOW_UNCONFIRMED"); v != "" {
		cfg.AllowUnconfirmed = parseBool(v)
	}
	if v := os.Getenv("MCP_LOG_FILE"); v != "" {
		cfg.LogFile = v
	}
	if v := os.Getenv("MCP_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("MCP_LOG_MAX_SIZE"); v != "" {
		cfg.LogMaxSize = parseInt(v)
	}
	if v := os.Getenv("MCP_LOG_MAX_BACKUPS"); v != "" {
		cfg.LogMaxBackups = parseInt(v)
	}
}

func parseInt(v string) int {
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}

func parseBool(v string) bool {
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false
	}
	return b
}
