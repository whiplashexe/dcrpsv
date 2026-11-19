// Package presence turns PS5 status changes into Discord Rich Presence updates.
package presence

import (
	"log"
	"time"

	"github.com/hugolgst/rich-go/client"

	"github.com/whiplashexe/dcrpsv/internal/covers"
	"github.com/whiplashexe/dcrpsv/internal/ps5"
)

// Options configure the presence.
type Options struct {
	AppID       string // Discord application (client) ID
	LargeImage  string // asset key uploaded in the Discord developer portal
	LargeText   string
	Covers      covers.Map // optional title ID -> image
	MaxFailures int        // consecutive poll failures before presence is cleared
}

// Manager keeps Discord in sync with the console.
type Manager struct {
	opt       Options
	loggedIn  bool
	current   string // title ID currently shown ("" = nothing)
	startedAt time.Time
	failures  int
}

func New(opt Options) *Manager {
	if opt.MaxFailures <= 0 {
		opt.MaxFailures = 3
	}
	return &Manager{opt: opt}
}

// Update applies one poll result (err != nil means the poll failed).
func (m *Manager) Update(st ps5.Status, err error) {
	if err != nil {
		m.failures++
		if m.failures == m.opt.MaxFailures {
			log.Printf("console unreachable (%v); clearing presence", err)
			m.clear()
		}
		return
	}
	m.failures = 0

	if !st.Running || st.TitleID == "" {
		m.clear()
		return
	}
	if st.TitleID == m.current {
		return // unchanged; Discord keeps the timer running by itself
	}

	if err := m.login(); err != nil {
		log.Printf("discord not available: %v", err)
		return
	}
	m.startedAt = time.Now()
	name := st.TitleName
	if name == "" {
		name = st.TitleID
	}
	act := client.Activity{
		Details:    name,
		State:      "Playing on PS5",
		LargeImage: m.opt.Covers.Lookup(st.TitleID, m.opt.LargeImage),
		LargeText:  m.opt.LargeText,
		Timestamps: &client.Timestamps{Start: &m.startedAt},
	}
	if err := client.SetActivity(act); err != nil {
		log.Printf("set activity failed: %v", err)
		m.loggedIn = false // force re-login next time (Discord may have restarted)
		return
	}
	m.current = st.TitleID
	log.Printf("now playing: %s (%s)", name, st.TitleID)
}

func (m *Manager) login() error {
	if m.loggedIn {
		return nil
	}
	if err := client.Login(m.opt.AppID); err != nil {
		return err
	}
	m.loggedIn = true
	return nil
}

func (m *Manager) clear() {
	if m.current == "" && !m.loggedIn {
		return
	}
	if m.loggedIn {
		client.Logout()
		m.loggedIn = false
	}
	if m.current != "" {
		log.Printf("stopped playing: %s", m.current)
	}
	m.current = ""
}

// Close releases the Discord connection.
func (m *Manager) Close() { m.clear() }
