package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const missionsURL = "https://valorant-api.com/v1/missions"

type MissionInfo struct {
	UUID               string         `json:"uuid"`
	Title              string         `json:"title"`
	Type               string         `json:"type"`
	XPGrant            int            `json:"xpGrant"`
	ProgressToComplete int            `json:"progressToComplete"`
	Objectives         map[string]int `json:"objectives"` // objectiveUuid -> value
}

// Built-in fallback definitions for common daily/weekly missions
var DefaultMissions = map[string]MissionInfo{
	"5163c5f2-4c28-9844-3d07-2ebdb22f98e6": {Title: "Get Headshots", ProgressToComplete: 5, XPGrant: 2000},
	"f3e5cfb8-4682-1402-995b-2bb548483f81": {Title: "Deal Damage", ProgressToComplete: 1000, XPGrant: 2000},
	"69502152-4467-36c5-8494-b1b51e065757": {Title: "Use Your Ultimate", ProgressToComplete: 5, XPGrant: 2000},
}

type missionsCache struct {
	Version  string                 `json:"version"`
	Missions map[string]MissionInfo `json:"missions"`
	CachedAt time.Time              `json:"cachedAt"`
}

type missionsAPIResponse struct {
	Status int `json:"status"`
	Data   []struct {
		UUID               string  `json:"uuid"`
		Title              *string `json:"title"`
		Type               *string `json:"type"`
		XPGrant            int     `json:"xpGrant"`
		ProgressToComplete int     `json:"progressToComplete"`
		Objectives         []struct {
			ObjectiveUUID string `json:"objectiveUuid"`
			Value         int    `json:"value"`
		} `json:"objectives"`
	} `json:"data"`
}

func LoadOrFetchMissions(currentVersion string) (map[string]MissionInfo, error) {
	dir, err := cacheDir()
	if err != nil {
		return DefaultMissions, nil
	}
	path := filepath.Join(dir, "missions.json")

	if data, err := os.ReadFile(path); err == nil {
		var mc missionsCache
		if json.Unmarshal(data, &mc) == nil && mc.Missions != nil && len(mc.Missions) > 0 {
			if currentVersion == "" || mc.Version == currentVersion {
				return mc.Missions, nil
			}
		}
	}

	missions, err := fetchMissions()
	if err != nil {
		return DefaultMissions, nil
	}

	mc := missionsCache{Version: currentVersion, Missions: missions, CachedAt: time.Now()}
	if data, err := json.MarshalIndent(mc, "", "  "); err == nil {
		_ = os.WriteFile(path, data, 0o644)
	}

	return missions, nil
}

func fetchMissions() (map[string]MissionInfo, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(missionsURL)
	if err != nil {
		return nil, fmt.Errorf("fetch missions: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mar missionsAPIResponse
	if err := json.Unmarshal(body, &mar); err != nil {
		return nil, err
	}

	missions := make(map[string]MissionInfo, len(mar.Data))
	for _, m := range mar.Data {
		title := "Mission"
		if m.Title != nil && *m.Title != "" {
			title = *m.Title
		}
		mType := ""
		if m.Type != nil {
			mType = *m.Type
		}
		objs := make(map[string]int)
		for _, o := range m.Objectives {
			objs[strings.ToLower(o.ObjectiveUUID)] = o.Value
		}

		missions[strings.ToLower(m.UUID)] = MissionInfo{
			UUID:               m.UUID,
			Title:              title,
			Type:               mType,
			XPGrant:            m.XPGrant,
			ProgressToComplete: m.ProgressToComplete,
			Objectives:         objs,
		}
	}

	if len(missions) == 0 {
		return DefaultMissions, nil
	}
	return missions, nil
}
