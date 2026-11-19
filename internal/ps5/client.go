// Package ps5 talks to the ps5-rpc payload running on a jailbroken console.
//
// Protocol (v1): the payload listens on a TCP port. Every connection receives
// exactly one line of JSON and is then closed. See docs/PROTOCOL.md.
package ps5

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Status is what the payload reports about the console.
type Status struct {
	Version   int    `json:"v"`
	Running   bool   `json:"running"`    // a game/app is in the foreground
	TitleID   string `json:"title_id"`   // e.g. PPSA01234
	TitleName string `json:"title_name"` // e.g. "Astro's Playroom"
}

// Client polls a payload.
type Client struct {
	Addr    string
	Timeout time.Duration
}

func New(host string, port int) *Client {
	return &Client{Addr: net.JoinHostPort(host, fmt.Sprint(port)), Timeout: 3 * time.Second}
}

// Poll fetches the current status once.
func (c *Client) Poll() (Status, error) {
	var st Status
	conn, err := net.DialTimeout("tcp", c.Addr, c.Timeout)
	if err != nil {
		return st, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(c.Timeout))

	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil && len(line) == 0 {
		return st, err
	}
	if err := json.Unmarshal(line, &st); err != nil {
		return st, fmt.Errorf("bad payload response: %w", err)
	}
	return st, nil
}
