package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"habit-idle/internal/db"
)

type app struct {
	t   *testing.T
	h   http.Handler
	now time.Time
}

func newApp(t *testing.T, now string) *app {
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	a := &app{t: t}
	a.setNow(now)
	a.h = New(Deps{DB: d, Now: func() time.Time { return a.now }})
	return a
}

func (a *app) setNow(s string) {
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	t, err := time.ParseInLocation("2006-01-02 15:04", s, loc)
	if err != nil {
		a.t.Fatal(err)
	}
	a.now = t
}

func (a *app) do(method, path string, body any, out any) int {
	a.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	rec := httptest.NewRecorder()
	a.h.ServeHTTP(rec, httptest.NewRequest(method, path, &buf))
	if out != nil && rec.Code < 300 {
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			a.t.Fatalf("%s %s: %v (%s)", method, path, err, rec.Body.String())
		}
	}
	return rec.Code
}

type todayResp struct {
	Day   string `json:"day"`
	Items []struct {
		Habit struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"habit"`
		Done bool `json:"done"`
	} `json:"items"`
}

func TestCriarHabitoMarcarEDesmarcarNoHoje(t *testing.T) {
	a := newApp(t, "2026-11-19 09:00")
	var created struct{ ID int }
	if code := a.do("POST", "/api/habits", map[string]any{
		"name": "Ler 20 páginas", "effort": 120, "schedule": map[string]any{"kind": "daily"},
	}, &created); code != 201 {
		t.Fatalf("POST = %d", code)
	}

	var today todayResp
	a.do("GET", "/api/today", nil, &today)
	if today.Day != "2026-11-19" || len(today.Items) != 1 || today.Items[0].Done {
		t.Fatalf("hoje: %+v", today)
	}

	if code := a.do("PUT", "/api/checks", map[string]any{"habitId": created.ID, "day": "2026-11-19", "done": true}, nil); code != 204 {
		t.Fatalf("PUT check = %d", code)
	}
	a.do("GET", "/api/today", nil, &today)
	if !today.Items[0].Done {
		t.Fatal("não ficou feito")
	}

	a.do("PUT", "/api/checks", map[string]any{"habitId": created.ID, "day": "2026-11-19", "done": false}, nil)
	a.do("GET", "/api/today", nil, &today)
	if today.Items[0].Done {
		t.Fatal("não desmarcou")
	}
}

func TestViradaDoDiaPadraoSeisHorasEConfiguravel(t *testing.T) {
	a := newApp(t, "2026-11-20 05:30")
	var today todayResp
	a.do("GET", "/api/today", nil, &today)
	if today.Day != "2026-11-19" {
		t.Fatalf("antes das 06:00: %s", today.Day)
	}
	if code := a.do("PUT", "/api/settings", map[string]any{"dayStart": "05:00"}, nil); code != 204 {
		t.Fatalf("PUT settings = %d", code)
	}
	a.do("GET", "/api/today", nil, &today)
	if today.Day != "2026-11-20" {
		t.Fatalf("com virada às 05:00: %s", today.Day)
	}
}

func TestEditarFrequenciaNaoMudaOPassadoEArquivarTiraDoHoje(t *testing.T) {
	a := newApp(t, "2026-11-16 09:00") // segunda
	var created struct{ ID int }
	a.do("POST", "/api/habits", map[string]any{"name": "Correr", "effort": 180, "schedule": map[string]any{"kind": "daily"}}, &created)
	a.do("PUT", "/api/checks", map[string]any{"habitId": created.ID, "day": "2026-11-16", "done": true}, nil)

	a.setNow("2026-11-17 09:00") // terça
	path := "/api/habits/" + itoa(created.ID)
	a.do("PUT", path, map[string]any{"name": "Correr", "effort": 180, "schedule": map[string]any{"kind": "days", "days": []int{3}}}, nil)
	var today todayResp
	a.do("GET", "/api/today", nil, &today)
	if len(today.Items) != 0 {
		t.Fatalf("terça não devia ter: %+v", today)
	}

	var detail struct {
		Stats struct {
			Total int `json:"total"`
		} `json:"stats"`
	}
	a.do("GET", path, nil, &detail)
	if detail.Stats.Total != 1 {
		t.Fatalf("detalhe: %+v", detail)
	}

	a.setNow("2026-11-18 09:00") // quarta
	a.do("GET", "/api/today", nil, &today)
	if len(today.Items) != 1 {
		t.Fatalf("quarta devia ter: %+v", today)
	}
	a.do("PUT", path, map[string]any{"name": "Correr", "effort": 180, "schedule": map[string]any{"kind": "days", "days": []int{3}}, "archived": true}, nil)
	a.do("GET", "/api/today", nil, &today)
	if len(today.Items) != 0 {
		t.Fatalf("arquivado continua: %+v", today)
	}
}

func TestHabitoInvalidoDa400(t *testing.T) {
	a := newApp(t, "2026-11-19 09:00")
	bad := []map[string]any{
		{"name": "", "effort": 120, "schedule": map[string]any{"kind": "daily"}},
		{"name": "x", "effort": 90, "schedule": map[string]any{"kind": "daily"}},
		{"name": "x", "effort": 60, "schedule": map[string]any{"kind": "days", "days": []int{}}},
		{"name": "x", "effort": 60, "schedule": map[string]any{"kind": "weekly", "times": 0}},
	}
	for _, b := range bad {
		if code := a.do("POST", "/api/habits", b, nil); code != 400 {
			t.Errorf("%v: %d", b, code)
		}
	}
}
