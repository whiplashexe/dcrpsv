package covers

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func fakeIGDB(t *testing.T, apiBody string) (*IGDB, *int32) {
	var apiCalls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("grant_type") != "client_credentials" {
			http.Error(w, "bad", 400)
			return
		}
		w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
	})
	mux.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&apiCalls, 1)
		if r.Header.Get("Authorization") != "Bearer tok" || r.Header.Get("Client-ID") != "id" {
			http.Error(w, "unauth", 401)
			return
		}
		w.Write([]byte(apiBody))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	g := NewIGDB("id", "secret")
	g.AuthURL, g.APIURL = srv.URL+"/token", srv.URL+"/games"
	return g, &apiCalls
}

func TestIGDBExactMatchOnly(t *testing.T) {
	g, _ := fakeIGDB(t, `[
	  {"name":"Ghost of Tsushima","cover":{"image_id":"wrong"}},
	  {"name":"Ghost of Yōtei","cover":{"image_id":"abc123"}}]`)
	u, err := g.Lookup("Ghost of Yōtei")
	if err != nil || !strings.HasSuffix(u, "/t_cover_big/abc123.jpg") {
		t.Fatalf("got %q, %v", u, err)
	}
	if u, _ := g.Lookup("Some Other Game"); u != "" {
		t.Fatalf("non-matching name must not get a cover, got %q", u)
	}
}

func TestResolverOrderAndCache(t *testing.T) {
	g, calls := fakeIGDB(t, `[{"name":"Game","cover":{"image_id":"zzz"}}]`)
	cache := filepath.Join(t.TempDir(), "cache.json")
	r := &Resolver{Static: Map{"PPSA1": "https://static"}, Fallback: "ps5", IGDB: g, CachePath: cache}

	if got := r.Cover("PPSA1", "Game"); got != "https://static" {
		t.Fatalf("static must win, got %q", got)
	}
	if atomic.LoadInt32(calls) != 0 {
		t.Fatal("static hit must not call IGDB")
	}
	first := r.Cover("PPSA2", "Game")
	if !strings.Contains(first, "zzz") {
		t.Fatalf("got %q", first)
	}
	// new Resolver, same cache file: no network call needed
	r2 := &Resolver{Fallback: "ps5", IGDB: g, CachePath: cache}
	if got := r2.Cover("PPSA2", "Game"); got != first || atomic.LoadInt32(calls) != 1 {
		t.Fatalf("cache not used: %q calls=%d", got, atomic.LoadInt32(calls))
	}
	if got := r2.Cover("PPSA3", "Nope"); got != "ps5" {
		t.Fatalf("fallback expected, got %q", got)
	}
	// a miss is remembered for the run: no second API call
	n := atomic.LoadInt32(calls)
	r2.Cover("PPSA3", "Nope")
	if atomic.LoadInt32(calls) != n {
		t.Fatal("miss was not remembered")
	}
}

func TestResolverWithoutIGDB(t *testing.T) {
	r := &Resolver{Fallback: "ps5"}
	if got := r.Cover("PPSA9", "X"); got != "ps5" {
		t.Fatalf("got %q", got)
	}
}
