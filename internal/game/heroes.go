package game

import "math"

type Stat int

const (
	ATK Stat = iota
	DEF
	HP
	VEL
	SOR
	nStats
)

var StatKeys = [nStats]string{"atk", "def", "hp", "vel", "sor"}

func ParseStat(s string) (Stat, bool) {
	for i, k := range StatKeys {
		if k == s {
			return Stat(i), true
		}
	}
	return 0, false
}

type HeroInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Class string `json:"class"`
	Spec  Stat   `json:"-"`
}

// Catalog na ordem de exibição. A referência dos chefes usa ATK, DEF e HP.
var Catalog = []HeroInfo{
	{"oren", "Oren", "Mago", ATK},
	{"kael", "Kael", "Cavaleiro", DEF},
	{"brom", "Brom", "Anão", HP},
	{"rasko", "Rasko", "Homem-lagarto", VEL},
	{"lyra", "Lyra", "Elfa", SOR},
}

func info(id string) (HeroInfo, bool) {
	for _, h := range Catalog {
		if h.ID == id {
			return h, true
		}
	}
	return HeroInfo{}, false
}

// Points por status, de um herói no nível dado com pontos extras.
func Points(spec Stat, level int, extra [nStats]int) [nStats]int {
	var p [nStats]int
	for s := Stat(0); s < nStats; s++ {
		if s == spec {
			p[s] = SpecPoints + LevelSpecPoint*(level-1)
		} else {
			p[s] = BasePoints + LevelPoints*(level-1)
		}
		p[s] += extra[s]
	}
	return p
}

// Fighter: valores de combate de um herói (já com o cansaço).
type Fighter struct {
	Damage, Reduction, Life, Speed, Crit float64
}

func NewFighter(p [nStats]int, fatigue int) Fighter {
	k := 1 - float64(fatigue)/100
	return Fighter{
		Damage:    DamagePerAtk * float64(p[ATK]) * k,
		Reduction: float64(p[DEF]) / (float64(p[DEF]) + DefHalf),
		Life:      LifePerHp * float64(p[HP]) * k,
		Speed:     BaseSpeed * (1 + float64(p[VEL])/SpeedPerVel),
		Crit:      math.Min(CritMax, CritPerSor*float64(p[SOR])),
	}
}

func (f Fighter) DPS() float64 { return f.Damage * f.Speed * (1 + f.Crit*(CritMult-1)) }

func (f Fighter) EffectiveLife() float64 { return f.Life / (1 - f.Reduction) }

func sumPoints(p [nStats]int) int {
	t := 0
	for _, v := range p {
		t += v
	}
	return t
}

// XPToNext: XP para ir do nível L para L+1.
func XPToNext(level int) int {
	return int(math.Round(XPBase * math.Pow(XPGrowth, float64(level-1))))
}
