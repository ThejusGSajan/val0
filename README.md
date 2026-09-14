[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://golang.org)
[![License: AGPLv3](https://img.shields.io/badge/License-AGPLv3-blue.svg?style=flat-square)](LICENSE)
[![Platform: Windows](https://img.shields.io/badge/Platform-Windows%20Only-0078D6?style=flat-square&logo=windows&logoColor=white)](https://github.com/ThejusGSajan/val0)

```text
██╗   ██╗ █████╗ ██╗      ██████╗ 
██║   ██║██╔══██╗██║     ██╔═████╗
██║   ██║███████║██║     ██║██╔██║
╚██╗ ██╔╝██╔══██║██║     ████╔╝██║
 ╚████╔╝ ██║  ██║███████╗╚██████╔╝
  ╚═══╝  ╚═╝  ╚═╝╚══════╝ ╚═════╝ 
```

> Terminal UI dashboard for Valorant

`val0` started as a fifty-line script to grab the skins from the store without loading up the entire game. Then, since nothing stays a fifty-line script for long, it grew past its original scope into a store fetcher with a wishlist mechanism, detailed match history, stats, BP and mission progress trackers, and a session tracker (cus why not, at that point), all without needing Electron, a browser tab, or a couple hundred MB of memory to show some text and images.

## Core Features

### 1. Store & Wishlist (`[1] Store`)
- **Daily Rotating Storefront (`[s] Shop`)**: View your 4 daily weapon skin offers.
- **Wishlist Match Alert**: Highlights matching daily store drops with an immediate alert banner (`⭐ WISHLIST MATCH!`) so you never miss a desired skin.
- **Skin Wishlist & Catalog Search (`[w] Wishlist`)**: A split-view manager displaying your active wishlist on top and a searchable catalog on the bottom. Features real-time as-you-type search filtering, live right-pane graphic preview sprites of highlighted skins, and one-key addition (`Enter`) or removal (`x` / `Delete`).
- **Night Market Integration (`[n] Night Market`)**: Dynamically unlocks when the Night Market event is active, presenting all 6 discounted offers, custom calculated discount percentages, and original vs. discounted VP pricing.
- **Terminal Graphic Previews**: High-performance skin sprite rendering using Sixel graphics (supported natively in Windows Terminal v1.22+ and modern terminal emulators) or high-density Unicode half-block ANSI fallbacks.

### 2. Match History & Scoreboard (`[2] Matches`)
- **Recent Matches Overview**: Chronological list of your last 20 matches across all game modes.
- **Match Summary Badges**: Immediate visual indicators for outcome (`WIN`, `LOSS`, `DRAW`), map name, queue type, agent played, final round score, K/D/A ratio, Ranked Rating delta (`+/- RR` for competitive games), and relative match age.
- **Detailed Match Scoreboard (`Enter`)**: Drill into any match to view the complete 10-player scoreboard split into friendly and enemy teams (or unified leaderboard for Deathmatch).
- **In-Depth Performance Breakdown**: Tracks ACS, Kills, Deaths, Assists, Headshot %, Average Damage per Round (ADR), Econ rating, and a round-by-round win/loss timeline (`■`/`□`).

### 3. Performance Analytics & Weapon Mastery (`[3] Stats`)
- **Agent Performance Metrics**: Aggregated performance statistics across all played agents in tactical modes (automatically filters out Deathmatch and other casual arcade modes for accurate combat averages). Tracks games played, win rate %, K/D ratio, ACS, ADR, and Headshot %.
- **Weapon Mastery Statistics**: Detailed kill breakdown for your top 6 weapons ranked by total eliminations across recent matches.
- **Ranked Rating Trend Sparkline**: Chronological signed sparkline (`▲`/`▼`/`─`) tracking competitive rating fluctuations across recent matches, accompanied by your current competitive rank badge and aggregate Net RR earned.

### 4. Battlepass & Active Missions (`[4] Progress`)
- **Act Battlepass Tracker**: Visual tier progress bar displaying current tier, total tiers, overall act progression percentage, current tier XP progress (`XPInCurrentTier` / `XPForNextTier`), and total accumulated XP.
- **Active Missions Breakdown**: Real-time status for all active daily and weekly mission contracts, displaying target objective progress bars, completion checkmarks, and XP reward values.

### 5. Intelligent Session Tracker (`[5] Session`)
- **Hybrid Session Resumption**: Automatically resumes your session if your last match occurred within the past 2 hours; otherwise anchors to midnight today for a fresh daily view.
- **Mode Filter Toggle (`m`)**: Switch on the fly between `Competitive Only` and `Comp + Unrated + Swiftplay + Spike Rush` (casual Deathmatch and custom matches are permanently excluded from session statistics).
- **Real-Time Summary Card**: Displays games played, W-L-D record, Win Rate %, Net RR pill, Current Rank & RR, Average KDA, K/D ratio, Average ACS, Average HS%, Average ADR, and elapsed session duration.
- **Session Match Breakdown Table**: Scrollable table detailing every match played during the session with agent icons, scores, combat stats, and individual RR deltas.
- **Manual Session Reset (`x`)**: Clear the active session and start a new tracking baseline at any point.
- **Silent Background Auto-Polling**: Every 30 seconds, `val0` silently checks for completed matches while on Tab 2 or Tab 5, updating stats seamlessly without UI freezes or loading flickers.

---

## Architecture & Authentication

`val0` uses a zero-credential, local-first authentication design that communicates strictly with the official Riot Client running locally on Windows:

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

### Authentication & Shard Discovery
1. **Zero Credential Entry**: You never provide your Riot username, password, or 2FA codes. No credentials are ever collected, transmitted, or stored on disk.
2. **Local Loopback Discovery**: `val0` reads `%LOCALAPPDATA%\Riot Games\Riot Client\Config\lockfile` to obtain the local client port and authorization secret, communicating strictly over `127.0.0.1` HTTPS to retrieve short-lived Entitlements and Access JWTs.
3. **4-Stage Region & Shard Auto-Detection**:
   - *Stage 1*: Parses `%LOCALAPPDATA%\VALORANT\Saved\Logs\ShooterGame.log` for active game server shard URLs (`pd.<shard>.a.pvp.net`).
   - *Stage 2*: Queries the local Riot Client `/product-session/v1/external-sessions` endpoint for active process arguments (`-ares-deployment`).
   - *Stage 3*: Queries the Riot Geo PAS endpoint (`riot-geo.pas.si.riotgames.com`) with the session token.
   - *Stage 4*: Performs active shard probing across standard regions (`na`, `eu`, `ap`, `kr`).
   - *Fallback*: If auto-detection fails, an interactive region selection modal allows manual shard selection.
4. **Read-Only & Vanguard Safe**: `val0` performs no code injection, modifies no memory or game files, and runs purely as an external diagnostic reader. Tokens are kept in transient memory and discarded upon exit.
5. **Windows Terminal Auto-Relaunch**: If `val0.exe` is launched via Windows Explorer (conhost.exe), it automatically detects the legacy console host and respawns cleanly inside Windows Terminal (`wt.exe`) to guarantee true-color and Sixel graphics support.

---

## How to Use

### Installation

#### Pre-built Binary (Windows x64)
Download the latest `val0.exe` from the [GitHub Releases page](https://github.com/ThejusGSajan/val0/releases) and place it anywhere in your Windows `PATH` (or double-click to launch).

#### Building from Source
Prerequisites: **Go 1.22** or later installed on Windows.

```powershell
# Clone the repository
git clone https://github.com/ThejusGSajan/val0.git
cd val0

# Compile the Windows binary
go build -o val0.exe .
```

### Usage

1. Start the **Riot Client** and log into your account (or start Valorant).
2. Launch `val0` from Windows Terminal, PowerShell, Command Prompt, or by double-clicking `val0.exe`:
   ```powershell
   .\val0.exe
   ```
3. If the Riot Client is not yet running, `val0` displays a connection prompt. Start the client, wait for login, and press `r` to connect (or `q` to exit).

---

## Keyboard Navigation

| Scope | Keybinding | Action |
|:---|:---|:---|
| **Global Navigation** | `1` | Switch to Store & Wishlist Tab |
| | `2` | Switch to Match History Tab |
| | `3` | Switch to Performance & Weapon Stats Tab |
| | `4` | Switch to Battlepass & Missions Tab |
| | `5` | Switch to Session Tracker Tab |
| | `h` / `l` or `Left` / `Right` | Navigate to Previous / Next Tab |
| | `r` | Invalidate caches and force refresh all API data |
| | `q` / `Ctrl+C` | Quit application |
| **Store Navigation** | `s` | Switch to Daily Shop view |
| | `w` | Switch to Wishlist & Skin Catalog view |
| | `n` | Switch to Night Market view (when active) |
| **Wishlist & Catalog** | `Tab` / `Shift+Tab` | Switch focus between Wishlist and Catalog Search |
| *(Wishlist Focused)* | `j` / `k` or `Down` / `Up` | Navigate wishlisted items |
| | `x` / `Delete` | Remove selected item from wishlist |
| *(Catalog Focused)* | *Type characters* | Filter skin catalog in real time |
| | `Backspace` | Delete search characters |
| | `Down` / `Up` | Navigate filtered catalog results |
| | `Enter` | Add selected skin to wishlist |
| | `Esc` | Return focus to Wishlist list |
| **Match History** | `j` / `k` or `Down` / `Up` | Scroll through recent match list |
| | `Enter` | Open Detailed Match Scoreboard for selected match |
| | `Esc` / `Backspace` / `Left` | Return from Scoreboard to Match List |
| **Performance Stats** | `j` / `k` or `Down` / `Up` / `s` / `w` | Scroll down / up through performance metrics |
| **Session Tracker** | `m` | Toggle mode filter (`Competitive Only` vs `Comp + Unrated + Swiftplay + Spike Rush`) |
| | `x` | Reset active session tracking |
| | `j` / `k` or `Down` / `Up` | Scroll through session match breakdown table |

---

## Configuration & Environment Variables

### Graphics Protocol Override
`val0` automatically probes terminal capabilities and selects the highest-fidelity graphics protocol supported:

| Flag / Variable | Options | Description |
|:---|:---|:---|
| `--graphics <proto>` | `sixel`, `kitty`, `iterm2`, `halfblock`, `none` | CLI flag to force a specific rendering protocol |
| `VAL0_GRAPHICS` | `sixel`, `kitty`, `iterm2`, `halfblock`, `none` | Environment variable equivalent |

### Local Storage Paths
All persistent configuration and cache files reside in standard Windows AppData:
- `%APPDATA%\val-tracker\wishlist.json`: Saved skin wishlist entries.
- `%APPDATA%\val-tracker\session.json`: Active session baseline and match tracking data.
- `%APPDATA%\val-tracker\prices.json`: Persisted skin VP pricing database recorded from live storefronts.
- `%APPDATA%\val-tracker\*.json`: Cached Riot & valorant-api.com metadata (`skins`, `weapons`, `agents`, `maps`, `ranks`, `contracts`, `missions`), automatically invalidated when client patch version changes.

---

## API Reference & Data Sources

| Layer | Service / Endpoint | Description |
|:---|:---|:---|
| **Local Client** | `GET https://127.0.0.1:<port>/entitlements/v1/token` | Access Bearer token, Entitlements JWT, and player PUUID |
| **Local Client** | `GET https://127.0.0.1:<port>/product-session/v1/external-sessions` | Local process metadata for deployment/shard auto-detection |
| **Remote Riot PVP** | `POST https://pd.<shard>.a.pvp.net/store/v3/storefront/{puuid}` | Daily rotating store offers and Night Market offers (payload: `{}`) |
| **Remote Riot PVP** | `GET https://pd.<shard>.a.pvp.net/store/v1/wallet/{puuid}` | Player balances (Valorant Points, Radianite, Kingdom Credits) |
| **Remote Riot PVP** | `GET https://pd.<shard>.a.pvp.net/match-history/v1/history/{puuid}` | Recent match history list and match IDs |
| **Remote Riot PVP** | `GET https://pd.<shard>.a.pvp.net/match-details/v1/matches/{matchId}` | Comprehensive 10-player match scoreboard, economy, and rounds |
| **Remote Riot PVP** | `GET https://pd.<shard>.a.pvp.net/mmr/v1/players/{puuid}/competitiveupdates` | Ranked Rating (RR) delta history and tier movements |
| **Remote Riot PVP** | `GET https://pd.<shard>.a.pvp.net/mmr/v1/players/{puuid}` | Current competitive rank, seasonal MMR, and tier progress |
| **Remote Riot PVP** | `GET https://pd.<shard>.a.pvp.net/contracts/v1/contracts/{puuid}` | Active battlepass tier progression and mission progress |
| **Remote Riot PVP** | `PUT https://pd.<shard>.a.pvp.net/name-service/v2/players` | Batch player name resolution (`GameName#TagLine`) & shard probe |
| **Remote Riot Shared** | `GET https://shared.<shard>.a.pvp.net/content-service/v3/content` | Active season, act IDs, and battlepass contract metadata |
| **Riot Geo PAS** | `PUT https://riot-geo.pas.si.riotgames.com/pas/v1/product/valorant` | Shard and live region affinity discovery |
| **Static Assets** | `GET https://valorant-api.com/v1/*` | Static game asset metadata (`skins`, `weapons`, `agents`, `maps`, `ranks`, `contracts`, `missions`, `version`) |

---

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) & [Lipgloss](https://github.com/charmbracelet/lipgloss) by Charmbracelet for the terminal UI runtime and styling primitives.
- [valorant-api.com](https://valorant-api.com) for maintaining community game metadata and asset graphics.
- [go-sixel](https://github.com/mattn/go-sixel) by mattn for fast terminal Sixel image encoding.
- The Valorant API community for reverse-engineering endpoints and data structures.

---

## Security & Legal Disclaimer

### Read-Only & Fair Use
`val0` is an external diagnostic companion that interfaces exclusively with local loopback and standard HTTP endpoints. It does not inject code into the Valorant game process, does not inspect or manipulate game memory, does not modify game assets, and provides no tactical in-game advantages nor does it interfere with Riot Vanguard.

### Trademark Notice
`val0` is not affiliated with or endorsed by Riot Games, Inc. Valorant and Riot Games are trademarks or registered trademarks of Riot Games, Inc.

### License
This project is licensed under the [AGPLv3 License](LICENSE).
