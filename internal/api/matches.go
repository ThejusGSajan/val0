package api

import (
	"encoding/json"
	"fmt"

	"github.com/ThejusGSajan/val0/internal/models"
)

// FetchMatchHistory retrieves recent match summaries for the player.
//
//	GET /match-history/v1/history/{puuid}?startIndex={start}&endIndex={end}&queue={queue}
func (c *Client) FetchMatchHistory(start, end int, queue string) (*models.MatchHistoryResponse, error) {
	url := c.pdURL(fmt.Sprintf("/match-history/v1/history/%s?startIndex=%d&endIndex=%d", c.session.PUUID, start, end))
	if queue != "" {
		url += fmt.Sprintf("&queue=%s", queue)
	}

	body, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch match history: %w", err)
	}

	var mhr models.MatchHistoryResponse
	if err := json.Unmarshal(body, &mhr); err != nil {
		return nil, fmt.Errorf("decode match history: %w", err)
	}
	return &mhr, nil
}

// FetchMatchDetails retrieves full match breakdown for a specific match ID.
//
//	GET /match-details/v1/matches/{matchId}
func (c *Client) FetchMatchDetails(matchID string) (*models.MatchDetails, error) {
	url := c.pdURL(fmt.Sprintf("/match-details/v1/matches/%s", matchID))
	body, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch match details: %w", err)
	}

	var md models.MatchDetails
	if err := json.Unmarshal(body, &md); err != nil {
		return nil, fmt.Errorf("decode match details: %w", err)
	}
	return &md, nil
}

// FetchPlayerNames resolves a list of PUUIDs into GameName#TagLine strings.
//
//	PUT /name-service/v2/players
func (c *Client) FetchPlayerNames(puuids []string) (map[string]string, error) {
	if len(puuids) == 0 {
		return make(map[string]string), nil
	}

	payload, err := json.Marshal(puuids)
	if err != nil {
		return nil, err
	}

	url := c.pdURL("/name-service/v2/players")
	body, err := c.doRequest("PUT", url, payload)
	if err != nil {
		return nil, fmt.Errorf("fetch player names: %w", err)
	}

	var players []models.PlayerNameResponse
	if err := json.Unmarshal(body, &players); err != nil {
		return nil, fmt.Errorf("decode player names: %w", err)
	}

	nameMap := make(map[string]string, len(players))
	for _, p := range players {
		if p.GameName != "" {
			nameMap[p.Subject] = fmt.Sprintf("%s#%s", p.GameName, p.TagLine)
		} else if p.DisplayName != "" {
			nameMap[p.Subject] = p.DisplayName
		}
	}
	return nameMap, nil
}
