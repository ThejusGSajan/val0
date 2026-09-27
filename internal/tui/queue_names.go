package tui

import (
	"strings"

	"github.com/ThejusGSajan/val0/internal/cache"
	"github.com/ThejusGSajan/val0/internal/models"
)

// QueueDisplayName translates internal Valorant QueueID strings
// into human-readable game mode names.
var queueDisplayNames = map[string]string{
	"competitive":       "Competitive",
	"unrated":           "Unrated",
	"deathmatch":        "Deathmatch",
	"skirmish":          "Skirmish",
	"spikerush":         "Spike Rush",
	"swiftplay":         "Swiftplay",
	"hurm":              "Team Deathmatch",
	"ggteam":            "Escalation",
	"onefa":             "Replication",
	"snowball":          "Snowball Fight",
	"newmap":            "New Map",
	"custom":            "Custom",
	"abilitydraftarena": "Gauntlet",
	"abilitydraft":      "Gauntlet",
	"gauntlet":          "Gauntlet",
	"":                  "Unknown",
}

// GetQueueDisplayName returns the human-readable name for a QueueID.
// Falls back to title-casing the raw ID if not in the map.
func GetQueueDisplayName(queueID string) string {
	if name, ok := queueDisplayNames[strings.ToLower(queueID)]; ok {
		return name
	}
	// Fallback: capitalize first letter
	if len(queueID) > 0 {
		return strings.ToUpper(queueID[:1]) + queueID[1:]
	}
	return "Unknown"
}

// ResolveQueueDisplayName resolves human-readable queue name from QueueID,
// falling back to inspecting the GameMode asset path if QueueID is empty or unknown.
func ResolveQueueDisplayName(queueID, gameMode string) string {
	if name, ok := queueDisplayNames[strings.ToLower(queueID)]; ok && name != "Unknown" {
		return name
	}
	// Fallback: inspect GameMode asset path
	if gameMode != "" {
		gmLower := strings.ToLower(gameMode)
		if strings.Contains(gmLower, "abilitydraftarena") || strings.Contains(gmLower, "abilitydraft") || strings.Contains(gmLower, "gauntlet") {
			return "Gauntlet"
		}
		if strings.Contains(gmLower, "skirmish") {
			return "Skirmish"
		}
		if strings.Contains(gmLower, "deathmatch") {
			return "Deathmatch"
		}
		if strings.Contains(gmLower, "hurm") {
			return "Team Deathmatch"
		}
		if strings.Contains(gmLower, "spikerush") {
			return "Spike Rush"
		}
		if strings.Contains(gmLower, "swiftplay") {
			return "Swiftplay"
		}
		if strings.Contains(gmLower, "onefa") {
			return "Replication"
		}
		if strings.Contains(gmLower, "ggteam") {
			return "Escalation"
		}
		if strings.Contains(gmLower, "snowball") {
			return "Snowball Fight"
		}
	}
	return GetQueueDisplayName(queueID)
}

// ResolveAgentDisplayName resolves the agent display name taking into account
// game modes like Gauntlet where agents are not chosen and must display as "-".
func ResolveAgentDisplayName(characterID string, agentsMap map[string]string, md *models.MatchDetails) string {
	if md != nil && md.IsGauntlet() {
		return "-"
	}
	return cache.GetAgentName(characterID, agentsMap)
}
