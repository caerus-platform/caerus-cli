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
	"io"
)

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
	ContainerPort int64               `json:"containerPort"`
	HostPort      int64               `json:"hostPort"`
	ServicePort   int64               `json:"servicePort"`
	Protocol      string              `json:"protocol"`
	Labels        []string            `json:"labels"`
}

// MarathonDocker ...
type MarathonDocker struct {
	Image          string             `json:"image"`
	Network        string             `json:"network"`
	Privileged     bool               `json:"privileged"`
	Parameters     []string           `json:"parameters"`
	ForcePullImage bool               `json:"forcePullImage"`
	PortsMapping   []PortMapping      `json:"portMappings"`
}

// MarathonContainer ...
type MarathonContainer struct {
	Type    string                    `json:"type"`
	Volumes []string                  `json:"volumes"`
	Docker  MarathonDocker            `json:"docker"`
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
			log.Debugf("Found [%s]", task.ID)
			container := fetchContainerByTask(task)
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
	defer closeGracefully(r.Body)
	body, _ := ioutil.ReadAll(r.Body)
	log.Debugf(string(body))
}

func (app MarathonApp) updateImage(image string, force bool) {
	u, _ := url.Parse(fmt.Sprintf(
		"%s/v2/apps/%s?force=%t",
		viper.GetString(MarathonHost), app.ID, force))
	app.Container.Docker.Image = image
	requestBody, _ := json.Marshal(app)
	log.Debugf("try update  [%s]'s image to [%s]", app.ID, image)

	putApp(u.String(), strings.NewReader(string(requestBody)))
}

func (app MarathonApp) scale(instances int, force bool) {
	u, _ := url.Parse(fmt.Sprintf(
		"%s/v2/apps/%s?force=%t",
		viper.GetString(MarathonHost), app.ID, force))
	log.Debugf("try scale  [%s]'s image to [%d]", app.ID, instances)

	requestBody := fmt.Sprintf(`{"instances": %d}`, instances)
	putApp(u.String(), strings.NewReader(requestBody))
}

func putApp(url string, body io.Reader) {
	client := &http.Client{}
	request, _ := http.NewRequest("PUT", url, body)
	r, err := client.Do(request)
	if err != nil {
		log.Panic(err)
	}
	defer closeGracefully(r.Body)
	resp, _ := ioutil.ReadAll(r.Body)
	log.Debugf(string(resp))
}

// MarathonCallApp ...
type MarathonCallApp struct {
	App MarathonApp                   `json:"app"`
}

// MarathonCallApps ...
type MarathonCallApps struct {
	Apps []MarathonApp                `json:"apps"`
}

func fetchContainerByTask(task MarathonTask) (container DockerContainer) {
	log.Debugf("Task is [%s]", task)
	containers, err := listContainers(task.Host)
	if err != nil {
		log.Panic(err)
	}
	if len(containers) == 0 {
		log.Fatalf("No containers found.")
	}

	for index, c := range containers {
		for _, mount := range c.Mounts {
			if contain := strings.Contains(mount.Source, task.ID); contain == true {
				renderContainer(containers[index])
				container = containers[index]
			}
		}
	}

	return
}

func fetchApp(id string) MarathonApp {
	u, _ := url.Parse(fmt.Sprintf(
		"%s/v2/apps/%s?embed=apps.tasks&embed=apps.deployments&embed=apps.counts&embed=apps.readiness",
		viper.GetString(MarathonHost), id))
	r, err := http.Get(u.String())
	if err != nil {
		log.Panic(err)
	}
	defer closeGracefully(r.Body)

	app := MarathonCallApp{}
	json.NewDecoder(r.Body).Decode(&app)
	return app.App
}

func fetchApps() []MarathonApp {
	u, _ := url.Parse(fmt.Sprintf(
		"%s/v2/apps?embed=apps.tasks&embed=apps.deployments&embed=apps.counts&embed=apps.readiness",
		viper.GetString(MarathonHost)))
	log.Debugf(u.String())
	r, err := http.Get(u.String())
	if err != nil {
		log.Panic(err)
	}
	defer closeGracefully(r.Body)

	apps := MarathonCallApps{}
	json.NewDecoder(r.Body).Decode(&apps)
	return apps.Apps
}

func renderTasks(tasks []MarathonTask) {
	log.Debugf("----------------------------------")
	for _, task := range tasks {
		log.Debug(task)
	}
	log.Debugf("----------------------------------")
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
	table.Append([]string{"Ports", func() string {
		var buffer bytes.Buffer
		for _, portMapping := range app.Container.Docker.PortsMapping {
			buffer.WriteString(fmt.Sprintf("%s:%s\n",
				portMapping.Protocol, strconv.FormatInt(portMapping.ContainerPort, 10)))
		}
		return buffer.String()
	}()})
	table.Render()

	renderTasks(app.Tasks)
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
			Before:      func(c *cli.Context) error {
				_, err := getConfig(MarathonHost)
				failOnError(err, "")
				return nil
			},
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
								app := fetchApp(id)

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

								app := fetchApp(id)

								app.restart(c.Bool("force"))
							},
						},
						{
							Name: "scale",
							Aliases: []string{"s"},
							Usage: "scale app, set to 0 to stop app.",
							Flags: []cli.Flag{
								cli.IntFlag{
									Name: "instances, n",
									Usage: "--instances [0-9]+ default is 0",
								},
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

								app := fetchApp(id)

								app.scale(c.Int("instances"), c.Bool("force"))
							},
						},
						{
							Name: "update",
							Aliases: []string{"u"},
							Usage: "update app",
							Flags: []cli.Flag{
								cli.BoolFlag{
									Name: "force, f",
									Usage: "--force true, default is false",
								},
								cli.StringFlag{
									Name: "image",
									Usage: "--image <image_name>",
								},
							},
							Action: func(c *cli.Context) {
								id := c.Args().First()
								if id == "" {
									log.Fatal("app_id is needed")
								}
								log.Debugf("App info: %s", id)
								image := c.String("image")
								if image == "" {
									log.Fatal("--image is needed")
								}

								app := fetchApp(id)

								app.updateImage(image, c.Bool("force"))
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
								app := fetchApp(id)
								taskID := c.String("task")
								if taskID == "" {
									length := len(app.Tasks)
									log.Debugf("Task is empty, check if app has any task...%d", length)
									if length > 0 {
										task := app.Tasks[0]
										log.Debugf("using %s", task.ID)
										container := fetchContainerByTask(task)
										runStreamLogs(task.Host, container.ID)
									}
								} else {
									log.Debugf("Not implemented yet.")
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
							log.Fatal("container_id is needed")
						}
						cmd := c.String("command")
						if cmd == "" {
							log.Fatal("command is needed")
						}

						port := c.String("port")
						user := c.String("user")
						key := c.String("key")
						app := fetchApp(id)

						if containers := app.containers(); len(containers) > 0 {
							container := containers[0]
							cmd = fmt.Sprintf("docker exec -it %s %s", container.ID, cmd)
							log.Debugf("[%s] - [%s] - [%s] - [%s] - [%s]", user, container.Host, port, key, cmd)
							runCommand(user, container.Host, port, key, cmd)
						} else {
							log.Fatal("No running tasks found.")
						}
					},
				},
				{
					Name:  "apps",
					Usage: "list all apps.",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "env, e",
							Value: "",
							Usage: "declare env for searching",
						},
					},
					Action: func(c *cli.Context) error {
						log.Debugf("Display all apps...")
						apps := fetchApps()

						renderApps(apps)

						return nil
					},
				},
			},
		},
	}
}
