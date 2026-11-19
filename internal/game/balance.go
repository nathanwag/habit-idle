package game

import "time"

// Números do jogo. Espelham docs/regras.md: mudou aqui, muda lá.

const (
	FolegoCap      = 16 * time.Hour // teto do saldo
	DayBonus       = time.Hour      // marcar a última missão do dia
	BonusMissions  = 2              // missões mínimas no dia para o bônus e o baú
	FatigueStep    = 10             // % por dia
	FatigueMax     = 30             // %
	projectionDays = 3              // até onde o "explorando até" olha
)

// Status: pontos → valores de combate.
const (
	BasePoints     = 4  // nível 1, cada status
	SpecPoints     = 10 // nível 1, especialidade
	LevelPoints    = 1  // por nível, cada status
	LevelSpecPoint = 2  // por nível, especialidade
	PowerPerPoint  = 10

	DamagePerAtk = 4.0
	DefHalf      = 40.0 // redução = pts / (pts + DefHalf)
	LifePerHp    = 30.0
	BaseSpeed    = 0.5  // golpes/s
	SpeedPerVel  = 20.0 // golpes/s = BaseSpeed × (1 + pts/SpeedPerVel)
	CritPerSor   = 0.01
	CritMax      = 0.6
	CritMult     = 2.0

	SquadSize = 3
)

// Torre.
const (
	FloorLife    = 1800.0
	FloorThreat  = 100.0
	FloorGrowth  = 1.07
	FloorCoins   = 10 // + FloorCoinsN × andar
	FloorCoinsN  = 3
	FloorXP      = 20 // + FloorXPN × andar
	FloorXPN     = 5
	ChestChance  = 0.20
	XPBase       = 100.0
	XPGrowth     = 1.12
	BossEvery    = 10
	BossKillSecs = 60.0 // a referência mata o chefe nisso
	BossDieSecs  = 75.0 // o chefe mataria a referência nisso
	BossMaxSecs  = 120.0
	BossTick     = 0.1
	BossCoins    = 250 // × k
	BossXP       = 200 // × k
	RecruitUntil = 40  // chefes que liberam recrutamento: 10, 20, 30, 40
)

// Equipamento e loja.
const (
	ItemBase        = 2.0
	ItemFloorDiv    = 10.0 // pontos × (1 + andar/ItemFloorDiv)
	ItemExtraShare  = 0.4
	ItemRefineStep  = 0.1
	RefineMax       = 10
	RareChance      = 0.25 // baú de andar ou de dia
	LegendaryChance = 0.15 // baú de chefe
	SellBase        = 10.0
	BuyBase         = 50.0
	BuyGrowth       = 1.15
)
