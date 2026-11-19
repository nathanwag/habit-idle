package db

import (
	"database/sql"
	"encoding/json"
	"time"

	"habit-idle/internal/core"
)

// Querier é o que *sql.DB e *sql.Tx têm em comum.
type Querier interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

func Settings(q Querier) (map[string]string, error) {
	rows, err := q.Query("SELECT key, value FROM setting")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

func SetSetting(q Querier, key, value string) error {
	_, err := q.Exec("INSERT INTO setting(key, value) VALUES(?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", key, value)
	return err
}

func Habits(q Querier) ([]core.Habit, error) {
	rows, err := q.Query("SELECT id, name, effort, created_day, ord FROM habit ORDER BY ord, id")
	if err != nil {
		return nil, err
	}
	var hs []core.Habit
	byID := map[int]int{}
	for rows.Next() {
		var h core.Habit
		if err := rows.Scan(&h.ID, &h.Name, &h.Effort, &h.CreatedDay, &h.Order); err != nil {
			rows.Close()
			return nil, err
		}
		byID[h.ID] = len(hs)
		hs = append(hs, h)
	}
	rows.Close()
	vr, err := q.Query("SELECT habit_id, from_day, schedule, archived FROM habit_version ORDER BY habit_id, from_day")
	if err != nil {
		return nil, err
	}
	defer vr.Close()
	for vr.Next() {
		var id int
		var v core.Version
		var sched string
		if err := vr.Scan(&id, &v.From, &sched, &v.Archived); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(sched), &v.Schedule); err != nil {
			return nil, err
		}
		if i, ok := byID[id]; ok {
			hs[i].Versions = append(hs[i].Versions, v)
		}
	}
	return hs, vr.Err()
}

// SaveHabit cria (ID 0) ou atualiza o hábito, regravando as versões.
func SaveHabit(q Querier, h core.Habit) (int, error) {
	if h.ID == 0 {
		res, err := q.Exec("INSERT INTO habit(name, effort, created_day, ord) VALUES(?, ?, ?, (SELECT COALESCE(MAX(ord), 0) + 1 FROM habit))", h.Name, h.Effort, h.CreatedDay)
		if err != nil {
			return 0, err
		}
		id, _ := res.LastInsertId()
		h.ID = int(id)
	} else if _, err := q.Exec("UPDATE habit SET name = ?, effort = ? WHERE id = ?", h.Name, h.Effort, h.ID); err != nil {
		return 0, err
	}
	if _, err := q.Exec("DELETE FROM habit_version WHERE habit_id = ?", h.ID); err != nil {
		return 0, err
	}
	for _, v := range h.Versions {
		s, _ := json.Marshal(v.Schedule)
		if _, err := q.Exec("INSERT INTO habit_version(habit_id, from_day, schedule, archived) VALUES(?, ?, ?, ?)", h.ID, v.From, string(s), v.Archived); err != nil {
			return 0, err
		}
	}
	return h.ID, nil
}

func DeleteHabit(q Querier, id int) error {
	_, err := q.Exec("DELETE FROM habit WHERE id = ?", id)
	return err
}

// LogEntry é um toque de marcar (Done) ou desmarcar, na ordem em que aconteceu.
type LogEntry struct {
	HabitID int
	Day     core.Day
	At      time.Time
	Done    bool
}

func AddCheck(q Querier, e LogEntry) error {
	_, err := q.Exec("INSERT INTO check_log(habit_id, day, at, done) VALUES(?, ?, ?, ?)", e.HabitID, e.Day, e.At.UTC().Format(time.RFC3339Nano), e.Done)
	return err
}

func CheckLog(q Querier) ([]LogEntry, error) {
	rows, err := q.Query("SELECT habit_id, day, at, done FROM check_log ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LogEntry
	for rows.Next() {
		var e LogEntry
		var at string
		if err := rows.Scan(&e.HabitID, &e.Day, &at, &e.Done); err != nil {
			return nil, err
		}
		e.At, _ = time.Parse(time.RFC3339Nano, at)
		out = append(out, e)
	}
	return out, rows.Err()
}

// Checks: o estado atual de cada hábito e dia (a última linha do registro).
func Checks(log []LogEntry) []core.Check {
	type key struct {
		id  int
		day core.Day
	}
	last := map[key]bool{}
	var order []key
	for _, e := range log {
		k := key{e.HabitID, e.Day}
		if _, seen := last[k]; !seen {
			order = append(order, k)
		}
		last[k] = e.Done
	}
	out := []core.Check{}
	for _, k := range order {
		if last[k] {
			out = append(out, core.Check{HabitID: k.id, Day: k.day})
		}
	}
	return out
}
