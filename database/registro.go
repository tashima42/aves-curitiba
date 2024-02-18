package database

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Registro struct {
	ID          int64     `db:"id"`
	WaID        int64     `db:"wa_id"`
	Tipo        string    `db:"tipo"`
	UsuarioID   string    `db:"usuario_id"`
	EspecieID   int64     `db:"especie_id"`
	Autor       string    `db:"autor"`
	Por         string    `db:"por"`
	Perfil      string    `db:"perfil"`
	Data        time.Time `db:"data"`
	Questionada bool      `db:"questionada"`
	Local       bool      `db:"local"`
	MunicipioID bool      `db:"municipio_id"`
	Comentarios int64     `db:"comentarios"`
	Likes       int64     `db:"likes"`
	Views       int64     `db:"views"`
	Grande      string    `db:"grande"`
	Enviado     string    `db:"enviado"`
	Link        string    `db:"link"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func CreateRegistroTxx(tx *sqlx.Tx, r *Registro) error {
	query := "INSERT INTO registros(wa_id, tipo, usuario_id, especie_id, autor, por, perfil, data, questionada, local, municipio_id, comentarios, likes, views, grande, enviado, link, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7);"
	_, err := tx.Exec(query, r.WaID, r.Tipo, r.UsuarioID, r.EspecieID, r.Autor, r.Por, r.Perfil, r.Data, r.Questionada, r.Local, r.MunicipioID, r.Comentarios, r.Likes, r.Views, r.Grande, r.Enviado, r.Link, time.Now(), time.Now())
	return err
}
