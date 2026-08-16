package api

import (
	"testing"

	"github.com/val-tracker/val-tracker/internal/models"
)

func TestFindActiveBattlepass(t *testing.T) {
	activeActID := "act-season-123"

	content := &models.ContentResponse{
		Seasons: []models.Season{
			{
				ID:       "old-act",
				Name:     "Act 1",
				Type:     "act",
				IsActive: false,
			},
			{
				ID:       activeActID,
				Name:     "Act 2",
				Type:     "act",
				IsActive: true,
			},
		},
	}

	contracts := &models.ContractsResponse{
		Contracts: []models.Contract{
			{
				ContractDefinitionID: "agent-contract-1",
				ContractProgression: models.ContractProgression{
					TotalProgressionEarned: 5000,
				},
				ProgressionLevelReached: 3,
			},
			{
				ContractDefinitionID: activeActID,
				ContractProgression: models.ContractProgression{
					TotalProgressionEarned:           120000,
					TotalProgressionTowardsNextLevel: 15000,
				},
				ProgressionLevelReached:     24,
				ProgressionTowardsNextLevel: 7500,
			},
		},
	}

	bp, found := FindActiveBattlepass(contracts, content)
	if !found {
		t.Fatal("expected to find active battlepass contract")
	}

	if bp.ContractDefinitionID != activeActID {
		t.Errorf("expected ContractDefinitionID %s, got %s", activeActID, bp.ContractDefinitionID)
	}
	if bp.ProgressionLevelReached != 24 {
		t.Errorf("expected level 24, got %d", bp.ProgressionLevelReached)
	}
}

func TestFindActiveBattlepass_FallbackHighestXP(t *testing.T) {
	// Active act not matching any definition ID
	content := &models.ContentResponse{
		Seasons: []models.Season{
			{
				ID:       "non-matching-act",
				Type:     "act",
				IsActive: true,
			},
		},
	}

	contracts := &models.ContractsResponse{
		Contracts: []models.Contract{
			{
				ContractDefinitionID: "agent-1",
				ContractProgression: models.ContractProgression{
					TotalProgressionEarned: 2000,
				},
			},
			{
				ContractDefinitionID: "bp-def-id",
				ContractProgression: models.ContractProgression{
					TotalProgressionEarned: 85000,
				},
				ProgressionLevelReached: 18,
			},
		},
	}

	bp, found := FindActiveBattlepass(contracts, content)
	if !found {
		t.Fatal("expected to find battlepass via highest XP fallback")
	}
	if bp.ContractDefinitionID != "bp-def-id" {
		t.Errorf("expected bp-def-id, got %s", bp.ContractDefinitionID)
	}
}
