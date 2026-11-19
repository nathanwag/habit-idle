package game

import "time"

// Números do jogo. Espelham docs/regras.md: mudou aqui, muda lá.

const (
	FolegoCap      = 16 * time.Hour // teto do saldo
	DayBonus       = time.Hour      // marcar a última missão do dia
	FatigueStep    = 10             // % por dia
	FatigueMax     = 30             // %
	projectionDays = 3              // até onde o "explorando até" olha
)
