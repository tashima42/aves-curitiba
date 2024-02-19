CREATE TABLE IF NOT EXISTS registros (
  id INTEGER PRIMARY KEY,
  'wa_id' TEXT NOT NULL UNIQUE,
  'tipo' TEXT NOT NULL,
  'usuario_id' TEXT NOT NULL,
  'especie_id' INTEGER NOT NULL,
  'autor' TEXT NOT NULL,
  'por' TEXT NOT NULL,
  'perfil' TEXT NOT NULL,
  'data' DATE NOT NULL,
  'questionada' INTEGER NOT NULL,
  'local' TEXT NOT NULL,
  'municipio_id' TEXT NOT NULL,
  'comentarios' INTEGER NOT NULL,
  'likes' INTEGER NOT NULL,
  'views' INTEGER NOT NULL,
  'grande' TEXT NOT NULL,
  'enviado' TEXT NOT NULL,
  'link' TEXT NOT NULL,
  'created_at' DATE NOT NULL,
  'updated_at' DATE NOT NULL,

  FOREIGN KEY('especie_id') REFERENCES 'especies'('id')
);
