// Package api is the backend's public HTTP API (see docs/architecture.md).
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/lukamrko/VibeRiichiMahjong/backend/internal/hand"
	"github.com/lukamrko/VibeRiichiMahjong/backend/internal/score"
	"github.com/lukamrko/VibeRiichiMahjong/backend/internal/tile"
)

// maxBodyBytes caps the request body size.
const maxBodyBytes = 64 << 10

// ErrorCode is a machine-readable error code in an ErrorResponse (closed list,
// see docs/architecture.md).
type ErrorCode string

// Error codes.
const (
	CodeInvalidJSON    ErrorCode = "invalid_json"
	CodeBodyTooLarge   ErrorCode = "body_too_large"
	CodeInvalidHand    ErrorCode = "invalid_hand"
	CodeInvalidContext ErrorCode = "invalid_context"
	CodeNotWinningHand ErrorCode = "not_winning_hand"
	CodeNoYaku         ErrorCode = "no_yaku"
	CodeInternal       ErrorCode = "internal"
)

// NewHandler returns the API's HTTP handler.
func NewHandler(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /score", func(w http.ResponseWriter, r *http.Request) {
		handleScore(w, r, logger)
	})
	return mux
}

// ScoreRequest is the body of POST /score.
type ScoreRequest struct {
	Hand        string         `json:"hand"`
	WinningTile string         `json:"winningTile"`
	Context     ContextRequest `json:"context"`
}

// ContextRequest is the WinContext part of a ScoreRequest.
type ContextRequest struct {
	RoundWind         string   `json:"roundWind"`
	SeatWind          string   `json:"seatWind"`
	WinType           string   `json:"winType"`
	Riichi            bool     `json:"riichi"`
	DoubleRiichi      bool     `json:"doubleRiichi"`
	Ippatsu           bool     `json:"ippatsu"`
	DoraIndicators    []string `json:"doraIndicators"`
	UraDoraIndicators []string `json:"uraDoraIndicators"`
	Honba             int      `json:"honba"`
	Haitei            bool     `json:"haitei"`
	Houtei            bool     `json:"houtei"`
	Rinshan           bool     `json:"rinshan"`
	Chankan           bool     `json:"chankan"`
	Tenhou            bool     `json:"tenhou"`
	Chiihou           bool     `json:"chiihou"`
}

// ScoreResponse is the 200 body of POST /score.
type ScoreResponse struct {
	Yaku        []YakuResponse `json:"yaku"`
	Dora        int            `json:"dora"`
	AkaDora     int            `json:"akaDora"`
	UraDora     int            `json:"uraDora"`
	Han         int            `json:"han"`
	Fu          *int           `json:"fu"` // null for yakuman
	FuBreakdown []FuResponse   `json:"fuBreakdown"`
	Limit       *string        `json:"limit"` // null below mangan
	Points      PointsResponse `json:"points"`
}

// YakuResponse is one yaku in a ScoreResponse.
type YakuResponse struct {
	Name string `json:"name"`
	Han  int    `json:"han"`
}

// FuResponse is one fu item in a ScoreResponse.
type FuResponse struct {
	Reason string `json:"reason"`
	Fu     int    `json:"fu"`
}

// PointsResponse holds the payments; which keys are present depends on the win
// (payments are never 0, so omitempty drops exactly the keys that do not apply).
type PointsResponse struct {
	Total             int `json:"total"`
	FromDiscarder     int `json:"fromDiscarder,omitempty"`
	FromDealer        int `json:"fromDealer,omitempty"`
	FromNonDealer     int `json:"fromNonDealer,omitempty"`
	FromEachNonDealer int `json:"fromEachNonDealer,omitempty"`
}

// ErrorResponse is the body of every non-2xx response.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody carries a machine-readable code and a human-readable message.
type ErrorBody struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func handleScore(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	var req ScoreRequest
	if status, code, err := decode(w, r, &req); err != nil {
		writeError(w, logger, status, code, err)
		return
	}
	h, err := hand.Parse(req.Hand, req.WinningTile)
	if err != nil {
		writeError(w, logger, http.StatusBadRequest, CodeInvalidHand, err)
		return
	}
	ctx, err := toWinContext(req.Context)
	if err != nil {
		writeError(w, logger, http.StatusBadRequest, CodeInvalidContext, err)
		return
	}
	res, err := score.Score(h, ctx)
	switch {
	case err == nil:
	case errors.Is(err, score.ErrInvalidContext):
		writeError(w, logger, http.StatusBadRequest, CodeInvalidContext, err)
		return
	case errors.Is(err, score.ErrNotWinningHand):
		writeError(w, logger, http.StatusUnprocessableEntity, CodeNotWinningHand, err)
		return
	case errors.Is(err, score.ErrNoYaku):
		writeError(w, logger, http.StatusUnprocessableEntity, CodeNoYaku, err)
		return
	default:
		writeError(w, logger, http.StatusInternalServerError, CodeInternal, err)
		return
	}
	writeJSON(w, logger, http.StatusOK, toResponse(res))
}

// decode reads exactly one JSON object with no unknown fields.
func decode(w http.ResponseWriter, r *http.Request, v any) (int, ErrorCode, error) {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	err := dec.Decode(v)
	if err == nil {
		if dec.Decode(&struct{}{}) != io.EOF {
			err = errors.New("unexpected data after the JSON object")
		}
	}
	var tooLarge *http.MaxBytesError
	switch {
	case err == nil:
		return 0, "", nil
	case errors.As(err, &tooLarge):
		return http.StatusRequestEntityTooLarge, CodeBodyTooLarge, fmt.Errorf("body exceeds %d bytes", tooLarge.Limit)
	default:
		return http.StatusBadRequest, CodeInvalidJSON, err
	}
}

func toWinContext(c ContextRequest) (score.WinContext, error) {
	round, err := score.ParseWind(c.RoundWind)
	if err != nil {
		return score.WinContext{}, fmt.Errorf("roundWind: %w", err)
	}
	seat, err := score.ParseWind(c.SeatWind)
	if err != nil {
		return score.WinContext{}, fmt.Errorf("seatWind: %w", err)
	}
	winType, err := score.ParseWinType(c.WinType)
	if err != nil {
		return score.WinContext{}, err
	}
	dora, err := parseIndicators(c.DoraIndicators)
	if err != nil {
		return score.WinContext{}, fmt.Errorf("doraIndicators: %w", err)
	}
	ura, err := parseIndicators(c.UraDoraIndicators)
	if err != nil {
		return score.WinContext{}, fmt.Errorf("uraDoraIndicators: %w", err)
	}
	return score.WinContext{
		RoundWind: round, SeatWind: seat, WinType: winType,
		Riichi: c.Riichi, DoubleRiichi: c.DoubleRiichi, Ippatsu: c.Ippatsu,
		DoraIndicators: dora, UraDoraIndicators: ura,
		Honba:  c.Honba,
		Haitei: c.Haitei, Houtei: c.Houtei, Rinshan: c.Rinshan, Chankan: c.Chankan,
		Tenhou: c.Tenhou, Chiihou: c.Chiihou,
	}, nil
}

func parseIndicators(ids []string) ([]tile.Tile, error) {
	out := make([]tile.Tile, 0, len(ids))
	for _, id := range ids {
		t, err := tile.Parse(id)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

func toResponse(r score.Result) ScoreResponse {
	resp := ScoreResponse{
		Yaku:        make([]YakuResponse, 0, len(r.Yaku)),
		Dora:        r.Dora,
		AkaDora:     r.AkaDora,
		UraDora:     r.UraDora,
		Han:         r.Han,
		FuBreakdown: make([]FuResponse, 0, len(r.FuBreakdown)),
		Points: PointsResponse{
			Total:             r.Points.Total,
			FromDiscarder:     r.Points.FromDiscarder,
			FromDealer:        r.Points.FromDealer,
			FromNonDealer:     r.Points.FromNonDealer,
			FromEachNonDealer: r.Points.FromEachNonDealer,
		},
	}
	for _, y := range r.Yaku {
		resp.Yaku = append(resp.Yaku, YakuResponse{Name: y.Yaku.String(), Han: y.Han})
	}
	for _, f := range r.FuBreakdown {
		resp.FuBreakdown = append(resp.FuBreakdown, FuResponse{Reason: f.Reason.String(), Fu: f.Fu})
	}
	if !r.IsYakuman() {
		fu := r.Fu
		resp.Fu = &fu
	}
	if r.Limit != score.NoLimit {
		limit := r.Limit.String()
		resp.Limit = &limit
	}
	return resp
}

func writeError(w http.ResponseWriter, logger *slog.Logger, status int, code ErrorCode, err error) {
	if status >= 500 {
		logger.Error("score request failed", "code", code, "err", err)
	} else {
		logger.Debug("score request rejected", "code", code, "err", err)
	}
	writeJSON(w, logger, status, ErrorResponse{Error: ErrorBody{Code: code, Message: err.Error()}})
}

func writeJSON(w http.ResponseWriter, logger *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Error("writing response", "err", err)
	}
}
