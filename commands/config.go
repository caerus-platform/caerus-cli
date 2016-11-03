package commands

import (
	"github.com/spf13/viper"
	"log"
	"github.com/urfave/cli"
)

const (
	CAERUS_API = "caerus_api"
	MARATHON_HOST = "marathon_host"
)

func Init() {
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
		log.Println(err)
		log.Println("Use default configuration.")
	}
	return viper.AllSettings()
}

func ConfigCommands() []cli.Command {
	log.SetPrefix("Config:\t")
	return []cli.Command{
		{
			Name: "config",
			Aliases: []string{"c"},
			Usage: "show config infomation",
			Action: func(c *cli.Context) {
				log.Println(viper.AllSettings())
			},
		},
	}
}
