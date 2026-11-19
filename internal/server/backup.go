package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"habit-idle/internal/core"
	"habit-idle/internal/db"
)

const backupApp = "habit-idle"

type backupCheck struct {
	HabitID int       `json:"habitId"`
	Day     core.Day  `json:"day"`
	At      time.Time `json:"at"`
	Done    bool      `json:"done"`
}

type backupEvent struct {
	At   time.Time       `json:"at"`
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

type backup struct {
	App      string            `json:"app"`
	Version  int               `json:"version"`
	Settings map[string]string `json:"settings"`
	Habits   []core.Habit      `json:"habits"`
	Checks   []backupCheck     `json:"checks"`
	Events   []backupEvent     `json:"events"`
}

func (s *srv) exportBackup(w http.ResponseWriter, r *http.Request) {
	b := backup{App: backupApp, Version: 1, Checks: []backupCheck{}, Events: []backupEvent{}}
	var err error
	if b.Settings, err = db.Settings(s.DB); err != nil {
		fail(w, err)
		return
	}
	delete(b.Settings, "lastSeen")
	if b.Habits, err = db.Habits(s.DB); err != nil {
		fail(w, err)
		return
	}
	log, err := db.CheckLog(s.DB)
	if err != nil {
		fail(w, err)
		return
	}
	for _, e := range log {
		b.Checks = append(b.Checks, backupCheck{e.HabitID, e.Day, e.At, e.Done})
	}
	evs, err := db.Events(s.DB)
	if err != nil {
		fail(w, err)
		return
	}
	for _, e := range evs {
		b.Events = append(b.Events, backupEvent{e.At, e.Kind, json.RawMessage(e.Data)})
	}
	w.Header().Set("Content-Disposition", `attachment; filename="habit-idle-`+s.Now().Format("2006-01-02")+`.json"`)
	writeJSON(w, http.StatusOK, b)
}

// importBackup troca tudo pelo conteúdo do arquivo, numa transação só.
func (s *srv) importBackup(w http.ResponseWriter, r *http.Request) {
	var b backup
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil || b.App != backupApp || b.Version != 1 {
		badRequest(w, "arquivo de backup inválido")
		return
	}
	err := db.InTx(s.DB, func(tx *sql.Tx) error {
		for _, t := range []string{"check_log", "habit_version", "habit", "event", "setting"} {
			if _, err := tx.Exec("DELETE FROM " + t); err != nil {
				return err
			}
		}
		for k, v := range b.Settings {
			if err := db.SetSetting(tx, k, v); err != nil {
				return err
			}
		}
		for _, h := range b.Habits {
			if _, err := tx.Exec("INSERT INTO habit(id, name, effort, created_day, ord) VALUES(?, ?, ?, ?, ?)", h.ID, h.Name, h.Effort, h.CreatedDay, h.Order); err != nil {
				return err
			}
			if _, err := db.SaveHabit(tx, h); err != nil {
				return err
			}
		}
		for _, c := range b.Checks {
			if err := db.AddCheck(tx, db.LogEntry{HabitID: c.HabitID, Day: c.Day, At: c.At, Done: c.Done}); err != nil {
				return err
			}
		}
		for _, e := range b.Events {
			if err := db.AddEvent(tx, db.Event{At: e.At, Kind: e.Kind, Data: string(e.Data)}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		badRequest(w, "não deu para importar: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
