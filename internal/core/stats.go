package core

import "math"

type DayState string

const (
	Done   DayState = "done"
	Missed DayState = "missed"
	Open   DayState = "open"   // hoje, ainda não feito
	Off    DayState = "off"    // fora da agenda
	Before DayState = "before" // antes do começo do hábito
	Future DayState = "future"
)

type GridDay struct {
	Day   Day      `json:"day"`
	State DayState `json:"state"`
}

type Stats struct {
	Streak int       `json:"streak"`
	Best   int       `json:"best"`
	Total  int       `json:"total"`
	Rate   int       `json:"rate"` // % dos dias (ou semanas, no semanal) cumpridos; -1 sem nada a cobrar
	Grid   []GridDay `json:"grid"` // semanas inteiras, de segunda a domingo, terminando na de today
}

// HabitStats usa a frequência atual do hábito. O começo é a criação ou o
// primeiro dia marcado antes dela (um hábito que já se fazia antes).
func HabitStats(h Habit, all []Check, today Day, weeks int) Stats {
	s := h.Current().Schedule
	done := map[Day]bool{}
	start := h.CreatedDay
	for _, c := range all {
		if c.HabitID != h.ID {
			continue
		}
		done[c.Day] = true
		if c.Day < start {
			start = c.Day
		}
	}
	st := Stats{Total: len(done), Rate: -1}

	first := AddDays(WeekOf(today), -7*(weeks-1))
	for i := 0; i < weeks*7; i++ {
		d := AddDays(first, i)
		st.Grid = append(st.Grid, GridDay{d, dayState(s, start, done, d, today)})
	}

	hits, due := 0, 0
	if s.Kind == Weekly {
		perWeek := map[Day]int{}
		for d := range done {
			perWeek[WeekOf(d)]++
		}
		run, cur := 0, WeekOf(today)
		for w := WeekOf(start); w <= cur; w = AddDays(w, 7) {
			met := perWeek[w] >= s.Times
			if met {
				run++
				st.Best = max(st.Best, run)
			} else if w != cur {
				run = 0
			}
			if w >= first && (w != cur || met) {
				due++
				if met {
					hits++
				}
			}
		}
		st.Streak = run
	} else {
		run := 0
		for d := start; d <= today; d = AddDays(d, 1) {
			if !IsScheduled(s, d) {
				continue
			}
			if done[d] {
				run++
				st.Best = max(st.Best, run)
			} else if d != today {
				run = 0
			}
		}
		st.Streak = run
		for _, g := range st.Grid {
			if g.State == Missed || (g.State == Done && IsScheduled(s, g.Day)) {
				due++
				if g.State == Done {
					hits++
				}
			}
		}
	}
	if due > 0 {
		st.Rate = int(math.Round(100 * float64(hits) / float64(due)))
	}
	return st
}

func dayState(s Schedule, start Day, done map[Day]bool, d, today Day) DayState {
	switch {
	case d > today:
		return Future
	case d < start:
		return Before
	case done[d]:
		return Done
	case d == today:
		if IsScheduled(s, d) {
			return Open
		}
		return Off
	case IsScheduled(s, d) && s.Kind != Weekly:
		return Missed
	}
	return Off
}
