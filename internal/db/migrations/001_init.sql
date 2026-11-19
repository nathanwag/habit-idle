CREATE TABLE setting (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

CREATE TABLE habit (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  effort INTEGER NOT NULL,          -- minutos de fôlego por check: 60, 120 ou 180
  created_day TEXT NOT NULL,
  ord INTEGER NOT NULL DEFAULT 0
);

-- Frequência e arquivamento valem a partir de from_day; o passado segue a versão antiga.
CREATE TABLE habit_version (
  habit_id INTEGER NOT NULL REFERENCES habit(id) ON DELETE CASCADE,
  from_day TEXT NOT NULL,
  schedule TEXT NOT NULL,           -- JSON: {"kind":"daily"} | {"kind":"days","days":[1,3]} | {"kind":"weekly","times":2}
  archived INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (habit_id, from_day)
);

-- Registro de marcar/desmarcar, nunca apagado: o fôlego depende da hora de cada toque.
CREATE TABLE check_log (
  id INTEGER PRIMARY KEY,
  habit_id INTEGER NOT NULL REFERENCES habit(id) ON DELETE CASCADE,
  day TEXT NOT NULL,
  at TEXT NOT NULL,
  done INTEGER NOT NULL
);
CREATE INDEX check_log_habit_day ON check_log(habit_id, day, id);

-- Escolhas do jogador (start, recruit, squad, buy, equip, sell, refine, challenge).
CREATE TABLE event (
  id INTEGER PRIMARY KEY,
  at TEXT NOT NULL,
  kind TEXT NOT NULL,
  data TEXT NOT NULL DEFAULT '{}'
);
