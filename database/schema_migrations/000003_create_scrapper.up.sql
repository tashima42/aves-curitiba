CREATE TABLE IF NOT EXISTS scrapper (
  id INTEGER PRIMARY KEY,
  'total' INTEGER NOT NULL,
  'per_page' INTEGER NOT NULL,
  'current_page' INTEGER NOT NULL,
  'created_at' DATE NOT NULL,
  'updated_at' DATE NOT NULL
)