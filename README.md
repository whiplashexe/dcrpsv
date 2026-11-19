# ps5-discord-rpc

Show the game you're playing on a jailbroken PS5 as Discord Rich Presence.

```
 PS5 (payload, TCP :9045)  <--poll--  PC (companion)  --IPC-->  Discord desktop app
```

Discord Rich Presence only works through the local Discord desktop app, so a
small companion program runs on your PC. The payload on the console is tiny and
stateless ([protocol](docs/PROTOCOL.md)): it just answers "what is running?".

## Status

| part               | state                                                   |
|--------------------|---------------------------------------------------------|
| companion (Go)     | working                                                 |
| payload networking | working on console (TCP :9045)                          |
| game detection     | verified on PS5 firmware 13.20 (Relapse), one game      |
| other firmware     | untested - reports welcome                              |

## Requirements

**Console**
- A jailbroken PS5 that can run ELF payloads (an ELF loader such as
  [ps5-payload-elfldr](https://github.com/john-tornblom/ps5-payload-elfldr) must be running).
- The payload does not survive a reboot. Jailbreak and send it again after every boot.

**PC** (Windows, macOS or Linux)
- The **Discord desktop app**, running and logged in. The web version does not work.
- Discord: *Settings -> Activity Privacy -> Share your detected activity* must be on.
- The PC and the PS5 on the same local network. A fixed IP (DHCP reservation) for the PS5 saves you from editing the address later.

**To build from source** (optional)
- Go 1.22+ for the companion.
- [ps5-payload-sdk](https://github.com/ps5-payload-dev/sdk) plus LLVM/clang 18 and lld for the payload.

## Setup

### 1. Create a Discord application (once)

1. Open <https://discord.com/developers/applications> and create an application.
   Its **name is what Discord shows** after "Playing" (e.g. `PlayStation 5`).
2. Copy the **Application ID** from *General Information*.
3. Under *Rich Presence -> Art Assets*, upload a square image named `ps5`.
   It is the fallback picture for games that have no cover.

### 2. Put the payload on the PS5

1. Jailbreak the console and start your ELF loader.
2. Find the loader's port. It differs between setups (commonly 9020 or 9021):
   ```bash
   nc -zv <PS5_IP> 9020 9021
   ```
3. Send the payload to the port that answers:
   ```bash
   nc -w 1 <PS5_IP> <LOADER_PORT> < ps5-rpc.elf
   ```
   (or use any payload manager that can send an ELF.)
4. Check it is alive. You should get one line of JSON:
   ```bash
   nc <PS5_IP> 9045
   ```
   ```json
   {"v":1,"running":true,"title_id":"PPSA26344","title_name":"Ghost of Yōtei"}
   ```
   `"running":false` on the home screen is normal.

`ps5-rpc.elf` comes from the Releases page, or build it yourself (below).

### 3. Run the companion on your PC

Download the binary for your OS from the Releases page (verify it with
`SHA256SUMS.txt`), then:

```bash
./ps5-discord-rpc-linux-amd64 --host <PS5_IP> --app-id <APPLICATION_ID>
```

On macOS you may have to allow an unsigned binary once (*System Settings ->
Privacy & Security -> Open Anyway*).

Start a game; your Discord profile should show it within a few seconds, with a
timer that restarts when the game changes. Close the game or turn the console
off and the presence is removed.

#### Options

| flag         | env var          | default      | meaning                                  |
|--------------|------------------|--------------|------------------------------------------|
| `--host`     | `PS5_HOST`       | (required)   | PS5 IP address                           |
| `--app-id`   | `DISCORD_APP_ID` | (required)   | Discord Application ID                   |
| `--port`     |                  | `9045`       | payload port                             |
| `--interval` |                  | `5s`         | how often to ask the console             |
| `--image`    |                  | `ps5`        | fallback art asset name                  |
| `--covers`   |                  | `covers.json`| your own title ID -> cover map           |
| `--igdb-client-id`     | `IGDB_CLIENT_ID`     | (off) | enables automatic covers        |
| `--igdb-client-secret` | `IGDB_CLIENT_SECRET` | (off) | enables automatic covers        |
| `--cache`    |                  | `covers.cache.json` | remembers looked-up covers  |

## Game covers

Covers come from a community catalog compiled into the binary
([`internal/covers/catalog.json`](internal/covers/catalog.json)). To add one
for yourself, copy `covers.example.json` to `covers.json` next to the program:

```json
{ "PPSA26344": "https://example.com/cover-1024.png" }
```

The key is the title ID (the companion prints it every time the game changes).
The value is a public `https` URL to a square image (1024x1024 is ideal), or
the name of an art asset you uploaded in the Developer Portal. Your file
overrides the catalog. Want everyone to benefit? See [CONTRIBUTING](CONTRIBUTING.md).

### Automatic covers (optional)

For games that are in neither your `covers.json` nor the catalog, the companion
can look the cover up on [IGDB](https://www.igdb.com/) by game name:

1. Create a free Twitch application at <https://dev.twitch.tv/console/apps>
   (any redirect URL, e.g. `http://localhost`) and copy its **Client ID** and a
   generated **Client Secret**. See the [IGDB guide](https://api-docs.igdb.com/#account-creation).
2. Run the companion with both values:
   ```bash
   ./ps5-discord-rpc --host <PS5_IP> --app-id <ID> \
     --igdb-client-id <TWITCH_ID> --igdb-client-secret <TWITCH_SECRET>
   ```
   (or set `IGDB_CLIENT_ID` / `IGDB_CLIENT_SECRET`).

Lookup order: `covers.json` -> catalog -> local cache -> IGDB. Found covers are
saved in `covers.cache.json` so IGDB is asked once per game. Only an **exact**
name match is accepted; otherwise the fallback image is used and the log tells
you to add the game to `covers.json`. Keep the secret private: never commit it
or paste it in issues.

## Troubleshooting

| symptom                                   | likely cause / fix                                                       |
|-------------------------------------------|--------------------------------------------------------------------------|
| `discord not available`                   | Discord desktop is not running, or was started as admin while this wasn't |
| nothing happens, no log lines             | wrong `--host`, payload not running (re-send it after a reboot), firewall |
| `nc <IP> 9045` shows `"running":false`    | no game in the foreground, or detection failed on your firmware - open an issue |
| presence shows but the picture is missing | missing `ps5` art asset, or the cover URL is not a public image (Discord can take a few minutes to process new assets) |
| wrong or raw title ID instead of a name   | the game's `param.json` was not found at the paths tried by the payload  |
| presence disappears after a few seconds   | 3 failed polls in a row clear it; check network stability                |

## Try it without a PS5

```bash
go run ./cmd/mockps5 -title PPSA01234 -name "Test Game"
```
```bash
go run ./cmd/companion --host 127.0.0.1 --app-id <APPLICATION_ID>
```

## Building from source

```bash
# companion
go build -o ps5-discord-rpc ./cmd/companion

# payload (needs ps5-payload-sdk installed, e.g. in /opt/ps5-payload-sdk)
export PS5_PAYLOAD_SDK=/opt/ps5-payload-sdk
make -C payload          # -> payload/ps5-rpc.elf
```

Do not run `make` inside the SDK folder (it is root-owned); build in your own
checkout. Tests: `go test ./...`.

## Privacy and safety

- The payload listens on your local network with **no authentication** and only
  reveals the current game's title ID and name. Do not expose port 9045 to the internet.
- Without IGDB the companion sends nothing anywhere except to your local Discord app. With IGDB enabled it sends only the game name to IGDB (and your Twitch credentials to Twitch). Cover
  images are fetched by Discord itself from the URLs you configure.

## Disclaimer

This is an unofficial homebrew tool, not affiliated with Sony or Discord.
It does not include, and will not accept contributions of, copyrighted game
data or anything that supports piracy. Use at your own risk; running payloads
on a console can violate its terms of service.

## Roadmap

- Test more games and firmware versions
- Grow the cover catalog
- Tray icon / config file

## License

[MIT](LICENSE)
