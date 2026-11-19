package covers

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestLookup(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c.json")
	os.WriteFile(p, []byte(`{"PPSA1":"https://x/y.png"}`), 0o600)
	m, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Lookup("PPSA1", "ps5"); got != "https://x/y.png" {
		t.Fatalf("got %q", got)
	}
	if got := m.Lookup("PPSA2", "ps5"); got != "ps5" {
		t.Fatalf("fallback got %q", got)
	}
	if _, err := Load(filepath.Join(t.TempDir(), "missing.json")); err != nil {
		t.Fatal("missing file should not error")
	}
}

// Guards community pull requests: valid title IDs and https URLs only.
func TestCatalogValid(t *testing.T) {
	m, err := Catalog()
	if err != nil {
		t.Fatalf("catalog.json is not valid JSON: %v", err)
	}
	re := regexp.MustCompile(`^PPSA\d{5}$`)
	for id, url := range m {
		if !re.MatchString(id) {
			t.Errorf("bad title id %q (want PPSAxxxxx)", id)
		}
		if !strings.HasPrefix(url, "https://") {
			t.Errorf("%s: cover must be an https URL, got %q", id, url)
		}
	}
}

func TestMerge(t *testing.T) {
	got := Merge(Map{"A": "1", "B": "2"}, Map{"B": "x"})
	if got["A"] != "1" || got["B"] != "x" {
		t.Fatalf("got %v", got)
	}
}
