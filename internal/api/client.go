package api

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/val-tracker/val-tracker/internal/auth"
	"github.com/val-tracker/val-tracker/internal/models"
)

// The static X-Riot-ClientPlatform header (base64-encoded JSON).
// This value is a well-known constant from the community API docs.
var clientPlatform = base64.StdEncoding.EncodeToString([]byte(`{
	"platformType": "PC",
	"platformOS": "Windows",
	"platformOSVersion": "10.0.19042.1.256.64bit",
	"platformChipset": "Unknown"
}`))

// Client wraps an authenticated HTTP client for Riot PD/Shared endpoints.
type Client struct {
	session    *models.Session
	httpClient *http.Client
}

// NewClient creates a new authenticated API client from a Session.
func NewClient(session *models.Session) *Client {
	return &Client{
		session: session,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// Session returns the underlying session.
func (c *Client) Session() *models.Session {
	return c.session
}

// doRequest performs an authenticated HTTP request against a Riot endpoint.
func (c *Client) doRequest(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.session.AccessToken)
	req.Header.Set("X-Riot-Entitlements-JWT", c.session.EntitlementToken)
	req.Header.Set("X-Riot-ClientPlatform", clientPlatform)
	req.Header.Set("X-Riot-ClientVersion", c.session.ClientVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s %s: %w", method, url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, url, string(respBody))
	}
	return respBody, nil
}

// pdURL builds a full URL against the Player Data service.
func (c *Client) pdURL(path string) string {
	return auth.PDBaseURL(c.session.Shard) + path
}

// sharedURL builds a full URL against the Shared service.
func (c *Client) sharedURL(path string) string {
	return auth.SharedBaseURL(c.session.Shard) + path
}
