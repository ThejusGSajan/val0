package models

import (
	"encoding/json"
	"testing"
)

func TestStorefrontJSONUnmarshal(t *testing.T) {
	sampleJSON := `{
		"SkinsPanelLayout": {
			"SingleItemOffers": ["uuid-1", "uuid-2"],
			"SingleItemStoreOffers": [
				{
					"OfferID": "offer-1",
					"IsDirectPurchase": true,
					"StartDate": "2026-08-16T00:00:00Z",
					"Cost": {
						"85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741": 1775
					},
					"Rewards": [
						{
							"ItemTypeID": "item-type-1",
							"ItemID": "skin-level-uuid-1",
							"Quantity": 1
						}
					]
				}
			],
			"SingleItemOffersRemainingDurationInSeconds": 43200
		},
		"BonusStore": {
			"BonusStoreOffers": [
				{
					"BonusOfferID": "bo-1",
					"Offer": {
						"OfferID": "offer-bo-1",
						"Cost": {
							"85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741": 2175
						},
						"Rewards": [
							{"ItemID": "skin-nm-1"}
						]
					},
					"DiscountPercent": 35,
					"DiscountCosts": {
						"85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741": 1413
					},
					"IsSeen": true
				}
			],
			"BonusStoreRemainingDurationInSeconds": 864000
		}
	}`

	var sf StorefrontResponse
	if err := json.Unmarshal([]byte(sampleJSON), &sf); err != nil {
		t.Fatalf("failed to unmarshal storefront JSON: %v", err)
	}

	if len(sf.SkinsPanelLayout.SingleItemOffers) != 2 {
		t.Errorf("expected 2 single item offers, got %d", len(sf.SkinsPanelLayout.SingleItemOffers))
	}
	if len(sf.SkinsPanelLayout.SingleItemStoreOffers) != 1 {
		t.Errorf("expected 1 store offer, got %d", len(sf.SkinsPanelLayout.SingleItemStoreOffers))
	}
	cost := sf.SkinsPanelLayout.SingleItemStoreOffers[0].Cost["85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"]
	if cost != 1775 {
		t.Errorf("expected cost 1775, got %d", cost)
	}
	if sf.BonusStore == nil || len(sf.BonusStore.BonusStoreOffers) != 1 {
		t.Fatalf("expected 1 bonus store offer, got %+v", sf.BonusStore)
	}
	if sf.BonusStore.BonusStoreOffers[0].DiscountPercent != 35 {
		t.Errorf("expected discount 35, got %d", sf.BonusStore.BonusStoreOffers[0].DiscountPercent)
	}
}

func TestEntitlementResponseUnmarshal(t *testing.T) {
	sampleJSON := `{
		"accessToken": "eySampleAccessToken",
		"token": "eySampleEntitlementsToken",
		"subject": "puuid-12345-67890",
		"issuer": "https://auth.riotgames.com"
	}`

	var ent EntitlementResponse
	if err := json.Unmarshal([]byte(sampleJSON), &ent); err != nil {
		t.Fatalf("failed to unmarshal entitlements: %v", err)
	}

	if ent.AccessToken != "eySampleAccessToken" {
		t.Errorf("expected accessToken 'eySampleAccessToken', got '%s'", ent.AccessToken)
	}
	if ent.Token != "eySampleEntitlementsToken" {
		t.Errorf("expected token 'eySampleEntitlementsToken', got '%s'", ent.Token)
	}
	if ent.Subject != "puuid-12345-67890" {
		t.Errorf("expected PUUID 'puuid-12345-67890', got '%s'", ent.Subject)
	}
}

func TestWalletResponse(t *testing.T) {
	sampleJSON := `{
		"Balances": {
			"85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741": 4350,
			"e59aa87c-4cbf-517a-5983-6e81511be9b7": 85,
			"f08d4ae3-939c-4576-ab26-09ce1f23bb37": 10000
		}
	}`

	var w WalletResponse
	if err := json.Unmarshal([]byte(sampleJSON), &w); err != nil {
		t.Fatalf("failed to unmarshal wallet: %v", err)
	}

	if w.VP() != 4350 {
		t.Errorf("expected VP 4350, got %d", w.VP())
	}
	if w.RP() != 85 {
		t.Errorf("expected RP 85, got %d", w.RP())
	}
	if w.KC() != 10000 {
		t.Errorf("expected KC 10000, got %d", w.KC())
	}
}

func TestMMRResponse(t *testing.T) {
	sampleJSON := `{
		"Subject": "player-123",
		"LatestCompetitiveUpdate": {
			"MatchID": "match-1",
			"TierAfterUpdate": 21,
			"RankedRatingAfterUpdate": 52,
			"RankedRatingEarned": 22
		},
		"QueueSkills": {
			"competitive": {
				"SeasonalInfoBySeasonID": {
					"season-1": {
						"CompetitiveTier": 21,
						"RankedRating": 52
					}
				}
			}
		}
	}`

	var mmr MMRResponse
	if err := json.Unmarshal([]byte(sampleJSON), &mmr); err != nil {
		t.Fatalf("failed to unmarshal mmr: %v", err)
	}

	tier, rr := mmr.GetCurrentCompetitiveInfo()
	if tier != 21 || rr != 52 {
		t.Errorf("expected tier 21, rr 52, got tier %d, rr %d", tier, rr)
	}
}

func TestContractsResponseWithMissions(t *testing.T) {
	sampleJSON := `{
		"Version": 1,
		"Subject": "player-123",
		"Contracts": [],
		"Missions": [
			{
				"ID": "mission-daily-1",
				"Objectives": {
					"obj-1": 5
				},
				"Complete": false
			},
			{
				"ID": "mission-daily-2",
				"Complete": true
			}
		]
	}`

	var cr ContractsResponse
	if err := json.Unmarshal([]byte(sampleJSON), &cr); err != nil {
		t.Fatalf("failed to unmarshal contracts with missions: %v", err)
	}

	if len(cr.Missions) != 2 {
		t.Fatalf("expected 2 missions, got %d", len(cr.Missions))
	}
	if cr.Missions[0].Objectives["obj-1"] != 5 {
		t.Errorf("expected objective count 5, got %d", cr.Missions[0].Objectives["obj-1"])
	}
	if !cr.Missions[1].Complete {
		t.Errorf("expected mission 2 to be complete")
	}
}
