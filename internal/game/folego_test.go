package game

import (
	"testing"
	"time"

	"habit-idle/internal/core"
)

var sp, _ = time.LoadLocation("America/Sao_Paulo")

var cal = core.Calendar{Loc: sp, DayStart: 6 * 60}

func at(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04", s, sp)
	if err != nil {
		panic(err)
	}
	return t
}

func daily(id, effort int, from core.Day) core.Habit {
	return core.Habit{ID: id, Name: "h", Effort: effort, CreatedDay: from, Order: id,
		Versions: []core.Version{{From: from, Schedule: core.Schedule{Kind: core.Daily}}}}
}

func touch(id int, day core.Day, when string, done bool) Touch {
	return Touch{HabitID: id, Day: day, At: at(when), Done: done}
}

func started(s string) *time.Time { t := at(s); return &t }

func mins(n int) time.Duration { return time.Duration(n) * time.Minute }

func TestCheckDaFolegoQueEGastoExplorando(t *testing.T) {
	in := FolegoInput{
		Cal:     cal,
		Habits:  []core.Habit{daily(1, 120, "2026-11-01"), daily(2, 60, "2026-11-01")},
		Touches: []Touch{touch(1, "2026-11-19", "2026-11-19 09:00", true)},
		Start:   started("2026-11-18 12:00"),
		Now:     at("2026-11-19 10:00"),
	}
	f := Folego(in)
	if f.Balance != mins(60) {
		t.Fatalf("saldo = %v, quer 1h", f.Balance)
	}
	if len(f.Segments) != 1 || !f.Segments[0].From.Equal(at("2026-11-19 09:00")) || !f.Segments[0].To.Equal(at("2026-11-19 10:00")) {
		t.Fatalf("segmentos = %+v", f.Segments)
	}
	if !f.Until.Equal(at("2026-11-19 11:00")) {
		t.Fatalf("até = %v", f.Until)
	}
}

func TestMarcarDiaPassadoNaoDaFolego(t *testing.T) {
	f := Folego(FolegoInput{
		Cal:     cal,
		Habits:  []core.Habit{daily(1, 120, "2026-11-01")},
		Touches: []Touch{touch(1, "2026-11-18", "2026-11-19 09:00", true)},
		Start:   started("2026-11-18 12:00"),
		Now:     at("2026-11-19 10:00"),
	})
	if f.Balance != 0 || len(f.Segments) != 0 {
		t.Fatalf("got %+v", f)
	}
}

func TestSemHeroiOFolegoNaoEGasto(t *testing.T) {
	f := Folego(FolegoInput{
		Cal:     cal,
		Habits:  []core.Habit{daily(1, 120, "2026-11-01"), daily(2, 60, "2026-11-01")},
		Touches: []Touch{touch(1, "2026-11-19", "2026-11-19 09:00", true)},
		Now:     at("2026-11-19 10:00"),
	})
	if f.Balance != mins(120) || len(f.Segments) != 0 {
		t.Fatalf("got %+v", f)
	}
}

func TestDesmarcarDepoisDeGastoViraDividaQueOProximoCheckPaga(t *testing.T) {
	in := FolegoInput{
		Cal:    cal,
		Habits: []core.Habit{daily(1, 120, "2026-11-01"), daily(2, 60, "2026-11-01"), daily(3, 180, "2026-11-01")},
		Touches: []Touch{
			touch(1, "2026-11-19", "2026-11-19 09:00", true),
			touch(1, "2026-11-19", "2026-11-19 12:00", false), // as 2h já foram gastas
			touch(2, "2026-11-19", "2026-11-19 13:00", true),
		},
		Start: started("2026-11-18 12:00"),
		Now:   at("2026-11-19 14:00"),
	}
	f := Folego(in)
	if f.Balance != -mins(60) {
		t.Fatalf("saldo = %v, quer -1h", f.Balance)
	}
	if len(f.Segments) != 1 || !f.Segments[0].To.Equal(at("2026-11-19 11:00")) {
		t.Fatalf("não devia explorar com dívida: %+v", f.Segments)
	}
	if !f.Until.IsZero() {
		t.Fatalf("até = %v", f.Until)
	}
}

func TestSaldoNuncaPassaDe16h(t *testing.T) {
	var hs []core.Habit
	var ts []Touch
	for i := 1; i <= 7; i++ {
		hs = append(hs, daily(i, 180, "2026-11-01"))
		ts = append(ts, touch(i, "2026-11-19", "2026-11-19 09:00", true))
	}
	f := Folego(FolegoInput{Cal: cal, Habits: hs, Touches: ts, Now: at("2026-11-19 09:30")})
	if f.Balance != 16*time.Hour {
		t.Fatalf("saldo = %v", f.Balance)
	}
}

func TestUltimaMissaoDaBonusEDesmarcarTiraOBonus(t *testing.T) {
	hs := []core.Habit{daily(1, 120, "2026-11-01"), daily(2, 60, "2026-11-01")}
	ts := []Touch{
		touch(1, "2026-11-19", "2026-11-19 09:00", true),
		touch(2, "2026-11-19", "2026-11-19 09:00", true),
	}
	f := Folego(FolegoInput{Cal: cal, Habits: hs, Touches: ts, Now: at("2026-11-19 09:30")})
	if f.Balance != mins(240) {
		t.Fatalf("com bônus = %v, quer 4h", f.Balance)
	}
	ts = append(ts, touch(2, "2026-11-19", "2026-11-19 09:10", false))
	f = Folego(FolegoInput{Cal: cal, Habits: hs, Touches: ts, Now: at("2026-11-19 09:30")})
	if f.Balance != mins(120) {
		t.Fatalf("sem bônus = %v, quer 2h", f.Balance)
	}
}

func TestSemanalNaoContaParaODiaCompleto(t *testing.T) {
	semanal := core.Habit{ID: 2, Effort: 60, CreatedDay: "2026-11-01", Versions: []core.Version{{From: "2026-11-01", Schedule: core.Schedule{Kind: core.Weekly, Times: 2}}}}
	hs := []core.Habit{daily(1, 120, "2026-11-01"), semanal}
	ts := []Touch{touch(1, "2026-11-19", "2026-11-19 09:00", true)}
	f := Folego(FolegoInput{Cal: cal, Habits: hs, Touches: ts, Now: at("2026-11-19 09:30")})
	if f.Balance != mins(180) {
		t.Fatalf("saldo = %v, quer 3h (2h + bônus)", f.Balance)
	}
}

func TestViradaCortaOQueSobrouPelaMetade(t *testing.T) {
	hs := []core.Habit{daily(1, 180, "2026-11-01"), daily(2, 60, "2026-11-01")}
	ts := []Touch{touch(1, "2026-11-19", "2026-11-19 09:00", true)}
	// Sem herói: as 3h sobram até a virada das 06:00 e viram 1h30.
	f := Folego(FolegoInput{Cal: cal, Habits: hs, Touches: ts, Now: at("2026-11-20 07:00")})
	if f.Balance != mins(90) {
		t.Fatalf("saldo = %v, quer 1h30", f.Balance)
	}
}

func TestExplorandoAteConsideraAVirada(t *testing.T) {
	var hs []core.Habit
	var ts []Touch
	for i := 1; i <= 3; i++ {
		hs = append(hs, daily(i, 180, "2026-11-01"))
		ts = append(ts, touch(i, "2026-11-19", "2026-11-19 22:00", true))
	}
	hs = append(hs, daily(9, 60, "2026-11-01")) // fica sem fazer: sem bônus
	f := Folego(FolegoInput{Cal: cal, Habits: hs, Touches: ts, Start: started("2026-11-18 12:00"), Now: at("2026-11-19 22:00")})
	// 9h às 22:00; às 06:00 sobra 1h, que cai para 30 min.
	if !f.Until.Equal(at("2026-11-20 06:30")) {
		t.Fatalf("até = %v", f.Until)
	}
}

func TestSemanalQueFaltouViraDividaNoDomingo(t *testing.T) {
	correr := core.Habit{ID: 1, Effort: 180, CreatedDay: "2026-11-01", Versions: []core.Version{{From: "2026-11-01", Schedule: core.Schedule{Kind: core.Weekly, Times: 2}}}}
	ts := []Touch{touch(1, "2026-11-18", "2026-11-18 09:00", true)} // 1 de 2 na semana de 16 a 22
	f := Folego(FolegoInput{Cal: cal, Habits: []core.Habit{correr}, Touches: ts, Now: at("2026-11-23 07:00")})
	// 3h ganhas; viradas de 19 a 23: 3h → 1h30 → 45 → 22 → 11 → 5 min; na de segunda, menos a falta de 3h.
	if f.Balance != mins(5)-mins(180) {
		t.Fatalf("saldo = %v", f.Balance)
	}
}

func TestCansacoSobeSemNadaFeitoEZeraNoDiaCompleto(t *testing.T) {
	hs := []core.Habit{daily(1, 60, "2026-11-01"), daily(2, 60, "2026-11-01")}
	ts := []Touch{
		touch(1, "2026-11-23", "2026-11-23 09:00", true), // dia 23: só uma
		touch(1, "2026-11-24", "2026-11-24 09:00", true), // dia 24: as duas
		touch(2, "2026-11-24", "2026-11-24 09:00", true),
	}
	// O cansaço só conta depois do início do jogo.
	f := Folego(FolegoInput{Cal: cal, Habits: hs, Touches: ts, Start: started("2026-11-19 08:00"), Now: at("2026-11-25 07:00")})
	// 19, 20, 21, 22 sem nada (+10 até 30); 23 parcial (−10); 24 completo (0).
	cases := map[string]int{
		"2026-11-20 05:59": 0,
		"2026-11-20 06:00": 10,
		"2026-11-22 06:00": 30,
		"2026-11-23 06:00": 30,
		"2026-11-24 06:00": 20,
		"2026-11-25 06:00": 0,
	}
	for when, want := range cases {
		if got := f.FatigueAt(at(when)); got != want {
			t.Errorf("cansaço em %s = %d, quer %d", when, got, want)
		}
	}
	if len(f.Complete) != 1 || f.Complete[0] != "2026-11-24" {
		t.Errorf("dias completos = %v", f.Complete)
	}
}
