package database

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Autor struct {
	ID           int64     `db:"id"`
	Nome         string    `db:"nome"`
	Perfil       string    `db:"perfil"`
	Cidade       string    `db:"cidade"`
	DataCadastro time.Time `db:"data_cadastro"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func CreateAutorTxx(tx *sqlx.Tx, a *Autor) error {
	query := "INSERT OR IGNORE INTO autores(nome, perfil, cidade, data_cadastro, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6);"
	_, err := tx.Exec(query, a.Nome, a.Perfil, a.Cidade, a.DataCadastro, time.Now(), time.Now())
	return err
}

func GetAutorInfoByPerfilTxx(tx *sqlx.Tx, perfil string) (*Autor, error) {
	var a Autor
	query := "SELECT autor as nome, perfil FROM registros WHERE perfil = $1 LIMIT 1;"
	err := tx.Get(&a, query, perfil)
	if err != nil {
		return nil, err
	}
	a.Perfil = perfil
	return &a, nil
}
