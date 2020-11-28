package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"sort"
)

func main() {
	app := cli.NewApp()
	app.Usage = "mq tool for client"
	app.Flags = initFlags()
	app.Commands = initCommands()

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
