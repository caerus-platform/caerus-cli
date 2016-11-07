package commands

import (
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

// define config variable name
const (
	CaerusAPI = "caerus_api"        // Deprecated: caerus config name
	MarathonHost = "marathon_host"  // marathon config name
)

// InitConfig used for init config from other place
func InitConfig() {
	loadConfig()
}

// setup default value and load config file
func loadConfig() map[string]interface{} {
	viper.SetDefault(CaerusAPI, "http://api.center.caerus.x")
	viper.SetDefault(MarathonHost, "http://marathon.caerus.x")

	viper.SetConfigName("config.yaml")
	viper.AddConfigPath("/etc/caerus")
	viper.AddConfigPath("$HOME/.caerus")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Noticef("%s, Use default configuration.", err)
	}
	return viper.AllSettings()
}

// ConfigCommands returns config commands
func ConfigCommands() []cli.Command {
	return []cli.Command{
		{
			Name: "config",
			Aliases: []string{"c"},
			Usage: "show config infomation",
			Action: func(c *cli.Context) {
				log.Debug(viper.AllSettings())
			},
		},
	}
}
