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

					return runScrapper(db, c.String("auth-cookie"), false, "")
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
					&cli.StringFlag{
						Name:     "write-path",
						Usage:    "write path",
						Required: true,
						Aliases:  []string{"w"},
						EnvVars:  []string{"WRITE_PATH"},
					},
				},
				Action: func(c *cli.Context) error {
					db, err := database.Open(c.String("db-path"), false)
					if err != nil {
						return err
					}
					defer database.Close(db)

					return runScrapper(db, c.String("auth-cookie"), true, c.String("write-path"))
				},
			},
			{
				Name: "html",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "db-path",
						Usage:    "path for the sqlite database",
						Required: true,
						Aliases:  []string{"d"},
						EnvVars:  []string{"DB_PATH"},
					},
					&cli.StringFlag{
						Name:     "html-path",
						Usage:    "path for the html files",
						Required: true,
						Aliases:  []string{"p"},
						EnvVars:  []string{"HTML_PATH"},
					},
				},
				Action: func(c *cli.Context) error {
					db, err := database.Open(c.String("db-path"), false)
					if err != nil {
						return err
					}
					defer database.Close(db)
					sc := scrapper.Scrapper{
						DB:       db,
						HTMLPath: c.String("html-path"),
					}
					// return nil
					return sc.ScrapeHTML()
					// return scrapper.Test()
				},
			},
		},
	}
}

func runScrapper(db *sqlx.DB, authCookie string, additionalData bool, writeToPath string) error {
	sc := scrapper.Scrapper{
		DB:          db,
		AuthCookie:  authCookie,
		WriteToPath: writeToPath,
	}
	if additionalData {
		return sc.ScrapeAdditionalData()
	}
	return sc.Scrape()
}
