package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// fixturesDir is the shared, language-agnostic fixture directory (repo root).
const fixturesDir = "../../../fixtures/hands"

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil))))
	t.Cleanup(srv.Close)
	return srv
}

func post(t *testing.T, srv *httptest.Server, body []byte) (*http.Response, []byte) {
	t.Helper()
	resp, err := http.Post(srv.URL+"/score", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp, data
}

// TestScoreFixtures POSTs every fixture input to a real HTTP server and
// compares the JSON response with the fixture's expected output exactly.
func TestScoreFixtures(t *testing.T) {
	srv := newServer(t)
	files, err := filepath.Glob(filepath.Join(fixturesDir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatalf("no fixture files in %s", fixturesDir)
	}
	n := 0
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var cases []struct {
			ID       string          `json:"id"`
			Input    json.RawMessage `json:"input"`
			Expected any             `json:"expected"`
		}
		if err := json.Unmarshal(data, &cases); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		for _, c := range cases {
			n++
			t.Run(c.ID, func(t *testing.T) {
				resp, body := post(t, srv, c.Input)
				if resp.StatusCode != http.StatusOK {
					t.Fatalf("status %d: %s", resp.StatusCode, body)
				}
				if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("Content-Type = %q", ct)
				}
				var got any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, c.Expected) {
					want, _ := json.Marshal(c.Expected)
					t.Errorf("response mismatch\n got: %s\nwant: %s", body, want)
				}
			})
		}
	}
	t.Logf("%d fixture cases", n)
}

const validRequest = `{
  "hand": "123m45699p345678s", "winningTile": "6s",
  "context": {
    "roundWind": "1z", "seatWind": "2z", "winType": "ron",
    "riichi": true, "doubleRiichi": false, "ippatsu": false,
    "doraIndicators": ["1z"], "uraDoraIndicators": ["7z"], "honba": 0,
    "haitei": false, "houtei": false, "rinshan": false, "chankan": false, "tenhou": false, "chiihou": false
  }
}`

func TestScoreErrors(t *testing.T) {
	srv := newServer(t)
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{"malformed JSON", `{"hand": `, 400, "invalid_json"},
		{"unknown field", strings.Replace(validRequest, `"hand"`, `"hnd": "x", "hand"`, 1), 400, "invalid_json"},
		{"trailing data", validRequest + `{}`, 400, "invalid_json"},
		{"wrong type", strings.Replace(validRequest, `"honba": 0`, `"honba": "0"`, 1), 400, "invalid_json"},
		{"missing hand", strings.Replace(validRequest, `"hand": "123m45699p345678s",`, ``, 1), 400, "invalid_hand"},
		{"bad MPSZ", strings.Replace(validRequest, `123m45699p345678s`, `123x`, 1), 400, "invalid_hand"},
		{"13 tiles", strings.Replace(validRequest, `123m45699p345678s`, `123m4569p345678s`, 1), 400, "invalid_hand"},
		{"winning tile not in hand", strings.Replace(validRequest, `"winningTile": "6s"`, `"winningTile": "7z"`, 1), 400, "invalid_hand"},
		{"bad wind", strings.Replace(validRequest, `"roundWind": "1z"`, `"roundWind": "5z"`, 1), 400, "invalid_context"},
		{"bad win type", strings.Replace(validRequest, `"winType": "ron"`, `"winType": "draw"`, 1), 400, "invalid_context"},
		{"bad dora indicator", strings.Replace(validRequest, `["1z"]`, `["1z2z"]`, 1), 400, "invalid_context"},
		{"contradiction", strings.Replace(validRequest, `"riichi": true`, `"riichi": false`, 1), 400, "invalid_context"},
		{"not a winning hand", strings.Replace(validRequest, `123m45699p345678s`, `123m45699p345679s`, 1), 422, "not_winning_hand"},
		{"no yaku", `{"hand": "123m456p234789s55z", "winningTile": "6p", "context": {
			"roundWind": "1z", "seatWind": "2z", "winType": "ron", "doraIndicators": ["1z"]}}`, 422, "no_yaku"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := post(t, srv, []byte(tt.body))
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d: %s", resp.StatusCode, tt.wantStatus, body)
			}
			var e struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.Unmarshal(body, &e); err != nil {
				t.Fatalf("error body %q: %v", body, err)
			}
			if e.Error.Code != tt.wantCode || e.Error.Message == "" {
				t.Errorf("error = %+v, want code %q with a message", e.Error, tt.wantCode)
			}
		})
	}
}

func TestScoreMethodNotAllowed(t *testing.T) {
	srv := newServer(t)
	resp, err := http.Get(srv.URL + "/score")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("GET /score status = %d, want 405", resp.StatusCode)
	}
}

func TestScoreBodyTooLarge(t *testing.T) {
	srv := newServer(t)
	body := `{"hand": "` + strings.Repeat("1", maxBodyBytes) + `"}`
	resp, _ := post(t, srv, []byte(body))
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", resp.StatusCode)
	}
}
