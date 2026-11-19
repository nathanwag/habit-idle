package main

import (
	"log"
	"net/http"
	"os"

	"habit-idle/internal/db"
	"habit-idle/internal/server"
	"habit-idle/web"
)

func main() {
	d, err := db.Open(db.DefaultPath())
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()
	addr := env("HOST", "127.0.0.1") + ":" + env("PORT", "3002")
	log.Printf("habit-idle em http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, server.New(server.Deps{DB: d, Web: web.FS})))
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
