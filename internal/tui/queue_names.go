package tui

import (
	"strings"
)

// QueueDisplayName translates internal Valorant QueueID strings
// into human-readable game mode names.
var queueDisplayNames = map[string]string{
	"competitive": "Competitive",
	"unrated":     "Unrated",
	"deathmatch":  "Deathmatch",
	"spikerush":   "Spike Rush",
	"swiftplay":   "Swiftplay",
	"hurm":        "Team Deathmatch",
	"ggteam":      "Escalation",
	"onefa":       "Replication",
	"snowball":    "Snowball Fight",
	"newmap":      "New Map",
	"custom":      "Custom",
	"":            "Unknown",
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
