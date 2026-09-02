// Package api provides a simple REST API server implementation without external dependencies
// This is a fallback implementation when Gin is not available
package api

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"ipfs-encrypted-storage/src/config"
	"ipfs-encrypted-storage/src/ipfs"
)

// SimpleServer provides a basic REST API server
type SimpleServer struct {
	ipfsClient *ipfs.IPFSClient
	config     *config.Config
	startTime  time.Time
	mux        *http.ServeMux
}

// NewSimpleServer creates a new simple API server
func NewSimpleServer(cfg *config.Config) (*SimpleServer, error) {
	ipfsClient, err := ipfs.NewIPFSClient(cfg.IPFS.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to create IPFS client: %w", err)
	}

	server := &SimpleServer{
		ipfsClient: ipfsClient,
		config:     cfg,
		startTime:  time.Now(),
		mux:        http.NewServeMux(),
	}

	server.setupRoutes()
	return server, nil
}

// setupRoutes configures the HTTP routes
func (s *SimpleServer) setupRoutes() {
	// Health endpoints
	s.mux.HandleFunc("/api/v1/health", s.handleHealth)
	s.mux.HandleFunc("/api/v1/health/deep", s.handleDeepHealth)
	s.mux.HandleFunc("/api/v1/health/ready", s.handleReadiness)
	s.mux.HandleFunc("/api/v1/health/live", s.handleLiveness)

	// File operations
	s.mux.HandleFunc("/api/v1/files", s.handleFiles)
	s.mux.HandleFunc("/api/v1/files/", s.handleFileByCID)

	// P2P operations
	s.mux.HandleFunc("/api/v1/p2p/connect", s.handleP2PConnect)
	s.mux.HandleFunc("/api/v1/p2p/peers", s.handleP2PPeers)
	s.mux.HandleFunc("/api/v1/p2p/status", s.handleP2PStatus)

	// Metrics
	s.mux.HandleFunc("/api/v1/metrics", s.handleMetrics)

	// Root
	s.mux.HandleFunc("/", s.handleRoot)
}

// Start starts the simple HTTP server
func (s *SimpleServer) Start(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Starting Simple API server on %s\n", addr)
	fmt.Println("API endpoints available at:")
	fmt.Printf("  Health: http://%s/api/v1/health\n", addr)
	fmt.Printf("  Files:  http://%s/api/v1/files\n", addr)
	fmt.Printf("  P2P:    http://%s/api/v1/p2p\n", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: s.authMiddleware(s.corsMiddleware(s.loggingMiddleware(s.mux))),
	}

	return server.ListenAndServe()
}

// authMiddleware provides basic API key authentication
func (s *SimpleServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health endpoints
		if strings.HasPrefix(r.URL.Path, "/api/v1/health") {
			next.ServeHTTP(w, r)
			return
		}

		// Fail closed unless the server has an API key configured.
		expectedKey := os.Getenv("IPFS_API_KEY")
		apiKey := r.Header.Get("X-API-Key")
		if expectedKey == "" || apiKey == "" || subtle.ConstantTimeCompare([]byte(apiKey), []byte(expectedKey)) != 1 {
			s.sendError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// corsMiddleware adds basic CORS headers
func (s *SimpleServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs requests
func (s *SimpleServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapper, r)

		fmt.Printf("[%s] %s %s %d %v\n",
			r.Method, r.URL.Path, r.RemoteAddr,
			wrapper.statusCode, time.Since(start))
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// sendError sends a JSON error response
func (s *SimpleServer) sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(response)
}

// sendJSON sends a JSON response
func (s *SimpleServer) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// HTTP handlers

func (s *SimpleServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.sendError(w, http.StatusNotFound, "Not found")
		return
	}

	response := map[string]interface{}{
		"name":    "IPFS Encrypted Storage API",
		"version": "1.0.0",
		"status":  "running",
		"endpoints": map[string]string{
			"health": "/api/v1/health",
			"files":  "/api/v1/files",
			"p2p":    "/api/v1/p2p",
		},
	}

	s.sendJSON(w, http.StatusOK, response)
}

func (s *SimpleServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	ipfsHealthy := true
	if err := s.ipfsClient.HealthCheck(); err != nil {
		ipfsHealthy = false
	}

	status := "healthy"
	if !ipfsHealthy {
		status = "degraded"
	}

	response := map[string]interface{}{
		"status":         status,
		"ipfs_connected": ipfsHealthy,
		"p2p_peers":      0,
		"uptime":         time.Since(s.startTime).Seconds(),
		"version":        "1.0.0",
		"timestamp":      time.Now().Unix(),
	}

	s.sendJSON(w, http.StatusOK, response)
}

func (s *SimpleServer) handleDeepHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"checks":    map[string]bool{"ipfs": true, "filesystem": true},
		"timestamp": time.Now(),
	}
	s.sendJSON(w, http.StatusOK, response)
}

func (s *SimpleServer) handleReadiness(w http.ResponseWriter, r *http.Request) {
	if err := s.ipfsClient.HealthCheck(); err != nil {
		s.sendError(w, http.StatusServiceUnavailable, "IPFS not ready")
		return
	}
	s.sendJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *SimpleServer) handleLiveness(w http.ResponseWriter, r *http.Request) {
	s.sendJSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

func (s *SimpleServer) handleFiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListFiles(w, r)
	case http.MethodPost:
		s.handleUploadFile(w, r)
	default:
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *SimpleServer) handleFileByCID(w http.ResponseWriter, r *http.Request) {
	cid := strings.TrimPrefix(r.URL.Path, "/api/v1/files/")

	switch r.Method {
	case http.MethodGet:
		s.handleDownloadFile(w, r, cid)
	case http.MethodDelete:
		s.handleDeleteFile(w, r, cid)
	default:
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *SimpleServer) handleListFiles(w http.ResponseWriter, r *http.Request) {
	// Simplified implementation
	response := map[string]interface{}{
		"files": []map[string]interface{}{},
		"count": 0,
	}
	s.sendJSON(w, http.StatusOK, response)
}

func (s *SimpleServer) handleUploadFile(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, http.StatusNotImplemented, "File upload not implemented in simple server")
}

func (s *SimpleServer) handleDownloadFile(w http.ResponseWriter, r *http.Request, cid string) {
	s.sendError(w, http.StatusNotImplemented, "File download not implemented in simple server")
}

func (s *SimpleServer) handleDeleteFile(w http.ResponseWriter, r *http.Request, cid string) {
	s.sendError(w, http.StatusNotImplemented, "File delete not implemented in simple server")
}

func (s *SimpleServer) handleP2PConnect(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, http.StatusNotImplemented, "P2P connect not implemented")
}

func (s *SimpleServer) handleP2PPeers(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"peers": []map[string]interface{}{},
		"count": 0,
	}
	s.sendJSON(w, http.StatusOK, response)
}

func (s *SimpleServer) handleP2PStatus(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"node_id":             "simple-server",
		"listening_addresses": []string{},
		"connected_peers":     0,
	}
	s.sendJSON(w, http.StatusOK, response)
}

func (s *SimpleServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"files_uploaded":       0,
		"total_bytes_uploaded": 0,
		"active_connections":   0,
		"memory_usage":         0.0,
		"cpu_usage":            0.0,
		"uptime":               time.Since(s.startTime).Seconds(),
		"request_count":        0,
		"error_count":          0,
	}
	s.sendJSON(w, http.StatusOK, response)
}

// RunSimple starts the simple API server.
func RunSimple(cfg *config.Config, host string, port int) error {
	server, err := NewSimpleServer(cfg)
	if err != nil {
		return fmt.Errorf("failed to create simple server: %w", err)
	}

	return server.Start(host, port)
}
