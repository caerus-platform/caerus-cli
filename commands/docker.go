package commands

import (
	"github.com/urfave/cli"
	"log"
	"net/url"
	"net/http"
	"bufio"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"strings"
)

func runStreamLogs(host string, id string) {
	u, _ := url.Parse(fmt.Sprintf(
		"http://%s:2375/containers/%s/logs?follow=true&stderr=true&timestamps=true&stdout=true",
		host, id))
	r, _ := http.Get(u.String())
	defer r.Body.Close()
	reader := bufio.NewReader(r.Body)
	for line := []byte{0}; len(line) > 0; {
		line, _, _ := reader.ReadLine()
		log.Println(string(line))
	}
}

type ExecCreate struct {
	AttachStdin  bool
	AttachStdout bool
	AttachStderr bool
	Cmd          []string
	DetachKeys   string
	Privileged   bool
	Tty          bool
}

func runExecCreate(host string, id string, cmd []string) {
	u, _ := url.Parse(fmt.Sprintf("http://%s:2375/containers/%s/exec", host, id))
	body, _ := json.Marshal(ExecCreate{
		Cmd: cmd,
	})
	r, err := http.Post(u.String(), "application/json", strings.NewReader(string(body)))
	if err == nil {
		log.Panic(err)
	}
	defer r.Body.Close()
	b, _ := ioutil.ReadAll(r.Body)
	log.Println(string(b))
}

type ExecStart struct {
	Detach bool
	Tty    bool
}

type ExecResize struct{}

type ExecInspect struct{}

func DockerCommands() []cli.Command {
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
						runStreamLogs(host, id)
					},
				},
				{
					Name: "cmd",
					Usage: "run cmd",
					Action: func(c *cli.Context) {
						id := c.Args().First()
						host := c.GlobalString("host")
						runExecCreate(host, id, []string{"ls"})
					},
				},
			},
		},
	}
}
