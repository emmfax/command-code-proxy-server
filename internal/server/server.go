package server

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/dev2k6/command-code-proxy-server/internal/proxy"
)

const defaultPort = "55990"
const defaultHost = "127.0.0.1"

// maxBodyBytes caps every request body (32 MB) at the edge so oversized
// uploads are rejected before reaching any handler.
const maxBodyBytes int64 = 32 << 20

// Server represents the HTTP server
type Server struct {
	Port    string
	Host    string
	Proxy   *proxy.Proxy
	Handler http.Handler
	httpSrv *http.Server
}

// NewServer creates a new server instance
func NewServer(proxy *proxy.Proxy) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", logger(proxy.HandleChatCompletions))
	mux.HandleFunc("/chat/completions", logger(proxy.HandleChatCompletions))
	mux.HandleFunc("/v1/responses", logger(proxy.HandleResponses))
	mux.HandleFunc("/v1/models", logger(proxy.HandleModels))
	mux.HandleFunc("/health", logger(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))

	return &Server{
		Port:    defaultPort,
		Host:    defaultHost,
		Proxy:   proxy,
		Handler: limitBody(mux),
	}
}

// SetPort sets the port for the server
func (s *Server) SetPort(port string) {
	if port != "" {
		s.Port = port
	}
}

// SetHost sets the host for the server
func (s *Server) SetHost(host string) {
	if host != "" {
		s.Host = host
	}
}

// GetPort returns the server port
func (s *Server) GetPort() string {
	return s.Port
}

// GetHost returns the server host
func (s *Server) GetHost() string {
	return s.Host
}

// limitBody caps request bodies at the edge for every route.
func limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		}
		next.ServeHTTP(w, r)
	})
}

// logger is a middleware for logging requests
func logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next(w, r)
		log.Printf("[%s] %s done in %v", r.Method, r.URL.Path, time.Since(start))
	}
}

// Start starts the HTTP server
func (s *Server) Start() {
	addr := net.JoinHostPort(s.Host, s.Port)
	s.httpSrv = &http.Server{
		Addr:              addr,
		Handler:           s.Handler,
		ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout:       120 * time.Second,
		// NOTE: ReadTimeout/WriteTimeout intentionally unset — SSE streams may
		// legitimately stay open for minutes while a model generates.
	}
	log.Printf("listening on %s", addr)
	if err := s.httpSrv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
