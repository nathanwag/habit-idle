package game

import (
	"encoding/json"
	"testing"
	"time"

	"habit-idle/internal/core"
)

func ev(when, kind string, data any) Event {
	b, _ := json.Marshal(data)
	return Event{At: at(when), Kind: kind, Data: b}
}

func explore(from string, d time.Duration) []Segment {
	f := at(from)
	return []Segment{{From: f, To: f.Add(d)}}
}

func TestPoderDoHeroiNivelUm(t *testing.T) {
	st := Play(PlayInput{
		Cal:    cal,
		Events: []Event{ev("2026-11-19 08:00", "start", map[string]string{"hero": "oren"})},
		Now:    at("2026-11-19 08:00"),
	})
	// 4 status com 4 pontos e a especialidade com 10: 26 pontos × 10.
	if st.Power != 260 {
		t.Fatalf("poder = %d", st.Power)
	}
}

func TestSeisHorasDeFolegoComOOrenNivelUm(t *testing.T) {
	st := Play(PlayInput{
		Cal:      cal,
		Events:   []Event{ev("2026-11-19 08:00", "start", map[string]string{"hero": "oren"})},
		Segments: explore("2026-11-19 09:00", 6*time.Hour),
		Now:      at("2026-11-19 16:00"),
	})
	// DPS 40 × 0,6 × 1,04 = 24,96. Andares 1 a 4: 72 + 77 + 83 + 88 min ≈ 320 min < 6h;
	// o 5º (nível 2) pede mais 63 min. Moedas 13+16+19+22; XP 25+30+35+40 = 130 (nível 2 aos 100).
	if st.Floor != 5 {
		t.Errorf("andar = %d, quer 5", st.Floor)
	}
	if st.Coins != 70 {
		t.Errorf("moedas = %d, quer 70", st.Coins)
	}
	if h := st.Hero("oren"); h == nil || h.Level != 2 || h.XP != 30 {
		t.Errorf("oren = %+v", h)
	}
}

func TestSemStartNaoHaJogo(t *testing.T) {
	st := Play(PlayInput{Cal: cal, Segments: explore("2026-11-19 09:00", time.Hour), Now: at("2026-11-19 16:00")})
	if st.Started || st.Floor != 1 || st.Coins != 0 {
		t.Fatalf("got %+v", st)
	}
}

func bosses(st *State) []LogEntry {
	var out []LogEntry
	for _, e := range st.Log {
		if e.Kind == "boss" {
			out = append(out, e)
		}
	}
	return out
}

func TestOrenNoNivelDaReferenciaVenceOChefeDoAndar10ELiberaRecrutamento(t *testing.T) {
	st := Play(PlayInput{
		Cal:      cal,
		Events:   []Event{ev("2026-11-19 08:00", "start", map[string]string{"hero": "oren"})},
		Segments: explore("2026-11-19 09:00", 16*time.Hour),
		Now:      at("2026-11-20 05:00"),
	})
	b := bosses(st)
	if len(b) != 1 || !b[0].Win || b[0].Floor != 10 {
		t.Fatalf("lutas = %+v (andar %d)", b, st.Floor)
	}
	if st.Recruits != 1 || st.Floor < 11 || st.Training {
		t.Fatalf("got andar %d, recrutas %d, treino %v", st.Floor, st.Recruits, st.Training)
	}
}

func TestKaelPerdeParaOChefeTreinaNoAndar9EDesafiaUmaVezPorDia(t *testing.T) {
	in := PlayInput{
		Cal:      cal,
		Events:   []Event{ev("2026-11-19 08:00", "start", map[string]string{"hero": "kael"})},
		Segments: explore("2026-11-19 09:00", 96*time.Hour),
		Now:      at("2026-11-23 09:00"),
	}
	st := Play(in)
	b := bosses(st)
	if len(b) != 1 || b[0].Win || !st.Training || st.Floor != 9 {
		t.Fatalf("lutas = %+v, andar %d, treino %v", b, st.Floor, st.Training)
	}
	lost := st.in.Cal.DayOf(b[0].At)

	// Desafiar no mesmo dia da derrota não faz nada.
	in.Events = append(in.Events, ev(b[0].At.Add(time.Minute).In(sp).Format("2006-01-02 15:04"), "challenge", nil))
	if got := bosses(Play(in)); len(got) != 1 {
		t.Fatalf("desafio no dia da derrota lutou: %+v", got)
	}
	// No dia seguinte, luta de novo.
	next := cal.Start(core.AddDays(lost, 1)).Add(time.Hour)
	in.Events = append(in.Events, ev(next.In(sp).Format("2006-01-02 15:04"), "challenge", nil))
	in.Now = next.Add(time.Hour)
	if got := bosses(Play(in)); len(got) != 2 {
		t.Fatalf("desafio no dia seguinte não lutou: %+v", got)
	}
}

func TestRecrutaEntraNoNivelDoMaisBaixoEOEsquadraoEscolheQuemGanhaXP(t *testing.T) {
	in := PlayInput{
		Cal:      cal,
		Events:   []Event{ev("2026-11-19 08:00", "start", map[string]string{"hero": "oren"})},
		Segments: explore("2026-11-19 09:00", 16*time.Hour),
		Now:      at("2026-11-20 05:00"),
	}
	st := Play(in)
	level := st.Hero("oren").Level
	in.Events = append(in.Events,
		ev("2026-11-20 05:00", "recruit", map[string]string{"hero": "lyra"}),
		ev("2026-11-20 05:00", "recruit", map[string]string{"hero": "brom"}), // só havia 1
	)
	in.Now = at("2026-11-20 05:01")
	st = Play(in)
	if st.Hero("brom") != nil || st.Hero("lyra") == nil || st.Hero("lyra").Level != level {
		t.Fatalf("heróis = %+v", st.Heroes)
	}
	if len(st.Squad) != 2 {
		t.Fatalf("esquadrão = %v", st.Squad)
	}

	in.Events = append(in.Events, ev("2026-11-20 05:01", "squad", map[string][]string{"heroes": {"lyra"}}))
	in.Segments = append(in.Segments, explore("2026-11-20 06:00", 2*time.Hour)...)
	in.Now = at("2026-11-20 09:00")
	st = Play(in)
	if st.Hero("oren").Level != level || st.Hero("oren").XP != Play(PlayInput{Cal: cal, Events: in.Events[:1], Segments: in.Segments[:1], Now: at("2026-11-20 05:00")}).Hero("oren").XP {
		t.Fatalf("oren fora do esquadrão ganhou XP")
	}
	if st.Hero("lyra").XP == 0 && st.Hero("lyra").Level == level {
		t.Fatalf("lyra não ganhou XP")
	}
}
