package api

import (
	"encoding/json"
	"fmt"

	"github.com/val-tracker/val-tracker/internal/models"
)

// FetchStorefront retrieves the daily shop + night market (if active).
//
//	GET https://pd.{shard}.a.pvp.net/store/v2/storefront/{puuid}
func (c *Client) FetchStorefront() (*models.StorefrontResponse, error) {
	url := c.pdURL(fmt.Sprintf("/store/v2/storefront/%s", c.session.PUUID))
	body, err := c.doRequest(url)
	if err != nil {
		return nil, fmt.Errorf("fetch storefront: %w", err)
	}

	var sf models.StorefrontResponse
	if err := json.Unmarshal(body, &sf); err != nil {
		return nil, fmt.Errorf("decode storefront: %w", err)
	}
	return &sf, nil
}
