package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/ThejusGSajan/val0/internal/models"
)

func TestParseLockfile(t *testing.T) {
	raw := "Riot Client:12345:55555:secretPass:https"
	modTime := time.Now()

	lf, err := parseLockfile(raw, modTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lf.Name != "Riot Client" {
		t.Errorf("expected Name 'Riot Client', got '%s'", lf.Name)
	}
	if lf.PID != "12345" {
		t.Errorf("expected PID '12345', got '%s'", lf.PID)
	}
	if lf.Port != "55555" {
		t.Errorf("expected Port '55555', got '%s'", lf.Port)
	}
	if lf.Password != "secretPass" {
		t.Errorf("expected Password 'secretPass', got '%s'", lf.Password)
	}
	if lf.Protocol != "https" {
		t.Errorf("expected Protocol 'https', got '%s'", lf.Protocol)
	}
}

func TestParseLockfile_Malformed(t *testing.T) {
	raw := "Riot Client:12345:55555"
	_, err := parseLockfile(raw, time.Now())
	if err == nil {
		t.Errorf("expected error for malformed lockfile, got nil")
	}
}

func TestProbeLiveness(t *testing.T) {
	// 1. Success case (HTTP 200)
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/entitlements/v1/token" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("parse test server url: %v", err)
	}

	lf := &models.Lockfile{
		Port:     u.Port(),
		Password: "testpassword",
	}

	if err := ProbeLiveness(lf); err != nil {
		t.Errorf("expected ProbeLiveness to succeed, got: %v", err)
	}

	// 2. Error case (HTTP 500)
	tsError := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer tsError.Close()

	uError, _ := url.Parse(tsError.URL)
	lfError := &models.Lockfile{
		Port:     uError.Port(),
		Password: "testpassword",
	}
	if err := ProbeLiveness(lfError); err == nil {
		t.Errorf("expected ProbeLiveness to fail on HTTP 500, got nil")
	}

	// 3. Unreachable case
	lfUnreachable := &models.Lockfile{
		Port:     "1", // unreachable port
		Password: "testpassword",
	}
	if err := ProbeLiveness(lfUnreachable); err == nil {
		t.Errorf("expected ProbeLiveness to fail on unreachable port, got nil")
	}
}

func TestReadLockfile_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	origLocalApp := os.Getenv("LOCALAPPDATA")
	defer os.Setenv("LOCALAPPDATA", origLocalApp)
	os.Setenv("LOCALAPPDATA", tempDir)

	_, err := ReadLockfile()
	if err != ErrLockfileNotFound {
		t.Errorf("expected ErrLockfileNotFound, got: %v", err)
	}
}

func TestRegionToShardMapping(t *testing.T) {
	tests := []struct {
		region   string
		expected string
	}{
		{"na", "na"},
		{"eu", "eu"},
		{"ap", "ap"},
		{"kr", "kr"},
		{"latam", "na"},
		{"br", "na"},
	}

	for _, tt := range tests {
		shard, ok := RegionToShard[tt.region]
		if !ok || shard != tt.expected {
			t.Errorf("for region %s: expected %s, got %s (ok=%v)", tt.region, tt.expected, shard, ok)
		}
	}
}

func TestURLBuilders(t *testing.T) {
	if got := PDBaseURL("na"); got != "https://pd.na.a.pvp.net" {
		t.Errorf("unexpected PDBaseURL: %s", got)
	}
	if got := SharedBaseURL("eu"); got != "https://shared.eu.a.pvp.net" {
		t.Errorf("unexpected SharedBaseURL: %s", got)
	}
}
