package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Test that the auth middleware allows requests with the correct API key
func TestAuthMiddleware_AllowsCorrectKey(t *testing.T) {
	t.Setenv("IPFS_API_KEY", "test-key-123")

	s := &SimpleServer{startTime: time.Now()}
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := s.authMiddleware(next)
	req := httptest.NewRequest("GET", "/api/v1/files", nil)
	req.Header.Set("X-API-Key", "test-key-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !nextCalled {
		t.Fatalf("expected next handler to be called on valid key")
	}
}

// Test that the auth middleware denies incorrect API key
func TestAuthMiddleware_DeniesWrongKey(t *testing.T) {
	t.Setenv("IPFS_API_KEY", "correct-key")

	s := &SimpleServer{startTime: time.Now()}
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := s.authMiddleware(next)
	req := httptest.NewRequest("GET", "/api/v1/files", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if nextCalled {
		t.Fatalf("next handler should not be called on invalid key")
	}
}

func TestAuthMiddleware_DeniesRequestsWhenKeyIsNotConfigured(t *testing.T) {
	t.Setenv("IPFS_API_KEY", "")

	s := &SimpleServer{startTime: time.Now()}
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := s.authMiddleware(next)
	req := httptest.NewRequest("GET", "/api/v1/files", nil)
	req.Header.Set("X-API-Key", "unconfigured-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if nextCalled {
		t.Fatal("next handler should not be called without a configured API key")
	}
}

// Health endpoints should bypass auth
func TestAuthMiddleware_HealthEndpointSkipsAuth(t *testing.T) {
	s := &SimpleServer{startTime: time.Now()}
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := s.authMiddleware(next)
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 on health endpoint skip, got %d", rr.Code)
	}
	if !nextCalled {
		t.Fatalf("health endpoint should bypass auth and reach next")
	}
}
