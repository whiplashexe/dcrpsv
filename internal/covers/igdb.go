package covers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
)

// IGDB looks up cover art by game name. It needs a free Twitch application
// (client ID + secret): https://api-docs.igdb.com/#account-creation
type IGDB struct {
	ClientID, Secret string
	HTTP             *http.Client
	AuthURL, APIURL  string // overridable for tests
	CoverSize        string // IGDB image size, default t_cover_big

	token string
	exp   time.Time
}

func NewIGDB(clientID, secret string) *IGDB {
	return &IGDB{
		ClientID: clientID, Secret: secret,
		HTTP:      &http.Client{Timeout: 5 * time.Second},
		AuthURL:   "https://id.twitch.tv/oauth2/token",
		APIURL:    "https://api.igdb.com/v4/games",
		CoverSize: "t_cover_big",
	}
}

func (g *IGDB) ensureToken() error {
	if g.token != "" && time.Now().Before(g.exp) {
		return nil
	}
	q := url.Values{
		"client_id": {g.ClientID}, "client_secret": {g.Secret},
		"grant_type": {"client_credentials"},
	}
	resp, err := g.HTTP.Post(g.AuthURL+"?"+q.Encode(), "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("twitch auth: %s (check client id/secret)", resp.Status)
	}
	var t struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil || t.AccessToken == "" {
		return errors.New("twitch auth: bad response")
	}
	g.token = t.AccessToken
	g.exp = time.Now().Add(time.Duration(t.ExpiresIn-60) * time.Second)
	return nil
}

type igdbGame struct {
	Name  string `json:"name"`
	Cover *struct {
		ImageID string `json:"image_id"`
	} `json:"cover"`
}

// Lookup returns a cover URL for the game, or "" when no confident match
// exists. Only an exact (case/punctuation-insensitive) name match is accepted:
// a wrong cover is worse than the fallback image.
func (g *IGDB) Lookup(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", nil
	}
	for attempt := 0; attempt < 2; attempt++ {
		if err := g.ensureToken(); err != nil {
			return "", err
		}
		body := fmt.Sprintf(`search "%s"; fields name,cover.image_id; limit 10;`,
			strings.NewReplacer(`\`, ``, `"`, ``).Replace(name))
		req, _ := http.NewRequest(http.MethodPost, g.APIURL, bytes.NewBufferString(body))
		req.Header.Set("Client-ID", g.ClientID)
		req.Header.Set("Authorization", "Bearer "+g.token)
		resp, err := g.HTTP.Do(req)
		if err != nil {
			return "", err
		}
		if resp.StatusCode == http.StatusUnauthorized && attempt == 0 {
			resp.Body.Close()
			g.token = "" // expired or revoked: re-auth once
			continue
		}
		var games []igdbGame
		err = json.NewDecoder(resp.Body).Decode(&games)
		status := resp.Status
		ok := resp.StatusCode == http.StatusOK
		resp.Body.Close()
		if !ok {
			return "", fmt.Errorf("igdb: %s", status)
		}
		if err != nil {
			return "", fmt.Errorf("igdb: bad response: %w", err)
		}
		want := normalize(name)
		for _, gm := range games {
			if gm.Cover != nil && gm.Cover.ImageID != "" && normalize(gm.Name) == want {
				return "https://images.igdb.com/igdb/image/upload/" + g.CoverSize + "/" + gm.Cover.ImageID + ".jpg", nil
			}
		}
		return "", nil
	}
	return "", nil
}

func normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
