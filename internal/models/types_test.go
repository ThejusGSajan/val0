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
