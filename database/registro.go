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
	Local       string    `db:"local"`
	LocalNome   string    `db:"local_nome"`
	LocalTipo   string    `db:"local_tipo"`
	MunicipioID int64     `db:"municipio_id"`
	Comentarios int64     `db:"comentarios"`
	Likes       int64     `db:"likes"`
	Views       int64     `db:"views"`
	Grande      string    `db:"grande"`
	Enviado     string    `db:"enviado"`
	Link        string    `db:"link"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type RegistroCustom struct {
	ID        int64
	Data      string
	Autor     string
	Especie   string
	LocalNome string
	LocalTipo string
}

func CreateRegistroTxx(tx *sqlx.Tx, r *Registro) error {
	query := "INSERT INTO registros_filtered_2(id, wa_id, tipo, usuario_id, especie_id, autor, por, perfil, data, questionada, local, municipio_id, comentarios, likes, views, grande, enviado, link, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20);"
	_, err := tx.Exec(query, r.ID, r.WaID, r.Tipo, r.UsuarioID, r.EspecieID, r.Autor, r.Por, r.Perfil, r.Data, r.Questionada, r.Local, r.MunicipioID, r.Comentarios, r.Likes, r.Views, r.Grande, r.Enviado, r.Link, time.Now(), time.Now())
	return err
}

func GetRegistrosTxx(tx *sqlx.Tx) ([]*Registro, error) {
	var registros []*Registro
	query := "SELECT id, wa_id, tipo, usuario_id, especie_id, autor, por, perfil, \"data\", questionada, \"local\", local_nome, local_tipo, municipio_id, comentarios, likes, views, grande, enviado, link, created_at, updated_at FROM registros_filtered;"
	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		registro := new(Registro)
		rows.Scan(&registro.ID, &registro.WaID, &registro.Tipo, &registro.UsuarioID, &registro.EspecieID, &registro.Autor, &registro.Por, &registro.Perfil, &registro.Data, &registro.Questionada, &registro.Local, &registro.LocalNome, &registro.LocalTipo, &registro.MunicipioID, &registro.Comentarios, &registro.Likes, &registro.Views, &registro.Grande, &registro.Enviado, &registro.Link, &registro.CreatedAt, &registro.UpdatedAt)
		registros = append(registros, registro)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return registros, nil
}

func GetFilteredRegistrosTxx(tx *sqlx.Tx) ([]*RegistroCustom, error) {
	var registros []*RegistroCustom
	query := "SELECT r.id, r.\"data\", r.autor, e.nvt FROM registros_filtered r INNER JOIN especies e ON r.especie_id = e.id WHERE r.\"data\" BETWEEN  \"2022-01-01\" AND \"2023-12-31\";"
	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		registro := new(RegistroCustom)
		rows.Scan(&registro.ID, &registro.Data, &registro.Autor, &registro.Especie)
		registros = append(registros, registro)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return registros, nil
}

func UpdateLocalTxx(tx *sqlx.Tx, r *RegistroCustom) error {
	query := "UPDATE registros_filtered SET local_nome = $1, local_tipo = $2 WHERE id = $3;"
	_, err := tx.Exec(query, r.LocalNome, r.LocalTipo, r.ID)
	return err
}

func GetNoLocalRegistrosTxx(tx *sqlx.Tx, limit int, skip int) ([]*Registro, error) {
	var registros []*Registro
	query := "SELECT r.id, r.wa_id FROM registros_filtered r WHERE r.local_nome IS NULL ORDER BY r.wa_id DESC LIMIT $1, $2;"
	// query := "SELECT r.id, r.wa_id FROM registros_filtered r WHERE r.wa_id = 5405787;"
	rows, err := tx.Query(query, skip, limit)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		registro := new(Registro)
		rows.Scan(&registro.ID, &registro.WaID)
		registros = append(registros, registro)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return registros, nil
}
