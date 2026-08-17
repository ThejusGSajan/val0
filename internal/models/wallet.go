package models

// Currency UUID constants for Valorant store wallet.
const (
	VPUUID   = "85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741" // Valorant Points
	RPUUID   = "e59aa87c-4cbf-517a-5983-6e81511be9b7" // Radianite Points
	KCUUID   = "f08d4ae3-939c-4576-ab26-09ce1f23bb37" // Kingdom Credits
	FreeUUID = "85ca954a-41f2-ce94-9b45-8ca3dd39a00e" // Free Agent Token
)

// WalletResponse is returned by GET https://pd.{shard}.a.pvp.net/store/v1/wallet/{puuid}
type WalletResponse struct {
	Balances map[string]int `json:"Balances"`
}

// VP returns the Valorant Points balance.
func (w *WalletResponse) VP() int {
	if w == nil || w.Balances == nil {
		return 0
	}
	return w.Balances[VPUUID]
}

// RP returns the Radianite Points balance.
func (w *WalletResponse) RP() int {
	if w == nil || w.Balances == nil {
		return 0
	}
	return w.Balances[RPUUID]
}

// KC returns the Kingdom Credits balance.
func (w *WalletResponse) KC() int {
	if w == nil || w.Balances == nil {
		return 0
	}
	return w.Balances[KCUUID]
}
