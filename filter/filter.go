package filter

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/tashima42/aves-curitiba/database"
)

type Filter struct {
	DB *sqlx.DB
}

func (f *Filter) Filter() error {
	slog.Info("starting tx")
	tx, err := f.DB.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	alreadyRegistered := map[string]bool{}
	registros, err := database.GetRegistrosTxx(tx)
	if err != nil {
		return err
	}
	for i, r := range registros {
		registro := *r
		slog.Info(fmt.Sprintf("%d - %d", i, registro.WaID))
		key := fmt.Sprintf("%d-%s-%s", registro.EspecieID, registro.Perfil, registro.Data)
		if _, ok := alreadyRegistered[key]; ok {
			continue
		}
		alreadyRegistered[key] = true
		database.CreateRegistroTxx(tx, r)
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
