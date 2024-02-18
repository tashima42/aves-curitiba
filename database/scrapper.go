package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type Scrapper struct {
	ID          int64     `db:"id"`
	Total       int64     `db:"total"`
	PerPage     int64     `db:"per_page"`
	CurrentPage int64     `db:"current_page"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func GetScrapperByID(ctx context.Context, db *sqlx.DB, id int64) (*Scrapper, error) {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Commit()
	return GetScrapperByIDTxx(tx, id)
}

func SetScrapperCurrentPageByID(ctx context.Context, db *sqlx.DB, id int64, currentPage int64) error {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Commit()
	return SetScrapperCurrentPageByIDTxx(tx, id, currentPage)
}

func GetScrapperByIDTxx(tx *sqlx.Tx, id int64) (*Scrapper, error) {
	var s Scrapper
	query := "SELECT id, total, per_page, current_page, created_at, updated_at FROM scrapper WHERE id=$1 LIMIT 1;"
	err := tx.Get(&s, query, id)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func SetScrapperCurrentPageByIDTxx(tx *sqlx.Tx, id int64, currentPage int64) error {
	query := "UPDATE scrapper SET current_page = $1, updated_at = $2  WHERE id=$3;"
	_, err := tx.Exec(query, currentPage, time.Now(), id)
	return err
}
