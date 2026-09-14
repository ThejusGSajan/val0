package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/ThejusGSajan/val0/internal/models"
)

var (
	wishlistMutex  sync.Mutex
	wishlistCache  []WishlistEntry
	wishlistLoaded bool
)

type WishlistEntry struct {
	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	Rarity  string `json:"rarity"`
	CostVP  int    `json:"costVP"`
	IconURL string `json:"iconURL"`
}

type wishlistFile struct {
	Skins []WishlistEntry `json:"skins"`
}

func wishlistPath() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "wishlist.json"), nil
}

// EnsureLoaded loads the wishlist into memory once.
func EnsureLoaded() {
	wishlistMutex.Lock()
	defer wishlistMutex.Unlock()
	if !wishlistLoaded {
		wishlistCache, _ = loadFromDisk()
		wishlistLoaded = true
	}
}

// InvalidateCache forces a reload on next access.
func InvalidateCache() {
	wishlistMutex.Lock()
	defer wishlistMutex.Unlock()
	wishlistLoaded = false
	wishlistCache = nil
}

func loadFromDisk() ([]WishlistEntry, error) {
	path, err := wishlistPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil // empty if doesn't exist
	}

	var wf wishlistFile
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, err
	}
	return wf.Skins, nil
}

// LoadWishlist reads the saved wishlist from disk (or cached memory).
func LoadWishlist() ([]WishlistEntry, error) {
	EnsureLoaded()
	wishlistMutex.Lock()
	defer wishlistMutex.Unlock()
	entries := make([]WishlistEntry, len(wishlistCache))
	copy(entries, wishlistCache)
	return entries, nil
}

// SaveWishlist writes the entire wishlist to disk.
func SaveWishlist(entries []WishlistEntry) error {
	wishlistMutex.Lock()
	defer wishlistMutex.Unlock()

	path, err := wishlistPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	wf := wishlistFile{Skins: entries}
	data, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	wishlistCache = entries
	wishlistLoaded = true
	return nil
}

// AddToWishlist adds a skin entry to the wishlist if not already present.
func AddToWishlist(entry WishlistEntry) error {
	entries, _ := LoadWishlist()
	for _, e := range entries {
		if e.UUID == entry.UUID {
			return nil // already in wishlist
		}
	}
	entries = append(entries, entry)
	err := SaveWishlist(entries)
	InvalidateCache()
	return err
}

// RemoveFromWishlist removes a skin by UUID from the wishlist.
func RemoveFromWishlist(uuid string) error {
	entries, _ := LoadWishlist()
	var updated []WishlistEntry
	for _, e := range entries {
		if e.UUID != uuid {
			updated = append(updated, e)
		}
	}
	err := SaveWishlist(updated)
	InvalidateCache()
	return err
}

// IsInWishlist returns true if a skin UUID is saved in the wishlist.
func IsInWishlist(uuid string) bool {
	EnsureLoaded()
	wishlistMutex.Lock()
	defer wishlistMutex.Unlock()
	for _, e := range wishlistCache {
		if e.UUID == uuid {
			return true
		}
	}
	return false
}

// ConvertResolvedSkinToWishlist helper.
func ConvertResolvedSkinToWishlist(s models.ResolvedSkin) WishlistEntry {
	return WishlistEntry{
		UUID:    s.UUID,
		Name:    s.DisplayName,
		Rarity:  s.Rarity,
		CostVP:  s.CostVP,
		IconURL: s.IconURL,
	}
}
