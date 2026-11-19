package game

import (
	"fmt"
	"io"
)

// Table imprime o ritmo esperado da torre com um herói de cada vez (o da
// referência do chefe), sem compras nem itens: base para ajustar balance.go.
func Table(w io.Writer, floors int) {
	fmt.Fprintln(w, "| Andar | Nível | Heróis | Min/andar | Min/andar (só Kael) | Moedas | XP |")
	fmt.Fprintln(w, "|---|---|---|---|---|---|---|")
	for n := 1; n <= floors; n++ {
		if n != 1 && n%5 != 0 {
			continue
		}
		lv := refLevel(n)
		heroes := min(SquadSize, 1+(n-1)/BossEvery)
		var squad []Fighter
		for i := 0; i < heroes; i++ {
			squad = append(squad, NewFighter(Points(Catalog[i].Spec, lv, [nStats]int{}), 0))
		}
		kael := NewFighter(Points(DEF, lv, [nStats]int{}), 0)
		fmt.Fprintf(w, "| %d | %d | %d | %.0f | %.0f | %d | %d |\n", n, lv, heroes,
			floorMinutes(n, squad), floorMinutes(n, []Fighter{kael}),
			FloorCoins+FloorCoinsN*n, FloorXP+FloorXPN*n)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "| Chefe | Nível da referência | Heróis | Poder recomendado | Vida | Dano/s |")
	fmt.Fprintln(w, "|---|---|---|---|---|---|")
	for k := 1; k <= floors/BossEvery; k++ {
		b := BossOf(k * BossEvery)
		fmt.Fprintf(w, "| %d | %d | %d | %d | %.0f | %.0f |\n", k*BossEvery, refLevel(k*BossEvery), min(SquadSize, k), b.Recommended, b.Life, b.DPS)
	}
}
