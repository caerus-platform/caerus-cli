package commands

import (
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

const (
	CAERUS_API = "caerus_api"
	MARATHON_HOST = "marathon_host"
)

func InitConfig() {
	loadConfig()
}

func loadConfig() map[string]interface{} {
	viper.SetDefault(CAERUS_API, "http://api.center.caerus.x")
	viper.SetDefault(MARATHON_HOST, "http://marathon.caerus.x")

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
