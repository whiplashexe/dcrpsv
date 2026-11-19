package covers

import (
	"encoding/json"
	"log"
	"os"
)

// Resolver picks a cover in this order:
// static (your covers.json over the community catalog) -> disk cache -> IGDB.
// IGDB hits are cached on disk; misses are remembered only for this run.
type Resolver struct {
	Static    Map
	Fallback  string
	IGDB      *IGDB // optional
	CachePath string

	cache  Map
	missed map[string]bool
}

func (r *Resolver) load() {
	if r.cache != nil {
		return
	}
	r.cache, _ = Load(r.CachePath)
	r.missed = map[string]bool{}
}

// Cover returns the image for a title. It never returns an error: any
// failure falls back to the default image.
func (r *Resolver) Cover(titleID, name string) string {
	r.load()
	if v, ok := r.Static[titleID]; ok && v != "" {
		return v
	}
	if v, ok := r.cache[titleID]; ok && v != "" {
		return v
	}
	if r.IGDB == nil || r.missed[titleID] {
		return r.Fallback
	}
	u, err := r.IGDB.Lookup(name)
	if err != nil {
		log.Printf("cover lookup failed for %q: %v", name, err)
		r.missed[titleID] = true
		return r.Fallback
	}
	if u == "" {
		log.Printf("no confident cover match for %q (%s); add it to covers.json", name, titleID)
		r.missed[titleID] = true
		return r.Fallback
	}
	log.Printf("cover found for %q", name)
	r.cache[titleID] = u
	r.save()
	return u
}

func (r *Resolver) save() {
	if r.CachePath == "" {
		return
	}
	b, _ := json.MarshalIndent(r.cache, "", "  ")
	if err := os.WriteFile(r.CachePath, b, 0o600); err != nil {
		log.Printf("cannot write cover cache: %v", err)
	}
}
