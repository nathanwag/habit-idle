package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

func DefaultPath() string {
	if p := os.Getenv("HABIT_IDLE_DB"); p != "" {
		return p
	}
	return "./data/habit-idle.db"
}

// Open abre o banco e aplica as migrations pendentes.
func Open(path string) (*sql.DB, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
	}
	d, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	// Uma conexão só: evita disputa de escrita e mantém vivo o banco ":memory:" dos testes.
	d.SetMaxOpenConns(1)
	d.SetConnMaxIdleTime(0)
	d.SetConnMaxLifetime(0)
	if err := migrate(d); err != nil {
		d.Close()
		return nil, fmt.Errorf("migrations: %w", err)
	}
	return d, nil
}

func migrate(d *sql.DB) error {
	if _, err := d.Exec("CREATE TABLE IF NOT EXISTS schema_migration (name TEXT PRIMARY KEY, applied_at TEXT NOT NULL)"); err != nil {
		return err
	}
	applied := map[string]bool{}
	rows, err := d.Query("SELECT name FROM schema_migration")
	if err != nil {
		return err
	}
	for rows.Next() {
		var n string
		rows.Scan(&n)
		applied[n] = true
	}
	rows.Close()
	files, err := fs.Glob(migrations, "migrations/*.sql")
	if err != nil {
		return err
	}
	slices.Sort(files)
	for _, file := range files {
		name := filepath.Base(file)
		if applied[name] {
			continue
		}
		text, _ := migrations.ReadFile(file)
		err := InTx(d, func(tx *sql.Tx) error {
			if _, err := tx.Exec(string(text)); err != nil {
				return err
			}
			_, err := tx.Exec("INSERT INTO schema_migration VALUES (?, ?)", name, time.Now().UTC().Format(time.RFC3339))
			return err
		})
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

func InTx(d *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
