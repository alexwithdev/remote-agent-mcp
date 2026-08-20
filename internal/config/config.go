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
}

// Load reads configuration from environment variables and flags, returning the
// merged result.
func Load(args []string) (*Config, error) {
	cfg := &Config{
		Addr:  defaultAddr,
		Root:  defaultRoot(),
		Shell: "/bin/bash",
	}
	applyEnv(cfg)

	fs := flag.NewFlagSet("remote-agent-mcp", flag.ContinueOnError)
	fs.StringVar(&cfg.Addr, "addr", cfg.Addr, "listen address (env MCP_ADDR)")
	fs.StringVar(&cfg.Root, "root", cfg.Root, "root directory for file access (env MCP_ROOT)")
	fs.StringVar(&cfg.Shell, "shell", cfg.Shell, "shell for the bash tool (env MCP_SHELL)")
	fs.StringVar(&cfg.Token, "token", cfg.Token, "optional bearer token (env MCP_TOKEN)")
	fs.BoolVar(&cfg.AllowAll, "allow-all", cfg.AllowAll, "allow full filesystem access (env MCP_ALLOW_ALL)")
	fs.BoolVar(&cfg.AllowUnconfirmed, "allow-unconfirmed", cfg.AllowUnconfirmed, "allow destructive actions without confirmation (env MCP_ALLOW_UNCONFIRMED)")

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
}

func parseBool(v string) bool {
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false
	}
	return b
}
