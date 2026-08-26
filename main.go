package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dev2k6/command-code-proxy-server/internal/proxy"
	"github.com/dev2k6/command-code-proxy-server/internal/server"
	"github.com/dev2k6/command-code-proxy-server/internal/update"
)

const appVersion = "v1.0.9"
const repositoryURL = "https://github.com/dev2k6/command-code-proxy-server"
const debugLogging = false

func main() {
	port := flag.String("port", "", "Port to run the server on (default: 55990)")
	host := flag.String("host", "", "Host to bind to (default: 127.0.0.1)")
	apiKey := flag.String("api-key", "", "API key for CommandCode (optional, can also be set via Authorization header)")
	authKeys := flag.String("auth-keys", "", "Comma-separated client keys allowed to use this proxy (empty = open). Also: env CCP_AUTH_KEYS")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(versionText())
		return
	}

	keys := *authKeys
	if keys == "" {
		keys = os.Getenv("CCP_AUTH_KEYS")
	}

	p := proxy.NewProxy(*apiKey)
	p.Debug = debugLogging
	p.SetAuthKeys(proxy.ParseKeyList(keys))

	srv := server.NewServer(p)
	srv.SetPort(*port)
	srv.SetHost(*host)

	printStartupInfo(srv, len(p.AuthKeys) > 0)

	srv.Start()
}

func versionText() string {
	latest, hasUpdate, err := update.LatestVersion(appVersion)
	if err != nil || !hasUpdate {
		return appVersion
	}
	return fmt.Sprintf("%s (latest: %s)", appVersion, latest)
}

func printStartupInfo(srv *server.Server, authEnabled bool) {
	fmt.Println("")
	fmt.Println("========================================")
	fmt.Println("  CommandCode Proxy Server")
	fmt.Println("========================================")
	fmt.Println("")
	fmt.Printf("  Version:     %s\n", versionText())
	fmt.Printf("  Repository:  %s\n", repositoryURL)
	fmt.Printf("  Host:        %s\n", srv.GetHost())
	fmt.Printf("  Port:        %s\n", srv.GetPort())
	if authEnabled {
		fmt.Println("  Client auth: ENABLED (whitelist via -auth-keys / CCP_AUTH_KEYS)")
	} else {
		fmt.Println("  Client auth: disabled (open proxy)")
	}
	fmt.Println("  Base URL:    https://api.commandcode.ai")
	fmt.Println("")
	fmt.Println("  Endpoints:")
	fmt.Println("    POST /v1/chat/completions  (OpenAI-compatible)")
	fmt.Println("    POST /chat/completions     (OpenAI-compatible alias)")
	fmt.Println("    POST /v1/responses         (OpenAI Responses-compatible)")
	fmt.Println("    GET  /v1/models            (list models)")
	fmt.Println("    GET  /health               (health check)")
	fmt.Println("")
	fmt.Printf("  Server running on http://%s:%s\n", srv.GetHost(), srv.GetPort())
	fmt.Println("")
	fmt.Println("  Press Ctrl+C to stop")
	fmt.Println("========================================")
}
