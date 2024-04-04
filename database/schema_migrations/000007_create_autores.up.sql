CREATE TABLE IF NOT EXISTS autores (
  id INTEGER PRIMARY KEY,
  nome TEXT NOT NULL,
  perfil TEXT NOT NULL,
  cidade TEXT,
  data_cadastro DATE NOT NULL,
  created_at DATE NOT NULL,
  updated_at DATE NOT NULL
);
