package commands

import (
	"github.com/spf13/viper"
	"fmt"
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
	viper.SetDefault(CAERUS_API, "http://localhost:9000")
	viper.SetDefault(MARATHON_HOST, "http://marathon.caerus.x")

	viper.SetConfigName("config.yaml")
	viper.AddConfigPath("/etc/caerus")
	viper.AddConfigPath("$HOME/.caerus")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Using default configuration.\n")
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
				fmt.Println(viper.AllSettings())
			},
		},
	}}
