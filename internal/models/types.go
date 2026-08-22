package models

import "time"

// ── Lockfile ────────────────────────────────────────────────────────

// Lockfile represents the parsed contents of the Riot Client lockfile.
// Format: name:pid:port:password:protocol
type Lockfile struct {
	Name     string
	PID      string
	Port     string
	Password string
	Protocol string
	ModTime  time.Time // os.Stat modification time
}

// ── Local Entitlement Token Response ────────────────────────────────

// EntitlementResponse is the JSON returned by
// GET https://127.0.0.1:{port}/entitlements/v1/token
type EntitlementResponse struct {
	AccessToken  string `json:"accessToken"`
	Token        string `json:"token"`         // X-Riot-Entitlements-JWT
	Subject      string `json:"subject"`       // PUUID
	Issuer       string `json:"issuer"`
	Entitlements []any  `json:"entitlements"`
}

// ── Session / Auth Context ──────────────────────────────────────────

// Session holds all auth state needed for downstream API calls.
type Session struct {
	AccessToken      string
	EntitlementToken string
	PUUID            string
	Region           string // "na", "eu", "ap", "kr"
	Shard            string // same as region for most cases
	ClientVersion    string // e.g. "release-13.02-shipping-17-5277781"
}

// ── Riot Geo Response ───────────────────────────────────────────────

// RiotGeoResponse is the JSON returned by
// PUT https://riot-geo.pas.si.riotgames.com/pas/v1/product/valorant
type RiotGeoResponse struct {
	Token      string            `json:"token"`
	Affinities map[string]string `json:"affinities"`
}

// ── Storefront (Daily Shop) ─────────────────────────────────────────

// StorefrontResponse is the JSON returned by
// GET https://pd.{shard}.a.pvp.net/store/v2/storefront/{puuid}
type StorefrontResponse struct {
	SkinsPanelLayout SkinsPanelLayout `json:"SkinsPanelLayout"`
	BonusStore       *BonusStore      `json:"BonusStore"`
}

type SkinsPanelLayout struct {
	SingleItemOffers                            []string               `json:"SingleItemOffers"`
	SingleItemStoreOffers                       []SingleItemStoreOffer `json:"SingleItemStoreOffers"`
	SingleItemOffersRemainingDurationInSeconds int                    `json:"SingleItemOffersRemainingDurationInSeconds"`
}

type SingleItemStoreOffer struct {
	OfferID          string         `json:"OfferID"`
	IsDirectPurchase bool           `json:"IsDirectPurchase"`
	StartDate        string         `json:"StartDate"`
	Cost             map[string]int `json:"Cost"`
	Rewards          []Reward       `json:"Rewards"`
}

type Reward struct {
	ItemTypeID string `json:"ItemTypeID"`
	ItemID     string `json:"ItemID"`
	Quantity   int    `json:"Quantity"`
}

// ── Night Market (BonusStore) ───────────────────────────────────────

type BonusStore struct {
	BonusStoreOffers                     []BonusStoreOffer `json:"BonusStoreOffers"`
	BonusStoreRemainingDurationInSeconds int               `json:"BonusStoreRemainingDurationInSeconds"`
}

type BonusStoreOffer struct {
	BonusOfferID    string               `json:"BonusOfferID"`
	Offer           SingleItemStoreOffer `json:"Offer"`
	DiscountPercent int                  `json:"DiscountPercent"`
	DiscountCosts   map[string]int       `json:"DiscountCosts"`
	DiscountedCost  map[string]int       `json:"DiscountedCost"`
	IsSeen          bool                 `json:"IsSeen"`
}

// ── Contracts / Battlepass ──────────────────────────────────────────

// ContractsResponse is the JSON returned by
// GET https://pd.{shard}.a.pvp.net/contracts/v1/contracts/{puuid}
type ContractsResponse struct {
	Version               int             `json:"Version"`
	Subject               string          `json:"Subject"`
	Contracts             []Contract      `json:"Contracts"`
	ProcessedMatches      []any           `json:"ProcessedMatches"`
	ActiveSpecialContract string          `json:"ActiveSpecialContract"`
	Missions              []Mission       `json:"Missions"`
	MissionMetadata       MissionMetadata `json:"MissionMetadata"`
}

type Mission struct {
	ID             string         `json:"ID"`
	Objectives     map[string]int `json:"Objectives"`
	Complete       bool           `json:"Complete"`
	ExpirationTime string         `json:"ExpirationTime"`
}

type MissionMetadata struct {
	NPECompleted     bool   `json:"NPECompleted"`
	WeeklyCheckpoint string `json:"WeeklyCheckpoint"`
	WeeklyRefillTime string `json:"WeeklyRefillTime"`
}

type Contract struct {
	ContractDefinitionID        string              `json:"ContractDefinitionID"`
	ContractProgression         ContractProgression `json:"ContractProgression"`
	ProgressionLevelReached     int                 `json:"ProgressionLevelReached"`
	ProgressionTowardsNextLevel int                 `json:"ProgressionTowardsNextLevel"`
}

type ContractProgression struct {
	TotalProgressionEarned           int            `json:"TotalProgressionEarned"`
	TotalProgressionTowardsNextLevel int            `json:"TotalProgressionTowardsNextLevel"`
	HighestRewardedLevel             map[string]any `json:"HighestRewardedLevel"`
}

// ── Content Service (Active Season) ─────────────────────────────────

// ContentResponse is the JSON returned by
// GET https://shared.{shard}.a.pvp.net/content-service/v3/content
type ContentResponse struct {
	Seasons []Season `json:"Seasons"`
	Events  []Event  `json:"Events"`
}

type Season struct {
	ID       string `json:"ID"`
	Name     string `json:"Name"`
	Type     string `json:"Type"` // "act" or "episode"
	IsActive bool   `json:"IsActive"`
}

type Event struct {
	ID       string `json:"ID"`
	Name     string `json:"Name"`
	IsActive bool   `json:"IsActive"`
}

// ── valorant-api.com Assets ─────────────────────────────────────────

// ValorantAPISkinsResponse wraps the /v1/weapons/skins endpoint.
type ValorantAPISkinsResponse struct {
	Status int         `json:"status"`
	Data   []SkinAsset `json:"data"`
}

// ── Skin Chroma (Variant) ───────────────────────────────────────────

type SkinChroma struct {
	UUID          string  `json:"uuid"`
	DisplayName   string  `json:"displayName"`
	DisplayIcon   *string `json:"displayIcon"`
	FullRender    *string `json:"fullRender"`
	Swatch        *string `json:"swatch"`
	StreamedVideo *string `json:"streamedVideo"`
}

type SkinAsset struct {
	UUID            string       `json:"uuid"`
	DisplayName     string       `json:"displayName"`
	ThemeUUID       string       `json:"themeUuid"`
	ContentTierUUID *string      `json:"contentTierUuid"` // nil for base skins
	DisplayIcon     *string      `json:"displayIcon"`     // URL, nullable
	Chromas         []SkinChroma `json:"chromas"`
	Levels          []SkinLevel  `json:"levels"`
}

type SkinLevel struct {
	UUID        string  `json:"uuid"`
	DisplayName *string `json:"displayName"`
	DisplayIcon *string `json:"displayIcon"`
}

// ValorantAPIVersionResponse wraps the /v1/version endpoint.
type ValorantAPIVersionResponse struct {
	Status int         `json:"status"`
	Data   VersionData `json:"data"`
}

type VersionData struct {
	RiotClientVersion string `json:"riotClientVersion"` // used as X-Riot-ClientVersion
	Branch            string `json:"branch"`
	BuildDate         string `json:"buildDate"`
}

// ── Content Tiers (Rarity) ──────────────────────────────────────────

type ContentTier struct {
	UUID           string `json:"uuid"`
	DevName        string `json:"devName"`       // "Select", "Deluxe", etc.
	DisplayName    string `json:"displayName"`   // "Select Edition", etc.
	Rank           int    `json:"rank"`
	HighlightColor string `json:"highlightColor"` // hex RGBA, e.g. "5a9fe233"
}

// ── Resolved Skin (display-ready after UUID → name + rarity lookup) ─

type ResolvedSkin struct {
	UUID          string
	DisplayName   string
	Rarity        string // "Select", "Deluxe", "Premium", "Exclusive", "Ultra"
	CostVP        int
	IconURL       string
	FullRenderURL string
	Sprite        string // pre-rendered ANSI block-character fallback
}
