package commands

import (
	"github.com/urfave/cli"
	"log"
	"net/url"
	"net/http"
	"bufio"
	"fmt"
	"encoding/json"
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

type ExecCreateResp struct {
	Id string
}

// TODO not working
func runExecCreate(host string, id string, cmd []string) ExecCreateResp {
	log.Println("Creating exec...")
	u, _ := url.Parse(fmt.Sprintf("http://%s:2375/containers/%s/exec", host, id))
	body, _ := json.Marshal(ExecCreate{
		AttachStdin: true,
		AttachStdout: true,
		AttachStderr: true,
		DetachKeys: "ctrl-p,ctrl-q,ctrl-c",
		Cmd: cmd,
	})
	r, err := http.Post(u.String(), "application/json", strings.NewReader(string(body)))
	if err != nil {
		log.Fatalln(err)
	}
	defer r.Body.Close()

	exec := ExecCreateResp{}
	json.NewDecoder(r.Body).Decode(&exec)
	log.Println("Create exec id", exec.Id)
	return exec
}

type ExecStart struct {
	Detach bool
	Tty    bool
}

// TODO not working
func runExecStart(host string, exec ExecCreateResp) {
	log.Println("Starting exec...")
	u, _ := url.Parse(fmt.Sprintf("http://%s:2375/exec/%s/start", host, exec.Id))
	body, _ := json.Marshal(ExecStart{
		Detach: false,
		Tty: true,
	})
	r, err := http.Post(u.String(), "application/json", strings.NewReader(string(body)))
	if err != nil {
		log.Fatalln(err)
	}
	defer r.Body.Close()
	// TODO not implemented
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
					Name:   "logs",
					Usage:  "logs id, logs for container id",
					Action: func(c *cli.Context) {
						id := c.Args().First()
						host := c.GlobalString("host")
						runStreamLogs(host, id)
					},
				},
				{
					Name:  "ssh",
					Usage: "ssh [container] bash or sh",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "command, c",
							Usage: `-c "command", -c "bash"`,
						},
						cli.StringFlag{
							Name: "port, p",
							Usage: "--port 22, default is 22",
							Value: "22",
						},
						cli.StringFlag{
							Name: "private-key, key",
							Usage: "--private-key ur/private/key",
						},
						cli.StringFlag{
							Name: "user, u",
							Usage: "--user root, default is root",
							Value: "root",
						},
					},
					Action: func(c *cli.Context) {
						host := c.GlobalString("host")
						if host == "" {
							log.Fatalln("docker --host is needed")
						}
						id := c.Args().First()
						if id == "" {
							log.Fatalln("container_id is needed")
						}
						cmd := c.String("command")
						if cmd == "" {
							log.Fatalln("command is needed")
						}

						cmd = fmt.Sprintf("docker exec -it %s %s", id, cmd)
						port := c.String("port")
						user := c.String("user")
						key := c.String("key")
						//log.Println(user, port, key, cmd)
						runCommand(user, host, port, key, cmd)
					},
				},
				{
					Name: "cmd",
					Usage: "run cmd",
					Action: func(c *cli.Context) {
						id := c.Args().First()
						if id == "" {
							log.Fatalln("container_id is needed")
						}
						host := c.GlobalString("host")
						if host == "" {
							log.Fatalln("docker --host is needed")
						}
						exec := runExecCreate(host, id, []string{"ls"})
						runExecStart(host, exec)
					},
				},
			},
		},
	}
}
