CREATE TABLE IF NOT EXISTS especies (
  id INTEGER PRIMARY KEY,
  'wa_id' TEXT NOT NULL,
  'nome' TEXT NOT NULL,
  'nvt' TEXT NOT NULL,
  'wiki_id' TEXT NOT NULL,
  'created_at' DATE NOT NULL,
  'updated_at' DATE NOT NULL
);
