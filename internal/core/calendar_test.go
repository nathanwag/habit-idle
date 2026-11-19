package core

import (
	"testing"
	"time"
)

var sp, _ = time.LoadLocation("America/Sao_Paulo")

func at(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04", s, sp)
	if err != nil {
		panic(err)
	}
	return t
}

func TestDayOfViraNoHorarioDoAjuste(t *testing.T) {
	cal := Calendar{Loc: sp, DayStart: 6 * 60}
	cases := map[string]string{
		"2026-11-19 05:59": "2026-11-18",
		"2026-11-19 06:00": "2026-11-19",
		"2026-11-19 23:30": "2026-11-19",
		"2026-11-20 01:00": "2026-11-19",
	}
	for in, want := range cases {
		if got := cal.DayOf(at(in)); got != want {
			t.Errorf("DayOf(%s) = %s, quer %s", in, got, want)
		}
	}
}

func TestDayOfUsaOFusoEnaoOUTC(t *testing.T) {
	cal := Calendar{Loc: sp, DayStart: 0}
	// 01:00 UTC do dia 20 ainda é 22:00 do dia 19 em São Paulo.
	if got := cal.DayOf(time.Date(2026, 11, 20, 1, 0, 0, 0, time.UTC)); got != "2026-11-19" {
		t.Fatalf("got %s", got)
	}
}

func TestWeekOfComecaNaSegunda(t *testing.T) {
	cases := map[string]string{
		"2026-11-16": "2026-11-16", // segunda
		"2026-11-19": "2026-11-16", // quinta
		"2026-11-22": "2026-11-16", // domingo
		"2026-11-01": "2026-10-26", // domingo, virada de mês
	}
	for in, want := range cases {
		if got := WeekOf(in); got != want {
			t.Errorf("WeekOf(%s) = %s, quer %s", in, got, want)
		}
	}
}

func TestAddDaysAtravessaMesEAno(t *testing.T) {
	if got := AddDays("2026-12-31", 1); got != "2027-01-01" {
		t.Fatalf("got %s", got)
	}
	if got := AddDays("2026-03-01", -1); got != "2026-02-28" {
		t.Fatalf("got %s", got)
	}
}
