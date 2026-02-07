package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/stuttgart-things/machinery-registry-api/internal/sync"
	"github.com/stuttgart-things/machinery-registry-api/internal/version"
)

// Server represents the HTTP API server
type Server struct {
	router *mux.Router
	http   *http.Server
	syncer *sync.Syncer
}

// NewServer creates and initializes a new HTTP server
func NewServer(syncer *sync.Syncer) *Server {
	s := &Server{
		router: mux.NewRouter(),
		syncer: syncer,
	}

	s.registerRoutes()
	s.applyMiddleware()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s.http = &http.Server{
		Addr:         ":" + port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// registerRoutes sets up all API routes
func (s *Server) registerRoutes() {
	s.router.HandleFunc("/health", s.healthCheck).Methods(http.MethodGet)
	s.router.HandleFunc("/", s.rootInfo).Methods(http.MethodGet)
	s.router.HandleFunc("/version", s.versionInfo).Methods(http.MethodGet)
	s.router.HandleFunc("/openapi", s.serveOpenAPI).Methods(http.MethodGet)
	s.router.HandleFunc("/openapi.yaml", s.serveOpenAPI).Methods(http.MethodGet)
	s.router.HandleFunc("/docs", s.serveDocs).Methods(http.MethodGet)

	s.router.HandleFunc("/api/v1/claims", s.listClaims).Methods(http.MethodGet)
	s.router.HandleFunc("/api/v1/claims/{name}", s.getClaim).Methods(http.MethodGet)
}

// applyMiddleware applies middleware to all routes
func (s *Server) applyMiddleware() {
	s.router.Use(errorHandlerMiddleware)
	s.router.Use(corsMiddleware)
	s.router.Use(requestIDMiddleware)
	s.router.Use(loggingMiddleware)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("HTTP API server starting on %s", s.http.Addr)
	return s.http.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Shutting down HTTP server...")
	return s.http.Shutdown(ctx)
}

// healthCheck returns server health status
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// rootInfo returns a minimal service index
func (s *Server) rootInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{
  "service": "machinery-registry-api",
  "version": "%s",
  "endpoints": [
    "/health",
    "/version",
    "/api/v1/claims",
    "/api/v1/claims/{name}",
    "/openapi.yaml",
    "/docs"
  ]
}`, version.Version)
}

// versionInfo returns build-time version metadata
func (s *Server) versionInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"version":"%s","commit":"%s","buildDate":"%s"}`,
		version.Version, version.Commit, version.BuildDate)
}

// serveOpenAPI serves the OpenAPI spec from docs/openapi.yaml
func (s *Server) serveOpenAPI(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean(filepath.Join("docs", "openapi.yaml"))
	if _, err := os.Stat(path); err == nil {
		w.Header().Set("Content-Type", "application/yaml")
		http.ServeFile(w, r, path)
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "openapi: 3.0.3\ninfo:\n  title: Machinery Registry API\n  version: 0.0.0\npaths: {}\n")
}

// serveDocs serves a Redoc viewer for the OpenAPI spec
func (s *Server) serveDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `<!doctype html>
<html>
  <head>
    <meta charset="utf-8"/>
    <title>Machinery Registry API Docs</title>
    <script src="https://cdn.jsdelivr.net/npm/redoc@next/bundles/redoc.standalone.js"></script>
  </head>
  <body>
    <redoc spec-url="/openapi.yaml"></redoc>
  </body>
</html>`)
}
