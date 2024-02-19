package database

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Especie struct {
	ID        int64     `db:"id"`
	WaID      string    `db:"wa_id"`
	Nome      string    `db:"nome"`
	Nvt       string    `db:"nvt"`
	WikiID    string    `db:"wiki_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func CreateEspecieTxx(tx *sqlx.Tx, e *Especie) (int64, error) {
	query := "INSERT INTO especies(wa_id, nome, nvt, wiki_id, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6);"
	res, err := tx.Exec(query, e.WaID, e.Nome, e.Nvt, e.WikiID, time.Now(), time.Now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetEspecieByWaIDTxx(tx *sqlx.Tx, waID string) (*Especie, error) {
	var e Especie
	query := "SELECT id, wa_id, nome, nvt, wiki_id, created_at, updated_at FROM especies WHERE wa_id=$1 LIMIT 1;"
	err := tx.Get(&e, query, waID)
	if err != nil {
		return nil, err
	}
	return &e, nil
}
