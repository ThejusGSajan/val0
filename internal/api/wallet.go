package api

import (
	"encoding/json"
	"fmt"

	"github.com/ThejusGSajan/val0/internal/models"
)

// FetchWallet retrieves VP, Radianite, Kingdom Credits, and Free Agent balances.
//
//	GET https://pd.{shard}.a.pvp.net/store/v1/wallet/{puuid}
func (c *Client) FetchWallet() (*models.WalletResponse, error) {
	url := c.pdURL(fmt.Sprintf("/store/v1/wallet/%s", c.session.PUUID))
	body, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch wallet: %w", err)
	}

	var wr models.WalletResponse
	if err := json.Unmarshal(body, &wr); err != nil {
		return nil, fmt.Errorf("decode wallet: %w", err)
	}
	return &wr, nil
}
