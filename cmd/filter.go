package cmd

import (
	"github.com/jmoiron/sqlx"
	"github.com/tashima42/aves-curitiba/database"
	"github.com/tashima42/aves-curitiba/filter"
	"github.com/urfave/cli/v2"
)

func FilterCommand() *cli.Command {
	return &cli.Command{
		Name:  "filter",
		Usage: "filter registros",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "db-path",
				Usage:    "path for the sqlite database",
				Required: true,
				Aliases:  []string{"d"},
				EnvVars:  []string{"DB_PATH"},
			},
		},
		Action: func(c *cli.Context) error {
			db, err := database.Open(c.String("db-path"), false)
			if err != nil {
				return err
			}
			defer database.Close(db)

			return runFilter(db)
		},
	}
}

func runFilter(db *sqlx.DB) error {
	fi := filter.Filter{
		DB: db,
	}
	return fi.Filter()
}
