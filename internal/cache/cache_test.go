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
