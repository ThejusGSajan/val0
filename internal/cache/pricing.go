package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/val-tracker/val-tracker/internal/models"
)

var (
	priceMutex  sync.RWMutex
	priceCache  map[string]int
	priceLoaded bool
)

// KnownMeleePrices maps curated special melee weapons to their official in-game VP costs.
var KnownMeleePrices = map[string]int{
	"Phaseguard Splitter":               5350,
	"Kuronami no Yaiba":                 5350,
	"VCT LOCK//IN Misericórdia":         5440,
	"VCT Champions 2021 Karambit":       5350,
	"VCT Champions 2022 Butterfly Knife": 5350,
	"VCT Champions 2023 Kunai":          5350,
	"VCT Champions 2024 Blade":          5350,
	"Power Fist":                        5950,
	"Ignite Fan":                        4710,
	"Arcane Melee":                      4350,
	"Broken Blade of the Ruined King":   4350,
	"Araxys Bio Harvester":              4350,
	"Prelude to Chaos Stinger":          4350,
	"Mystbloom Kunai":                   4350,
	"Singularity Knife":                 4350,
	"Oni Katana":                        5350,
	"Evori's Spellcaster":               4950,
	"Neo Frontier Axe":                  4350,
}

// IsMeleeSkin determines whether a skin asset represents a melee weapon.
func IsMeleeSkin(s models.SkinAsset) bool {
	lowerPath := strings.ToLower(s.AssetPath)
	if strings.Contains(lowerPath, "equippables/melee") || strings.Contains(lowerPath, "/melee/") {
		return true
	}
	// Fallback on DisplayName keywords
	name := strings.ToLower(s.DisplayName)
	keywords := []string{
		"knife", "blade", "scythe", "splitter", "dagger", "sword", "axe",
		"karambit", "fist", "mace", "baton", "wand", "hammer", "claw",
		"kunai", "katana", "relic stone", "comb", "cane", "bat",
	}
	for _, kw := range keywords {
		if strings.Contains(name, kw) {
			return true
		}
	}
	return false
}

// BuildThemeGunTierMap creates a mapping from ThemeUUID to the gun contentTierUUID.
// Melee skins are excluded to prevent them from overwriting the bundle's base gun tier.
func BuildThemeGunTierMap(skins []models.SkinAsset) map[string]string {
	themeMap := make(map[string]string)
	for _, s := range skins {
		if s.ThemeUUID == "" || s.ContentTierUUID == nil || *s.ContentTierUUID == "" {
			continue
		}
		if !IsMeleeSkin(s) {
			themeMap[s.ThemeUUID] = *s.ContentTierUUID
		}
	}
	return themeMap
}

// ResolveSkinTierUUID returns the effective content tier UUID for a skin asset.
// For melee skins, it attempts to inherit the base gun tier from the parent collection/theme.
func ResolveSkinTierUUID(s models.SkinAsset, themeGunTiers map[string]string) string {
	if IsMeleeSkin(s) && themeGunTiers != nil {
		if parentTier, ok := themeGunTiers[s.ThemeUUID]; ok && parentTier != "" {
			return parentTier
		}
	}
	if s.ContentTierUUID != nil {
		return *s.ContentTierUUID
	}
	return ""
}

// ResolveSkinRarity returns the human-readable rarity name for a skin.
func ResolveSkinRarity(s models.SkinAsset, themeGunTiers map[string]string) string {
	tierUUID := ResolveSkinTierUUID(s, themeGunTiers)
	switch tierUUID {
	case models.TierSelectUUID, models.TierSelectUUIDAlt:
		return "Select"
	case models.TierDeluxeUUID, models.TierDeluxeUUIDAlt:
		return "Deluxe"
	case models.TierPremiumUUID:
		return "Premium"
	case models.TierExclusiveUUID:
		return "Exclusive"
	case models.TierUltraUUID, models.TierUltraUUIDAlt:
		return "Ultra"
	default:
		return ""
	}
}

func pricesPath() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "prices.json"), nil
}

type pricesFile struct {
	Prices map[string]int `json:"prices"`
}

func ensurePricesLoaded() {
	priceMutex.Lock()
	defer priceMutex.Unlock()
	if priceLoaded {
		return
	}
	priceCache = make(map[string]int)
	priceLoaded = true

	path, err := pricesPath()
	if err != nil {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var pf pricesFile
	if err := json.Unmarshal(data, &pf); err == nil && pf.Prices != nil {
		priceCache = pf.Prices
	}
}

// RecordSkinPrice persists an authoritative skin price (e.g. from live Riot storefront) to prices.json.
func RecordSkinPrice(uuid string, price int) {
	if uuid == "" || price <= 0 {
		return
	}
	ensurePricesLoaded()

	priceMutex.Lock()
	defer priceMutex.Unlock()

	if oldPrice, exists := priceCache[uuid]; exists && oldPrice == price {
		return
	}
	priceCache[uuid] = price

	if path, err := pricesPath(); err == nil {
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		pf := pricesFile{Prices: priceCache}
		if data, err := json.MarshalIndent(pf, "", "  "); err == nil {
			_ = os.WriteFile(path, data, 0o644)
		}
	}
}

// GetCachedSkinPrice retrieves a recorded live store price from memory/disk cache.
func GetCachedSkinPrice(uuid string) int {
	if uuid == "" {
		return 0
	}
	ensurePricesLoaded()

	priceMutex.RLock()
	defer priceMutex.RUnlock()
	return priceCache[uuid]
}

// ResolveSkinPrice implements the 5-layer pricing resolution engine.
func ResolveSkinPrice(s models.SkinAsset, themeGunTiers map[string]string) int {
	// Layer 1: Live Riot store price cache
	if p := GetCachedSkinPrice(s.UUID); p > 0 {
		return p
	}
	// Layer 2: Curated special melee overrides
	if p, ok := KnownMeleePrices[s.DisplayName]; ok {
		return p
	}
	// Layer 3 & 4: Formulaic calculation
	tierUUID := ResolveSkinTierUUID(s, themeGunTiers)
	isMelee := IsMeleeSkin(s)
	switch tierUUID {
	case models.TierSelectUUID, models.TierSelectUUIDAlt:
		if isMelee {
			return 1750
		}
		return 875
	case models.TierDeluxeUUID, models.TierDeluxeUUIDAlt:
		if isMelee {
			return 2550
		}
		return 1275
	case models.TierPremiumUUID:
		if isMelee {
			return 3550
		}
		return 1775
	case models.TierExclusiveUUID:
		if isMelee {
			return 4350
		}
		return 2175
	case models.TierUltraUUID, models.TierUltraUUIDAlt:
		if isMelee {
			return 4950
		}
		return 2475
	}
	return 0
}
