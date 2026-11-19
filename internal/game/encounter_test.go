package game

import (
	"testing"
	"time"
)

func startWith(hero string) []Event {
	return []Event{ev("2026-11-19 08:00", "start", map[string]string{"hero": hero})}
}

func TestOPrimeiroInimigoApareceDepoisDe5MinDeCaminhada(t *testing.T) {
	in := PlayInput{Cal: cal, Events: startWith("oren"), Segments: explore("2026-11-19 09:00", 4*time.Minute), Now: at("2026-11-19 10:00")}
	st := Play(in)
	if st.Enc.Enemy != 1 || st.Enc.Phase != "walk" || st.Coins != 0 {
		t.Fatalf("4 min: %+v, moedas %d", st.Enc, st.Coins)
	}
	in.Segments = explore("2026-11-19 09:00", 6*time.Minute)
	if st = Play(in); st.Enc.Enemy != 1 || st.Enc.Phase != "fight" {
		t.Fatalf("6 min: %+v", st.Enc)
	}
}

func TestMoedasEntramACadaInimigoEFechamOTotalDoAndar(t *testing.T) {
	in := PlayInput{Cal: cal, Events: startWith("oren"), Segments: explore("2026-11-19 09:00", 24*time.Hour)}
	midFloor := false
	for m := 0; m < 24*60; m++ {
		in.Now = at("2026-11-19 09:00").Add(time.Duration(m) * time.Minute)
		st := Play(in)
		if st.Floor == 1 && st.Enc.Enemy > 2 && st.Coins > 0 && st.Coins < 13 {
			midFloor = true
		}
		if st.Floor == 2 {
			// Andar 1 paga 13 moedas (10 + 3×1), divididas entre os 12 inimigos.
			if st.Coins != 13 || st.Enc.Enemy != 1 {
				t.Fatalf("ao chegar no andar 2: moedas %d, %+v", st.Coins, st.Enc)
			}
			if !midFloor {
				t.Fatal("as moedas só entraram no fim do andar")
			}
			return
		}
	}
	t.Fatal("não passou do andar 1 em 24h")
}

func TestQualquerHeroiInicialPassaOsAndares1a5SemMorrer(t *testing.T) {
	for _, h := range Catalog {
		st := Play(PlayInput{Cal: cal, Events: startWith(h.ID), Segments: explore("2026-11-19 09:00", 48*time.Hour), Now: at("2026-11-22 09:00")})
		if st.Floor < 6 && !st.Training {
			t.Errorf("%s ficou no andar %d em 48h", h.ID, st.Floor)
		}
		for _, e := range st.Log {
			if e.Kind == "death" && e.Floor <= 5 {
				t.Errorf("%s morreu no andar %d", h.ID, e.Floor)
			}
		}
	}
}

func TestEsquadraoQueMorreVoltaAoPrimeiroInimigoDoAndar(t *testing.T) {
	// Cansaço absurdo (só possível em teste) deixa o herói sem vida nem dano.
	st := Play(PlayInput{
		Cal: cal, Events: startWith("oren"),
		Segments: explore("2026-11-19 09:00", 3*time.Hour),
		Fatigue:  []FatigueChange{{At: at("2026-11-19 08:00"), Pct: 97}},
		Now:      at("2026-11-19 13:00"),
	})
	died := false
	for _, e := range st.Log {
		died = died || (e.Kind == "death" && e.Floor == 1)
	}
	if !died || st.Floor != 1 || st.Enc.Enemy != 1 {
		t.Fatalf("morreu %v, andar %d, %+v", died, st.Floor, st.Enc)
	}
}
