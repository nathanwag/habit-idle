package core

import (
	"slices"
	"sort"
)

type Kind string

const (
	Daily  Kind = "daily"
	Days   Kind = "days"   // dias fixos da semana (0 = domingo)
	Weekly Kind = "weekly" // X vezes por semana, em qualquer dia
)

type Schedule struct {
	Kind  Kind  `json:"kind"`
	Days  []int `json:"days,omitempty"`
	Times int   `json:"times,omitempty"`
}

func (s Schedule) Equal(o Schedule) bool {
	return s.Kind == o.Kind && s.Times == o.Times && slices.Equal(s.Days, o.Days)
}

// Version: frequência e arquivamento a partir de From. O passado segue a
// versão de cada dia, porque o jogo recalcula os dias antigos.
type Version struct {
	From     Day      `json:"from"`
	Schedule Schedule `json:"schedule"`
	Archived bool     `json:"archived"`
}

type Habit struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Effort     int       `json:"effort"` // minutos de fôlego por check
	CreatedDay Day       `json:"createdDay"`
	Order      int       `json:"order"`
	Versions   []Version `json:"versions"`
}

// Check: o hábito está marcado naquele dia.
type Check struct {
	HabitID int `json:"habitId"`
	Day     Day `json:"day"`
}

// AsOf é a versão que vale em day, ou false se o hábito ainda não existia ou
// estava arquivado.
func (h Habit) AsOf(day Day) (Schedule, bool) {
	var cur *Version
	for i := range h.Versions {
		if h.Versions[i].From <= day {
			cur = &h.Versions[i]
		}
	}
	if cur == nil || cur.Archived {
		return Schedule{}, false
	}
	return cur.Schedule, true
}

// Current é a frequência mais recente (para as telas).
func (h Habit) Current() Version {
	return h.Versions[len(h.Versions)-1]
}

// Revise registra uma versão a partir de day se a frequência ou o arquivado
// mudaram. Duas mudanças no mesmo dia: vale a última.
func Revise(versions []Version, s Schedule, archived bool, day Day) []Version {
	if len(versions) > 0 {
		last := versions[len(versions)-1]
		if last.Archived == archived && last.Schedule.Equal(s) {
			return versions
		}
	}
	out := []Version{}
	for _, v := range versions {
		if v.From != day {
			out = append(out, v)
		}
	}
	return append(out, Version{From: day, Schedule: s, Archived: archived})
}

func IsScheduled(s Schedule, day Day) bool {
	return s.Kind != Days || slices.Contains(s.Days, Weekday(day))
}

type TodayItem struct {
	Habit        Habit    `json:"habit"`
	Schedule     Schedule `json:"schedule"`
	Done         bool     `json:"done"`
	DoneThisWeek int      `json:"doneThisWeek"`
}

// Mission: conta para o dia completo. O semanal é cobrado na semana.
func (i TodayItem) Mission() bool { return i.Schedule.Kind != Weekly }

// TodayList: hábitos ativos que valem em day, na ordem escolhida.
func TodayList(habits []Habit, checks []Check, day Day) []TodayItem {
	monday := WeekOf(day)
	sunday := AddDays(monday, 6)
	out := []TodayItem{}
	for _, h := range habits {
		s, ok := h.AsOf(day)
		if !ok || !IsScheduled(s, day) {
			continue
		}
		it := TodayItem{Habit: h, Schedule: s}
		for _, c := range checks {
			if c.HabitID != h.ID || c.Day < monday || c.Day > sunday {
				continue
			}
			it.DoneThisWeek++
			if c.Day == day {
				it.Done = true
			}
		}
		out = append(out, it)
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].Habit.Order < out[b].Habit.Order })
	return out
}

// Miss: quantas vezes faltaram de um semanal numa semana.
type Miss struct {
	Habit   Habit
	Missing int
}

// WeekMisses da semana que começa em monday. Só conta quem foi semanal a
// semana inteira (criado ou mudado no meio dela não teria como cumprir).
// Fazer a mais num não compensa outro.
func WeekMisses(habits []Habit, checks []Check, monday Day) []Miss {
	sunday := AddDays(monday, 6)
	var out []Miss
	for _, h := range habits {
		s, ok := h.AsOf(sunday)
		whole := ok && s.Kind == Weekly
		for d := monday; whole && d < sunday; d = AddDays(d, 1) {
			ds, ok := h.AsOf(d)
			whole = ok && ds.Kind == Weekly
		}
		if !whole {
			continue
		}
		done := 0
		for _, c := range checks {
			if c.HabitID == h.ID && c.Day >= monday && c.Day <= sunday {
				done++
			}
		}
		if s.Times > done {
			out = append(out, Miss{Habit: h, Missing: s.Times - done})
		}
	}
	return out
}
