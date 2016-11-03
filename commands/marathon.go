package commands

import (
	"github.com/urfave/cli"
	"fmt"
	"net/http"
	"encoding/json"
	"net/url"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
	"strings"
	"bytes"
	"github.com/spf13/viper"
	"log"
)

type DockerContainerHost struct {
	Ip string                         `json:"ip"`
}

type DockerContainer struct {
	Id      string
	State   string
	Status  string
	Image   string
	Command string
	Host    DockerContainerHost       `json:"host"`
}

type LastTaskFailure struct {
	Timestamp string                  `json:"timestamp"`
	Message   string                  `json:"message"`
}

type MarathonTask struct {
	Id    string                      `json:"id"`
	Host  string                      `json:"host"`
	Ports []int64                     `json:"ports"`
	State string                      `json:"state"`
}

type PortMapping struct {
	Protocol      string              `json:"protocol"`
	ContainerPort int64               `json:"containerPort"`
}

type MarathonDocker struct {
	Image        string               `json:"image"`
	PortsMapping []PortMapping        `json:"portMappings"`
}

type MarathonContainer struct {
	Docker MarathonDocker             `json:"docker"`
}

type MarathonApp struct {
	Id              string            `json:"id"`
	Instances       int64             `json:"instances"`
	Cpus            float64           `json:"cpus"`
	Constraints     [][]string        `json:"constraints"`
	Labels          map[string]string `json:"labels"`
	MEM             int64             `json:"mem"`
	Container       MarathonContainer `json:"container"`
	Tasks           []MarathonTask    `json:"tasks"`
	Env             map[string]string `json:"env"`
	LastTaskFailure LastTaskFailure   `json:"lastTaskFailure"`
}

func (app MarathonApp) containers() (containers []DockerContainer) {
	length := len(app.Tasks)
	containers = []DockerContainer{}
	if length > 0 {
		for _, task := range app.Tasks {
			log.Println("Found", task.Id)
			container := fetchContainerByTask(viper.GetString(CAERUS_API), task.Id)
			containers = append(containers, container)
		}
	} else {
		log.Println("Task is empty, check if app has any task...", length)
	}
	return
}

type MarathonCallApp struct {
	App MarathonApp                   `json:"app"`
}

type MarathonCallApps struct {
	Apps []MarathonApp                `json:"apps"`
}

func fetchContainerByTask(host string, taskId string) DockerContainer {
	u, _ := url.Parse(fmt.Sprintf("%s/api/v1/marathon/%s/container", host, taskId))
	r, _ := http.Get(u.String())
	defer r.Body.Close()
	container := DockerContainer{}
	json.NewDecoder(r.Body).Decode(&container)
	return container
}

func fetchApp(host string, id string) MarathonApp {
	u, _ := url.Parse(fmt.Sprintf("%s/api/v1/marathon/app?id=%s", host, id))
	r, _ := http.Get(u.String())
	defer r.Body.Close()
	app := MarathonCallApp{}
	json.NewDecoder(r.Body).Decode(&app)
	return app.App
}

func fetchApps(host string) []MarathonApp {
	u, err := url.Parse(host + "/api/v1/marathon/apps")
	if err != nil {
		log.Panic(err)
	}

	r, err := http.Get(u.String())
	if err != nil {
		os.Exit(1)
		log.Panic(err)
	}

	defer r.Body.Close()

	apps := MarathonCallApps{}
	json.NewDecoder(r.Body).Decode(&apps)
	return apps.Apps
}

func renderTasks(tasks []MarathonTask) {
	for _, task := range tasks {
		log.Println(task)
	}
}

func renderApp(app MarathonApp) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Append([]string{"ID", app.Id})
	table.Append([]string{"Image", app.Container.Docker.Image})
	table.Append([]string{"Env", func() string {
		var buffer bytes.Buffer
		for k, v := range app.Env {
			buffer.WriteString(fmt.Sprintf("%s:%s\n", k, v))
		}
		return buffer.String()
	}()})
	table.Append([]string{"Instances", strconv.FormatInt(app.Instances, 10)})
	table.Append([]string{"Cpus", strconv.FormatFloat(app.Cpus, 'f', 1, 64)})
	table.Append([]string{"Mem", strconv.FormatInt(app.MEM, 10)})
	table.Append([]string{"Constraints", func() string {
		var buffer bytes.Buffer
		for _, constraint := range app.Constraints {
			buffer.WriteString(strings.Join(constraint, ":"))
			buffer.WriteString("\n")
		}
		return buffer.String()
	}()})
	table.Append([]string{"Labels", func() string {
		var buffer bytes.Buffer
		for k, v := range app.Labels {
			buffer.WriteString(fmt.Sprintf("%s:%s\n", k, v))
		}
		return buffer.String()
	}()})
	table.Append([]string{"Labels", func() string {
		var buffer bytes.Buffer
		for _, portMapping := range app.Container.Docker.PortsMapping {
			buffer.WriteString(fmt.Sprintf("%s:%s\n",
				portMapping.Protocol, strconv.FormatInt(portMapping.ContainerPort, 10)))
		}
		return buffer.String()
	}()})
	table.Render()

	log.Println("----------------------------------")
	renderTasks(app.Tasks)
	log.Println("----------------------------------")
	log.Println("Last failure:", app.LastTaskFailure)
}

func renderApps(apps []MarathonApp) {
	log.Println("Rendering table...")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Image"})
	table.SetFooter([]string{"Total", strconv.FormatInt(int64(len(apps)), 10)})

	for _, app := range apps {
		table.Append([]string{
			app.Id,
			app.Container.Docker.Image,
		})
	}
	table.Render()
}

func MarathonCommands() []cli.Command {
	log.SetPrefix("Marathon:\t")
	return []cli.Command{
		{
			Name:        "marathon",
			Aliases:     []string{"m"},
			Usage:       "options for marathon",
			Subcommands: []cli.Command{
				{
					Name:  "app",
					Usage: "show app info, operate app",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "restart, r",
							Usage: "restart app",
						},
					},
					Action: func(c *cli.Context) error {
						id := c.Args().First()
						log.Println("App info:", id)
						app := fetchApp(viper.GetString(CAERUS_API), id)

						renderApp(app)

						return nil
					},
					Subcommands: []cli.Command{
						{
							Name: "logs",
							Usage: "tail logs for docker",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name: "task",
									Usage: "--task taskId, which task. default is the first.",
									Value: "",
								},
							},
							Action: func(c *cli.Context) {
								id := c.Args().First()
								app := fetchApp(viper.GetString(CAERUS_API), id)
								taskId := c.String("task")
								if taskId == "" {
									length := len(app.Tasks)
									log.Println("Task is empty, check if app has any task...", length)
									if length > 0 {
										task := app.Tasks[0]
										log.Println("using", task.Id)
										container := fetchContainerByTask(viper.GetString(CAERUS_API), task.Id)
										runStreamLogs(task.Host, container.Id)
									}
								} else {
									log.Println("Not implemented yet.")
									//container := fetchContainerByTask(host, taskId)
									//docker.StreamLogs(container.Host, container.Id)
								}
							},
						},
					},
				},
				{
					Name: "ssh",
					Usage: "ssh to a marathon app, default is the first docker in tasks",
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
						id := c.Args().First()
						if id == "" {
							log.Fatalln("container_id is needed")
						}
						cmd := c.String("command")
						if cmd == "" {
							log.Fatalln("command is needed")
						}

						port := c.String("port")
						user := c.String("user")
						key := c.String("key")
						app := fetchApp(viper.GetString(CAERUS_API), id)
						if containers := app.containers(); len(containers) > 0 {
							container := containers[0]
							cmd = fmt.Sprintf("docker exec -it %s %s", container.Id, cmd)
							//log.Println(user, container.Host.Ip, port, key, cmd)
							runCommand(user, container.Host.Ip, port, key, cmd)
						} else {
							log.Fatalln("No running tasks found.")
						}
					},
				},
				{
					Name:  "apps",
					Usage:
					`list all apps.`,
					//[filter] - name::regex exp: marathon apps id::/caerus`,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "env, e",
							Value: "",
							Usage: "declare env for searching",
						},
					},
					Action: func(c *cli.Context) error {
						log.Println("Display all apps...")
						apps := fetchApps(viper.GetString(CAERUS_API))

						renderApps(apps)

						return nil
					},
				},
			},
		},
	}
}
