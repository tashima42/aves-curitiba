package database

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Especie struct {
	ID        int64     `db:"id"`
	WaID      int64     `db:"wa_id"`
	Nome      string    `db:"nome"`
	Nvt       string    `db:"nvt"`
	WikiID    string    `db:"wiki_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func CreateEspecieTxx(tx *sqlx.Tx, e *Especie) error {
	query := "INSERT INTO especies(wa_id, nome, nvt, wiki_id, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7);"
	_, err := tx.Exec(query, e.WaID, e.Nome, e.Nvt, e.WikiID, time.Now(), time.Now())
	return err
}
