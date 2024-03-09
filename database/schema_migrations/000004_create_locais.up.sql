CREATE TABLE IF NOT EXISTS locais (
  id INTEGER PRIMARY KEY,
  'tipo' TEXT NOT NULL,
  'nome' TEXT NOT NULL,
  'created_at' DATE NOT NULL,
  'updated_at' DATE NOT NULL
);
