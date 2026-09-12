[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-informational?style=flat-square)](https://github.com/val-tracker/val-tracker)
[![Build Status](https://img.shields.io/github/actions/workflow/status/val-tracker/val-tracker/ci.yml?branch=main&style=flat-square)](https://github.com/val-tracker/val-tracker/actions)
[![Release](https://img.shields.io/github/v/release/val-tracker/val-tracker?style=flat-square)](https://github.com/val-tracker/val-tracker/releases)

```text
██╗   ██╗ █████╗ ██╗      ██████╗ 
██║   ██║██╔══██╗██║     ██╔═████╗
██║   ██║███████║██║     ██║██╔██║
╚██╗ ██╔╝██╔══██║██║     ████╔╝██║
 ╚████╔╝ ██║  ██║███████╗╚██████╔╝
  ╚═══╝  ╚═╝  ╚═╝╚══════╝ ╚═════╝ 
```

> No-nonsense terminal dashboard for Valorant

---

## Origin & Motivation

Like many players, this project started out of a very simple daily friction: wanting to check what skins appeared in the daily rotating store or whether the Night Market had finally rolled something good, without having to launch the full game. Booting the heavy 3D game client, waiting through anti-cheat initialization, loading intro screens, and spinning up shaders just to check a shop timer felt unnecessary when you just wanted a quick glance from your desktop.

What began as a tiny Go script to read local authentication tokens and query the storefront endpoint gradually expanded. First came high-resolution terminal skin previews, then a wishlist tracker to notify when favorite skins dropped, followed by match history inspection, round-by-round combat scoreboards, aggregate headshot analytics, and battlepass mission progression.

`val0` is the result of that evolution: a lightweight, fast, keyboard-driven companion app built specifically for the terminal. It provides instant access to your stats, cosmetics, and match records in milliseconds while consuming negligible system resources.

---

## Core Features

### 1. Store & Night Market
- **Daily Storefront**: View your 4 daily rotating weapon offers alongside precise countdown timers until the next shop refresh.
- **Terminal Graphics Previews**: Automatic weapon render previews utilizing Sixel, Kitty Graphics Protocol, iTerm2 Inline Images, or ANSI half-block rendering based on your terminal's capabilities.
- **Night Market Integration**: When active, automatically populates the Night Market sub-view with custom discounted pricing and calculated percentage deductions.
- **Local Skin Wishlist**: Browse the complete Valorant weapon skin catalog, search for desired cosmetics, add them to your local wishlist, and receive instant alert banners whenever a wishlisted item rotates into your daily shop.

### 2. Match History & Deep Scoreboard
- **Match Overview**: Chronological list of recent competitive and casual matches displaying match outcome (WIN, LOSS, DRAW), map, agent played, final round score, K/D/A, and Ranked Rating (RR) delta.
- **Deep Match Scoreboard**: Inspect any match with a single keystroke to view a complete 10-player scoreboard breakdown featuring Average Combat Score (ACS), K/D/A ratios, Average Damage per Round (ADR), Econ Rating, and round win distribution.

### 3. Performance Analytics & Sparklines
- **Agent Performance Metrics**: Aggregated performance metrics across all played agents, including games played, win rate percentage, K/D, ACS, ADR, and Headshot percentage.
- **Weapon Mastery Statistics**: Detailed breakdown of weapon kill distributions, including precision headshot, bodyshot, and legshot percentages.
- **Trend Sparklines**: Real-time terminal sparklines visualizing match-by-match headshot accuracy trends and competitive Ranked Rating fluctuations over your recent match history.

### 4. Battlepass Progression & Active Missions
- **Act Battlepass Tracker**: Visual tier progress bar indicating current tier, tier completion percentage, remaining XP required for the next unlock, and total accumulated XP.
- **Active Mission Breakdown**: Real-time status of all active daily and weekly missions with individual progress indicators, target thresholds, and XP reward values.

---

## Architecture & Authentication

`val0` uses a local-first, zero-credential authentication mechanism that interfaces directly with the official Riot Client running on your machine:

```text
+---------------------+         +----------------------+         +-----------------------+
|  Local Riot Client  | <-----> |     val0 Engine      | <-----> |   Riot PVP Endpoints  |
|  (Lockfile & TLS)   |         |   (Bubble Tea TUI)   |         | (Store, MMR, Matches) |
+---------------------+         +----------------------+         +-----------------------+
                                           |
                                           v
                                +----------------------+
                                |  valorant-api.com    |
                                |  (Skins & Metadata)  |
                                +----------------------+
```

### Why the Lockfile Route?
- **Zero Credential Entry**: You never enter your Riot username, password, or two-factor authentication codes into `val0`. No credentials are ever collected, stored, or transmitted.
- **Local Loopback TLS**: `val0` discovers the active Riot Client process via the local `lockfile` generated in `%LOCALAPPDATA%\Riot Games\Riot Client\Config\lockfile`. It communicates strictly with the local client on `127.0.0.1` over loopback HTTPS using standard HTTP Basic authentication to obtain short-lived session tokens (Entitlements JWT and Access Bearer token).
- **Non-Invasive & Read-Only**: `val0` acts entirely as a read-only spectator. It does not inject DLLs, hook game memory, modify game files, or interact with Vanguard anti-cheat in any way.
- **Transient Memory**: Session tokens remain exclusively in transient application memory during execution and expire naturally when the application terminates.

---

## How to Use

### Installation

#### Pre-built Binaries
Download the pre-compiled binary for your operating system and architecture from the [GitHub Releases](https://github.com/val-tracker/val-tracker/releases) page. Extract the archive and place the `val0` binary anywhere in your system `PATH`.

#### Building from Source
Prerequisites: **Go 1.22** or later.

```bash
# Clone the repository
git clone https://github.com/val-tracker/val-tracker.git
cd val-tracker

# Download dependencies and build binary
go build -o val0 .
```

### Usage

1. Launch the official **Riot Client** and log into your account.
2. Run `val0` in your terminal:
   ```bash
   ./val0
   ```
3. If the Riot Client is closed or has not yet authenticated, `val0` will present a connection error screen indicating that the lockfile could not be found. Simply open the Riot Client, wait for login to complete, and press `r` to retry the connection (or `q` to quit).

### Keyboard Navigation

| Scope | Keybinding | Action |
|:---|:---|:---|
| **Global** | `1` | Switch to Store & Wishlist Tab |
| | `2` | Switch to Match History Tab |
| | `3` | Switch to Performance Stats Tab |
| | `4` | Switch to Battlepass & Missions Tab |
| | `h` / `l` or `Left` / `Right` | Navigate to Previous / Next Tab |
| | `r` | Refresh all data and invalidate image caches |
| | `q` / `Ctrl+C` | Exit application |
| **Store Navigation** | `s` | Switch to Daily Shop view |
| | `w` | Switch to Wishlist & Catalog view |
| | `n` | Switch to Night Market view (when active) |
| **Wishlist & Catalog** | `Tab` / `Shift+Tab` | Toggle focus between Wishlist and Catalog Search |
| | `j` / `k` or `Down` / `Up` | Navigate items in active section |
| | `Enter` | Add selected skin from search results to wishlist |
| | `x` / `Delete` | Remove selected item from wishlist |
| | `Esc` | Return focus to wishlist list |
| **Match History** | `j` / `k` or `Down` / `Up` | Scroll through recent match records |
| | `Enter` | Open Deep Match Scoreboard for selected match |
| | `Esc` / `Backspace` / `Left` | Return from Scoreboard to Match List |
| **Performance Stats** | `j` / `k` or `Down` / `Up` | Scroll down / up through performance metrics |

### Configuration & Environment Variables

`val0` automatically detects the most optimal graphics protocol supported by your terminal emulator. You can explicitly override this behavior via command-line flags or environment variables:

| Parameter | Type | Options | Description |
|:---|:---|:---|:---|
| `--graphics` | CLI Flag | `sixel`, `kitty`, `iterm2`, `halfblock`, `none` | Explicitly force a graphics rendering protocol |
| `VAL0_GRAPHICS` | Environment Variable | `sixel`, `kitty`, `iterm2`, `halfblock`, `none` | Equivalent environment override |

#### Local Storage & Cache Paths
- **Windows**: `%APPDATA%\val-tracker\`
- **Linux / macOS**: `~/.config/val-tracker/`

These directories persist `wishlist.json` (user-selected wishlist items) and `skins.json` (cached static skin assets validated against the remote client version to minimize network requests).

---

## API Reference & Data Sources

`val0` aggregates data across local loopback services, remote Riot PVP infrastructure, and community static endpoints:

| Layer | Service / Endpoint | Description |
|:---|:---|:---|
| **Local Client** | `GET https://127.0.0.1:<port>/entitlements/v1/token` | Fetches Access Bearer token, Entitlements JWT, and player PUUID |
| **Remote Riot PVP** | `GET https://pd.<shard>.a.pvp.net/store/v2/storefront/{puuid}` | Daily rotating store offers and active Night Market offers |
| **Remote Riot PVP** | `GET https://pd.<shard>.a.pvp.net/store/v1/wallet/{puuid}` | Player balances (Valorant Points, Radianite, Kingdom Credits) |
| **Remote Riot PVP** | `GET https://pd.<shard>.a.pvp.net/match-history/v1/history/{puuid}` | Recent match history list and match IDs |
| **Remote Riot PVP** | `GET https://pd.<shard>.a.pvp.net/match-details/v1/matches/{matchId}` | Comprehensive 10-player match scoreboard, economy, and rounds |
| **Remote Riot PVP** | `GET https://pd.<shard>.a.pvp.net/mmr/v1/players/{puuid}/competitiveupdates` | Ranked Rating (RR) delta history and tier movements |
| **Remote Riot PVP** | `GET https://pd.<shard>.a.pvp.net/mmr/v1/players/{puuid}` | Current competitive rank, leaderboards, and seasonal MMR |
| **Remote Riot PVP** | `GET https://pd.<shard>.a.pvp.net/contracts/v1/contracts/{puuid}` | Active battlepass tier progress, total XP, and mission progress |
| **Static Assets** | `GET https://valorant-api.com/v1/weapons/skins` | Weapon skin names, rarity tiers, and weapon icon sprites |
| **Static Assets** | `GET https://valorant-api.com/v1/version` | Live client version metadata for cache invalidation |

---

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) & [Lipgloss](https://github.com/charmbracelet/lipgloss) by Charmbracelet for the terminal UI runtime and styling primitives.
- [valorant-api.com](https://valorant-api.com) for maintaining open-access static game metadata and asset graphics.
- [go-sixel](https://github.com/mattn/go-sixel) by mattn for high-speed terminal Sixel image encoding.
- The Valorant API community documentation contributors for reverse-engineering endpoints and data models.

---

## Security & Legal Disclaimer

### Read-Only & Non-Invasive Guarantee
`val0` is an external diagnostic dashboard that communicates with public and loopback HTTP APIs. It does not inject code into the game process, does not manipulate memory, does not modify game assets, and does not provide any in-game tactical advantages. It strictly adheres to fair-use principles and safe execution alongside Riot Vanguard.

### Trademark Notice
`val0` is not endorsed by Riot Games and does not reflect the views or opinions of Riot Games or anyone officially involved in producing or managing Riot Games properties. Riot Games and Valorant are trademarks or registered trademarks of Riot Games, Inc. Valorant © Riot Games, Inc.

### License
This project is open-source software licensed under the [MIT License](LICENSE).
