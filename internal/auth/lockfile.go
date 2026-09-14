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
	"regexp"
	"strings"
	"time"

	"github.com/ThejusGSajan/val0/internal/models"
)

var (
	ErrLockfileNotFound = errors.New("lockfile not found — please start the Riot Client")
)

// lockfilePath returns the absolute path to the Riot Client lockfile.
func lockfilePath() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"),
		"Riot Games", "Riot Client", "Config", "lockfile")
}

// ReadLockfile reads and validates the lockfile.
// Returns ErrLockfileNotFound if absent.
func ReadLockfile() (*models.Lockfile, error) {
	path := lockfilePath()

	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrLockfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("stat lockfile: %w", err)
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

// ProbeLiveness attempts a lightweight GET to the local Riot Client.
// Returns nil if the client is alive, or an error if unreachable.
func ProbeLiveness(lf *models.Lockfile) error {
	url := fmt.Sprintf("https://127.0.0.1:%s/entitlements/v1/token", lf.Port)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("build probe request: %w", err)
	}
	req.Header.Set("Authorization", basicAuth(lf.Password))

	resp, err := insecureClient().Do(req)
	if err != nil {
		return fmt.Errorf("riot client not reachable — please start the Riot Client: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("riot client returned status %d — please restart the Riot Client", resp.StatusCode)
	}
	return nil
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

// DetectRegion attempts to resolve the player's region and shard using:
//  1. Local Valorant ShooterGame.log analysis (most reliable & fast)
//  2. Local Riot Client external-sessions endpoint
//  3. Riot Geo PAS endpoint
func DetectRegion(lf *models.Lockfile, accessToken, puuid string) (string, error) {
	// Strategy 1: Parse ShooterGame.log
	if region := detectRegionFromLog(); region != "" {
		return region, nil
	}

	// Strategy 2: Query local external-sessions
	if lf != nil {
		if region := detectRegionFromLocalSessions(lf); region != "" {
			return region, nil
		}
	}

	// Strategy 3: Riot Geo API
	if region, err := detectRegionFromGeo(accessToken); err == nil && region != "" {
		return region, nil
	}

	// Strategy 4: Active Shard Probing
	if puuid != "" {
		if region := probeActiveShard(accessToken, puuid); region != "" {
			return region, nil
		}
	}

	return "", fmt.Errorf("could not auto-detect region")
}

// detectRegionFromLog scans ShooterGame.log for active region/shard URLs.
func detectRegionFromLog() string {
	logPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "VALORANT", "Saved", "Logs", "ShooterGame.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		return ""
	}

	content := string(data)
	// Match https://pd.ap.a.pvp.net or https://glz-ap-1.ap.a.pvp.net
	patterns := []string{
		`https://pd\.([a-z0-9-]+)\.a\.pvp\.net`,
		`https://glz-[a-z0-9-]+-1\.([a-z0-9-]+)\.a\.pvp\.net`,
		`-ares-deployment=([a-z0-9-]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			region := strings.ToLower(matches[1])
			if _, ok := RegionToShard[region]; ok {
				return region
			}
		}
	}
	return ""
}

// detectRegionFromLocalSessions queries the local Riot Client session.
func detectRegionFromLocalSessions(lf *models.Lockfile) string {
	url := fmt.Sprintf("https://127.0.0.1:%s/product-session/v1/external-sessions", lf.Port)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", basicAuth(lf.Password))

	resp, err := insecureClient().Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	re := regexp.MustCompile(`-ares-deployment=([a-zA-Z0-9-]+)`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) > 1 {
		region := strings.ToLower(matches[1])
		if _, ok := RegionToShard[region]; ok {
			return region
		}
	}
	return ""
}

// detectRegionFromGeo queries the Riot Geo endpoint.
func detectRegionFromGeo(accessToken string) (string, error) {
	url := "https://riot-geo.pas.si.riotgames.com/pas/v1/product/valorant"
	req, err := http.NewRequest("PUT", url, strings.NewReader(`{"id_token":"`+accessToken+`"}`))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("geo status %d", resp.StatusCode)
	}

	var geo models.RiotGeoResponse
	if err := json.NewDecoder(resp.Body).Decode(&geo); err != nil {
		return "", err
	}

	if region, ok := geo.Affinities["live"]; ok {
		return region, nil
	}
	return "", fmt.Errorf("no live affinity")
}

// probeActiveShard tests each shard with a lightweight GET request.
func probeActiveShard(accessToken, puuid string) string {
	client := &http.Client{Timeout: 3 * time.Second}
	for _, shard := range []string{"ap", "eu", "na", "kr"} {
		url := fmt.Sprintf("https://pd.%s.a.pvp.net/name-service/v2/players", shard)
		req, err := http.NewRequest("PUT", url, strings.NewReader(`["`+puuid+`"]`))
		if err != nil {
			continue
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return shard
			}
		}
	}
	return ""
}

// BuildSession orchestrates lockfile → tokens → region → Session.
func BuildSession() (*models.Session, error) {
	lf, err := ReadLockfile()
	if err != nil {
		return nil, err
	}

	// Verify the Riot Client process is actually running
	if err := ProbeLiveness(lf); err != nil {
		return nil, err
	}

	ent, err := FetchEntitlements(lf)
	if err != nil {
		return nil, fmt.Errorf("fetch entitlements: %w", err)
	}

	region, err := DetectRegion(lf, ent.AccessToken, ent.Subject)
	if err != nil {
		// Region detection fallback: empty region triggers Bubble Tea selector
		return &models.Session{
			AccessToken:      ent.AccessToken,
			EntitlementToken: ent.Token,
			PUUID:            ent.Subject,
			Region:           "",
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
