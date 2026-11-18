<div align="center">

<img src="https://i.imgur.com/w4HuCXH.png" alt="NASDUCK — memecoins in GTA 6" width="100%">

# NASDUCK

**Memecoin economy layer for GTA VI.**
Launch, discover, search, buy, sell and trade tokens from the in-game market terminal — without ever leaving the game.

![status](https://img.shields.io/badge/status-active-brightgreen)
![engine](https://img.shields.io/badge/engine-rust-orange)
![platform](https://img.shields.io/badge/platform-windows%20%7C%20linux%20%7C%20macos-lightgrey)
![psn](https://img.shields.io/badge/psn-account%20link%20required-003791)
![license](https://img.shields.io/badge/license-MIT-blue)

[Website](https://nasduck.org/) · [Install](#installation) · [Link PSN](#linking-a-playstation-network-account) · [How it works](#how-it-works) · [CLI reference](#market-terminal-reference) · [FAQ](#faq)

</div>

---

## Contents

| | | |
| --- | --- | --- |
| [What NASDUCK is](#what-nasduck-is) | [The problem](#the-problem) | [How it works](#how-it-works) |
| [Requirements](#requirements) | [Installation](#installation) | [Linking PSN](#linking-a-playstation-network-account) |
| [Quick start](#quick-start) | [Terminal reference](#market-terminal-reference) | [Configuration](#configuration) |
| [Recording and traces](#recording-monitoring-and-traces) | [Bindings](#bindings) | [Exit codes](#exit-codes) |
| [Security model](#security-model) | [FAQ](#faq) | [Roadmap](#roadmap) |
| [Disclaimer](#disclaimer) | [Contributing](#contributing) | [License](#license) |

---

<img src="https://i.imgur.com/K0kyqOk.png" alt="Economy layer" width="100%">

## What NASDUCK is

NASDUCK brings a full memecoin economy into GTA VI. Players open the in-game market terminal, browse what launched in the last hour, search a ticker, read a book, buy, sell, take profit — and never touch a second screen.

The market is not a companion app. It is a surface inside the game, and NASDUCK is the engine that drives it: it attaches to the terminal session, mirrors every cell of the screen, routes orders, settles balances, and exposes the whole thing to scripts, bots and desktop tools through one Rust core and four language bindings.

| Half | What it is |
| --- | --- |
| **Surface** | The in-game market terminal — the native screen where players launch and trade. |
| **Engine** | `nasduck` — the Rust runtime that attaches to the surface, resolves identity, routes orders, records and replays sessions, and exposes a stable API. |

You can drive the surface by hand in-game, or drive it from the CLI, or from Python, JavaScript, Rust and Go. It is the same engine and the same session either way.

## The problem

Every memecoin venue built for a game so far has made the same mistake: it asked players to leave the game.

| | Outside the game | Inside the game |
| --- | --- | --- |
| **Steps to a trade** | tab out, open browser, connect wallet, paste contract, wait for load | press one button |
| **Time to first fill** | 20–40 seconds | under a second |
| **Context** | lost — you were in a firefight | kept — same screen, same session |
| **Failure mode** | stale page, wrong contract, dead link | none of the above |

NASDUCK inverts the old shape. The venue is a screen in the game. Discovery, search, price, book depth, position, exit — all of it is rendered by the game itself, in the game's own art direction.

That constraint is also what makes the engine interesting: to put a live market inside a terminal surface you need deterministic screen reads, retrying locators, snapshot assertions, structured failure traces and a recording format. Which is exactly what `nasduck` is.

<img src="https://i.imgur.com/q9uQKct.png" alt="Four layers: surface, transport, identity, settlement" width="100%">

## How it works

Four layers, each independently testable, each with its own contract.

```mermaid
flowchart TB
    X["4 · SETTLEMENT<br/>order router · matching · balance ledger · token supply"]
    I["3 · IDENTITY<br/>PSN link · per-session order key · entitlements"]
    T["2 · TRANSPORT<br/>local bridge · cell grid · backend attach · PTY or LAN"]
    S["1 · SURFACE<br/>in-game market terminal · launch · book · positions"]
    S --> T --> I --> X
```

**1. Surface.** The market terminal is a real text-mode application rendered by the game. Because it is a cell grid and not a bitmap, every element — ticker, price, depth, size, button — is addressable. NASDUCK reads the grid, not pixels, so a reshade or a resolution change does not break anything.

**2. Transport.** A local bridge publishes the terminal as a session. The engine attaches to it with the configured backend and gets a normal session handle: rows, columns, cells, styles, cursor, title, scrollback. From that point the surface is indistinguishable from a shell.

**3. Identity.** The terminal only serves a player it can name. Linking a PSN account binds the session to your identity, resolves your in-game wallet and derives a per-session order key. Orders are signed locally; the key never leaves your machine.

**4. Settlement.** Orders hit the router, get matched against the book, and settle into the ledger. Token supply, launch curves and balances are engine state — the surface is only ever a view of it.

Nothing in the stack is fire-and-forget. Every wait is bounded, every assertion is retried, every failure produces an artifact:

```console
$ nasduck run market-terminal
$ nasduck expect text "MARKET" --timeout 15000
$ nasduck click text "DISCOVER"
$ nasduck expect text "24h"
```

Read [references/cli.md](references/cli.md) for the full surface contract.

## Requirements

| Requirement | Detail |
| --- | --- |
| Game | **GTA VI**, market terminal unlocked after the in-game brokerage tutorial |
| Account | **PlayStation Network** account in good standing, linked to the account you play on |
| Host | **Windows, Linux or macOS** — same machine or same LAN as the console |
| Toolchain | **Rust 1.90+** only when building from source; binaries need nothing |

Console support is attach-only: the engine never patches the game and never writes to game files.

## Installation

### CLI

| Platform | Command |
| --- | --- |
| Homebrew | `brew tap shiplaone2q/nasduck https://github.com/shiplaone2q/nasduck` then `brew install nasduck` |
| macOS / Linux | `curl --proto '=https' --tlsv1.2 -LsSf https://raw.githubusercontent.com/shiplaone2q/nasduck/main/install/install.sh \| sh` |
| Windows | `irm https://raw.githubusercontent.com/shiplaone2q/nasduck/main/install/install.ps1 \| iex` |
| Manual | binary from [Releases](https://github.com/bobblockhain/nasduck-opensource/releases) |

### Libraries

| Language | Install | Reference |
| --- | --- | --- |
| Rust 1.90+ | `cargo add nasduck-rs@0.1.1` | [docs.rs](https://docs.rs/nasduck-rs/latest/nasduck/) |
| Python 3.8+ | `pip install nasduck` | [Python API](bindings/python/README.md) |
| Node 20+ | `npm install @shiplaone2q/nasduck` | [JavaScript API](bindings/js/README.md) |
| Go 1.26+ | `go get github.com/shiplaone2q/nasduck/bindings/go@v0.1.1` | [Go API](bindings/go/README.md) |

Add the Rust `recording-raster` feature for APNG, GIF and MP4 capture. Go, Python and JavaScript all go through one [native adapter crate](crates/nasduck-ffi/README.md); set `NASDUCK__NATIVE_LIBRARY` to pin a specific library.

<img src="https://i.imgur.com/tedUein.png" alt="Link PSN — device code approval" width="100%">

## Linking a PlayStation Network account

Linking is a one-time device-code flow. It takes about forty seconds and no password is ever entered into NASDUCK.

```mermaid
sequenceDiagram
    participant P as Player
    participant E as nasduck
    participant B as Browser
    participant N as PSN
    P->>E: nasduck psn link
    E->>P: short code + URL
    P->>B: opens URL, confirms code
    B->>N: approval on PSN side
    N-->>E: approval lands
    E->>E: exchange for refresh token, derive session key
    E->>P: linked account, scopes, key fingerprint
```

1. **Run the command.** The engine prints a short code and the link URL, and opens the URL in your default browser when one is available.

```console
$ nasduck psn link

  Opening psn.nasduck.org/link
  Enter code:  K7QM-4XDA

  Waiting for approval...            (expires in 300s)
```

2. **Approve on the PSN side.** Open the URL on your phone or console browser, sign in to PlayStation Network if you are not already, and confirm the code matches. The approval screen names the exact scopes being requested.
3. **Confirm on the terminal.** The engine polls until the approval lands, then exchanges it for a refresh token and prints the resolved account:

```console
  Linked: bobblockhain (PSN ID)
  Scopes: psn:identity psn:presence nasduck:market
  Key:    ed25519:9f2c…a71b  (session only)
```

4. **Attach the console.** With the account linked, the bridge accepts the terminal session:

```console
$ nasduck psn status
$ nasduck run market-terminal --wait-ready
```

| Command | Description |
| --- | --- |
| `psn link` | Start the device-code flow. |
| `psn status` | Show the linked account, scopes and key fingerprint. |
| `psn refresh` | Force a token refresh. |
| `psn unlink` | Revoke the token and drop the session key. |

| Property | Value |
| --- | --- |
| Token storage | platform config directory, `0600`, or the OS keychain when present |
| Scopes | `psn:identity`, `psn:presence`, `nasduck:market` |
| Accounts per install | one; use separate config files and `--session` names for more |
| Revocation | immediate and server-side, invalidates outstanding session keys |

## Quick start

```console
$ nasduck run market-terminal
$ nasduck expect text "MARKET"
$ nasduck click text "DISCOVER"
$ nasduck find text "$" --full
$ nasduck screenshot -o discovery.svg
```

Launching a token, from the terminal or from code:

```console
$ nasduck click text "LAUNCH"
$ nasduck submit "DUCK/SOL"
$ nasduck expect text "SUPPLY"
$ nasduck submit "1000000"
$ nasduck click text "CONFIRM"
```

The same flow in Python:

```python
import asyncio
from nasduck import Nasduck

async def main():
    async with Nasduck.ephemeral() as terminal:
        await terminal.run("market-terminal")
        await terminal.get_by_text("MARKET").expect()
        await terminal.get_by_text("DISCOVER").click()
        await terminal.get_by_text("24h").expect()

asyncio.run(main())
```

JavaScript:

```js
import { Nasduck } from "@shiplaone2q/nasduck";

const terminal = Nasduck.ephemeral();

try {
  await terminal.run("market-terminal");
  await terminal.getByText("MARKET").expect();
  await terminal.getByText("DISCOVER").click();
  await terminal.getByText("24h").expect();
} finally {
  await terminal.closeQuiet();
}
```

Rust:

```rust
use nasduck::{OpenOptions, Operation, Session};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let terminal = Session::new("market");
    terminal.open(OpenOptions::default())?;
    terminal.execute(Operation::Submit {
        data: Some("market-terminal".into()),
    })?;
    terminal.get_by_text("MARKET").expect()?;
    terminal.close()?;
    Ok(())
}
```

Go:

```go
terminal, err := nasduck.Ephemeral("market", nasduck.ClientOptions{})
if err != nil {
    log.Fatal(err)
}
defer terminal.CloseQuiet()

if _, err := terminal.Run("market-terminal", nil, nasduck.SpawnOptions{}); err != nil {
    log.Fatal(err)
}
if err := terminal.GetByText("MARKET", nasduck.TextSelectorOptions{}).Expect(nasduck.LocatorExpectOptions{}); err != nil {
    log.Fatal(err)
}
```

<img src="https://i.imgur.com/ZYhev0d.png" alt="Market terminal" width="100%">

## Market terminal reference

### Screens

| Screen | What it does |
| --- | --- |
| `DISCOVER` | Live feed of recent launches, ordered by age, volume or momentum. |
| `SEARCH` | Ticker and issuer lookup with fuzzy fallback. |
| `BOOK` | Depth, spread, last, and the trade ticker for one token. |
| `LAUNCH` | Token creation: ticker, supply, curve, description, image hash. |
| `POSITIONS` | Open size, average entry, unrealized PnL, and exit buttons. |
| `WALLET` | In-game balance, deposits, withdrawals and pending settlements. |

### Global options

| Option | Description |
| --- | --- |
| `--session NAME` | Select a session. Default: `default` or `NASDUCK__SESSION`. |
| `--json` | Print JSON. |
| `--verbose`, `-v` | Write a session log. |
| `--failure-artifacts DIR` | Write structured assertion artifacts. |
| `--failure-artifact-mode MODE` | Select `none`, `text`, `html` or `all`. |
| `--diagnostic-context KEY=VALUE` | Attach caller context to failure details. |

### Sessions

| Command | Description |
| --- | --- |
| `open [options]` | Open a shell. |
| `run [options] PROGRAM [ARGS...]` | Run a program. |
| `restart [--graceful-timeout MS]` | Restart the session. |
| `sessions` | List sessions. |
| `close [--all]` | Close one or all sessions. |
| `daemon start` / `status` / `stop` | Manage the session daemon. |

### Locators

| Command | Purpose |
| --- | --- |
| `find text TEXT` | Return current matches and cell spans. |
| `expect text TEXT` | Retry until the locator passes. |
| `click text TEXT` | Retry, then click the middle cell. |
| `highlight text TEXT` | Mark matches in screenshots and the live monitor. |

| Locator option | Effect |
| --- | --- |
| `--regex` | Treat the text as a regular expression. |
| `--full` | Include scrollback. |
| `--whitespace exact\|normalize` | Choose whitespace matching. |
| `--after-text`, `--before-text` | Search relative to an anchor. |
| `--match any\|unique\|first\|last`, `--nth N` | Select among matches. |
| `--fg`, `--bg`, `--bold`, `--inverse` | Filter by cell style. |

`expect text` also accepts `--not` and `--timeout MS`.

### Read state

| Command | Description |
| --- | --- |
| `state` | Print session state and visible text. |
| `text [--full]` | Print terminal text. |
| `cells X Y [W H]` | Return cells and styles. |
| `get output` | Return the last command output. |
| `get clipboard` | Return the session clipboard. |

### Wait and assert

| Command | Description |
| --- | --- |
| `wait idle` | Wait for the screen to stop changing. |
| `wait ready` | Wait for the terminal prompt. |
| `expect text TEXT` | Retry until the locator passes. |
| `expect snapshot NAME` | Assert a stored snapshot. |

## Configuration

`nasduck.toml`, resolved from the current directory, then the platform config directory, then `~/.nasduck`. Override with `--config PATH` or `NASDUCK__CONFIG`.

```toml
[psn]
account = "bobblockhain"
scopes = ["psn:identity", "psn:presence", "nasduck:market"]

[market]
default_screen = "discover"
quote = "usd"
slippage_bps = 75
confirm_over = 500

[recording]
directory = "./casts"

[trace]
mode = "on-failure"
directory = "./traces"

[profiles.default]
scrollback = 10000

[profiles.default.colors]
background = "#000000"
foreground = "#c0c0c0"
red = "#800000"
```

| Key | Effect |
| --- | --- |
| `market.quote` | Quote currency for prices and PnL. |
| `market.slippage_bps` | Slippage tolerance in basis points. |
| `market.confirm_over` | Size above which a second confirmation is required; `0` confirms everything. |
| `recording.directory` | Where recordings land. |
| `trace.mode` | `off`, `on-failure` or `on`. |

## Recording, monitoring and traces

Every session can be captured. Recordings are useful for bug reports, for stream overlays and for demo material.

```console
$ nasduck record start run.gif --fps 20 --zoom 2
$ nasduck record stop
$ nasduck monitor --interactive
```

| Extension | Format | Note |
| --- | --- | --- |
| `.png`, `.apng` | still / animated PNG | transparency supported |
| `.gif` | GIF | transparency supported |
| `.mp4` | MP4 | needs `ffmpeg`, no transparency |
| `.cast` | asciinema | no canvas options |

Failures write a trace when `[trace] mode` is `"on-failure"`:

```console
$ nasduck --failure-artifacts ./artifacts expect text "POSITIONS"
error: assertion failed after 5000ms
  trace: ./artifacts/trace.md
```

| Artifact | Audience |
| --- | --- |
| `trace.md` | read this first |
| `trace.json`, `timeline.json` | structured evidence for agents |
| `trace.html`, `session.cast` | humans |

![nasduck trace viewer showing a failed terminal assertion](static/trace-viewer.png)

## Bindings

All four bindings expose the same surface and share one native adapter. Sessions run in-process; none of them shell out to the CLI.

| Surface | Reference |
| --- | --- |
| Rust | [docs.rs](https://docs.rs/nasduck-rs/latest/nasduck/) |
| Python | [bindings/python/README.md](bindings/python/README.md) |
| JavaScript | [bindings/js/README.md](bindings/js/README.md) |
| Go | [bindings/go/README.md](bindings/go/README.md) |

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Success |
| `1` | Wait or assertion failed |
| `2` | Invalid usage |
| `3` | No session |
| `4` | Daemon or IPC error |
| `5` | Internal error |

## Security model

| Property | How it holds |
| --- | --- |
| Credentials | never stored — device-code approval, refresh token only |
| Order signing | local, per session, key never transmitted |
| Scopes | identity, presence, market — no friends, trophies or store |
| Game files | untouched — attach-only, no patching, no injected DLLs |
| Revocation | `nasduck psn unlink` kills the token and every session key |
| Waits | every locator and network call is bounded by a timeout |

## FAQ

**Does this need a jailbreak or a modded console?**
No. The engine attaches to the terminal bridge from a normal machine on the same network.

**Will it get me banned?**
It reads a surface the game already renders and talks to official endpoints. Treat it like a second controller, not a trainer.

**Can I trade from a script?**
Yes. The bindings expose the same locators the CLI uses, and `--json` gives you machine-readable output everywhere.

**Does it work on PS5 and PS4?**
The bridge is console-agnostic; anything that can host the terminal works.

**Do I need the engine running to trade in-game?**
No. The surface works on its own. The engine adds automation, recording and external tooling.

**Multi-account?**
One account per install. Keep separate config files and separate `--session` names.

**Where do recordings and traces go?**
Wherever `[recording] directory` and `[trace] directory` point. Nothing is uploaded.

**The public API is in preview.**
Limits and error codes can change; the changelog in the repository is authoritative.

## Roadmap

| State | Item |
| --- | --- |
| done | Market terminal surface and locators |
| done | PSN device-code linking |
| done | Recording, monitoring, structured traces |
| next | Limit orders and stop triggers |
| next | Portfolio screen with cost basis export |
| later | Webhook and websocket feed for external bots |
| later | Snapshot packs per screen for regression suites |

## Disclaimer

NASDUCK is an unofficial fan project. Not affiliated with, endorsed by or sponsored by Rockstar Games or Sony Interactive Entertainment. GTA, Grand Theft Auto and PlayStation Network are trademarks of their respective owners. Tokens traded through NASDUCK are not securities and carry no expectation of profit.

The hosted API is a preview surface and is provided as-is.

## Contributing

Issues and pull requests are welcome. Most contributions require agreeing to the CLA; the bot decorates your PR with what to do. This project follows the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).

## License

[MIT](LICENSE)
