package cmd

import (
	"github.com/jmoiron/sqlx"
	"github.com/tashima42/aves-curitiba/database"
	"github.com/tashima42/aves-curitiba/scrapper"
	"github.com/urfave/cli/v2"
)

func ScrapperCommand() *cli.Command {
	return &cli.Command{
		Name:  "scrapper",
		Usage: "start the wikiaves scrapper",
		Subcommands: []*cli.Command{
			{
				Name: "wa",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "db-path",
						Usage:    "path for the sqlite database",
						Required: true,
						Aliases:  []string{"d"},
						EnvVars:  []string{"DB_PATH"},
					},
					&cli.StringFlag{
						Name:     "auth-cookie",
						Usage:    "wikiaves auth cookie",
						Required: true,
						Aliases:  []string{"a"},
						EnvVars:  []string{"AUTH_COOKIE"},
					},
				},
				Action: func(c *cli.Context) error {
					db, err := database.Open(c.String("db-path"), false)
					if err != nil {
						return err
					}
					defer database.Close(db)

					return runScrapper(db, c.String("auth-cookie"), false)
				},
			},
			{
				Name: "wa-additional",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "db-path",
						Usage:    "path for the sqlite database",
						Required: true,
						Aliases:  []string{"d"},
						EnvVars:  []string{"DB_PATH"},
					},
					&cli.StringFlag{
						Name:     "auth-cookie",
						Usage:    "wikiaves auth cookie",
						Required: true,
						Aliases:  []string{"a"},
						EnvVars:  []string{"AUTH_COOKIE"},
					},
				},
				Action: func(c *cli.Context) error {
					db, err := database.Open(c.String("db-path"), false)
					if err != nil {
						return err
					}
					defer database.Close(db)

					return runScrapper(db, c.String("auth-cookie"), true)
				},
			},
			{
				Name: "csv",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "db-path",
						Usage:    "path for the sqlite database",
						Required: true,
						Aliases:  []string{"d"},
						EnvVars:  []string{"DB_PATH"},
					},
					&cli.StringFlag{
						Name:     "csv-path",
						Required: true,
						Aliases:  []string{"c"},
						EnvVars:  []string{"CSV_PATH"},
					},
				},
				Action: func(c *cli.Context) error {
					db, err := database.Open(c.String("db-path"), false)
					if err != nil {
						return err
					}
					defer database.Close(db)
					sc := scrapper.Scrapper{DB: db}
					return sc.CSVAdditionalData(c.String("csv-path"))
				},
			},
		},
	}
}

func runScrapper(db *sqlx.DB, authCookie string, additionalData bool) error {
	sc := scrapper.Scrapper{
		DB:         db,
		AuthCookie: authCookie,
	}
	if additionalData {
		return sc.ScrapeAdditionalData()
	}
	return sc.Scrape()
}
