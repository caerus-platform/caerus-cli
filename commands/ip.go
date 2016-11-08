package commands

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/urfave/cli"
	"fmt"
)

// IPInfo ...
type IPInfo struct {
	IP       string    `json:"ip"`
	Hostname string    `json:"hostname"`
	City     string    `json:"city"`
	Region   string    `json:"region"`
	Country  string    `json:"country"`
	Loc      string    `json:"loc"`
	Org      string    `json:"org"`
	Postal   string    `json:"postal"`
}

func info(ip string) {
	r, _ := http.Get(fmt.Sprintf("http://ipinfo.io/%s/json", ip))
	defer Close(r.Body)
	body, _ := ioutil.ReadAll(r.Body)

	ipInfo := IPInfo{}
	json.Unmarshal([]byte(string(body)), &ipInfo)
	log.Debug(ipInfo)
}

// IPCommands returns ip commands
func IPCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "ip",
			Usage:  "get info about ip address",
			Action: func(c *cli.Context) {
				info(c.Args().Get(0))
			},
		},
	}
}
