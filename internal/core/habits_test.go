package core

import (
	"reflect"
	"testing"
)

var daily = Schedule{Kind: Daily}

func habit(id int, from Day, s Schedule) Habit {
	return Habit{ID: id, Name: "h", Effort: 120, CreatedDay: from, Order: id, Versions: []Version{{From: from, Schedule: s}}}
}

func ids(items []TodayItem) []int {
	out := []int{}
	for _, it := range items {
		out = append(out, it.Habit.ID)
	}
	return out
}

func TestTodayListRespeitaDiasFixosEOrdem(t *testing.T) {
	seg := habit(1, "2026-11-01", Schedule{Kind: Days, Days: []int{1}})
	tudo := habit(2, "2026-11-01", daily)
	tudo.Order = 0
	// 2026-11-16 é segunda; 2026-11-17 é terça.
	if got := ids(TodayList([]Habit{seg, tudo}, nil, "2026-11-16")); !reflect.DeepEqual(got, []int{2, 1}) {
		t.Errorf("segunda: %v", got)
	}
	if got := ids(TodayList([]Habit{seg, tudo}, nil, "2026-11-17")); !reflect.DeepEqual(got, []int{2}) {
		t.Errorf("terça: %v", got)
	}
}

func TestTodayListUsaAVersaoDoDia(t *testing.T) {
	h := habit(1, "2026-11-01", daily)
	h.Versions = Revise(h.Versions, Schedule{Kind: Days, Days: []int{3}}, false, "2026-11-10")
	h.Versions = Revise(h.Versions, daily, true, "2026-11-20")
	cases := map[Day][]int{
		"2026-10-31": {},  // antes de existir
		"2026-11-09": {1}, // ainda diário
		"2026-11-16": {},  // segunda, só quarta
		"2026-11-18": {1}, // quarta
		"2026-11-20": {},  // arquivado
	}
	for day, want := range cases {
		if got := ids(TodayList([]Habit{h}, nil, day)); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: %v, quer %v", day, got, want)
		}
	}
}

func TestReviseSemMudancaNaoCriaVersaoEMesmoDiaSubstitui(t *testing.T) {
	v := []Version{{From: "2026-11-01", Schedule: daily}}
	if got := Revise(v, daily, false, "2026-11-05"); len(got) != 1 {
		t.Fatalf("sem mudança criou versão: %v", got)
	}
	v = Revise(v, Schedule{Kind: Weekly, Times: 2}, false, "2026-11-05")
	v = Revise(v, Schedule{Kind: Weekly, Times: 3}, false, "2026-11-05")
	if len(v) != 2 || v[1].Schedule.Times != 3 {
		t.Fatalf("got %v", v)
	}
}

func TestTodayListMarcaFeitoEContaASemana(t *testing.T) {
	h := habit(1, "2026-11-01", Schedule{Kind: Weekly, Times: 3})
	checks := []Check{{1, "2026-11-15"}, {1, "2026-11-16"}, {1, "2026-11-18"}}
	got := TodayList([]Habit{h}, checks, "2026-11-18")[0]
	if !got.Done || got.DoneThisWeek != 2 {
		t.Fatalf("got %+v", got)
	}
}
