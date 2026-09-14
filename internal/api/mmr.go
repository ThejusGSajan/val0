package api

import (
	"encoding/json"
	"fmt"

	"github.com/ThejusGSajan/val0/internal/models"
)

// FetchMMR retrieves the player's current competitive rank and RR.
//
//	GET https://pd.{shard}.a.pvp.net/mmr/v1/players/{puuid}
func (c *Client) FetchMMR() (*models.MMRResponse, error) {
	url := c.pdURL(fmt.Sprintf("/mmr/v1/players/%s", c.session.PUUID))
	body, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch mmr: %w", err)
	}

	var mr models.MMRResponse
	if err := json.Unmarshal(body, &mr); err != nil {
		return nil, fmt.Errorf("decode mmr: %w", err)
	}
	return &mr, nil
}

// FetchCompetitiveUpdates retrieves per-match RR changes.
//
//	GET https://pd.{shard}.a.pvp.net/mmr/v1/players/{puuid}/competitiveupdates?startIndex={start}&endIndex={end}&queue=competitive
func (c *Client) FetchCompetitiveUpdates(start, end int) (*models.CompetitiveUpdatesResponse, error) {
	url := c.pdURL(fmt.Sprintf("/mmr/v1/players/%s/competitiveupdates?startIndex=%d&endIndex=%d&queue=competitive", c.session.PUUID, start, end))
	body, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch competitive updates: %w", err)
	}

	var cur models.CompetitiveUpdatesResponse
	if err := json.Unmarshal(body, &cur); err != nil {
		return nil, fmt.Errorf("decode competitive updates: %w", err)
	}
	return &cur, nil
}
