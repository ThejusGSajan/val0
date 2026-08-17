package models

// MMRResponse is returned by GET https://pd.{shard}.a.pvp.net/mmr/v1/players/{puuid}
type MMRResponse struct {
	Subject                 string                `json:"Subject"`
	QueueSkills             map[string]QueueSkill `json:"QueueSkills"`
	LatestCompetitiveUpdate CompetitiveUpdate     `json:"LatestCompetitiveUpdate"`
}

type QueueSkill struct {
	TotalGamesNeededForRating         int                     `json:"TotalGamesNeededForRating"`
	TotalGamesNeededForLeaderboard    int                     `json:"TotalGamesNeededForLeaderboard"`
	CurrentSeasonGamesNeededForRating int                     `json:"CurrentSeasonGamesNeededForRating"`
	SeasonalInfoBySeasonID            map[string]SeasonalInfo `json:"SeasonalInfoBySeasonID"`
}

type SeasonalInfo struct {
	SeasonID                   string         `json:"SeasonID"`
	NumberOfWins               int            `json:"NumberOfWins"`
	NumberOfWinsWithPlacements int            `json:"NumberOfWinsWithPlacements"`
	NumberOfGames              int            `json:"NumberOfGames"`
	Rank                       int            `json:"Rank"`
	CompetitiveTier            int            `json:"CompetitiveTier"`
	RankedRating               int            `json:"RankedRating"`
	WinsByTier                 map[string]int `json:"WinsByTier"`
	GamesNeededForRating       int            `json:"GamesNeededForRating"`
	TotalWinsNeededForRank     int            `json:"TotalWinsNeededForRank"`
}

type CompetitiveUpdate struct {
	MatchID                  string `json:"MatchID"`
	MapID                    string `json:"MapID"`
	SeasonID                 string `json:"SeasonID"`
	MatchStartTime           int64  `json:"MatchStartTime"`
	TierAfterUpdate          int    `json:"TierAfterUpdate"`
	TierBeforeUpdate         int    `json:"TierBeforeUpdate"`
	RankedRatingAfterUpdate  int    `json:"RankedRatingAfterUpdate"`
	RankedRatingBeforeUpdate int    `json:"RankedRatingBeforeUpdate"`
	RankedRatingEarned       int    `json:"RankedRatingEarned"`
	CompetitiveMovement      string `json:"CompetitiveMovement"`
}

type CompetitiveUpdatesResponse struct {
	Matches []CompetitiveUpdateMatch `json:"Matches"`
}

type CompetitiveUpdateMatch struct {
	MatchID                  string `json:"MatchID"`
	MapID                    string `json:"MapID"`
	SeasonID                 string `json:"SeasonID"`
	MatchStartTime           int64  `json:"MatchStartTime"`
	TierAfterUpdate          int    `json:"TierAfterUpdate"`
	TierBeforeUpdate         int    `json:"TierBeforeUpdate"`
	RankedRatingAfterUpdate  int    `json:"RankedRatingAfterUpdate"`
	RankedRatingBeforeUpdate int    `json:"RankedRatingBeforeUpdate"`
	RankedRatingEarned       int    `json:"RankedRatingEarned"`
	CompetitiveMovement      string `json:"CompetitiveMovement"`
}

// GetCurrentCompetitiveInfo extracts tier and RR for the player.
func (m *MMRResponse) GetCurrentCompetitiveInfo() (tier int, rr int) {
	if m == nil {
		return 0, 0
	}
	if m.LatestCompetitiveUpdate.TierAfterUpdate > 0 {
		return m.LatestCompetitiveUpdate.TierAfterUpdate, m.LatestCompetitiveUpdate.RankedRatingAfterUpdate
	}
	if comp, ok := m.QueueSkills["competitive"]; ok {
		for _, sInfo := range comp.SeasonalInfoBySeasonID {
			if sInfo.CompetitiveTier > 0 {
				return sInfo.CompetitiveTier, sInfo.RankedRating
			}
		}
	}
	return 0, 0
}
