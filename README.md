# ps5-discord-rpc

Show the game you're playing on a jailbroken PS5 as Discord Rich Presence.

```
 PS5 (payload, TCP :9045)  <--poll--  PC (companion)  --IPC-->  Discord client
```

Discord Rich Presence only works through the local Discord client, so a
companion app on your PC is required. The payload on the console is tiny and
stateless ([protocol](docs/PROTOCOL.md)).

## Status

| part              | state                                                        |
|-------------------|--------------------------------------------------------------|
| companion (Go)    | working, tested against a mock                               |
| payload networking| working on console (TCP :9045)                               |
| game detection    | verified on PS5 (fw 13.20, Relapse) with Ghost of Yotei      |

## Download

Prebuilt companion binaries (Windows, macOS, Linux) and `ps5-rpc.elf` are on the
Releases page. Verify downloads against `SHA256SUMS.txt`.

## Try it without a PS5

```bash
go run ./cmd/mockps5 -title PPSA01234 -name "Test Game"
go run ./cmd/companion --host 127.0.0.1 --app-id <DISCORD_APP_ID>
```

Create the app ID at https://discord.com/developers/applications and upload a
Rich Presence art asset named `ps5` (or pass `--image`).

## Real use

1. Build and send `payload/ps5-rpc.elf` with the ps5-payload-sdk (`make -C payload`).
2. `go run ./cmd/companion --host <PS5_IP> --app-id <ID>`

## Game covers

Covers come from a community catalog compiled into the binary
([`internal/covers/catalog.json`](internal/covers/catalog.json), see
[CONTRIBUTING](CONTRIBUTING.md)). Put your own entries in `covers.json`
(see `covers.example.json`); they override the catalog. Games without a cover
use the `ps5` art asset.

## Roadmap

- Test more games and firmware versions
- Grow the cover catalog
- Tray icon / config file / prebuilt releases
