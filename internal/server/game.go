package server

import (
	"encoding/json"
	"net/http"
	"slices"
	"time"

	"habit-idle/internal/core"
	"habit-idle/internal/db"
	"habit-idle/internal/game"
)

var eventKinds = []string{"start", "recruit", "squad", "challenge", "buy", "equip", "unequip", "sell", "refine"}

type snapshot struct {
	day    core.Day
	items  []core.TodayItem
	folego game.FolegoResult
	game   *game.State
}

// compute recalcula fôlego e jogo a partir do banco.
func (s *srv) compute() (snapshot, error) {
	cal, err := s.calendar()
	if err != nil {
		return snapshot{}, err
	}
	habits, err := db.Habits(s.DB)
	if err != nil {
		return snapshot{}, err
	}
	log, err := db.CheckLog(s.DB)
	if err != nil {
		return snapshot{}, err
	}
	evs, err := db.Events(s.DB)
	if err != nil {
		return snapshot{}, err
	}
	now := s.Now()
	var start *time.Time
	events := make([]game.Event, 0, len(evs))
	for _, e := range evs {
		if e.Kind == "start" && start == nil {
			t := e.At
			start = &t
		}
		events = append(events, game.Event{ID: e.ID, At: e.At, Kind: e.Kind, Data: json.RawMessage(e.Data)})
	}
	touches := make([]game.Touch, 0, len(log))
	for _, e := range log {
		touches = append(touches, game.Touch{HabitID: e.HabitID, Day: e.Day, At: e.At, Done: e.Done})
	}
	f := game.Folego(game.FolegoInput{Cal: cal, Habits: habits, Touches: touches, Start: start, Now: now})
	st := game.Play(game.PlayInput{Cal: cal, Events: events, Segments: f.Segments, Fatigue: f.Fatigue, Complete: f.Complete, Now: now})
	day := cal.DayOf(now)
	return snapshot{day: day, items: core.TodayList(habits, db.Checks(log), day), folego: f, game: st}, nil
}

func (s *srv) state(w http.ResponseWriter, r *http.Request) {
	snap, err := s.compute()
	if err != nil {
		fail(w, err)
		return
	}
	set, err := s.settings()
	if err != nil {
		fail(w, err)
		return
	}
	f := snap.folego
	now := s.Now()
	exploring := snap.game.Started && f.Balance > 0
	var until any
	if !f.Until.IsZero() {
		until = f.Until
	}
	missions, done := 0, 0
	for _, it := range snap.items {
		if it.Mission() {
			missions++
			if it.Done {
				done++
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"day": snap.day, "now": now, "items": snap.items,
		"missions": missions, "done": done,
		"folego": map[string]any{
			"balance": int(f.Balance.Minutes()), "until": until, "exploring": exploring, "cap": int(game.FolegoCap.Minutes()),
		},
		"game":     snap.game,
		"catalog":  game.Catalog,
		"lastSeen": set["lastSeen"],
	})
}

func (s *srv) postEvent(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Kind string          `json:"kind"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		badRequest(w, err.Error())
		return
	}
	if !slices.Contains(eventKinds, b.Kind) {
		badRequest(w, "evento desconhecido")
		return
	}
	if len(b.Data) == 0 {
		b.Data = json.RawMessage("{}")
	}
	if err := db.AddEvent(s.DB, db.Event{At: s.Now(), Kind: b.Kind, Data: string(b.Data)}); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("{}"))
}

func (s *srv) seen(w http.ResponseWriter, r *http.Request) {
	if err := db.SetSetting(s.DB, "lastSeen", s.Now().UTC().Format(time.RFC3339)); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *srv) resetGame(w http.ResponseWriter, r *http.Request) {
	if err := db.ResetGame(s.DB); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
