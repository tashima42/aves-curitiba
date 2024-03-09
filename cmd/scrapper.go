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
			&cli.BoolFlag{
				Name:     "additional-data",
				Usage:    "Scrape additional data for existing records",
				Required: false,
				Value:    false,
				Aliases:  []string{"t"},
				EnvVars:  []string{"ADDITIONAL_DATA"},
			},
		},
		Action: run,
	}
}

func run(c *cli.Context) error {
	db, err := database.Open(c.String("db-path"), false)
	if err != nil {
		return err
	}
	defer database.Close(db)

	return runScrapper(db, c.String("auth-cookie"), c.Bool("additional-data"))
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
