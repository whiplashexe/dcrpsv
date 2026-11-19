// Package covers maps PS5 title IDs to cover image URLs.
//
// covers.json looks like:
//
//	{ "PPSA26344": "https://example.com/ghost-of-yotei.png" }
//
// The value may also be a Discord asset key (no scheme) that you uploaded in
// the developer portal under Rich Presence -> Art Assets.
package covers

import (
	_ "embed"
	"encoding/json"
	"os"
)

//go:embed catalog.json
var catalogJSON []byte

type Map map[string]string

// Load reads a covers file. A missing path or file yields an empty map.
func Load(path string) (Map, error) {
	m := Map{}
	if path == "" {
		return m, nil
	}
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return m, nil
	}
	if err != nil {
		return m, err
	}
	return m, json.Unmarshal(b, &m)
}

// Lookup returns the image for a title, or fallback if none is configured.
func (m Map) Lookup(titleID, fallback string) string {
	if v, ok := m[titleID]; ok && v != "" {
		return v
	}
	return fallback
}

// Catalog returns the community catalog compiled into the binary.
func Catalog() (Map, error) {
	m := Map{}
	return m, json.Unmarshal(catalogJSON, &m)
}

// Merge returns a copy of base overridden by over.
func Merge(base, over Map) Map {
	out := Map{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range over {
		out[k] = v
	}
	return out
}
