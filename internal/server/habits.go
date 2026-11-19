package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"habit-idle/internal/core"
	"habit-idle/internal/db"
)

var defaults = map[string]string{"dayStart": "06:00", "tz": "America/Sao_Paulo"}

func (s *srv) settings() (map[string]string, error) {
	got, err := db.Settings(s.DB)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for k, v := range defaults {
		out[k] = v
	}
	for k, v := range got {
		out[k] = v
	}
	return out, nil
}

func (s *srv) calendar() (core.Calendar, error) {
	set, err := s.settings()
	if err != nil {
		return core.Calendar{}, err
	}
	loc, err := time.LoadLocation(set["tz"])
	if err != nil {
		return core.Calendar{}, err
	}
	m, err := hhmm(set["dayStart"])
	if err != nil {
		return core.Calendar{}, err
	}
	return core.Calendar{Loc: loc, DayStart: m}, nil
}

func hhmm(s string) (int, error) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil || h < 0 || h > 23 || m < 0 || m > 59 || len(s) != 5 {
		return 0, fmt.Errorf("horário inválido: %q", s)
	}
	return h*60 + m, nil
}

func (s *srv) getSettings(w http.ResponseWriter, r *http.Request) {
	set, err := s.settings()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, set)
}

func (s *srv) putSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DayStart *string `json:"dayStart"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		badRequest(w, err.Error())
		return
	}
	if body.DayStart != nil {
		if _, err := hhmm(*body.DayStart); err != nil {
			badRequest(w, err.Error())
			return
		}
		if err := db.SetSetting(s.DB, "dayStart", *body.DayStart); err != nil {
			fail(w, err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *srv) today(w http.ResponseWriter, r *http.Request) {
	cal, err := s.calendar()
	if err != nil {
		fail(w, err)
		return
	}
	habits, err := db.Habits(s.DB)
	if err != nil {
		fail(w, err)
		return
	}
	log, err := db.CheckLog(s.DB)
	if err != nil {
		fail(w, err)
		return
	}
	day := cal.DayOf(s.Now())
	writeJSON(w, http.StatusOK, map[string]any{"day": day, "items": core.TodayList(habits, db.Checks(log), day)})
}

type habitBody struct {
	Name     string        `json:"name"`
	Effort   int           `json:"effort"`
	Schedule core.Schedule `json:"schedule"`
	Archived bool          `json:"archived"`
}

func (b habitBody) validate() string {
	b.Name = strings.TrimSpace(b.Name)
	switch {
	case b.Name == "":
		return "nome vazio"
	case b.Effort != 60 && b.Effort != 120 && b.Effort != 180:
		return "esforço deve ser 60, 120 ou 180"
	case b.Schedule.Kind == core.Days && len(b.Schedule.Days) == 0:
		return "escolha pelo menos um dia"
	case b.Schedule.Kind == core.Weekly && (b.Schedule.Times < 1 || b.Schedule.Times > 7):
		return "vezes por semana entre 1 e 7"
	case b.Schedule.Kind != core.Daily && b.Schedule.Kind != core.Days && b.Schedule.Kind != core.Weekly:
		return "frequência inválida"
	}
	for _, d := range b.Schedule.Days {
		if d < 0 || d > 6 {
			return "dia da semana inválido"
		}
	}
	return ""
}

func decodeHabit(w http.ResponseWriter, r *http.Request) (habitBody, bool) {
	var b habitBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		badRequest(w, err.Error())
		return b, false
	}
	if msg := b.validate(); msg != "" {
		badRequest(w, msg)
		return b, false
	}
	b.Name = strings.TrimSpace(b.Name)
	if b.Schedule.Kind != core.Days {
		b.Schedule.Days = nil
	}
	if b.Schedule.Kind != core.Weekly {
		b.Schedule.Times = 0
	}
	return b, true
}

func (s *srv) createHabit(w http.ResponseWriter, r *http.Request) {
	b, ok := decodeHabit(w, r)
	if !ok {
		return
	}
	cal, err := s.calendar()
	if err != nil {
		fail(w, err)
		return
	}
	day := cal.DayOf(s.Now())
	id, err := db.SaveHabit(s.DB, core.Habit{
		Name: b.Name, Effort: b.Effort, CreatedDay: day,
		Versions: []core.Version{{From: day, Schedule: b.Schedule}},
	})
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (s *srv) findHabit(w http.ResponseWriter, r *http.Request) (core.Habit, []core.Habit, bool) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	habits, err := db.Habits(s.DB)
	if err != nil {
		fail(w, err)
		return core.Habit{}, nil, false
	}
	for _, h := range habits {
		if h.ID == id {
			return h, habits, true
		}
	}
	http.Error(w, "hábito não encontrado", http.StatusNotFound)
	return core.Habit{}, nil, false
}

func (s *srv) updateHabit(w http.ResponseWriter, r *http.Request) {
	h, _, ok := s.findHabit(w, r)
	if !ok {
		return
	}
	b, ok := decodeHabit(w, r)
	if !ok {
		return
	}
	cal, err := s.calendar()
	if err != nil {
		fail(w, err)
		return
	}
	h.Name, h.Effort = b.Name, b.Effort
	h.Versions = core.Revise(h.Versions, b.Schedule, b.Archived, cal.DayOf(s.Now()))
	if _, err := db.SaveHabit(s.DB, h); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *srv) deleteHabit(w http.ResponseWriter, r *http.Request) {
	h, _, ok := s.findHabit(w, r)
	if !ok {
		return
	}
	if err := db.DeleteHabit(s.DB, h.ID); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *srv) listHabits(w http.ResponseWriter, r *http.Request) {
	cal, err := s.calendar()
	if err != nil {
		fail(w, err)
		return
	}
	habits, err := db.Habits(s.DB)
	if err != nil {
		fail(w, err)
		return
	}
	log, err := db.CheckLog(s.DB)
	if err != nil {
		fail(w, err)
		return
	}
	checks := db.Checks(log)
	today := cal.DayOf(s.Now())
	type row struct {
		core.Habit
		Current core.Version `json:"current"`
		Streak  int          `json:"streak"`
	}
	out := []row{}
	for _, h := range habits {
		out = append(out, row{h, h.Current(), core.HabitStats(h, checks, today, 1).Streak})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *srv) habitDetail(w http.ResponseWriter, r *http.Request) {
	h, _, ok := s.findHabit(w, r)
	if !ok {
		return
	}
	cal, err := s.calendar()
	if err != nil {
		fail(w, err)
		return
	}
	log, err := db.CheckLog(s.DB)
	if err != nil {
		fail(w, err)
		return
	}
	today := cal.DayOf(s.Now())
	writeJSON(w, http.StatusOK, map[string]any{
		"habit": h, "current": h.Current(), "today": today,
		"stats": core.HabitStats(h, db.Checks(log), today, 12),
	})
}

func (s *srv) putCheck(w http.ResponseWriter, r *http.Request) {
	var b struct {
		HabitID int      `json:"habitId"`
		Day     core.Day `json:"day"`
		Done    bool     `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		badRequest(w, err.Error())
		return
	}
	if _, err := time.Parse("2006-01-02", b.Day); err != nil {
		badRequest(w, "dia inválido")
		return
	}
	cal, err := s.calendar()
	if err != nil {
		fail(w, err)
		return
	}
	if b.Day > cal.DayOf(s.Now()) {
		badRequest(w, "dia no futuro")
		return
	}
	habits, err := db.Habits(s.DB)
	if err != nil {
		fail(w, err)
		return
	}
	found := false
	for _, h := range habits {
		found = found || h.ID == b.HabitID
	}
	if !found {
		http.Error(w, "hábito não encontrado", http.StatusNotFound)
		return
	}
	if err := db.AddCheck(s.DB, db.LogEntry{HabitID: b.HabitID, Day: b.Day, At: s.Now(), Done: b.Done}); err != nil {
		fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
