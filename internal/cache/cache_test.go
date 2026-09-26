package cache

import (
	"testing"

	"github.com/ThejusGSajan/val0/internal/models"
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

func TestWishlistOperations(t *testing.T) {
	entry := WishlistEntry{
		UUID:   "test-uuid-prime-vandal",
		Name:   "Prime Vandal",
		CostVP: 1775,
		Rarity: "Premium",
	}

	// Clean before test
	_ = RemoveFromWishlist(entry.UUID)

	if IsInWishlist(entry.UUID) {
		t.Errorf("expected false before adding")
	}

	if err := AddToWishlist(entry); err != nil {
		t.Fatalf("failed to add to wishlist: %v", err)
	}

	if !IsInWishlist(entry.UUID) {
		t.Errorf("expected true after adding to wishlist")
	}

	if err := RemoveFromWishlist(entry.UUID); err != nil {
		t.Fatalf("failed to remove from wishlist: %v", err)
	}

	if IsInWishlist(entry.UUID) {
		t.Errorf("expected false after removing from wishlist")
	}
}

func TestGetWeaponName(t *testing.T) {
	// Test Outlaw and Warden default fallbacks
	if name := GetWeaponName("5f0786ac-4366-2d39-96bd-2586c6734f07", nil); name != "Outlaw" {
		t.Errorf("expected Outlaw, got %s", name)
	}
	if name := GetWeaponName("8db0a1bf-4a50-832a-4566-faaaa6d250ca", nil); name != "Warden" {
		t.Errorf("expected Warden, got %s", name)
	}

	// Test uppercase UUID handling
	if name := GetWeaponName("9C82E19D-4575-0200-1A81-3EACF00CF872", nil); name != "Vandal" {
		t.Errorf("expected Vandal for uppercase UUID, got %s", name)
	}

	// Test custom weapons map override
	customMap := map[string]string{
		"custom-uuid-1": "Custom Blaster",
	}
	if name := GetWeaponName("custom-uuid-1", customMap); name != "Custom Blaster" {
		t.Errorf("expected Custom Blaster, got %s", name)
	}

	// Test unknown weapon fallback
	if name := GetWeaponName("00000000-0000-0000-0000-000000000000", nil); name != "Weapon" {
		t.Errorf("expected Weapon for unknown UUID, got %s", name)
	}
}

