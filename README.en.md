# Viticulture Web

[中文](README.md) · **English**

Run a vineyard in your browser with 2–6 friends: plant vines, make wine, play visitors, and fill orders. Supports Essential Edition and optional expansions. The game interface is primarily Chinese.

[Quick start](#quick-start) · [Features](#features) · [Version history](#version-history) · [Development guide](#development-guide) · [Buy me a HEYTEA](#buy-me-a-heytea)

## Quick start

### Download and launch

1. Open the [Releases page](https://github.com/NewYinbao/viticulture-essential-web/releases/latest) and download the **Windows x64 ZIP**.
2. Extract the entire archive and double-click **Start-Viticulture.cmd**.
3. Open **http://localhost:3015** in your browser.

No Go or Node.js installation is required. LAN play works without internet access. This repository is private; downloads require repository access.

### Build from source

Install **Git and Go 1.25+** and ensure git and go are available in your terminal. Building the game does not require Node.js. Access to the private repository and GitHub authentication are required.

Run in Windows PowerShell:

~~~powershell
git clone https://github.com/NewYinbao/viticulture-essential-web.git
cd viticulture-essential-web
powershell -NoProfile -File scripts/build-continuation.ps1
.\Start-Viticulture.cmd
~~~

The executable is built at dist/Viticulture-Continuation.exe. Select LAN mode when launching, then open **http://localhost:3015**. Source launches store saves in runtime/continuation-data; preserve that directory when upgrading.

On macOS / Linux, clone the repository, enter its directory, then build and run:

~~~sh
go build -trimpath -o viticulture ./cmd/viticulture
./viticulture -addr 0.0.0.0:3015 -data ./runtime/continuation-data
~~~

### Play with friends

The host creates a room with a name and seat password. Friends on the same LAN open `http://HOST_IP:3015` and join using the room code and their own names and passwords. The host selects expansions before starting; no separate launcher is needed. Use “另开账号” (Open another account) to sign in separately in another tab of the same browser. Refresh preserves that tab’s login; use your name and password to recover your seat after closing and reopening.

The download package stores saves in its `data` folder. Close the launch window to stop the server; stop the previous version and back up that folder before upgrading. Returning to an existing seat requires its original name and password.

[Running and save details](docs/guides/running.md) · [Temporary public sharing](docs/guides/cloudflare.md)

## Features

| Content | Support |
| --- | --- |
| Essential Edition | 2–6 players, 76 visitors, private hands, and local saves |
| Tuscany Essential | Four-season board and influence; optional 36 structures and 11 special worker types |
| Moor Visitors | 40 visitors added to the base decks |
| Visit from the Rhine Valley | 80 replacement visitors; the EE board excludes 4 Tuscany-dependent cards; not mixed with EE/Moor |
| Multiplayer and interface | Seat passwords, reconnecting, board-local card panels, action requirements, rules, and beginner guidance |

Disabled expansions stay out of actions, rules, and guidance. World, Bordeaux, extra modules from the original Tuscany, and Automa are outside the current scope.

## Version history

Three major development stages are summarized below. Linked notes describe the changes and validation scope.

| Stage | Highlights | Details |
| --- | --- | --- |
| Expansions and multiplayer | Tuscany, Moor, Rhine, seat passwords, and configuration-aware UI; available as Windows release v2026.09.12 | [Release Notes](docs/versions/06-expansions-563db1b/release-notes.md) |
| Usability and guidance | Action availability hints, board-local card panels, planting checks, and beginner help | [Release Notes](docs/versions/05-guidance-80e9f7c/release-notes.md) |
| Base game and architecture | EE rules, rule corrections, and separate game, server, storage, and web layers | [Release Notes](docs/versions/02-refactor-f9227db/release-notes.md) |

## Development guide

### Architecture

The backend uses Go; the frontend uses plain HTML, CSS, and JavaScript. One executable embeds the web interface and assets. The server applies game rules and sends each player their permitted view of the game.

```text
cmd/viticulture/   Application entry point
internal/game/    Rules, decks, and domain tests
internal/server/  HTTP, player sessions, and live updates
internal/store/   Save persistence
internal/sharing/ Optional public sharing
web/static/       Frontend and assets
tests/            Browser and native checks
scripts/          Build and check scripts
docs/             Detailed guides and historical records
```

[Detailed architecture](docs/guides/architecture.md)

### Local development

Follow [Build from source](#build-from-source) to obtain and run the project. Restart or rebuild the Go program after editing embedded frontend files. Frontend checks and browser tests also require **Node.js 22+**; run npm ci in the project directory to install development dependencies.

[Full build and testing instructions](docs/guides/testing.md)

## Buy me a HEYTEA

If this project brings a good game night to you and your friends, scan with WeChat and buy me a HEYTEA 🍵

<img src="docs/assets/buy-me-a-heytea-wechat.png" alt="WeChat donation QR code" width="270" height="270">

Choose a preset amount after scanning. Thank you for your support.

## About this project

This is an unofficial implementation of Viticulture and is not affiliated with its publisher. Game and asset rights belong to their respective owners. No open-source license has been added to this repository.
