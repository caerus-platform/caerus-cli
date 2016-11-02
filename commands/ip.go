package commands

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
)

type IPInfo struct {
	Ip       string    `json:"ip"`
	Hostname string    `json:"hostname"`
	City     string    `json:"city"`
	Region   string    `json:"region"`
	Country  string    `json:"country"`
	Loc      string    `json:"loc"`
	Org      string    `json:"org"`
	Postal   string    `json:"postal"`
}

func info(ip string) {
	r, _ := http.Get("http://ipinfo.io/8.8.8.8/json")
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)

	ipInfo := IPInfo{}
	json.Unmarshal([]byte(string(body)), &ipInfo)
	fmt.Println(ipInfo)
}

func IpCommands() []cli.Command {
	return []cli.Command{
		{
			Name: "ip",
			Usage: "get info about ip address",
			Action: func(c *cli.Context) {
				info(c.Args().Get(0))
			},
		},
	}
}
