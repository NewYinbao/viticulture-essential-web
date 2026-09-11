# Viticulture Web

[中文](README.md) · **English**

Bring your vineyard to the browser. Gather 2–6 friends in one room to plant vines, make wine, play visitors, and fill orders.

Built with Go and plain HTML/CSS/JavaScript. A single executable embeds the web interface, Chinese card descriptions, and SVG artwork. LAN play needs neither Node.js nor an internet connection. This is the English project guide; the game interface is currently primarily Chinese.

[Quick start](#quick-start) · [Version history](docs/README.md) · [Development guide](docs/guides/architecture.md) · [Buy me a HEYTEA](#buy-me-a-heytea)

## Features

| Content | Support |
| --- | --- |
| Essential Edition | 2–6 players, 76 visitors, finite decks, private hands, and local saves |
| Tuscany Essential | Four-season board, individual season transitions, influence; optional 36 structures and 11 special worker types |
| Moor Visitors | 40 visitors added to the EE visitor decks |
| Visit from the Rhine Valley | 80 replacement visitors; 4 Tuscany-dependent cards excluded on the EE board; never mixed with EE/Moor |
| Player identity | Passwords per room seat, password changes, logout, and session revocation; a matching name alone cannot claim another seat |
| Interface and help | Board-local card panels, hover/focus hints, inspectable planting conditions, contextual rules, and an optional beginner guide |

Disabled expansions stay out of in-game actions, rules, and guidance. World, Bordeaux, extra modules from the original Tuscany, and Automa are outside the current scope.

## Quick start

### Windows

Building from source requires **Go 1.25+**. Run from the project root:

```powershell
powershell -NoProfile -File scripts/build-continuation.ps1
.\Start-Viticulture-Continuation.cmd
```

If the current executable is already built, double-click `Start-Viticulture-Continuation.cmd`.

1. The host opens `http://localhost:3015`, enters a name and seat password, and creates a room.
2. Friends open `http://HOST_LAN_IP:3015`, then join with the room code and their own names and passwords.
3. The host selects expansions before starting. Configuration is locked once setup begins. Returning to an existing seat requires its original name and password in that room.

Saves live in `runtime/continuation-data`. Close the launch window or press Ctrl+C to stop the server. Only one process may use a save directory at a time. Executables, real saves, and session credentials are excluded from Git.

### Other platforms / run from source

```sh
go run ./cmd/viticulture -addr 0.0.0.0:3015 -data ./runtime/continuation-data
```

Legacy launchers remain for compatibility. They do not automatically upgrade old executables or migrate games. See the [running guide](docs/guides/running.md) for version switching and initial password enrollment for old seats. Temporary public sharing must be explicitly enabled as described in the [Cloudflare guide](docs/guides/cloudflare.md).

## Documentation and versions

Versions are organized by actual Git commits. Each includes a **development record, validation record, and Release Notes**. Directory numbers provide chronological ordering; they are not new release tags.

- [Documentation and version index](docs/README.md)
- [Current gameplay version: expansions, passwords, and UI · 563db1b](docs/versions/06-expansions-563db1b/release-notes.md)
- [Architecture](docs/guides/architecture.md) · [HTTP API](docs/guides/api.md) · [Build and testing](docs/guides/testing.md)

The `563db1b` acceptance run passed 12/12 steps, including 24 configurations, 24 natural games, 36 structure fixtures, and representative complex interactions. Two independent functional reviews and a separate engineering review were completed. GitHub CI, including the Go race detector, passed. The natural-game strategy does not play visitors; this is not exhaustive coverage of every card combination. See the [version's validation record](docs/versions/06-expansions-563db1b/validation.md) for precise boundaries. Detailed development and validation records are in Chinese; Release Notes include English summaries.

## Development

```sh
go test -tags "ee_rule_audit audit_diff" ./...
go vet ./...
npm ci
npx playwright install chromium
npm run test:units
npm run format:check
npm run test:ui
```

Node.js 22+ is only needed for development checks. On Windows, `scripts/check.ps1` runs Go audit tests, vet, and gofmt checks. The full expansion browser suite requires Windows Edge; see the [testing guide](docs/guides/testing.md).

```text
cmd/viticulture/   Application entry point
internal/game/    Rules, decks, and domain tests
internal/server/  Passwords, HTTP, player views, and live updates
internal/store/   Save persistence
internal/sharing/ Optional public sharing
web/static/       Plain web frontend and local assets
tests/            Browser and native verification
docs/             Current guides, version records, and documentation assets
```

## Buy me a HEYTEA

If this project brings a good game night to you and your friends, scan with WeChat and buy me a HEYTEA 🍵

<img src="docs/assets/buy-me-a-heytea-wechat.png" alt="WeChat donation QR code" width="270" height="270">

Thank you for your support. Choose a preset amount on the WeChat payment screen. [QR code and preset amount maintenance notes](docs/guides/support.md)

## About this project

This is an unofficial implementation of Viticulture and is not affiliated with its publisher. Version records retain rule references and interpretation boundaries. Reference implementations do not grant permission to redistribute commercial card artwork. No open-source license has been added to this repository; public visibility alone does not grant redistribution rights.
