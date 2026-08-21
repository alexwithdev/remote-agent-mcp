// Command remote-agent-mcp is an MCP server that exposes file and shell tools
// over Streamable HTTP for remote machine troubleshooting.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"remote-agent-mcp/internal/config"
	"remote-agent-mcp/internal/logging"
	"remote-agent-mcp/internal/server"
	"remote-agent-mcp/internal/tools"
)

func main() {
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}

	logger, closer, err := logging.New(logging.Config{
		Level:      cfg.LogLevel,
		File:       cfg.LogFile,
		MaxSize:    cfg.LogMaxSize,
		MaxBackups: cfg.LogMaxBackups,
	})
	if err != nil {
		slog.Error("logging", "error", err)
		os.Exit(1)
	}
	defer closer.Close()

	shell := tools.ResolveShell(cfg.Shell)
	if shell != cfg.Shell {
		logger.Info("shell not found, falling back", "requested", cfg.Shell, "using", shell)
	}

	ts := tools.NewToolSet(tools.Options{
		Root:             cfg.Root,
		Shell:            shell,
		AllowAll:         cfg.AllowAll,
		AllowUnconfirmed: cfg.AllowUnconfirmed,
		Logger:           logger,
	})

	s := server.New(ts)
	h := server.Handler(s, cfg.Token)

	logger.Info("listening",
		"addr", cfg.Addr,
		"root", cfg.Root,
		"allowAll", cfg.AllowAll,
		"allowUnconfirmed", cfg.AllowUnconfirmed,
		"token", cfg.Token != "",
	)
	if err := http.ListenAndServe(cfg.Addr, h); err != nil {
		logger.Error("server", "error", err)
		os.Exit(1)
	}
}
