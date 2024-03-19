package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type Registro struct {
	ID             int64     `db:"id"`
	WaID           int64     `db:"wa_id"`
	Tipo           string    `db:"tipo"`
	UsuarioID      string    `db:"usuario_id"`
	EspecieID      int64     `db:"especie_id"`
	Autor          string    `db:"autor"`
	Por            string    `db:"por"`
	Perfil         string    `db:"perfil"`
	Data           time.Time `db:"data"`
	DataPublicacao time.Time `db:"data_publicacao"`
	Questionada    bool      `db:"questionada"`
	Local          string    `db:"local"`
	LocalNome      string    `db:"local_nome"`
	LocalTipo      string    `db:"local_tipo"`
	MunicipioID    int64     `db:"municipio_id"`
	Comentarios    int64     `db:"comentarios"`
	Likes          int64     `db:"likes"`
	Views          int64     `db:"views"`
	Grande         string    `db:"grande"`
	Enviado        string    `db:"enviado"`
	Link           string    `db:"link"`
	Acao           string    `db:"acao"`
	Scrapped       bool      `db:"scrapped"`
	Assunto        string    `db:"assunto"`
	Sexo           string    `db:"sexo"`
	Idade          string    `db:"idade"`
	Observacoes    string    `db:"observacoes"`
	Camera         string    `db:"camera"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
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
	query := "INSERT INTO registros_filtered_2(id, wa_id, tipo, usuario_id, especie_id, autor, por, perfil, data, questionada, local, municipio_id, comentarios, likes, views, grande, enviado, link, data_publicacao, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21);"
	_, err := tx.Exec(query, r.ID, r.WaID, r.Tipo, r.UsuarioID, r.EspecieID, r.Autor, r.Por, r.Perfil, r.Data, r.Questionada, r.Local, r.MunicipioID, r.Comentarios, r.Likes, r.Views, r.Grande, r.Enviado, r.Link, r.DataPublicacao, time.Now(), time.Now())
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

func GetNoLocalRegistros(ctx context.Context, db *sqlx.DB, limit int) ([]*Registro, error) {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Commit()
	return GetNoLocalRegistrosTxx(tx, limit)
}
func GetNoLocalRegistrosTxx(tx *sqlx.Tx, limit int) ([]*Registro, error) {
	var registros []*Registro
	query := "SELECT r.id, r.wa_id FROM registros_filtered r WHERE r.scrapped IS NULL ORDER BY r.wa_id DESC LIMIT $1;"
	rows, err := tx.Query(query, limit)
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

func GetFilteredRegistros(ctx context.Context, db *sqlx.DB, limit int) ([]*Registro, error) {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Commit()
	return GetFilteredRegistrosTxx(tx, limit)
}

func GetFilteredRegistrosTxx(tx *sqlx.Tx, limit int) ([]*Registro, error) {
	var registros []*Registro
	query := `SELECT
		r.id,
		r.wa_id,
		r.tipo,
		r.usuario_id,
		r.especie_id,
		r.autor,
		r.por,
		r.perfil,
		r.data,
		r.questionada,
		r.local,
		r.local_nome,
		r.local_tipo,
		r.municipio_id,
		r.comentarios,
		r.likes,
		r.views,
		r.grande,
		r.enviado,
		r.link,
		r.acao,
		r.scrapped,
		r.assunto,
		r.sexo,
		r.idade,
		r.observacoes,
		r.camera,
		r.created_at,
		r.updated_at
	FROM registros_filtered r
	WHERE r.scrapped = FALSE
	LIMIT $1;`
	rows, err := tx.Query(query, limit)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		registro := new(Registro)
		rows.Scan(
			&registro.ID,
			&registro.WaID,
			&registro.Tipo,
			&registro.UsuarioID,
			&registro.EspecieID,
			&registro.Autor,
			&registro.Por,
			&registro.Perfil,
			&registro.Data,
			&registro.Questionada,
			&registro.Local,
			&registro.LocalNome,
			&registro.LocalTipo,
			&registro.MunicipioID,
			&registro.Comentarios,
			&registro.Likes,
			&registro.Views,
			&registro.Grande,
			&registro.Enviado,
			&registro.Link,
			&registro.Acao,
			&registro.Scrapped,
			&registro.Assunto,
			&registro.Sexo,
			&registro.Idade,
			&registro.Observacoes,
			&registro.Camera,
			&registro.CreatedAt,
			&registro.UpdatedAt)
		registros = append(registros, registro)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return registros, nil
}

func SetScrappedTxx(tx *sqlx.Tx, id int64) error {
	query := "UPDATE registros_filtered SET scrapped=1 WHERE id=$1"
	_, err := tx.Exec(query, id)
	return err
}

func AddAdditionalInfoTxx(tx *sqlx.Tx, r *Registro) error {
	query := `UPDATE registros_filtered SET 
		assunto = $1,
		acao = $2,
		sexo = $3,
		idade = $4,
		observacoes = $5,
		camera = $6,
		local_nome = $7,
		local_tipo = $8,
		scrapped = $9
	WHERE id = $10;
	`
	_, err := tx.Exec(query, r.Assunto, r.Acao, r.Sexo, r.Idade, r.Observacoes, r.Camera, r.LocalNome, r.LocalTipo, r.Scrapped, r.ID)
	return err
}

func AddADataPublicacaoInfoTxx(tx *sqlx.Tx, r *Registro) error {
	query := `UPDATE registros_filtered SET data_publicacao = $1, scrapped = TRUE WHERE id = $2;
	`
	_, err := tx.Exec(query, r.DataPublicacao, r.ID)
	return err
}
