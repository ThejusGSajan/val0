package cache

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDefaultMissionsFallback(t *testing.T) {
	if len(DefaultMissions) == 0 {
		t.Fatal("expected DefaultMissions to contain entries")
	}

	headshotsUUID := "5163c5f2-4c28-9844-3d07-2ebdb22f98e6"
	m, ok := DefaultMissions[headshotsUUID]
	if !ok {
		t.Fatalf("expected headshots mission %s in DefaultMissions", headshotsUUID)
	}

	if m.Title != "Get Headshots" {
		t.Errorf("expected Title 'Get Headshots', got '%s'", m.Title)
	}
	if m.ProgressToComplete != 5 {
		t.Errorf("expected ProgressToComplete 5, got %d", m.ProgressToComplete)
	}
	if m.XPGrant != 2000 {
		t.Errorf("expected XPGrant 2000, got %d", m.XPGrant)
	}
}

func TestMissionsAPIResponseUnmarshal(t *testing.T) {
	rawJSON := `{
		"status": 200,
		"data": [
			{
				"uuid": "5163c5f2-4c28-9844-3d07-2ebdb22f98e6",
				"title": "Get Headshots",
				"type": "EAresMissionType::Daily",
				"xpGrant": 2000,
				"progressToComplete": 5,
				"objectives": [
					{
						"objectiveUuid": "04ff6167-4d76-8051-789a-dc853b05f23d",
						"value": 5
					}
				]
			}
		]
	}`

	var mar missionsAPIResponse
	if err := json.Unmarshal([]byte(rawJSON), &mar); err != nil {
		t.Fatalf("failed to unmarshal missions API response: %v", err)
	}

	if len(mar.Data) != 1 {
		t.Fatalf("expected 1 mission item, got %d", len(mar.Data))
	}

	item := mar.Data[0]
	if item.Title == nil || *item.Title != "Get Headshots" {
		t.Errorf("expected title 'Get Headshots', got %v", item.Title)
	}
	if len(item.Objectives) != 1 {
		t.Fatalf("expected 1 objective, got %d", len(item.Objectives))
	}
	if strings.ToLower(item.Objectives[0].ObjectiveUUID) != "04ff6167-4d76-8051-789a-dc853b05f23d" {
		t.Errorf("unexpected objective UUID: %s", item.Objectives[0].ObjectiveUUID)
	}
}
