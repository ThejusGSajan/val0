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

const weaponsURL = "https://valorant-api.com/v1/weapons"

// DefaultWeaponNames provides built-in fallback mapping for Valorant weapons.
var DefaultWeaponNames = map[string]string{
	"ae3de142-4275-51ad-7780-9794a9a12ee1": "Classic",
	"e3367e9b-4217-aeab-3b20-57f404a09a40": "Shorty",
	"42da8c32-4b2d-ab7c-d911-a010e6a31d57": "Frenzy",
	"4492f7f1-4887-00ce-b603-fb8235b8e052": "Ghost",
	"1baa85b4-4c70-1284-64bb-6481dfc3bb4e": "Sheriff",
	"f7e1b454-4ad4-1609-41b9-07e7d477b812": "Stinger",
	"462080d1-4035-2937-7c09-27aa2a5c27a7": "Spectre",
	"910be178-4f55-5ca8-69f6-509cc2a83ac4": "Bucky",
	"ec84293c-4576-360a-0147-f28180c45da2": "Judge",
	"ae12384a-4a6f-a888-8f80-21880e2d4803": "Bulldog",
	"4ade7faa-4cf1-8376-7ab3-0d60174e0259": "Guardian",
	"ee8e8d15-496b-07ac-e5f6-8fae5d4c7b1a": "Phantom",
	"9c82e19d-4575-0200-1a81-3eacf00cf872": "Vandal",
	"c4883e50-4494-202c-3ec3-6b8a9284f00b": "Marshal",
	"55d8a0f4-4274-ca67-fe2c-06ab45efdf58": "Outlaw",
	"a03b24d3-4319-996d-0f8c-94bbfba1dfc7": "Operator",
	"55d8a0f4-4274-ca67-fe2c-06ab45efdf50": "Ares",
	"63e6c2b6-4a8e-869c-3d4c-e38355226584": "Odin",
	"2f59173c-4edd-c673-0986-40a2a023bd2e": "Melee",
}

type weaponsCache struct {
	Version  string            `json:"version"`
	Weapons  map[string]string `json:"weapons"`
	CachedAt time.Time         `json:"cachedAt"`
}

type weaponsAPIResponse struct {
	Status int `json:"status"`
	Data   []struct {
		UUID        string `json:"uuid"`
		DisplayName string `json:"displayName"`
	} `json:"data"`
}

// LoadOrFetchWeapons returns map from weapon UUID -> DisplayName.
func LoadOrFetchWeapons(currentVersion string) (map[string]string, error) {
	dir, err := cacheDir()
	if err != nil {
		return DefaultWeaponNames, nil
	}
	path := filepath.Join(dir, "weapons.json")

	if data, err := os.ReadFile(path); err == nil {
		var wc weaponsCache
		if json.Unmarshal(data, &wc) == nil && wc.Weapons != nil && len(wc.Weapons) > 0 {
			if currentVersion == "" || wc.Version == currentVersion {
				return wc.Weapons, nil
			}
		}
	}

	weapons, err := fetchWeapons()
	if err != nil {
		return DefaultWeaponNames, nil
	}

	wc := weaponsCache{Version: currentVersion, Weapons: weapons, CachedAt: time.Now()}
	if data, err := json.MarshalIndent(wc, "", "  "); err == nil {
		_ = os.WriteFile(path, data, 0o644)
	}

	return weapons, nil
}

func fetchWeapons() (map[string]string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(weaponsURL)
	if err != nil {
		return nil, fmt.Errorf("fetch weapons: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var wr weaponsAPIResponse
	if err := json.Unmarshal(body, &wr); err != nil {
		return nil, err
	}

	weapons := make(map[string]string, len(wr.Data))
	for _, w := range wr.Data {
		weapons[strings.ToLower(w.UUID)] = w.DisplayName
	}

	if len(weapons) == 0 {
		return DefaultWeaponNames, nil
	}
	return weapons, nil
}

// GetWeaponName resolves weapon UUID to human-friendly name.
func GetWeaponName(uuid string, weaponsMap map[string]string) string {
	uuidLower := strings.ToLower(uuid)
	if weaponsMap != nil {
		if name, ok := weaponsMap[uuidLower]; ok && name != "" {
			return name
		}
	}
	if name, ok := DefaultWeaponNames[uuidLower]; ok {
		return name
	}
	return "Weapon"
}
