package cache

import (
	"testing"

	"github.com/val-tracker/val-tracker/internal/models"
)

func TestCalculateBattlepassTier(t *testing.T) {
	levelXPs := DefaultBattlepassLevelXPs()
	if len(levelXPs) != 55 {
		t.Fatalf("expected 55 levels in default curve, got %d", len(levelXPs))
	}

	// 1. Tier 1 at 0 XP (level 0 completed for free, progressing towards Tier 2)
	curTier, xpInTier, xpForTier, completed := CalculateBattlepassTier(0, levelXPs)
	if curTier != 2 {
		t.Errorf("at 0 XP: expected curTier=2 (progressing to tier 2), got %d", curTier)
	}
	if xpInTier != 0 {
		t.Errorf("at 0 XP: expected xpInTier=0, got %d", xpInTier)
	}
	if xpForTier != 2000 {
		t.Errorf("at 0 XP: expected xpForTier=2000, got %d", xpForTier)
	}
	if completed {
		t.Errorf("at 0 XP: expected completed=false")
	}

	// 2. Mid-tier progression: 1000 XP (halfway to Tier 2)
	curTier, xpInTier, xpForTier, completed = CalculateBattlepassTier(1000, levelXPs)
	if curTier != 2 || xpInTier != 1000 || xpForTier != 2000 || completed {
		t.Errorf("at 1000 XP: expected tier 2 (1000/2000), got tier %d (%d/%d, completed=%v)", curTier, xpInTier, xpForTier, completed)
	}

	// 3. Exactly 2000 XP: Tier 2 completed, progressing to Tier 3 (which requires 2750 XP)
	curTier, xpInTier, xpForTier, completed = CalculateBattlepassTier(2000, levelXPs)
	if curTier != 3 {
		t.Errorf("at 2000 XP: expected curTier=3, got %d", curTier)
	}
	if xpInTier != 0 {
		t.Errorf("at 2000 XP: expected xpInTier=0, got %d", xpInTier)
	}
	if xpForTier != 2750 {
		t.Errorf("at 2000 XP: expected xpForTier=2750, got %d", xpForTier)
	}
	if completed {
		t.Errorf("at 2000 XP: expected completed=false")
	}

	// 4. Completed pass: > 1,157,500 XP
	curTier, xpInTier, xpForTier, completed = CalculateBattlepassTier(1200000, levelXPs)
	if curTier != 55 {
		t.Errorf("completed pass: expected curTier=55, got %d", curTier)
	}
	if !completed {
		t.Errorf("completed pass: expected completed=true")
	}
	if xpInTier != 36500 || xpForTier != 36500 {
		t.Errorf("completed pass: expected xpInTier=36500, xpForTier=36500, got %d, %d", xpInTier, xpForTier)
	}
}

func TestFindBattlepassContractAsset(t *testing.T) {
	activeSeasonID := "8102cd81-43a0-d0d7-bd59-47b8fe9bed1b"
	contracts := []models.ContractAsset{
		{
			UUID:        "agent-contract-uuid",
			DisplayName: "Jett Contract",
			Content: models.ContractAssetContent{
				RelationType: "Agent",
				RelationUUID: "agent-jett-uuid",
			},
		},
		{
			UUID:        "bp-contract-uuid",
			DisplayName: "Season 2026 // Act V",
			Content: models.ContractAssetContent{
				RelationType: "Season",
				RelationUUID: activeSeasonID,
			},
		},
	}

	found := FindBattlepassContractAsset(contracts, activeSeasonID)
	if found == nil {
		t.Fatalf("expected to find battlepass asset for season %s", activeSeasonID)
	}
	if found.UUID != "bp-contract-uuid" {
		t.Errorf("expected UUID bp-contract-uuid, got: %s", found.UUID)
	}

	notFound := FindBattlepassContractAsset(contracts, "non-existent-season")
	if notFound != nil {
		t.Errorf("expected nil for non-existent season, got: %+v", notFound)
	}
}

func TestGetBattlepassLevelXPs(t *testing.T) {
	// From asset
	asset := &models.ContractAsset{
		Content: models.ContractAssetContent{
			Chapters: []models.ContractChapter{
				{
					Levels: []models.ContractLevel{
						{XP: 0},
						{XP: 2000},
						{XP: 2750},
					},
				},
				{
					Levels: []models.ContractLevel{
						{XP: 3500},
						{XP: 4250},
					},
				},
			},
		},
	}

	xps := GetBattlepassLevelXPs(asset)
	if len(xps) != 5 {
		t.Fatalf("expected 5 levels extracted, got %d", len(xps))
	}
	if xps[0] != 0 || xps[1] != 2000 || xps[2] != 2750 || xps[3] != 3500 || xps[4] != 4250 {
		t.Errorf("unexpected XP curve extracted: %v", xps)
	}

	// Nil fallback
	fallback := GetBattlepassLevelXPs(nil)
	if len(fallback) != 55 {
		t.Fatalf("expected 55 fallback levels, got %d", len(fallback))
	}
}
