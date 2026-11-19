package server

import (
	"database/sql"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"time"
)

type Deps struct {
	DB  *sql.DB
	Web fs.FS
	Now func() time.Time
}

type srv struct{ Deps }

func New(d Deps) http.Handler {
	if d.Now == nil {
		d.Now = time.Now
	}
	s := &srv{d}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("GET /api/settings", s.getSettings)
	mux.HandleFunc("PUT /api/settings", s.putSettings)
	mux.HandleFunc("GET /api/today", s.today)
	mux.HandleFunc("GET /api/habits", s.listHabits)
	mux.HandleFunc("POST /api/habits", s.createHabit)
	mux.HandleFunc("GET /api/habits/{id}", s.habitDetail)
	mux.HandleFunc("PUT /api/habits/{id}", s.updateHabit)
	mux.HandleFunc("DELETE /api/habits/{id}", s.deleteHabit)
	mux.HandleFunc("PUT /api/checks", s.putCheck)
	if d.Web != nil {
		mux.Handle("/", http.FileServerFS(d.Web))
	}
	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func badRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
}

func fail(w http.ResponseWriter, err error) {
	log.Print(err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro interno"})
}
