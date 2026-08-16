package auth

// RegionToShard maps region codes to the PD/GLZ shard used in URLs.
// For most regions, the shard is the region itself.
// Exception: latam and br route through the "na" shard.
var RegionToShard = map[string]string{
	"na":    "na",
	"eu":    "eu",
	"ap":    "ap",
	"kr":    "kr",
	"latam": "na",
	"br":    "na",
}

// PDBaseURL returns the Player Data base URL for a given shard.
//
//	e.g. "https://pd.na.a.pvp.net"
func PDBaseURL(shard string) string {
	return "https://pd." + shard + ".a.pvp.net"
}

// SharedBaseURL returns the Shared Services base URL.
//
//	e.g. "https://shared.na.a.pvp.net"
func SharedBaseURL(shard string) string {
	return "https://shared." + shard + ".a.pvp.net"
}
