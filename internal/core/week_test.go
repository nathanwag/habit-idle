package core

import (
	"reflect"
	"testing"
)

func TestWeekMissesSoContaQuemFoiSemanalASemanaInteira(t *testing.T) {
	correr := habit(1, "2026-11-01", Schedule{Kind: Weekly, Times: 3})
	correr.Effort = 180
	novo := habit(2, "2026-11-18", Schedule{Kind: Weekly, Times: 2}) // criado no meio da semana
	ler := habit(3, "2026-11-01", daily)
	checks := []Check{{1, "2026-11-16"}, {1, "2026-11-22"}, {1, "2026-11-23"}}
	got := WeekMisses([]Habit{correr, novo, ler}, checks, "2026-11-16")
	want := []Miss{{Habit: correr, Missing: 1}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
}

func TestWeekMissesFazerAMaisNaoGeraFalta(t *testing.T) {
	h := habit(1, "2026-11-01", Schedule{Kind: Weekly, Times: 1})
	checks := []Check{{1, "2026-11-16"}, {1, "2026-11-17"}}
	if got := WeekMisses([]Habit{h}, checks, "2026-11-16"); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}
