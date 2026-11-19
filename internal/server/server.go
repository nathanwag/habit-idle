package server

import (
	"database/sql"
	"encoding/json"
	"io/fs"
	"net/http"
	"time"
)

type Deps struct {
	DB  *sql.DB
	Web fs.FS
	Now func() time.Time
}

func New(d Deps) http.Handler {
	if d.Now == nil {
		d.Now = time.Now
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
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
