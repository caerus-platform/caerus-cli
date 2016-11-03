package main

import (
	"github.com/urfave/cli"
	"os"
	"log"
	"./commands"
)

func main() {
	log.SetPrefix("Caerus:\t")
	cli.VersionPrinter = func(c *cli.Context) {
		log.Printf("version=%s\n", c.App.Version)
	}

	app := cli.NewApp()
	app.Name = "Caerus Command Helper"
	app.Version = "0.0.1.rc1"
	app.Usage = "^_^"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "host, H",
			Usage: "setup caerus api `HOST`",
		},
	}

	app.Commands = []cli.Command{}

	app.Commands = append(app.Commands, commands.IpCommands()...)
	app.Commands = append(app.Commands, commands.DockerCommands()...)
	app.Commands = append(app.Commands, commands.MarathonCommands()...)
	app.Commands = append(app.Commands, commands.SshCommands()...)
	app.Commands = append(app.Commands, commands.ConfigCommands()...)

	app.Before = func(c *cli.Context) error {
		commands.Init()
		return nil
	}

	app.Run(os.Args)
}
