package auth

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/val-tracker/val-tracker/internal/models"
)

var (
	ErrLockfileNotFound = errors.New("lockfile not found — please start the Riot Client")
	ErrLockfileStale    = errors.New("lockfile is stale (>1 hour) — please restart the Riot Client")
)

// lockfilePath returns the absolute path to the Riot Client lockfile.
func lockfilePath() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"),
		"Riot Games", "Riot Client", "Config", "lockfile")
}

// ReadLockfile reads and validates the lockfile.
// Returns ErrLockfileNotFound if absent, ErrLockfileStale if older than 1h.
func ReadLockfile() (*models.Lockfile, error) {
	path := lockfilePath()

	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrLockfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("stat lockfile: %w", err)
	}

	// Constraint: reject if modification time > 1 hour ago.
	if time.Since(info.ModTime()) > time.Hour {
		return nil, ErrLockfileStale
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read lockfile: %w", err)
	}

	return parseLockfile(string(data), info.ModTime())
}

// parseLockfile splits the colon-delimited line:
// name:pid:port:password:protocol
func parseLockfile(raw string, modTime time.Time) (*models.Lockfile, error) {
	parts := strings.Split(strings.TrimSpace(raw), ":")
	if len(parts) < 5 {
		return nil, fmt.Errorf("malformed lockfile: expected 5 fields, got %d", len(parts))
	}
	return &models.Lockfile{
		Name:     parts[0],
		PID:      parts[1],
		Port:     parts[2],
		Password: parts[3],
		Protocol: parts[4],
		ModTime:  modTime,
	}, nil
}

// ── Token Extraction ────────────────────────────────────────────────

// insecureClient returns an *http.Client that skips TLS verification.
// Required because the local Riot Client uses a self-signed certificate.
func insecureClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

// basicAuth builds the "Basic <base64>" header value for the lockfile.
func basicAuth(password string) string {
	cred := "riot:" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(cred))
}

// FetchEntitlements calls the local endpoint:
//
//	GET https://127.0.0.1:{port}/entitlements/v1/token
//
// and returns the parsed EntitlementResponse.
func FetchEntitlements(lf *models.Lockfile) (*models.EntitlementResponse, error) {
	url := fmt.Sprintf("https://127.0.0.1:%s/entitlements/v1/token", lf.Port)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", basicAuth(lf.Password))

	resp, err := insecureClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("local entitlements request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("entitlements returned %d: %s", resp.StatusCode, body)
	}

	var ent models.EntitlementResponse
	if err := json.NewDecoder(resp.Body).Decode(&ent); err != nil {
		return nil, fmt.Errorf("decode entitlements: %w", err)
	}
	return &ent, nil
}

// ── Region Detection ────────────────────────────────────────────────

// DetectRegion uses the Riot Geo endpoint:
//
//	PUT https://riot-geo.pas.si.riotgames.com/pas/v1/product/valorant
//
// with the access token as a Bearer header. Returns "na", "eu", "ap", "kr", etc.
// Falls back to "" if detection fails (caller must then prompt the user).
func DetectRegion(accessToken string) (string, error) {
	url := "https://riot-geo.pas.si.riotgames.com/pas/v1/product/valorant"

	req, err := http.NewRequest("PUT", url, strings.NewReader(`{"id_token":"`+accessToken+`"}`))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("riot geo request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("riot geo returned %d", resp.StatusCode)
	}

	var geo models.RiotGeoResponse
	if err := json.NewDecoder(resp.Body).Decode(&geo); err != nil {
		return "", fmt.Errorf("decode geo response: %w", err)
	}

	// The affinities map has keys like "live", "pbe".
	// The "live" key gives us the production shard.
	if region, ok := geo.Affinities["live"]; ok {
		return region, nil
	}
	return "", fmt.Errorf("no 'live' affinity in geo response")
}

// BuildSession orchestrates lockfile → tokens → region → Session.
func BuildSession() (*models.Session, error) {
	lf, err := ReadLockfile()
	if err != nil {
		return nil, err
	}

	ent, err := FetchEntitlements(lf)
	if err != nil {
		return nil, fmt.Errorf("fetch entitlements: %w", err)
	}

	region, err := DetectRegion(ent.AccessToken)
	if err != nil {
		// Region detection failed — return session with empty region.
		// The TUI will drop the user into a manual selection prompt.
		return &models.Session{
			AccessToken:      ent.AccessToken,
			EntitlementToken: ent.Token,
			PUUID:            ent.Subject,
			Region:           "", // signals "needs manual selection"
		}, nil
	}

	shard := region
	if s, ok := RegionToShard[region]; ok {
		shard = s
	}

	return &models.Session{
		AccessToken:      ent.AccessToken,
		EntitlementToken: ent.Token,
		PUUID:            ent.Subject,
		Region:           region,
		Shard:            shard,
	}, nil
}
