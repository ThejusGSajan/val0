package cache

import (
	"testing"

	"github.com/val-tracker/val-tracker/internal/models"
)

func TestBuildSkinLookup(t *testing.T) {
	levelUUID := "level-123"
	skinUUID := "skin-456"
	tierUUID := "tier-789"

	skins := []models.SkinAsset{
		{
			UUID:            skinUUID,
			DisplayName:     "Prime Vandal",
			ContentTierUUID: &tierUUID,
			Levels: []models.SkinLevel{
				{
					UUID: levelUUID,
				},
			},
		},
	}

	lookup := BuildSkinLookup(skins)

	// Lookup by skin UUID
	if s, ok := lookup[skinUUID]; !ok || s.DisplayName != "Prime Vandal" {
		t.Errorf("expected Prime Vandal by skin UUID, got: %+v", s)
	}

	// Lookup by level UUID (storefront items use level UUIDs)
	if s, ok := lookup[levelUUID]; !ok || s.DisplayName != "Prime Vandal" {
		t.Errorf("expected Prime Vandal by level UUID, got: %+v", s)
	}
}

func TestGetRankName(t *testing.T) {
	ranksMap := map[int]string{
		21: "Ascendant 1",
		27: "Radiant",
	}

	if name := GetRankName(21, ranksMap); name != "Ascendant 1" {
		t.Errorf("expected Ascendant 1, got %s", name)
	}
	if name := GetRankName(19, ranksMap); name != "Diamond 2" {
		t.Errorf("expected Diamond 2 from fallback, got %s", name)
	}
	if name := GetRankName(0, nil); name != "Unranked" {
		t.Errorf("expected Unranked for tier 0, got %s", name)
	}
}
