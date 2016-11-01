package main

import (
	"github.com/urfave/cli"
	"os"
	"./modules/ip"
	"./modules/docker"
	"./modules/marathon"
	"./modules/test"
	"fmt"
)

var (
	host = "http://localhost:9000"
	//host = "http://api.center.caerus.x"
	Revision = ">___________,.<"
)

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("version=%s revision=%s\n", c.App.Version, Revision)
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

	app.Commands = append(app.Commands, ip.Commands()...)
	app.Commands = append(app.Commands, docker.Commands()...)
	app.Commands = append(app.Commands, marathon.Commands(host)...)
	app.Commands = append(app.Commands, test.Commands()...)

	app.Run(os.Args)
}
