package game

import (
	"encoding/json"
	"math"
	"time"
)

type Slot string

const (
	Weapon Slot = "arma"
	Armor  Slot = "armadura"
	Amulet Slot = "amuleto"
)

var Rarities = []string{"comum", "rara", "épica", "lendária"}

type Extra struct {
	Stat string `json:"stat"`
	Pts  int    `json:"pts"`
}

type Item struct {
	ID     int         `json:"id"`
	Name   string      `json:"name"`
	Slot   Slot        `json:"slot"`
	Rarity int         `json:"rarity"`
	Floor  int         `json:"floor"`
	Main   string      `json:"main"`
	Extras []Extra     `json:"extras"`
	Refine int         `json:"refine"`
	Hero   string      `json:"hero,omitempty"` // quem está usando
	Value  int         `json:"value"`          // moedas ao vender
	Pts    [nStats]int `json:"pts"`
}

// Dois nomes por tipo e raridade: itens iguais (mesmo nome e raridade) se refinam.
var itemNames = map[Slot][4][2]string{
	Weapon: {{"Espada curta", "Cajado de pinho"}, {"Espada rúnica", "Cajado de ônix"}, {"Lâmina do eclipse", "Cetro estelar"}, {"Fúria do dragão", "Bastão do arquimago"}},
	Armor:  {{"Gibão de couro", "Cota de malha"}, {"Armadura de placas", "Manto reforçado"}, {"Couraça de obsidiana", "Manto das marés"}, {"Égide do titã", "Véu da aurora"}},
	Amulet: {{"Pingente de osso", "Anel de cobre"}, {"Amuleto de jade", "Anel de prata"}, {"Olho do basilisco", "Anel do vento"}, {"Coração da fênix", "Selo do destino"}},
}

func rarityMult(r int) float64 { return []float64{1, 2, 3.5, 6}[r] }

func (it *Item) base() int {
	return int(math.Round(ItemBase * rarityMult(it.Rarity) * (1 + float64(it.Floor)/ItemFloorDiv)))
}

// Points do item, já com o refino.
func (it *Item) Points() [nStats]int {
	var p [nStats]int
	k := 1 + ItemRefineStep*float64(it.Refine)
	main, _ := ParseStat(it.Main)
	p[main] += int(math.Round(float64(it.base()) * k))
	for _, e := range it.Extras {
		st, _ := ParseStat(e.Stat)
		p[st] += int(math.Round(float64(e.Pts) * k))
	}
	return p
}

func rollItem(id int, kind string, floor int, sd uint64) *Item {
	r := rng(sd)
	it := &Item{ID: id, Floor: floor}
	x := r.Float64()
	if kind == "boss" {
		it.Rarity = 2
		if x < LegendaryChance {
			it.Rarity = 3
		}
	} else if x < RareChance {
		it.Rarity = 1
	}
	it.Slot = []Slot{Weapon, Armor, Amulet}[r.IntN(3)]
	it.Name = itemNames[it.Slot][it.Rarity][r.IntN(2)]
	switch it.Slot {
	case Weapon:
		it.Main = StatKeys[ATK]
	case Armor:
		it.Main = StatKeys[DEF]
	default:
		it.Main = StatKeys[[]Stat{SOR, VEL}[r.IntN(2)]]
	}
	n := []int{0, 1, 1 + r.IntN(2), 2}[it.Rarity]
	used := map[string]bool{it.Main: true}
	for len(it.Extras) < n {
		st := StatKeys[r.IntN(int(nStats))]
		if used[st] {
			continue
		}
		used[st] = true
		it.Extras = append(it.Extras, Extra{Stat: st, Pts: int(math.Round(float64(it.base()) * ItemExtraShare))})
	}
	return it
}

func (s *State) chest(t time.Time, kind string, floor int, sd uint64) {
	s.nextItem++
	it := rollItem(s.nextItem, kind, floor, sd)
	s.Items = append(s.Items, it)
	s.addLog(LogEntry{At: t, Kind: "chest", Floor: floor, Item: it.ID})
}

func (s *State) item(id int) *Item {
	for _, it := range s.Items {
		if it.ID == id {
			return it
		}
	}
	return nil
}

func (s *State) dropItem(id int) {
	out := s.Items[:0]
	for _, it := range s.Items {
		if it.ID != id {
			out = append(out, it)
		}
	}
	s.Items = out
}

// BuyCost: preço do próximo ponto comprado num status de um herói.
func BuyCost(bought int) int {
	// O épsilon evita que 57,5 vire 57,4999… e arredonde para baixo.
	return int(math.Round(BuyBase*math.Pow(BuyGrowth, float64(bought)) + 1e-9))
}

func sellValue(it *Item) int {
	return int(math.Round(SellBase * rarityMult(it.Rarity) * (1 + float64(it.Floor)/ItemFloorDiv)))
}

func (s *State) unequip(it *Item) {
	if h := s.Hero(it.Hero); h != nil {
		delete(h.Equip, it.Slot)
	}
	it.Hero = ""
}

func (s *State) applyItemEvent(e Event) {
	var d struct {
		Hero string
		Stat string
		Item int
		With int
	}
	json.Unmarshal(e.Data, &d)
	switch e.Kind {
	case "buy":
		h := s.Hero(d.Hero)
		st, ok := ParseStat(d.Stat)
		if h == nil || !ok {
			return
		}
		cost := BuyCost(h.Bought[st])
		if s.Coins < cost {
			return
		}
		s.Coins -= cost
		h.Bought[st]++
	case "equip":
		h, it := s.Hero(d.Hero), s.item(d.Item)
		if h == nil || it == nil {
			return
		}
		s.unequip(it)
		if old := s.item(h.Equip[it.Slot]); old != nil {
			s.unequip(old)
		}
		h.Equip[it.Slot] = it.ID
		it.Hero = h.ID
	case "unequip":
		if it := s.item(d.Item); it != nil {
			s.unequip(it)
		}
	case "sell":
		it := s.item(d.Item)
		if it == nil {
			return
		}
		s.unequip(it)
		s.Coins += sellValue(it)
		s.dropItem(it.ID)
	case "refine":
		a, b := s.item(d.Item), s.item(d.With)
		if a == nil || b == nil || a == b || a.Name != b.Name || a.Rarity != b.Rarity || a.Refine >= RefineMax {
			return
		}
		s.unequip(b)
		a.Refine++
		s.dropItem(b.ID)
	}
}
