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
}

type State struct {
	Started     bool         `json:"started"`
	Heroes      []*HeroState `json:"heroes"` // na ordem em que entraram
	Squad       []string     `json:"squad"`
	Floor       int          `json:"floor"`    // andar sendo explorado
	Progress    float64      `json:"progress"` // 0 a 1 do andar atual
	FloorMins   float64      `json:"floorMins"`
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
	s := &State{in: in, Floor: 1, Heroes: []*HeroState{}, Squad: []string{}, Items: []*Item{}, Log: []LogEntry{}, passes: map[int]int{}}

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

func floorLife(n int) float64   { return FloorLife * math.Pow(FloorGrowth, float64(n-1)) }
func floorThreat(n int) float64 { return FloorThreat * math.Pow(FloorGrowth, float64(n-1)) }

// floorMinutes: o dano decide a velocidade; esquadrão frágil demais descansa entre as lutas.
func floorMinutes(n int, fs []Fighter) float64 {
	dps, ehp := 0.0, 0.0
	for _, f := range fs {
		dps += f.DPS()
		ehp += f.EffectiveLife()
	}
	if dps <= 0 {
		return math.Inf(1)
	}
	return floorLife(n) / dps * math.Max(1, floorThreat(n)/ehp)
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

func (s *State) run(a, b time.Time) {
	if !s.Started || len(s.Squad) == 0 {
		return
	}
	t := a
	for {
		total := floorMinutes(s.Floor, s.fighters(t))
		need := total * (1 - s.Progress)
		left := b.Sub(t).Minutes()
		if left < need {
			s.Progress += left / total
			return
		}
		t = t.Add(time.Duration(need * float64(time.Minute)))
		s.clearFloor(t)
	}
}

func (s *State) clearFloor(t time.Time) {
	n := s.Floor
	s.Progress = 0
	s.passes[n]++
	coins := FloorCoins + FloorCoinsN*n
	xp := FloorXP + FloorXPN*n
	s.Coins += coins
	s.addLog(LogEntry{At: t, Kind: "floor", Floor: n, Coins: coins, XP: xp})
	s.giveXP(t, xp)
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
	s.Progress = 0
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
	if s.Started && len(s.Squad) > 0 {
		s.FloorMins = floorMinutes(s.Floor, s.fighters(s.in.Now))
	}
}
