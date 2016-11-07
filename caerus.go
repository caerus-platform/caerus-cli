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
	app.Version = "0.0.2-rc.1"
	app.Usage = "^_^"

	app.Commands = []cli.Command{}

	app.Commands = append(app.Commands, commands.IPCommands()...)
	app.Commands = append(app.Commands, commands.DockerCommands()...)
	app.Commands = append(app.Commands, commands.MarathonCommands()...)
	//app.Commands = append(app.Commands, commands.SSHCommands()...)
	//app.Commands = append(app.Commands, commands.ConfigCommands()...)
	//app.Commands = append(app.Commands, commands.UICommands()...)

	app.Before = func(c *cli.Context) error {
		commands.InitLogger()
		commands.InitConfig()
		return nil
	}

	app.Run(os.Args)
}
