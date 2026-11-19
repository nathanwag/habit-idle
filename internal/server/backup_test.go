package server

import (
	"encoding/json"
	"testing"
)

func TestBackupExportadoRestauraOMesmoJogoEmOutroBanco(t *testing.T) {
	a := newApp(t, "2026-11-19 08:00")
	var h struct{ ID int }
	a.do("POST", "/api/habits", map[string]any{"name": "Ler", "effort": 180, "schedule": map[string]any{"kind": "daily"}}, &h)
	a.do("POST", "/api/events", map[string]any{"kind": "start", "data": map[string]string{"hero": "lyra"}}, nil)
	a.do("PUT", "/api/settings", map[string]any{"dayStart": "05:30"}, nil)
	a.setNow("2026-11-19 09:00")
	a.do("PUT", "/api/checks", map[string]any{"habitId": h.ID, "day": "2026-11-19", "done": true}, nil)
	a.setNow("2026-11-19 12:00")

	var backup json.RawMessage
	if code := a.do("GET", "/api/backup", nil, &backup); code != 200 {
		t.Fatalf("export = %d", code)
	}
	var want stateResp
	a.do("GET", "/api/state", nil, &want)

	b := newApp(t, "2026-11-19 12:00")
	b.do("POST", "/api/habits", map[string]any{"name": "Some", "effort": 60, "schedule": map[string]any{"kind": "daily"}}, nil)
	if code := b.do("PUT", "/api/backup", backup, nil); code != 204 {
		t.Fatalf("import = %d", code)
	}
	var got stateResp
	b.do("GET", "/api/state", nil, &got)
	if got.Game.Floor != want.Game.Floor || got.Game.Coins != want.Game.Coins || got.Folego.Balance != want.Folego.Balance || got.Game.Heroes[0].ID != "lyra" {
		t.Fatalf("restaurado %+v, quer %+v", got, want)
	}
	var habits []struct{ Name string }
	b.do("GET", "/api/habits", nil, &habits)
	if len(habits) != 1 || habits[0].Name != "Ler" {
		t.Fatalf("hábitos = %+v", habits)
	}
}

func TestBackupInvalidoNaoApagaNada(t *testing.T) {
	a := newApp(t, "2026-11-19 08:00")
	a.do("POST", "/api/habits", map[string]any{"name": "Ler", "effort": 180, "schedule": map[string]any{"kind": "daily"}}, nil)
	if code := a.do("PUT", "/api/backup", map[string]any{"app": "outro"}, nil); code != 400 {
		t.Fatalf("got %d", code)
	}
	var habits []struct{ Name string }
	a.do("GET", "/api/habits", nil, &habits)
	if len(habits) != 1 {
		t.Fatal("apagou")
	}
}
