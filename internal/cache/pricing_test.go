package cache

import (
	"testing"

	"github.com/val-tracker/val-tracker/internal/models"
)

func TestIsMeleeSkin(t *testing.T) {
	tests := []struct {
		name      string
		assetPath string
		dispName  string
		expected  bool
	}{
		{
			name:      "Melee by AssetPath equippables/melee",
			assetPath: "ShooterGame/Content/Equippables/Melee/Scythe/Scythe_Asset",
			dispName:  "Soulstrife Scythe",
			expected:  true,
		},
		{
			name:      "Melee by AssetPath /melee/",
			assetPath: "ShooterGame/Content/Items/Melee/Dagger",
			dispName:  "Custom Dagger",
			expected:  true,
		},
		{
			name:      "Melee by DisplayName keyword knife",
			assetPath: "",
			dispName:  "Reaver Knife",
			expected:  true,
		},
		{
			name:      "Melee by DisplayName keyword blade",
			assetPath: "",
			dispName:  "Broken Blade of the Ruined King",
			expected:  true,
		},
		{
			name:      "Melee by DisplayName keyword splitter",
			assetPath: "",
			dispName:  "Phaseguard Splitter",
			expected:  true,
		},
		{
			name:      "Melee by DisplayName keyword karambit",
			assetPath: "",
			dispName:  "Champions 2021 Karambit",
			expected:  true,
		},
		{
			name:      "Melee by DisplayName keyword fist",
			assetPath: "",
			dispName:  "Power Fist",
			expected:  true,
		},
		{
			name:      "Gun Vandal",
			assetPath: "ShooterGame/Content/Equippables/Guns/Rifles/Vandal/Vandal_Asset",
			dispName:  "Reaver Vandal",
			expected:  false,
		},
		{
			name:      "Gun Phantom",
			assetPath: "ShooterGame/Content/Equippables/Guns/Rifles/Phantom/Phantom_Asset",
			dispName:  "Prime//2.0 Phantom",
			expected:  false,
		},
		{
			name:      "Gun Ghost",
			assetPath: "ShooterGame/Content/Equippables/Guns/Pistols/Luger/Luger_Asset",
			dispName:  "Soulstrife Ghost",
			expected:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := models.SkinAsset{
				DisplayName: tc.dispName,
				AssetPath:   tc.assetPath,
			}
			res := IsMeleeSkin(s)
			if res != tc.expected {
				t.Errorf("%s: expected IsMeleeSkin=%v, got %v", tc.name, tc.expected, res)
			}
		})
	}
}

func TestResolveSkinPrice_Melee(t *testing.T) {
	premiumTier := models.TierPremiumUUID
	exclusiveTier := models.TierExclusiveUUID
	themeUUID := "soulstrife-theme-uuid"

	skins := []models.SkinAsset{
		{
			UUID:            "soulstrife-ghost-uuid",
			DisplayName:     "Soulstrife Ghost",
			ThemeUUID:       themeUUID,
			ContentTierUUID: &premiumTier,
			AssetPath:       "ShooterGame/Content/Equippables/Guns/Pistols/Luger/Luger_Asset",
		},
		{
			UUID:            "soulstrife-scythe-uuid",
			DisplayName:     "Soulstrife Scythe",
			ThemeUUID:       themeUUID,
			ContentTierUUID: &exclusiveTier, // Riot API assigns Exclusive by default
			AssetPath:       "ShooterGame/Content/Equippables/Melee/Scythe/Scythe_Asset",
		},
	}

	themeGunTiers := BuildThemeGunTierMap(skins)

	// 1. Soulstrife Scythe inherits Premium from Soulstrife Ghost -> 3,550 VP
	scythe := skins[1]
	tier := ResolveSkinTierUUID(scythe, themeGunTiers)
	if tier != models.TierPremiumUUID {
		t.Errorf("expected Scythe to inherit Premium tier, got: %s", tier)
	}
	rarity := ResolveSkinRarity(scythe, themeGunTiers)
	if rarity != "Premium" {
		t.Errorf("expected Scythe rarity to be Premium, got: %s", rarity)
	}
	price := ResolveSkinPrice(scythe, themeGunTiers)
	if price != 3550 {
		t.Errorf("expected Soulstrife Scythe price to be 3550 VP, got: %d", price)
	}

	// 2. Curated override: Phaseguard Splitter -> 5,350 VP
	phaseguard := models.SkinAsset{
		UUID:            "phaseguard-splitter-uuid",
		DisplayName:     "Phaseguard Splitter",
		ContentTierUUID: &exclusiveTier,
		AssetPath:       "ShooterGame/Content/Equippables/Melee/Splitter/Splitter_Asset",
	}
	pgPrice := ResolveSkinPrice(phaseguard, nil)
	if pgPrice != 5350 {
		t.Errorf("expected Phaseguard Splitter to resolve to 5350 VP, got: %d", pgPrice)
	}

	// 3. Curated override: Broken Blade of the Ruined King -> 4,350 VP
	ruinedBlade := models.SkinAsset{
		UUID:            "ruined-blade-uuid",
		DisplayName:     "Broken Blade of the Ruined King",
		ContentTierUUID: &exclusiveTier,
		AssetPath:       "ShooterGame/Content/Equippables/Melee/Sword/Sword_Asset",
	}
	bladePrice := ResolveSkinPrice(ruinedBlade, nil)
	if bladePrice != 4350 {
		t.Errorf("expected Broken Blade of the Ruined King to resolve to 4350 VP, got: %d", bladePrice)
	}
}

func TestResolveSkinPrice_Gun(t *testing.T) {
	selectTier := models.TierSelectUUID
	deluxeTier := models.TierDeluxeUUID
	premiumTier := models.TierPremiumUUID
	exclusiveTier := models.TierExclusiveUUID
	ultraTier := models.TierUltraUUID

	cases := []struct {
		name     string
		skin     models.SkinAsset
		expPrice int
		expTier  string
	}{
		{
			name: "Reaver Vandal (Premium)",
			skin: models.SkinAsset{
				UUID:            "reaver-vandal-uuid",
				DisplayName:     "Reaver Vandal",
				ContentTierUUID: &premiumTier,
				AssetPath:       "ShooterGame/Content/Equippables/Guns/Rifles/Vandal/Reaver_Asset",
			},
			expPrice: 1775,
			expTier:  "Premium",
		},
		{
			name: "Prime Phantom (Premium)",
			skin: models.SkinAsset{
				UUID:            "prime-phantom-uuid",
				DisplayName:     "Prime//2.0 Phantom",
				ContentTierUUID: &premiumTier,
				AssetPath:       "ShooterGame/Content/Equippables/Guns/Rifles/Phantom/Prime_Asset",
			},
			expPrice: 1775,
			expTier:  "Premium",
		},
		{
			name: "Select Gun",
			skin: models.SkinAsset{
				UUID:            "select-gun-uuid",
				DisplayName:     "Galleria Classic",
				ContentTierUUID: &selectTier,
				AssetPath:       "ShooterGame/Content/Equippables/Guns/Pistols/Base/Galleria_Asset",
			},
			expPrice: 875,
			expTier:  "Select",
		},
		{
			name: "Deluxe Gun",
			skin: models.SkinAsset{
				UUID:            "deluxe-gun-uuid",
				DisplayName:     "Aristocrat Vandal",
				ContentTierUUID: &deluxeTier,
				AssetPath:       "ShooterGame/Content/Equippables/Guns/Rifles/Vandal/Aristocrat_Asset",
			},
			expPrice: 1275,
			expTier:  "Deluxe",
		},
		{
			name: "Exclusive Gun",
			skin: models.SkinAsset{
				UUID:            "exclusive-gun-uuid",
				DisplayName:     "Glitchpop Vandal",
				ContentTierUUID: &exclusiveTier,
				AssetPath:       "ShooterGame/Content/Equippables/Guns/Rifles/Vandal/Glitchpop_Asset",
			},
			expPrice: 2175,
			expTier:  "Exclusive",
		},
		{
			name: "Ultra Gun",
			skin: models.SkinAsset{
				UUID:            "ultra-gun-uuid",
				DisplayName:     "Elderflame Vandal",
				ContentTierUUID: &ultraTier,
				AssetPath:       "ShooterGame/Content/Equippables/Guns/Rifles/Vandal/Elderflame_Asset",
			},
			expPrice: 2475,
			expTier:  "Ultra",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			price := ResolveSkinPrice(tc.skin, nil)
			if price != tc.expPrice {
				t.Errorf("%s: expected price %d, got %d", tc.name, tc.expPrice, price)
			}
			rarity := ResolveSkinRarity(tc.skin, nil)
			if rarity != tc.expTier {
				t.Errorf("%s: expected rarity %s, got %s", tc.name, tc.expTier, rarity)
			}
		})
	}
}

func TestLivePriceCache(t *testing.T) {
	testUUID := "test-live-price-skin-uuid"
	testPrice := 2175

	RecordSkinPrice(testUUID, testPrice)
	cached := GetCachedSkinPrice(testUUID)
	if cached != testPrice {
		t.Fatalf("expected cached price %d, got: %d", testPrice, cached)
	}

	// Layer 1 test: live price cache overrides formula calculation
	premiumTier := models.TierPremiumUUID
	skin := models.SkinAsset{
		UUID:            testUUID,
		DisplayName:     "Discounted Special Skin",
		ContentTierUUID: &premiumTier, // Normally 1775
		AssetPath:       "ShooterGame/Content/Equippables/Guns/Rifles/Vandal/Special_Asset",
	}

	resolved := ResolveSkinPrice(skin, nil)
	if resolved != testPrice {
		t.Errorf("expected Layer 1 cached price %d to take precedence over formula 1775, got %d", testPrice, resolved)
	}
}
