package commands

import (
	"github.com/urfave/cli"
	"net/url"
	"net/http"
	"bufio"
	"fmt"
	"encoding/json"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
	"strings"
)

// DockerMount ...
type DockerMount struct {
	Source string
}

// DockerContainer ...
type DockerContainer struct {
	Command string
	ID      string `json:"Id"`
	Image   string
	Mounts  []DockerMount
	Names   []string
	State   string
	Status  string
	Host    string `json:"-"`
}

func (container DockerContainer) setHost(host string) {
	container.Host = host
}

// list docker containers for host
func listContainers(host string) (containers []DockerContainer, err error) {
	u, _ := url.Parse(fmt.Sprintf("http://%s:2375/containers/json?all=1", host))
	r, err := http.Get(u.String())
	defer Close(r.Body)
	json.NewDecoder(r.Body).Decode(&containers)
	for _, container := range containers {
		container.setHost(host)
	}
	return
}

func runStreamLogs(host string, id string) {
	u, _ := url.Parse(fmt.Sprintf(
		"http://%s:2375/containers/%s/logs?follow=true&stderr=true&timestamps=true&stdout=true",
		host, id))
	r, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer Close(r.Body)
	reader := bufio.NewReader(r.Body)
	for line := []byte{0}; len(line) > 0; {
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		println(string(line))
	}
}

func renderContainers(containers []DockerContainer) {
	log.Debugf("Rendering table...")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Image", "State", "Status", "Names"})
	table.SetFooter([]string{"", "", "", "Total", strconv.FormatInt(int64(len(containers)), 10)})

	for _, container := range containers {
		table.Append([]string{
			container.ID[:12],
			container.Image,
			container.State,
			container.Status,
			strings.Join(container.Names, ","),
		})
	}
	table.Render()
}

// DockerCommands returns docker commands
func DockerCommands() []cli.Command {
	return []cli.Command{
		{
			Name:        "docker",
			Aliases:     []string{"d"},
			Usage:       "options for docker",
			Subcommands: []cli.Command{
				{
					Name: "containers",
					Aliases: []string{"c"},
					Usage: "containers for host",
					Action: func(c *cli.Context) {
						host := c.Args().First()
						containers, err := listContainers(host)
						if err != nil {
							log.Fatalf("List containers for %s error: %s", host, err)
						}
						renderContainers(containers)
					},
				},
				{
					Name:   "logs",
					Usage:  "host_id container_id, logs for container on host",
					Action: func(c *cli.Context) {
						host := c.Args().First()
						id := c.Args().Get(1)
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
						host := c.Args().First()
						if host == "" {
							log.Fatal("docker_host is needed")
						}
						id := c.Args().Get(1)
						if id == "" {
							log.Fatal("container_id is needed")
						}
						cmd := c.String("command")
						if cmd == "" {
							log.Fatal("command is needed")
						}

						cmd = fmt.Sprintf("docker exec -it %s %s", id, cmd)
						port := c.String("port")
						user := c.String("user")
						key := c.String("key")
						//log.Debugf(user, port, key, cmd)
						runCommand(user, host, port, key, cmd)
					},
				},
			},
		},
	}
}
