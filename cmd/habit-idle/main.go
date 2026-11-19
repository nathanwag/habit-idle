package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"habit-idle/internal/db"
	"habit-idle/internal/game"
	"habit-idle/internal/server"
	"habit-idle/web"
	_ "time/tzdata"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "balance" {
		game.Table(os.Stdout, 50)
		return
	}
	d, err := db.Open(db.DefaultPath())
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()
	if len(os.Args) > 1 && os.Args[1] == "demo" {
		if err := demo(d, time.Now()); err != nil {
			log.Fatal(err)
		}
		return
	}
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
