package game

import (
	"fmt"
	"io"
)

// Table imprime o ritmo esperado da torre: o esquadrão de referência e o Kael
// sozinho (o herói inicial que menos bate), sem compras nem itens. Base para
// ajustar balance.go.
func Table(w io.Writer, floors int) {
	fmt.Fprintln(w, "| Andar | Nível | Heróis | Vida do inimigo | Min/andar | Vida por luta | Min/andar (Kaels) | Vida por luta (Kaels) | Moedas | XP |")
	fmt.Fprintln(w, "|---|---|---|---|---|---|---|---|---|---|")
	for n := 1; n <= floors; n++ {
		if n != 1 && n%5 != 0 {
			continue
		}
		e := EnemyOf(n)
		dps, ehp, heroes := refPower(n)
		rt, rl := pace(dps, ehp, e)
		// Só Kaels, com o mesmo número de heróis da referência.
		kf := NewFighter(Points(DEF, refLevel(n), [nStats]int{}), 0)
		kt, kl := pace(kf.DPS()*float64(heroes), kf.EffectiveLife()*float64(heroes), e)
		fmt.Fprintf(w, "| %d | %d | %d | %.0f | %.0f | %.0f%% | %.0f | %.0f%% | %d | %d |\n",
			n, refLevel(n), heroes, e.Life, rt, rl*100, kt, kl*100, FloorCoins+FloorCoinsN*n, FloorXP+FloorXPN*n)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "| Chefe | Nível da referência | Heróis | Poder recomendado | Vida | Dano/s |")
	fmt.Fprintln(w, "|---|---|---|---|---|---|")
	for k := 1; k <= floors/BossEvery; k++ {
		b := BossOf(k * BossEvery)
		fmt.Fprintf(w, "| %d | %d | %d | %d | %.0f | %.0f |\n", k*BossEvery, refLevel(k*BossEvery), min(SquadSize, k), b.Recommended, b.Life, b.DPS)
	}
}

// pace: minutos por andar e fração da vida efetiva perdida em cada luta.
func pace(dps, ehp float64, e Enemy) (float64, float64) {
	secs := e.Life / dps
	return EnemiesPerFloor * (WalkMins + secs/60), e.Hit * EnemySpeed * secs / ehp
}
