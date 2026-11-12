package hand

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// fixturesDir is the shared, language-agnostic fixture directory (repo root).
const fixturesDir = "../../../fixtures/hands"

// TestFixtureHandsDecompose checks that every fixture hand parses and has at
// least one decomposition with the expected form. Scoring itself is #4.
func TestFixtureHandsDecompose(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(fixturesDir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatalf("no fixture files found in %s", fixturesDir)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var cases []struct {
			ID    string `json:"id"`
			Input struct {
				Hand        string `json:"hand"`
				WinningTile string `json:"winningTile"`
			} `json:"input"`
			Expected struct {
				Yaku []struct {
					Name string `json:"name"`
				} `json:"yaku"`
			} `json:"expected"`
		}
		if err := json.Unmarshal(data, &cases); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		for _, c := range cases {
			t.Run(c.ID, func(t *testing.T) {
				h, err := Parse(c.Input.Hand, c.Input.WinningTile)
				if err != nil {
					t.Fatalf("Parse(%q, %q): %v", c.Input.Hand, c.Input.WinningTile, err)
				}
				ds := h.Decompositions()
				if len(ds) == 0 {
					t.Fatalf("hand %q has no decomposition", c.Input.Hand)
				}
				want := Standard
				for _, y := range c.Expected.Yaku {
					switch y.Name {
					case "chiitoitsu":
						want = Chiitoitsu
					case "kokushi_musou", "kokushi_musou_juusanmen":
						want = Kokushi
					}
				}
				found := false
				for _, d := range ds {
					if d.Form == want {
						found = true
					}
				}
				if !found {
					t.Errorf("hand %q: no %v decomposition in %v", c.Input.Hand, want, ds)
				}
			})
		}
	}
}
