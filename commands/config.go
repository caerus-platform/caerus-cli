package commands

import (
	"github.com/spf13/viper"
	"github.com/urfave/cli"
	"os/user"
	"html/template"
	"bytes"
	"fmt"
	"os"
	"github.com/pkg/errors"
)

// define config variable name
const (
	CaerusAPI = "caerus_api"        // Deprecated: caerus config
	MarathonHost = "marathon_host"  // marathon config
	MQHost = "mq_host"              // mq config
	PrivateKey = "private_key"      // ssh private_key
)

var configStr = `---
# Caerus Command Line Interface
# BY Daniel Wei (danielwii0326@gmail.com)

# Caerus API 地址
# example: http://api.center.caerus.x
caerus_api: {{ .caerus_api }}

# Marathon API 地址
# example: http://marathon.caerus.x
marathon_host: {{ .marathon_host }}

# MQ 相关配置
# example: amqp://username:password@caerus.x:5672
mq_host: {{ .mq_host }}

# SSH 相关配置
private_key: ~/.ssh/id_rsa
`

// InitConfig used for init config from other place
func InitConfig() {
	loadConfig()
}

// setup default value and load config file
func loadConfig() map[string]interface{} {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	//viper.AddConfigPath("/etc/caerus")
	viper.AddConfigPath("$HOME/.caerus")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Noticef("%s, Use default configuration.", err)
	}
	return viper.AllSettings()
}

func getConfig(key string) (value string, err error) {
	value = viper.GetString(key)
	if value == "" {
		err = errors.New(key + " need to set in config.")
	}
	return
}

// ConfigCommands returns config commands
func ConfigCommands() []cli.Command {
	return []cli.Command{
		{
			Name:        "config",
			Aliases:     []string{"c"},
			Usage:       "config operations",
			Subcommands: []cli.Command{
				{
					Name: "init",
					Usage: "init new config file, remove old one.",
					Action: func(c *cli.Context) {
						usr, err := user.Current()
						failOnError(err, "Get current user error!")

						configFile := fmt.Sprintf("%s/.caerus/config.yml", usr.HomeDir)
						log.Debugf("Remove old config file if exists.")
						os.Remove(configFile)

						log.Debugf("Create %s", configFile)
						os.MkdirAll(fmt.Sprintf("%s/.caerus", usr.HomeDir), 0755)
						f, err := os.Create(configFile)
						failOnError(err, "Failed to create config file.")

						tmpl, err := template.New("config").Parse(configStr)
						failOnError(err, "Parse template error!")

						log.Debugf("Render with %s", viper.AllSettings())

						buf := bytes.NewBufferString("")
						err = tmpl.Execute(buf, viper.AllSettings())
						failOnError(err, "Render template error!")
						log.Debugf("Template is %s", buf.String())

						_, err = f.Write(buf.Bytes())
						failOnError(err, "Failed to write to file.")
					},
				},
				{
					Name:    "info",
					Aliases: []string{"i"},
					Usage:   "show configs",
					Action:  func(c *cli.Context) {
						log.Debug(viper.AllSettings())
					},
				},
			},
		},
	}
}
