package docker

import (
	"github.com/urfave/cli"
	"fmt"
	"net/url"
	"net/http"
	"bufio"
	"log"
)

func StreamLogs(host string, id string) {
	u, _ := url.Parse(fmt.Sprintf(
		"http://%s:2375/containers/%s/logs?follow=true&stderr=true&timestamps=true&stdout=true",
		host, id))
	r, _ := http.Get(u.String())
	defer r.Body.Close()
	reader := bufio.NewReader(r.Body)
	for {
		line, _, _ := reader.ReadLine()
		log.Println(string(line))
	}
}

func Commands() []cli.Command {
	return []cli.Command{
		{
			Name:        "docker",
			Aliases:     []string{"d"},
			Usage:       "options for docker",
			Flags:       []cli.Flag{
				cli.StringFlag{
					Name: "host, H",
					Usage: "host for app",
				},
			},
			Subcommands: []cli.Command{
				{
					Name: "logs",
					Usage: "logs id, logs for container id",
					Action: func(c *cli.Context) {
						id := c.Args().First()
						host := c.GlobalString("host")
						StreamLogs(host, id)
					},
				},
			},
		},
	}
}
