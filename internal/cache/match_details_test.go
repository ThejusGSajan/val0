package cache

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ThejusGSajan/val0/internal/models"
)

func TestMatchDetailsCache(t *testing.T) {
	testID := "test-cache-match-123"
	md := &models.MatchDetails{
		MatchInfo: models.MatchInfo{MatchID: testID, QueueID: "competitive"},
	}

	// 1. Initial lookup -> miss
	if _, ok := GetCachedMatchDetails(testID); ok {
		t.Error("expected initial cache miss")
	}

	// 2. Save
	if err := SaveCachedMatchDetails(md); err != nil {
		t.Fatalf("failed to save cache: %v", err)
	}

	// Clean up disk after test
	defer func() {
		p, err := matchCacheFilePath(testID)
		if err == nil {
			_ = os.Remove(p)
		}
	}()

	// 3. Lookup -> hit
	cached, ok := GetCachedMatchDetails(testID)
	if !ok || cached == nil {
		t.Fatalf("expected cache hit after save")
	}
	if cached.MatchInfo.MatchID != testID {
		t.Errorf("expected match ID %s, got %s", testID, cached.MatchInfo.MatchID)
	}

	// 4. Empty ID safety
	if _, ok := GetCachedMatchDetails(""); ok {
		t.Error("expected empty matchID lookup to return false")
	}
	if err := SaveCachedMatchDetails(nil); err != nil {
		t.Errorf("expected nil save to return nil, got %v", err)
	}
}

func TestMatchDetailsCache_Eviction(t *testing.T) {
	// Add more than MaxCacheSize items
	numItems := MaxCacheSize + 5
	createdIDs := make([]string, numItems)

	for i := 0; i < numItems; i++ {
		id := fmt.Sprintf("test-evict-match-%03d", i)
		createdIDs[i] = id
		md := &models.MatchDetails{
			MatchInfo: models.MatchInfo{MatchID: id, QueueID: "spikerush"},
		}
		if err := SaveCachedMatchDetails(md); err != nil {
			t.Fatalf("failed to save match %s: %v", id, err)
		}
	}

	defer func() {
		for _, id := range createdIDs {
			p, err := matchCacheFilePath(id)
			if err == nil {
				_ = os.Remove(p)
			}
		}
	}()

	// Allow async disk evictions to complete
	time.Sleep(100 * time.Millisecond)

	// In memory, oldest items (e.g. createdIDs[0]) should have been evicted from memMatchCache
	memMatchCacheMu.RLock()
	oldestID := createdIDs[0]
	_, inMem := memMatchCache[oldestID]
	memMatchCacheMu.RUnlock()

	if inMem {
		t.Errorf("expected %s to be evicted from in-memory cache", oldestID)
	}

	// On disk, oldest items should also be evicted since total > MaxCacheSize
	if _, ok := GetCachedMatchDetails(oldestID); ok {
		t.Errorf("expected %s to also be evicted from disk cache after exceeding MaxCacheSize", oldestID)
	}

	// For a retained item, evict it from memory to test disk reload
	retainedID := createdIDs[numItems-1]
	memMatchCacheMu.Lock()
	delete(memMatchCache, retainedID)
	memMatchCacheMu.Unlock()

	cached, ok := GetCachedMatchDetails(retainedID)
	if !ok || cached == nil {
		t.Errorf("expected %s to reload from disk cache", retainedID)
	}
}
