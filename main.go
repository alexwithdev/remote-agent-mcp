// Command remote-agent-mcp is an MCP server that exposes file and shell tools
// over Streamable HTTP for remote machine troubleshooting.
package main

import (
	"log"
	"net/http"
	"os"

	"remote-agent-mcp/internal/config"
	"remote-agent-mcp/internal/server"
	"remote-agent-mcp/internal/tools"
)

func main() {
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	shell := tools.ResolveShell(cfg.Shell)
	if shell != cfg.Shell {
		log.Printf("shell %q not found, falling back to %q", cfg.Shell, shell)
	}

	ts := tools.NewToolSet(tools.Options{
		Root:             cfg.Root,
		Shell:            shell,
		AllowAll:         cfg.AllowAll,
		AllowUnconfirmed: cfg.AllowUnconfirmed,
	})

	s := server.New(ts)
	h := server.Handler(s, cfg.Token)

	log.Printf("%s listening on %s (root=%q allowAll=%v allowUnconfirmed=%v token=%v)",
		"remote-agent-mcp", cfg.Addr, cfg.Root, cfg.AllowAll, cfg.AllowUnconfirmed, cfg.Token != "")
	if err := http.ListenAndServe(cfg.Addr, h); err != nil {
		log.Fatalf("server: %v", err)
	}
}
