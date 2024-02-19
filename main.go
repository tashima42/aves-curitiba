package main

import (
	"log"
	"os"

	"github.com/tashima42/aves-curitiba/cmd"
	"github.com/urfave/cli/v2"
)

var version = "dev"

func main() {
	app := cli.App{
		Name:                   "aves-curitiba",
		Usage:                  "control and run the aves-curitiba scrapper",
		UseShortOptionHandling: true,
		Version:                version,
		Commands:               []*cli.Command{cmd.DBCommand(), cmd.ScrapperCommand()},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
