package server

import "testing"

type stateResp struct {
	Day    string `json:"day"`
	Folego struct {
		Balance   int    `json:"balance"` // minutos
		Until     string `json:"until"`
		Exploring bool   `json:"exploring"`
	} `json:"folego"`
	Game struct {
		Started bool `json:"started"`
		Floor   int  `json:"floor"`
		Coins   int  `json:"coins"`
		Heroes  []struct {
			ID    string `json:"id"`
			Level int    `json:"level"`
		} `json:"heroes"`
	} `json:"game"`
	LastSeen string `json:"lastSeen"`
}

func TestComecarOJogoEExplorarComOFolegoDoCheck(t *testing.T) {
	a := newApp(t, "2026-11-19 08:00")
	var h struct{ ID int }
	a.do("POST", "/api/habits", map[string]any{"name": "Ler", "effort": 180, "schedule": map[string]any{"kind": "daily"}}, &h)
	a.do("POST", "/api/habits", map[string]any{"name": "Água", "effort": 60, "schedule": map[string]any{"kind": "daily"}}, nil)

	if code := a.do("POST", "/api/events", map[string]any{"kind": "start", "data": map[string]string{"hero": "oren"}}, nil); code != 201 {
		t.Fatalf("start = %d", code)
	}
	a.setNow("2026-11-19 09:00")
	a.do("PUT", "/api/checks", map[string]any{"habitId": h.ID, "day": "2026-11-19", "done": true}, nil)

	a.setNow("2026-11-19 10:30")
	var st stateResp
	a.do("GET", "/api/state", nil, &st)
	if !st.Game.Started || len(st.Game.Heroes) != 1 || st.Game.Heroes[0].ID != "oren" {
		t.Fatalf("jogo = %+v", st.Game)
	}
	// 3h às 09:00, 1h30 gasta: o andar 1 leva umas 2h, mas cada inimigo vencido já pagou.
	if st.Folego.Balance != 90 || !st.Folego.Exploring || st.Game.Floor != 1 || st.Game.Coins < 1 || st.Game.Coins >= 13 {
		t.Fatalf("estado = %+v", st)
	}
}

func TestEventoDesconhecidoDa400(t *testing.T) {
	a := newApp(t, "2026-11-19 08:00")
	if code := a.do("POST", "/api/events", map[string]any{"kind": "hack", "data": map[string]any{}}, nil); code != 400 {
		t.Fatalf("got %d", code)
	}
}

func TestVistoGuardaAUltimaVisita(t *testing.T) {
	a := newApp(t, "2026-11-19 08:00")
	if code := a.do("PUT", "/api/seen", nil, nil); code != 204 {
		t.Fatalf("seen = %d", code)
	}
	var st stateResp
	a.do("GET", "/api/state", nil, &st)
	if st.LastSeen == "" {
		t.Fatal("sem última visita")
	}
}
