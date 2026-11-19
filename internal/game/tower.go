package game

import (
	"encoding/json"
	"hash/fnv"
	"math"
	"math/rand/v2"
	"sort"
	"time"

	"habit-idle/internal/core"
)

// Event: uma escolha do jogador, gravada com a hora.
type Event struct {
	ID   int             `json:"id"`
	At   time.Time       `json:"at"`
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

type PlayInput struct {
	Cal      core.Calendar
	Events   []Event
	Segments []Segment       // quando o esquadrão explorou (do Folego)
	Fatigue  []FatigueChange // do Folego
	Complete []core.Day      // dias completos (do Folego): baú na virada
	Now      time.Time
}

type HeroState struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Class   string       `json:"class"`
	Spec    string       `json:"spec"`
	Level   int          `json:"level"`
	XP      int          `json:"xp"`
	XPNext  int          `json:"xpNext"`
	Bought  [nStats]int  `json:"bought"`
	Points  [nStats]int  `json:"points"`
	Power   int          `json:"power"`
	InSquad bool         `json:"inSquad"`
	Equip   map[Slot]int `json:"equip"` // espaço → id do item
	Costs   [nStats]int  `json:"costs"` // preço do próximo ponto de cada status
	spec    Stat
}

type LogEntry struct {
	At    time.Time `json:"at"`
	Kind  string    `json:"kind"` // floor, boss, level, chest, recruit
	Floor int       `json:"floor,omitempty"`
	Hero  string    `json:"hero,omitempty"`
	Win   bool      `json:"win,omitempty"`
	Coins int       `json:"coins,omitempty"`
	XP    int       `json:"xp,omitempty"`
	Item  int       `json:"item,omitempty"`
	Enemy int       `json:"enemy,omitempty"`
}

type State struct {
	Started     bool         `json:"started"`
	Heroes      []*HeroState `json:"heroes"` // na ordem em que entraram
	Squad       []string     `json:"squad"`
	Floor       int          `json:"floor"` // andar sendo explorado
	Enc         Encounter    `json:"enc"`
	Coins       int          `json:"coins"`
	Power       int          `json:"power"`
	Fatigue     int          `json:"fatigue"`
	Training    bool         `json:"training"` // perdeu para o chefe e treina no andar anterior
	NextBoss    int          `json:"nextBoss"`
	Recommended int          `json:"recommended"`
	Recruits    int          `json:"recruits"` // recrutamentos liberados e não usados
	LastFight   core.Day     `json:"lastFight"`
	CanFight    bool         `json:"canFight"`
	Beaten      int          `json:"beaten"` // chefes vencidos
	Items       []*Item      `json:"items"`
	Log         []LogEntry   `json:"log"`

	in       PlayInput
	nextItem int
	passes   map[int]int // vezes que cada andar foi vencido (semente do baú)
	fights   int
	deaths   int
}

func (s *State) Hero(id string) *HeroState {
	for _, h := range s.Heroes {
		if h.ID == id {
			return h
		}
	}
	return nil
}

// Play recalcula o jogo inteiro a partir dos eventos e da exploração.
func Play(in PlayInput) *State {
	s := &State{in: in, Floor: 1, Heroes: []*HeroState{}, Squad: []string{}, Items: []*Item{}, Log: []LogEntry{}, passes: map[int]int{},
		Enc: Encounter{Enemy: 1, Of: EnemiesPerFloor, Phase: "walk", left: WalkMins, Life: map[string]float64{}, MaxLife: map[string]float64{}}}

	type step struct {
		at    time.Time
		event *Event
		chest core.Day
	}
	var steps []step
	for i := range in.Events {
		steps = append(steps, step{at: in.Events[i].At, event: &in.Events[i]})
	}
	for _, d := range in.Complete {
		steps = append(steps, step{at: in.Cal.Start(core.AddDays(d, 1)), chest: d})
	}
	sort.SliceStable(steps, func(a, b int) bool { return steps[a].at.Before(steps[b].at) })

	var cursor time.Time
	for _, st := range steps {
		if st.at.After(in.Now) {
			break
		}
		s.explore(cursor, st.at)
		cursor = st.at
		if st.event != nil {
			s.apply(*st.event)
		} else if s.Started {
			s.chest(st.at, "day", s.Floor, seed("day", st.chest))
		}
	}
	s.explore(cursor, in.Now)
	s.finish()
	return s
}

func seed(parts ...any) uint64 {
	h := fnv.New64a()
	for _, p := range parts {
		b, _ := json.Marshal(p)
		h.Write(b)
		h.Write([]byte{0})
	}
	return h.Sum64()
}

func rng(sd uint64) *rand.Rand { return rand.New(rand.NewPCG(sd, sd^0x9e3779b97f4a7c15)) }

func (s *State) fatigueAt(t time.Time) int {
	pct := 0
	for _, c := range s.in.Fatigue {
		if c.At.After(t) {
			break
		}
		pct = c.Pct
	}
	return pct
}

func (s *State) squad() []*HeroState {
	var out []*HeroState
	for _, id := range s.Squad {
		out = append(out, s.Hero(id))
	}
	return out
}

func (s *State) fighters(t time.Time) []Fighter {
	f := s.fatigueAt(t)
	var out []Fighter
	for _, h := range s.squad() {
		out = append(out, NewFighter(s.points(h), f))
	}
	return out
}

func (s *State) points(h *HeroState) [nStats]int {
	extra := h.Bought
	for _, id := range h.Equip {
		if it := s.item(id); it != nil {
			ip := it.Points()
			for i := range extra {
				extra[i] += ip[i]
			}
		}
	}
	return Points(h.spec, h.Level, extra)
}

// Enemy: os status de um inimigo comum do andar n.
type Enemy struct {
	Life, Hit float64 // vida e dano por golpe (antes da DEF do herói)
}

// refPower: dano por segundo e vida efetiva do esquadrão "na média" do andar
// n: heróis no nível que o XP dos andares anteriores dá, um a mais a cada 10
// andares (até 3), cada um com a média dos 5 heróis (justo com qualquer um).
func refPower(n int) (dps, ehp float64, heroes int) {
	level := refLevel(n)
	heroes = min(SquadSize, 1+(n-1)/BossEvery)
	for _, h := range Catalog {
		f := NewFighter(Points(h.Spec, level, [nStats]int{}), 0)
		dps += f.DPS()
		ehp += f.EffectiveLife()
	}
	k := float64(heroes) / float64(len(Catalog))
	return dps * k, ehp * k, heroes
}

// EnemyOf: a referência mata o inimigo em FightSecs e perde EnemyBite da vida
// efetiva nisso. Nos primeiros andares o inimigo bate bem menos.
func EnemyOf(n int) Enemy {
	dps, ehp, _ := refPower(n)
	hit := EnemyBite * ehp / FightSecs / EnemySpeed
	if n <= SafeFloors {
		hit *= SafeFactor
	}
	return Enemy{Life: dps * FightSecs, Hit: hit}
}

// Encounter: onde o esquadrão está dentro do andar.
type Encounter struct {
	Enemy     int                `json:"enemy"` // 1 a EnemiesPerFloor
	Of        int                `json:"of"`
	Phase     string             `json:"phase"` // walk, fight, rest
	EnemyLife float64            `json:"enemyLife"`
	EnemyMax  float64            `json:"enemyMax"`
	Life      map[string]float64 `json:"life"` // vida atual de cada herói
	MaxLife   map[string]float64 `json:"maxLife"`

	left float64 // minutos que faltam da caminhada ou do descanso
	acc  map[string]float64
	eAcc float64
	r    *rand.Rand
	fs   []Fighter // valem a luta inteira; um evento zera e eles são refeitos
	hit  float64
}

// fightFighters: os valores de combate da luta atual, calculados uma vez.
func (s *State) fightFighters(t time.Time) []Fighter {
	if s.Enc.fs == nil || len(s.Enc.fs) != len(s.Squad) {
		s.Enc.fs = s.fighters(t)
	}
	return s.Enc.fs
}

func (s *State) explore(from, to time.Time) {
	for _, seg := range s.in.Segments {
		a, b := seg.From, seg.To
		if a.Before(from) {
			a = from
		}
		if b.After(to) {
			b = to
		}
		if a.Before(b) {
			s.run(a, b)
		}
	}
}

// heal enche a vida de todos (começo de andar ou depois de cair).
func (s *State) heal(t time.Time) {
	fs := s.fighters(t)
	for i, h := range s.squad() {
		s.Enc.Life[h.ID] = fs[i].Life
		s.Enc.MaxLife[h.ID] = fs[i].Life
	}
}

func (s *State) run(a, b time.Time) {
	if !s.Started || len(s.Squad) == 0 {
		return
	}
	t := a
	for t.Before(b) {
		left := b.Sub(t).Minutes()
		switch s.Enc.Phase {
		case "walk", "rest":
			step := math.Min(left, s.Enc.left)
			if s.Enc.Phase == "walk" {
				s.regen(t, step)
			}
			s.Enc.left -= step
			t = t.Add(time.Duration(step * float64(time.Minute)))
			if s.Enc.left > 1e-9 {
				return
			}
			if s.Enc.Phase == "rest" {
				s.heal(t)
				s.Enc.Phase, s.Enc.left = "walk", WalkMins
				continue
			}
			s.startFight(t)
		case "fight":
			if left*60 < 1 {
				return
			}
			t = t.Add(time.Second)
			s.tick(t)
		}
	}
}

func (s *State) regen(t time.Time, minutes float64) {
	fs := s.fighters(t)
	for i, h := range s.squad() {
		max := fs[i].Life
		s.Enc.MaxLife[h.ID] = max
		life, ok := s.Enc.Life[h.ID]
		switch {
		case !ok:
			s.Enc.Life[h.ID] = max // entrou no meio do andar
		case life > 0:
			s.Enc.Life[h.ID] = math.Min(max, life+max*RegenPerMin*minutes)
		}
	}
}

func (s *State) startFight(t time.Time) {
	e := EnemyOf(s.Floor)
	s.Enc.Phase = "fight"
	s.Enc.EnemyLife, s.Enc.EnemyMax = e.Life, e.Life
	s.Enc.acc = map[string]float64{}
	s.Enc.eAcc = 0
	s.Enc.r = rng(seed("enemy", s.Floor, s.passes[s.Floor], s.Enc.Enemy, s.deaths))
	s.Enc.fs = s.fighters(t)
	s.Enc.hit = e.Hit
}

// tick: um segundo de luta. Os heróis vivos golpeiam no ritmo da VEL; o
// inimigo bate no primeiro herói vivo, com o dano reduzido pela DEF dele.
func (s *State) tick(t time.Time) {
	fs := s.fightFighters(t)
	squad := s.squad()
	target := -1
	for i, h := range squad {
		s.Enc.MaxLife[h.ID] = fs[i].Life
		life, ok := s.Enc.Life[h.ID]
		if !ok {
			life = fs[i].Life
			s.Enc.Life[h.ID] = life
		}
		if life <= 0 {
			continue
		}
		if target < 0 {
			target = i
		}
		s.Enc.acc[h.ID] += fs[i].Speed
		for s.Enc.acc[h.ID] >= 1 {
			s.Enc.acc[h.ID]--
			dmg := fs[i].Damage
			if s.Enc.r.Float64() < fs[i].Crit {
				dmg *= CritMult
			}
			s.Enc.EnemyLife -= dmg
		}
	}
	if s.Enc.EnemyLife <= 0 {
		s.kill(t)
		return
	}
	if target < 0 {
		s.die(t)
		return
	}
	s.Enc.eAcc += EnemySpeed
	for s.Enc.eAcc >= 1 {
		s.Enc.eAcc--
		s.Enc.Life[squad[target].ID] -= s.Enc.hit * (1 - fs[target].Reduction)
	}
	for _, h := range squad {
		if s.Enc.Life[h.ID] > 0 {
			return
		}
	}
	s.die(t)
}

// share: a parte do total do andar que o inimigo i (a partir de 0) paga; a
// soma dos EnemiesPerFloor dá o total exato.
func share(total, i int) int {
	return total*(i+1)/EnemiesPerFloor - total*i/EnemiesPerFloor
}

func (s *State) kill(t time.Time) {
	n, i := s.Floor, s.Enc.Enemy-1
	s.Coins += share(FloorCoins+FloorCoinsN*n, i)
	s.giveXP(t, share(FloorXP+FloorXPN*n, i))
	s.Enc.Phase, s.Enc.left = "walk", WalkMins
	if s.Enc.Enemy < EnemiesPerFloor {
		s.Enc.Enemy++
		return
	}
	s.clearFloor(t)
}

// die: todos caíram. Descansam e recomeçam o andar do primeiro inimigo; o que
// já ganharam fica.
func (s *State) die(t time.Time) {
	s.deaths++
	s.addLog(LogEntry{At: t, Kind: "death", Floor: s.Floor, Enemy: s.Enc.Enemy})
	s.Enc.Enemy = 1
	s.Enc.Phase, s.Enc.left = "rest", DeathRestMins
}

func (s *State) clearFloor(t time.Time) {
	n := s.Floor
	s.passes[n]++
	s.Enc.Enemy = 1
	s.addLog(LogEntry{At: t, Kind: "floor", Floor: n, Coins: FloorCoins + FloorCoinsN*n, XP: FloorXP + FloorXPN*n})
	if r := rng(seed("chest", n, s.passes[n])); r.Float64() < ChestChance {
		s.chest(t, "floor", n, seed("item", n, s.passes[n]))
	}
	switch {
	case s.Training:
	case (n+1)%BossEvery == 0:
		s.fight(t, n+1)
	default:
		s.Floor = n + 1
	}
	s.heal(t)
}

func (s *State) giveXP(t time.Time, xp int) {
	for _, h := range s.squad() {
		h.XP += xp
		for h.XP >= XPToNext(h.Level) {
			h.XP -= XPToNext(h.Level)
			h.Level++
			s.addLog(LogEntry{At: t, Kind: "level", Hero: h.ID, Floor: h.Level})
		}
	}
}

func (s *State) addLog(e LogEntry) {
	s.Log = append(s.Log, e)
}

// Boss: os status do chefe do andar 10k.
type Boss struct {
	Life, DPS   float64
	Recommended int
}

// refLevel: o nível que um herói alcança com o XP dos andares antes do chefe.
func refLevel(bossFloor int) int {
	xp := 0
	for n := 1; n < bossFloor; n++ {
		xp += FloorXP + FloorXPN*n
	}
	level := 1
	for xp >= XPToNext(level) {
		xp -= XPToNext(level)
		level++
	}
	return level
}

func BossOf(bossFloor int) Boss {
	k := bossFloor / BossEvery
	level := refLevel(bossFloor)
	var b Boss
	life, dps, pts := 0.0, 0.0, 0
	for i := 0; i < min(SquadSize, k); i++ {
		p := Points(Catalog[i].Spec, level, [nStats]int{})
		f := NewFighter(p, 0)
		dps += f.DPS()
		life += f.EffectiveLife()
		pts += sumPoints(p)
	}
	b.Life = dps * BossKillSecs
	b.DPS = life / BossDieSecs
	b.Recommended = pts * PowerPerPoint
	return b
}

// Fight simula a luta em passos de BossTick. O chefe bate no primeiro herói vivo.
func Fight(fs []Fighter, b Boss, r *rand.Rand) bool {
	life := make([]float64, len(fs))
	acc := make([]float64, len(fs))
	for i, f := range fs {
		life[i] = f.Life
	}
	boss := b.Life
	for t := 0.0; t < BossMaxSecs; t += BossTick {
		target := -1
		for i := range fs {
			if life[i] <= 0 {
				continue
			}
			if target < 0 {
				target = i
			}
			acc[i] += fs[i].Speed * BossTick
			for acc[i] >= 1 {
				acc[i]--
				dmg := fs[i].Damage
				if r.Float64() < fs[i].Crit {
					dmg *= CritMult
				}
				boss -= dmg
			}
		}
		if target < 0 {
			return false
		}
		if boss <= 0 {
			return true
		}
		life[target] -= b.DPS * BossTick * (1 - fs[target].Reduction)
	}
	return false
}

func (s *State) fight(t time.Time, bossFloor int) {
	day := s.in.Cal.DayOf(t)
	s.fights++
	win := Fight(s.fighters(t), BossOf(bossFloor), rng(seed("boss", bossFloor, day, s.fights)))
	s.LastFight = day
	k := bossFloor / BossEvery
	if !win {
		s.Training = true
		s.Floor = bossFloor - 1
		s.addLog(LogEntry{At: t, Kind: "boss", Floor: bossFloor})
		return
	}
	s.Training = false
	s.Beaten++
	s.Floor = bossFloor + 1
	coins, xp := BossCoins*k, BossXP*k
	s.Coins += coins
	s.addLog(LogEntry{At: t, Kind: "boss", Floor: bossFloor, Win: true, Coins: coins, XP: xp})
	s.giveXP(t, xp)
	s.chest(t, "boss", bossFloor, seed("boss-item", bossFloor))
	if bossFloor <= RecruitUntil && len(s.Heroes)+s.Recruits < len(Catalog) {
		s.Recruits++
	}
}

func (s *State) apply(e Event) {
	s.Enc.fs = nil // esquadrão, pontos ou itens podem mudar
	switch e.Kind {
	case "start":
		var d struct{ Hero string }
		json.Unmarshal(e.Data, &d)
		if hi, ok := info(d.Hero); ok && !s.Started {
			s.Started = true
			s.addHero(hi, 1)
		}
	case "recruit":
		var d struct{ Hero string }
		json.Unmarshal(e.Data, &d)
		hi, ok := info(d.Hero)
		if !ok || s.Recruits == 0 || s.Hero(d.Hero) != nil {
			return
		}
		level := s.Heroes[0].Level
		for _, h := range s.Heroes {
			level = min(level, h.Level)
		}
		s.Recruits--
		s.addHero(hi, level)
		s.addLog(LogEntry{At: e.At, Kind: "recruit", Hero: hi.ID})
	case "squad":
		var d struct{ Heroes []string }
		json.Unmarshal(e.Data, &d)
		if len(d.Heroes) == 0 || len(d.Heroes) > SquadSize {
			return
		}
		seen := map[string]bool{}
		for _, id := range d.Heroes {
			if s.Hero(id) == nil || seen[id] {
				return
			}
			seen[id] = true
		}
		s.Squad = d.Heroes
	case "challenge":
		if s.Training && s.LastFight != s.in.Cal.DayOf(e.At) {
			s.fight(e.At, s.Floor+1)
		}
	default:
		s.applyItemEvent(e)
	}
}

func (s *State) addHero(hi HeroInfo, level int) {
	h := &HeroState{ID: hi.ID, Name: hi.Name, Class: hi.Class, Spec: StatKeys[hi.Spec], spec: hi.Spec, Level: level, Equip: map[Slot]int{}}
	s.Heroes = append(s.Heroes, h)
	if len(s.Squad) < SquadSize {
		s.Squad = append(s.Squad, h.ID)
	}
}

// finish preenche os campos derivados para as telas.
func (s *State) finish() {
	s.Fatigue = s.fatigueAt(s.in.Now)
	inSquad := map[string]bool{}
	for _, id := range s.Squad {
		inSquad[id] = true
	}
	total := 0
	for _, h := range s.Heroes {
		h.Points = s.points(h)
		h.XPNext = XPToNext(h.Level)
		h.Power = int(math.Round(float64(sumPoints(h.Points)*PowerPerPoint) * (1 - float64(s.Fatigue)/100)))
		h.InSquad = inSquad[h.ID]
		for i := range h.Costs {
			h.Costs[i] = BuyCost(h.Bought[i])
		}
		if h.InSquad {
			total += sumPoints(h.Points)
		}
	}
	s.Power = int(math.Round(float64(total*PowerPerPoint) * (1 - float64(s.Fatigue)/100)))
	s.NextBoss = (s.Floor/BossEvery + 1) * BossEvery
	if s.Training {
		s.NextBoss = s.Floor + 1
	}
	s.Recommended = BossOf(s.NextBoss).Recommended
	for _, it := range s.Items {
		it.Value = sellValue(it)
		it.Pts = it.Points()
	}
	s.CanFight = s.Training && s.LastFight != s.in.Cal.DayOf(s.in.Now)
	if s.Started {
		fs := s.fighters(s.in.Now)
		for i, h := range s.squad() {
			s.Enc.MaxLife[h.ID] = fs[i].Life
			if _, ok := s.Enc.Life[h.ID]; !ok {
				s.Enc.Life[h.ID] = fs[i].Life
			}
		}
	}
}
