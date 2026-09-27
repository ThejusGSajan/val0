package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ThejusGSajan/val0/internal/models"
)

func TestClient_DoRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-access-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("X-Riot-Entitlements-JWT") != "test-entitlement-token" {
			http.Error(w, "missing entitlement", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	session := &models.Session{
		AccessToken:      "test-access-token",
		EntitlementToken: "test-entitlement-token",
		ClientVersion:    "release-13.06",
	}
	client := NewClient(session)

	data, err := client.doRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("expected successful request, got: %v", err)
	}
	if !strings.Contains(string(data), `"status":"ok"`) {
		t.Errorf("unexpected response body: %s", string(data))
	}
}

func TestClient_DoRequest_BadClaimsAutoRetry(t *testing.T) {
	// 1. Mock local Riot Client TLS server for entitlements
	localAuthServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/entitlements/v1/token" {
			http.NotFound(w, r)
			return
		}
		resp := models.EntitlementResponse{
			AccessToken:  "refreshed-access-token",
			Token:        "refreshed-entitlement-token",
			Subject:      "test-puuid",
			Entitlements: []any{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer localAuthServer.Close()

	_, localPort, err := net.SplitHostPort(localAuthServer.Listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to split host/port: %v", err)
	}

	// 2. Set up temporary LOCALAPPDATA with mock lockfile pointing to localAuthServer
	tempDir := t.TempDir()
	t.Setenv("LOCALAPPDATA", tempDir)

	lockfileDir := filepath.Join(tempDir, "Riot Games", "Riot Client", "Config")
	if err := os.MkdirAll(lockfileDir, 0o755); err != nil {
		t.Fatalf("failed to create mock lockfile dir: %v", err)
	}
	lockfilePath := filepath.Join(lockfileDir, "lockfile")
	lockfileContent := fmt.Sprintf("RiotClient:1234:%s:mockpassword:https", localPort)
	if err := os.WriteFile(lockfilePath, []byte(lockfileContent), 0o644); err != nil {
		t.Fatalf("failed to write mock lockfile: %v", err)
	}

	// 3. Mock Riot API server that returns BAD_CLAIMS on stale token, but succeeds on refreshed token
	var reqCount int32
	remoteAPIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&reqCount, 1)
		authHeader := r.Header.Get("Authorization")

		if count == 1 {
			// First request: stale token -> BAD_CLAIMS
			if authHeader != "Bearer stale-token" {
				t.Errorf("expected first request to have stale token, got %s", authHeader)
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"errorCode": "BAD_CLAIMS", "message": "Failure validating/decoding RSO Access Token"}`))
			return
		}

		// Second request: retry with refreshed token -> OK
		if authHeader != "Bearer refreshed-access-token" {
			t.Errorf("expected retried request to have refreshed token, got %s", authHeader)
		}
		if r.Header.Get("X-Riot-Entitlements-JWT") != "refreshed-entitlement-token" {
			t.Errorf("expected refreshed entitlement header, got %s", r.Header.Get("X-Riot-Entitlements-JWT"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": "success"}`))
	}))
	defer remoteAPIServer.Close()

	session := &models.Session{
		AccessToken:      "stale-token",
		EntitlementToken: "stale-entitlement",
		ClientVersion:    "release-13.06",
	}
	client := NewClient(session)

	// 4. Execute request
	body, err := client.doRequest("GET", remoteAPIServer.URL, nil)
	if err != nil {
		t.Fatalf("expected request to succeed after auto-retry, got error: %v", err)
	}

	if !strings.Contains(string(body), `"result": "success"`) {
		t.Errorf("expected success response, got: %s", string(body))
	}

	// Verify session tokens were updated in-place
	if client.Session().AccessToken != "refreshed-access-token" {
		t.Errorf("expected session AccessToken updated to refreshed-access-token, got %s", client.Session().AccessToken)
	}
	if client.Session().EntitlementToken != "refreshed-entitlement-token" {
		t.Errorf("expected session EntitlementToken updated, got %s", client.Session().EntitlementToken)
	}
	if atomic.LoadInt32(&reqCount) != 2 {
		t.Errorf("expected exactly 2 requests (initial + 1 retry), got %d", atomic.LoadInt32(&reqCount))
	}
}

func TestClient_DoRequest_BadClaimsNoInfiniteLoop(t *testing.T) {
	// Mock Riot API server that ALWAYS returns BAD_CLAIMS
	var reqCount int32
	remoteAPIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errorCode": "BAD_CLAIMS", "message": "Failure validating/decoding RSO Access Token"}`))
	}))
	defer remoteAPIServer.Close()

	// Empty LOCALAPPDATA with no lockfile
	tempDir := t.TempDir()
	t.Setenv("LOCALAPPDATA", tempDir)

	session := &models.Session{
		AccessToken:      "stale-token",
		EntitlementToken: "stale-entitlement",
		ClientVersion:    "release-13.06",
	}
	client := NewClient(session)

	_, err := client.doRequest("GET", remoteAPIServer.URL, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "BAD_CLAIMS") {
		t.Errorf("expected BAD_CLAIMS in error message, got: %v", err)
	}

	// Should not retry if lockfile is missing
	if count := atomic.LoadInt32(&reqCount); count != 1 {
		t.Errorf("expected 1 request attempt when lockfile missing, got %d", count)
	}
}
