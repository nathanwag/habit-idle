package core

import "testing"

func TestStatsDiarioPulaDiasForaDaAgendaEHojeAbertoNaoQuebra(t *testing.T) {
	// Seg, qua e sex. 2026-11-16 é segunda.
	h := habit(1, "2026-11-02", Schedule{Kind: Days, Days: []int{1, 3, 5}})
	checks := []Check{
		{1, "2026-11-02"}, {1, "2026-11-04"}, {1, "2026-11-06"}, // 3 seguidos
		// 09 (seg) faltou
		{1, "2026-11-11"}, {1, "2026-11-13"}, {1, "2026-11-16"},
	}
	s := HabitStats(h, checks, "2026-11-18", 12) // quarta, ainda não feito
	if s.Streak != 3 || s.Best != 3 || s.Total != 6 {
		t.Fatalf("got %+v", s)
	}
}

func TestStatsSemanalContaSemanasNaMeta(t *testing.T) {
	h := habit(1, "2026-11-02", Schedule{Kind: Weekly, Times: 2})
	checks := []Check{
		{1, "2026-11-02"}, {1, "2026-11-05"}, // semana 02: ok
		{1, "2026-11-09"}, {1, "2026-11-10"}, // semana 09: ok
		{1, "2026-11-16"}, // semana 16, atual: em andamento
	}
	s := HabitStats(h, checks, "2026-11-18", 12)
	if s.Streak != 2 || s.Best != 2 {
		t.Fatalf("got %+v", s)
	}
}

func TestStatsGradeTemAsSemanasPedidasEATaxa(t *testing.T) {
	h := habit(1, "2026-11-16", daily)
	checks := []Check{{1, "2026-11-16"}, {1, "2026-11-18"}}
	s := HabitStats(h, checks, "2026-11-18", 12)
	if len(s.Grid) != 12*7 {
		t.Fatalf("grade com %d dias", len(s.Grid))
	}
	last := s.Grid[len(s.Grid)-7:] // semana de 16 a 22
	want := []DayState{Done, Missed, Done, Future, Future, Future, Future}
	for i, d := range last {
		if d.State != want[i] {
			t.Errorf("%s = %s, quer %s", d.Day, d.State, want[i])
		}
	}
	if s.Grid[0].State != Before {
		t.Errorf("primeiro dia = %s", s.Grid[0].State)
	}
	// 2 feitos de 3 dias que já contam (16, 17 e hoje feito).
	if s.Rate != 67 {
		t.Errorf("taxa = %d", s.Rate)
	}
}
