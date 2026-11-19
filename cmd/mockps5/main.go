// mockps5 imitates the payload so you can test the companion without a console.
//
//	go run ./cmd/mockps5 -title PPSA01234 -name "Test Game"
//	go run ./cmd/mockps5 -idle     # simulates the home screen
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
)

func main() {
	port := flag.Int("port", 9045, "listen port")
	id := flag.String("title", "PPSA01234", "title ID")
	name := flag.String("name", "Mock Game", "title name")
	idle := flag.Bool("idle", false, "report no running game")
	flag.Parse()

	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("mock PS5 payload on %s", l.Addr())
	for {
		c, err := l.Accept()
		if err != nil {
			continue
		}
		b, _ := json.Marshal(map[string]any{
			"v": 1, "running": !*idle, "title_id": *id, "title_name": *name,
		})
		c.Write(append(b, '\n'))
		c.Close()
	}
}
