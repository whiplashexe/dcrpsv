// companion runs on your PC. It polls the ps5-rpc payload and mirrors the
// current game to Discord Rich Presence.
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/your-name/ps5-discord-rpc/internal/covers"
	"github.com/your-name/ps5-discord-rpc/internal/presence"
	"github.com/your-name/ps5-discord-rpc/internal/ps5"
)

func main() {
	host := flag.String("host", env("PS5_HOST", ""), "PS5 IP address (or PS5_HOST)")
	port := flag.Int("port", 9045, "payload TCP port")
	appID := flag.String("app-id", env("DISCORD_APP_ID", ""), "Discord application ID (or DISCORD_APP_ID)")
	image := flag.String("image", "ps5", "Discord large image asset key")
	coversPath := flag.String("covers", "covers.json", "title ID -> cover image map (optional)")
	interval := flag.Duration("interval", 5*time.Second, "poll interval")
	flag.Parse()

	if *host == "" || *appID == "" {
		flag.Usage()
		log.Fatal("--host and --app-id are required")
	}

	catalog, err := covers.Catalog()
	if err != nil {
		log.Fatalf("built-in catalog: %v", err)
	}
	mine, err := covers.Load(*coversPath)
	if err != nil {
		log.Fatalf("covers: %v", err)
	}
	cv := covers.Merge(catalog, mine) // your covers.json wins over the catalog
	log.Printf("covers: %d from catalog, %d from %s", len(catalog), len(mine), *coversPath)

	c := ps5.New(*host, *port)
	m := presence.New(presence.Options{AppID: *appID, LargeImage: *image, LargeText: "PlayStation 5", Covers: cv})
	defer m.Close()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	log.Printf("polling %s every %s", c.Addr, *interval)
	t := time.NewTicker(*interval)
	defer t.Stop()
	for {
		m.Update(c.Poll())
		select {
		case <-sig:
			return
		case <-t.C:
		}
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
