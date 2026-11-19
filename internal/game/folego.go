package game

import (
	"sort"
	"time"

	"habit-idle/internal/core"
)

// Touch: um marcar (Done) ou desmarcar, na hora em que aconteceu.
type Touch struct {
	HabitID int
	Day     core.Day
	At      time.Time
	Done    bool
}

type FolegoInput struct {
	Cal     core.Calendar
	Habits  []core.Habit
	Touches []Touch    // em ordem de acontecimento
	Start   *time.Time // primeiro herói; sem ele o saldo não é gasto
	Now     time.Time
}

// Segment: intervalo [From, To) em que o esquadrão explorou.
type Segment struct {
	From, To time.Time
}

type FatigueChange struct {
	At  time.Time
	Pct int
}

type FolegoResult struct {
	Balance  time.Duration // negativo = dívida
	Segments []Segment
	Until    time.Time // até quando explora sem novos checks (zero se não explora)
	Fatigue  []FatigueChange
	Complete []core.Day // dias que fecharam com todas as missões
}

// FatigueAt é o cansaço (%) vigente no instante t.
func (f FolegoResult) FatigueAt(t time.Time) int {
	pct := 0
	for _, c := range f.Fatigue {
		if c.At.After(t) {
			break
		}
		pct = c.Pct
	}
	return pct
}

type key struct {
	id  int
	day core.Day
}

type sim struct {
	in        FolegoInput
	balance   time.Duration
	clock     time.Time
	state     map[key]bool          // check atual de cada hábito e dia
	granted   map[key]time.Duration // fôlego que o check atual deu
	bonus     map[core.Day]bool     // dia completo já pagou o bônus
	fatigue   int
	result    FolegoResult
	exploring bool
}

// Folego recalcula o saldo de fôlego do começo até in.Now.
func Folego(in FolegoInput) FolegoResult {
	s := &sim{in: in, state: map[key]bool{}, granted: map[key]time.Duration{}, bonus: map[core.Day]bool{}}
	touches := append([]Touch(nil), in.Touches...)
	sort.SliceStable(touches, func(a, b int) bool { return touches[a].At.Before(touches[b].At) })

	var begin time.Time
	if len(touches) > 0 {
		begin = touches[0].At
	}
	if in.Start != nil && (begin.IsZero() || in.Start.Before(begin)) {
		begin = *in.Start
	}

	if begin.IsZero() || begin.After(in.Now) {
		return s.result
	}
	s.clock = begin
	next := s.nextTurn(begin)
	for _, t := range touches {
		if t.At.After(in.Now) {
			break
		}
		for !next.After(t.At) {
			s.advance(next)
			s.turn(next)
			next = s.nextTurn(next)
		}
		s.advance(t.At)
		s.apply(t)
	}
	for !next.After(in.Now) {
		s.advance(next)
		s.turn(next)
		next = s.nextTurn(next)
	}
	s.advance(in.Now)
	s.result.Balance = s.balance
	s.result.Until = s.project(next)
	return s.result
}

// nextTurn: a próxima virada depois de t.
func (s *sim) nextTurn(t time.Time) time.Time {
	return s.in.Cal.Start(core.AddDays(s.in.Cal.DayOf(t), 1))
}

func (s *sim) started(t time.Time) bool {
	return s.in.Start != nil && !t.Before(*s.in.Start)
}

// advance gasta o saldo de s.clock até t.
func (s *sim) advance(t time.Time) {
	from := s.clock
	s.clock = t
	if s.in.Start == nil || s.balance <= 0 {
		return
	}
	if from.Before(*s.in.Start) {
		from = *s.in.Start
	}
	if !from.Before(t) {
		return
	}
	end := t
	if s.balance < end.Sub(from) {
		end = from.Add(s.balance)
	}
	s.balance -= end.Sub(from)
	s.addSegment(from, end)
}

func (s *sim) addSegment(from, to time.Time) {
	if !to.After(from) {
		return
	}
	segs := s.result.Segments
	if n := len(segs); n > 0 && segs[n-1].To.Equal(from) {
		segs[n-1].To = to
		return
	}
	s.result.Segments = append(segs, Segment{from, to})
}

func (s *sim) gain(d time.Duration) {
	s.balance += d
	if s.balance > FolegoCap {
		s.balance = FolegoCap
	}
}

func (s *sim) checks() []core.Check {
	out := []core.Check{}
	for k, done := range s.state {
		if done {
			out = append(out, core.Check{HabitID: k.id, Day: k.day})
		}
	}
	return out
}

func (s *sim) habit(id int) (core.Habit, bool) {
	for _, h := range s.in.Habits {
		if h.ID == id {
			return h, true
		}
	}
	return core.Habit{}, false
}

func (s *sim) apply(t Touch) {
	k := key{t.HabitID, t.Day}
	if s.state[k] == t.Done {
		return
	}
	s.state[k] = t.Done
	sameDay := s.in.Cal.DayOf(t.At) == t.Day
	if t.Done {
		if h, ok := s.habit(t.HabitID); ok && sameDay {
			s.granted[k] = time.Duration(h.Effort) * time.Minute
			s.gain(s.granted[k])
		}
	} else {
		s.balance -= s.granted[k]
		delete(s.granted, k)
	}
	if sameDay {
		complete := s.complete(t.Day)
		switch {
		case complete && !s.bonus[t.Day]:
			s.bonus[t.Day] = true
			s.gain(DayBonus)
		case !complete && s.bonus[t.Day]:
			s.bonus[t.Day] = false
			s.balance -= DayBonus
		}
	}
}

// complete: o dia vale o bônus (missões suficientes) e todas estão feitas.
// Com uma missão só, completar o dia seria fácil demais.
func (s *sim) complete(day core.Day) bool {
	missions, done := s.missions(day)
	return missions >= BonusMissions && done == missions
}

func (s *sim) missions(day core.Day) (missions, done int) {
	for _, it := range core.TodayList(s.in.Habits, s.checks(), day) {
		if it.Mission() {
			missions++
			if it.Done {
				done++
			}
		}
	}
	return
}

// turn: a virada do dia em t, fechando o dia anterior.
func (s *sim) turn(t time.Time) {
	closed := s.in.Cal.DayOf(t.Add(-time.Minute))
	if s.balance > 0 {
		s.balance = (s.balance / 2).Truncate(time.Minute)
	}
	if core.Weekday(closed) == 0 {
		for _, m := range core.WeekMisses(s.in.Habits, s.checks(), core.WeekOf(closed)) {
			s.balance -= time.Duration(m.Habit.Effort*m.Missing) * time.Minute
		}
	}
	missions, done := s.missions(closed)
	any := false
	for k, v := range s.state {
		any = any || (v && k.day == closed)
	}
	before := s.fatigue
	switch {
	case !s.started(t):
	case missions == 0:
	case !any:
		s.fatigue = min(FatigueMax, s.fatigue+FatigueStep)
	case done == missions:
		s.fatigue = 0
		if missions >= BonusMissions {
			s.result.Complete = append(s.result.Complete, closed)
		}
	default:
		s.fatigue = max(0, s.fatigue-FatigueStep)
	}
	if s.fatigue != before {
		s.result.Fatigue = append(s.result.Fatigue, FatigueChange{At: t, Pct: s.fatigue})
	}
}

// project: até quando o saldo atual dura, com as viradas no caminho.
func (s *sim) project(next time.Time) time.Time {
	if s.balance <= 0 || !s.started(s.in.Now) {
		return time.Time{}
	}
	bal, clock := s.balance, s.in.Now
	for i := 0; i < projectionDays; i++ {
		if bal <= next.Sub(clock) {
			return clock.Add(bal)
		}
		bal -= next.Sub(clock)
		bal = (bal / 2).Truncate(time.Minute)
		clock = next
		next = s.nextTurn(next)
		if bal <= 0 {
			return clock
		}
	}
	return clock.Add(bal)
}
