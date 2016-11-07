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
	"io/ioutil"
)

// DockerContainerHost ...
type DockerContainerHost struct {
	IP string                         `json:"ip"`
}

// DockerContainer ...
type DockerContainer struct {
	ID      string
	State   string
	Status  string
	Image   string
	Command string
	Host    DockerContainerHost       `json:"host"`
}

// LastTaskFailure ...
type LastTaskFailure struct {
	Timestamp string                  `json:"timestamp"`
	Message   string                  `json:"message"`
}

// MarathonTask ...
type MarathonTask struct {
	ID    string                      `json:"id"`
	Host  string                      `json:"host"`
	Ports []int64                     `json:"ports"`
	State string                      `json:"state"`
}

// PortMapping ...
type PortMapping struct {
	Protocol      string              `json:"protocol"`
	ContainerPort int64               `json:"containerPort"`
}

// MarathonDocker ...
type MarathonDocker struct {
	Image        string               `json:"image"`
	PortsMapping []PortMapping        `json:"portMappings"`
}

// MarathonContainer ...
type MarathonContainer struct {
	Docker MarathonDocker             `json:"docker"`
}

// MarathonApp ...
type MarathonApp struct {
	ID              string            `json:"id"`
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
	if len(app.Tasks) > 0 {
		for _, task := range app.Tasks {
			log.Debugf("Found", task.ID)
			container := fetchContainerByTask(viper.GetString(CaerusAPI), task.ID)
			containers = append(containers, container)
		}
	} else {
		log.Debugf("Task is empty, check if app has any task...")
	}
	return
}

func (app MarathonApp) restart(force bool) {
	u, _ := url.Parse(fmt.Sprintf("%s/v2/apps/%s/restart?force=%t", viper.GetString(MarathonHost), app.ID, force))
	r, err := http.Post(u.String(), "application/json", nil)
	if err != nil {
		log.Panic(err)
	}
	defer Close(r.Body)
	body, _ := ioutil.ReadAll(r.Body)
	log.Debugf(string(body))
}

// MarathonCallApp ...
type MarathonCallApp struct {
	App MarathonApp                   `json:"app"`
}

// MarathonCallApps ...
type MarathonCallApps struct {
	Apps []MarathonApp                `json:"apps"`
}

func fetchContainerByTask(host string, taskID string) DockerContainer {
	u, _ := url.Parse(fmt.Sprintf("%s/api/v1/marathon/%s/container", host, taskID))
	r, err := http.Get(u.String())
	if err != nil {
		log.Panic(err)
	}
	defer Close(r.Body)
	container := DockerContainer{}
	json.NewDecoder(r.Body).Decode(&container)
	return container
}

func fetchApp(host string, id string) MarathonApp {
	u, _ := url.Parse(fmt.Sprintf("%s/api/v1/marathon/app?id=%s", host, id))
	r, err := http.Get(u.String())
	if err != nil {
		log.Panic(err)
	}
	defer Close(r.Body)

	app := MarathonCallApp{}
	json.NewDecoder(r.Body).Decode(&app)
	return app.App
}

func fetchApps(host string) []MarathonApp {
	u, _ := url.Parse(host + "/api/v1/marathon/apps")
	r, err := http.Get(u.String())
	if err != nil {
		log.Panic(err)
	}
	defer Close(r.Body)

	apps := MarathonCallApps{}
	json.NewDecoder(r.Body).Decode(&apps)
	return apps.Apps
}

func renderTasks(tasks []MarathonTask) {
	for _, task := range tasks {
		log.Debug(task)
	}
}

func renderApp(app MarathonApp) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Append([]string{"ID", app.ID})
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

	log.Debugf("----------------------------------")
	renderTasks(app.Tasks)
	log.Debugf("----------------------------------")
	log.Debugf("Last failure: %s", app.LastTaskFailure)
}

func renderApps(apps []MarathonApp) {
	log.Debugf("Rendering table...")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Image"})
	table.SetFooter([]string{"Total", strconv.FormatInt(int64(len(apps)), 10)})

	for _, app := range apps {
		table.Append([]string{
			app.ID,
			app.Container.Docker.Image,
		})
	}
	table.Render()
}

// MarathonCommands returns marathon commands
func MarathonCommands() []cli.Command {
	return []cli.Command{
		{
			Name:        "marathon",
			Aliases:     []string{"m"},
			Usage:       "options for marathon",
			Subcommands: []cli.Command{
				{
					Name:  "app",
					Usage: "show app info and other operations like: restart, update image etc",
					Subcommands: []cli.Command{
						{
							Name: "info",
							Aliases: []string{"i"},
							Usage: "show app info",
							Action: func(c *cli.Context) {
								id := c.Args().First()
								if id == "" {
									log.Fatal("app_id is needed")
								}
								log.Debugf("App info: %s", id)
								app := fetchApp(viper.GetString(CaerusAPI), id)

								renderApp(app)
							},
						},
						{
							Name: "restart",
							Aliases: []string{"r"},
							Usage: "restart app",
							Flags: []cli.Flag{
								cli.BoolFlag{
									Name: "force, f",
									Usage: "--force true, default is false",
								},
							},
							Action: func(c *cli.Context) {
								id := c.Args().First()
								if id == "" {
									log.Fatal("app_id is needed")
								}
								log.Debugf("App info: %s", id)

								app := fetchApp(viper.GetString(CaerusAPI), id)

								app.restart(c.Bool("force"))
							},
						},
					},
				},
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
						if id == "" {
							log.Fatal("app_id is needed")
						}
						app := fetchApp(viper.GetString(CaerusAPI), id)
						taskID := c.String("task")
						if taskID == "" {
							length := len(app.Tasks)
							log.Debugf("Task is empty, check if app has any task...", length)
							if length > 0 {
								task := app.Tasks[0]
								log.Debugf("using", task.ID)
								container := fetchContainerByTask(viper.GetString(CaerusAPI), task.ID)
								runStreamLogs(task.Host, container.ID)
							}
						} else {
							log.Debugf("Not implemented yet.")
							//container := fetchContainerByTask(host, taskId)
							//docker.StreamLogs(container.Host, container.Id)
						}
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
							log.Fatal("container_id is needed")
						}
						cmd := c.String("command")
						if cmd == "" {
							log.Fatal("command is needed")
						}

						port := c.String("port")
						user := c.String("user")
						key := c.String("key")
						app := fetchApp(viper.GetString(CaerusAPI), id)

						if containers := app.containers(); len(containers) > 0 {
							container := containers[0]
							cmd = fmt.Sprintf("docker exec -it %s %s", container.ID, cmd)
							//log.Debugf()(user, container.Host.Ip, port, key, cmd)
							runCommand(user, container.Host.IP, port, key, cmd)
						} else {
							log.Fatal("No running tasks found.")
						}
					},
				},
				{
					Name:  "apps",
					Usage: "list all apps.",
					//[filter] - name::regex exp: marathon apps id::/caerus`,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "env, e",
							Value: "",
							Usage: "declare env for searching",
						},
					},
					Action: func(c *cli.Context) error {
						log.Debugf("Display all apps...")
						apps := fetchApps(viper.GetString(CaerusAPI))

						renderApps(apps)

						return nil
					},
				},
			},
		},
	}
}
