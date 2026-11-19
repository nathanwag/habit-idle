package main

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"time"

	"habit-idle/internal/core"
	"habit-idle/internal/db"
)

// demo enche um banco novo com 20 dias de hábitos marcados e um jogo começado,
// para ver as telas com dados. Nunca rode apontando para o banco de verdade.
func demo(d *sql.DB, now time.Time) error {
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	cal := core.Calendar{Loc: loc, DayStart: 6 * 60}
	today := cal.DayOf(now)
	first := core.AddDays(today, -20)
	habits := []struct {
		name   string
		effort int
		s      core.Schedule
	}{
		{"Ler 20 páginas", 120, core.Schedule{Kind: core.Daily}},
		{"Beber 2 L de água", 60, core.Schedule{Kind: core.Daily}},
		{"Estudar inglês", 120, core.Schedule{Kind: core.Daily}},
		{"Meditar", 60, core.Schedule{Kind: core.Daily}},
		{"Flexões 3×15", 120, core.Schedule{Kind: core.Days, Days: []int{1, 3, 5}}},
		{"Correr", 180, core.Schedule{Kind: core.Weekly, Times: 2}},
	}
	var hs []core.Habit
	for _, h := range habits {
		id, err := db.SaveHabit(d, core.Habit{Name: h.name, Effort: h.effort, CreatedDay: first, Versions: []core.Version{{From: first, Schedule: h.s}}})
		if err != nil {
			return err
		}
		hs = append(hs, core.Habit{ID: id, Effort: h.effort, Versions: []core.Version{{From: first, Schedule: h.s}}})
	}
	if err := db.AddEvent(d, db.Event{At: cal.Start(first).Add(time.Hour), Kind: "start", Data: `{"hero":"kael"}`}); err != nil {
		return err
	}
	r := rand.New(rand.NewPCG(7, 7))
	for day := first; day <= today; day = core.AddDays(day, 1) {
		for i, h := range hs {
			s, _ := h.AsOf(day)
			if !core.IsScheduled(s, day) || (s.Kind == core.Weekly && r.Float64() < 0.7) || r.Float64() < 0.2 {
				continue
			}
			at := cal.Start(day).Add(time.Duration(2+i*2+r.IntN(3)) * time.Hour)
			if at.After(now) {
				continue
			}
			if err := db.AddCheck(d, db.LogEntry{HabitID: h.ID, Day: day, At: at, Done: true}); err != nil {
				return err
			}
		}
	}
	if err := db.SetSetting(d, "lastSeen", now.Add(-10*time.Hour).UTC().Format(time.RFC3339)); err != nil {
		return err
	}
	fmt.Printf("demo: %d hábitos, jogo desde %s\n", len(hs), first)
	return nil
}
