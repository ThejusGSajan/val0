package api

import (
	"encoding/json"
	"fmt"

	"github.com/val-tracker/val-tracker/internal/models"
)

// FetchContracts retrieves all player contracts (including Battlepass).
//
//	GET https://pd.{shard}.a.pvp.net/contracts/v1/contracts/{puuid}
func (c *Client) FetchContracts() (*models.ContractsResponse, error) {
	url := c.pdURL(fmt.Sprintf("/contracts/v1/contracts/%s", c.session.PUUID))
	body, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch contracts: %w", err)
	}

	var cr models.ContractsResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, fmt.Errorf("decode contracts: %w", err)
	}
	return &cr, nil
}

// FetchContent retrieves the active season/act metadata.
//
//	GET https://shared.{shard}.a.pvp.net/content-service/v3/content
func (c *Client) FetchContent() (*models.ContentResponse, error) {
	url := c.sharedURL("/content-service/v3/content")
	body, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch content: %w", err)
	}

	var cr models.ContentResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, fmt.Errorf("decode content: %w", err)
	}
	return &cr, nil
}

// FindActiveBattlepass identifies the battlepass contract from the
// contracts list by cross-referencing with the active season.
// Returns the matching Contract and true, or zero-value and false.
func FindActiveBattlepass(
	contracts *models.ContractsResponse,
	content *models.ContentResponse,
) (models.Contract, bool) {
	if contracts == nil || content == nil {
		return models.Contract{}, false
	}

	// Find the currently active act
	var activeActID string
	for _, s := range content.Seasons {
		if s.IsActive && s.Type == "act" {
			activeActID = s.ID
			break
		}
	}

	// The battlepass ContractDefinitionID matches the active act's ID
	// in most implementations.
	if activeActID != "" {
		for _, c := range contracts.Contracts {
			if c.ContractDefinitionID == activeActID {
				return c, true
			}
		}
	}

	// Fallback: return the contract with the highest total progression
	// (the battlepass accumulates significantly more XP than character contracts).
	var best models.Contract
	var bestXP int
	for _, c := range contracts.Contracts {
		xp := c.ContractProgression.TotalProgressionEarned
		if xp > bestXP {
			bestXP = xp
			best = c
		}
	}
	if bestXP > 0 {
		return best, true
	}
	return models.Contract{}, false
}
