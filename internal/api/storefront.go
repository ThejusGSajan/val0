package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/val-tracker/val-tracker/internal/auth"
	"github.com/val-tracker/val-tracker/internal/models"
)

// FetchStorefront retrieves the daily shop + night market (if active).
//
// Uses the modern v3 endpoint:
//	POST https://pd.{shard}.a.pvp.net/store/v3/storefront/{puuid}
//
// If the primary shard returns 404 (wrong region), it automatically probes
// other shards (ap, eu, na, kr) to find the player's true shard and updates the session.
func (c *Client) FetchStorefront() (*models.StorefrontResponse, error) {
	shardsToTry := []string{c.session.Shard}
	for _, s := range []string{"ap", "eu", "na", "kr"} {
		if s != c.session.Shard {
			shardsToTry = append(shardsToTry, s)
		}
	}

	var lastErr error
	for _, shard := range shardsToTry {
		url := fmt.Sprintf("%s/store/v3/storefront/%s", auth.PDBaseURL(shard), c.session.PUUID)
		body, err := c.doRequest("POST", url, strings.NewReader("{}"))
		if err == nil {
			var sf models.StorefrontResponse
			if err := json.Unmarshal(body, &sf); err == nil {
				// Shard verified! Update session if it changed
				c.session.Shard = shard
				c.session.Region = shard
				return &sf, nil
			}
		}
		lastErr = err
	}

	return nil, fmt.Errorf("fetch storefront: %w", lastErr)
}
