package game

import (
	"testing"
	"time"

	"habit-idle/internal/core"
)

func TestComprarPontoCobraPrecoCrescenteENaoDeixaFicarNegativo(t *testing.T) {
	in := PlayInput{
		Cal:      cal,
		Events:   []Event{ev("2026-11-19 08:00", "start", map[string]string{"hero": "oren"})},
		Segments: explore("2026-11-19 09:00", 7*time.Hour),
		Now:      at("2026-11-19 18:00"),
	}
	st := Play(in)
	before, coins := st.Hero("oren").Points[ATK], st.Coins
	if coins < 50 || coins-50 >= 58 {
		t.Fatalf("cenário: %d moedas, preciso de 50 a 107", coins)
	}
	in.Events = append(in.Events,
		ev("2026-11-19 18:00", "buy", map[string]string{"hero": "oren", "stat": "atk"}), // 50
		ev("2026-11-19 18:00", "buy", map[string]string{"hero": "oren", "stat": "atk"}), // 58: não tem
	)
	in.Now = at("2026-11-19 18:01")
	st = Play(in)
	if st.Coins != coins-50 {
		t.Fatalf("moedas = %d, quer %d", st.Coins, coins-50)
	}
	if got := st.Hero("oren").Points[ATK]; got != before+1 {
		t.Fatalf("ATK = %d, quer %d", got, before+1)
	}
	if BuyCost(1) != 58 {
		t.Fatalf("segundo ponto = %d", BuyCost(1))
	}
}

// Trinta dias completos: um baú por dia, na virada.
func monthOfChests() PlayInput {
	var days []core.Day
	for d := core.Day("2026-11-01"); d <= "2026-11-30"; d = core.AddDays(d, 1) {
		days = append(days, d)
	}
	return PlayInput{
		Cal:      cal,
		Events:   []Event{ev("2026-10-31 08:00", "start", map[string]string{"hero": "kael"})},
		Complete: days,
		Now:      at("2026-12-01 12:00"),
	}
}

func TestDiaCompletoDaBauDeAndarComumOuRaro(t *testing.T) {
	st := Play(monthOfChests())
	if len(st.Items) != 30 {
		t.Fatalf("itens = %d", len(st.Items))
	}
	for _, it := range st.Items {
		if it.Rarity > 1 {
			t.Fatalf("baú de andar com %s", Rarities[it.Rarity])
		}
	}
}

func TestEquiparSomaOsPontosDoItemEVenderDaMoedas(t *testing.T) {
	in := monthOfChests()
	st := Play(in)
	it := st.Items[0]
	before := st.Hero("kael").Points
	in.Events = append(in.Events, ev("2026-12-01 12:00", "equip", map[string]any{"hero": "kael", "item": it.ID}))
	in.Now = at("2026-12-01 12:01")
	st = Play(in)
	after := st.Hero("kael").Points
	ip := st.item(it.ID).Points()
	for i := range after {
		if after[i] != before[i]+ip[i] {
			t.Fatalf("status %s: %d, quer %d", StatKeys[i], after[i], before[i]+ip[i])
		}
	}

	in.Events = append(in.Events, ev("2026-12-01 12:01", "sell", map[string]any{"item": it.ID}))
	in.Now = at("2026-12-01 12:02")
	st = Play(in)
	if st.item(it.ID) != nil || st.Hero("kael").Points != before {
		t.Fatal("item vendido continua")
	}
	// comum: 10 × 1; rara: 10 × 2 (andar 1: × 1,1).
	want := map[int]int{0: 11, 1: 22}[it.Rarity]
	if st.Coins != want {
		t.Fatalf("moedas = %d, quer %d", st.Coins, want)
	}
}

func TestRefinarJuntaDoisIguaisEOItemFicaMaisForte(t *testing.T) {
	in := monthOfChests()
	st := Play(in)
	var a, b *Item
	for i, x := range st.Items {
		for _, y := range st.Items[i+1:] {
			if a == nil && x.Name == y.Name && x.Rarity == y.Rarity {
				a, b = x, y
			}
		}
	}
	if a == nil {
		t.Fatal("30 baús sem um par igual")
	}
	base := a.Points()
	in.Events = append(in.Events, ev("2026-12-01 12:00", "refine", map[string]any{"item": a.ID, "with": b.ID}))
	in.Now = at("2026-12-01 12:01")
	st = Play(in)
	got := st.item(a.ID)
	if got == nil || got.Refine != 1 || st.item(b.ID) != nil || len(st.Items) != 29 {
		t.Fatalf("refino = %+v, itens %d", got, len(st.Items))
	}
	main, _ := ParseStat(got.Main)
	if got.Points()[main] <= base[main] && base[main] >= 5 {
		t.Fatalf("não ficou mais forte: %v → %v", base, got.Points())
	}
}
