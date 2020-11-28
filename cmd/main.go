package main

import (
	"github.com/urfave/cli/v2"
	"gomq/cmd/service"
	"log"
	"os"
	"sort"
)

func main() {
	app := cli.NewApp()
	app.Usage = "mq tool for client"
	app.Flags = service.Flags()
	app.Commands = service.Commands()

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
