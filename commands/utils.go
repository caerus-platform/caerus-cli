package commands

import (
	"io"
	"fmt"
	"os/user"
	"github.com/fatih/color"
)

var keyColor = color.New(color.FgRed).SprintfFunc()
var valueColor = color.New(color.FgGreen).SprintfFunc()
var errColor = color.New(color.FgGreen).SprintfFunc()

func closeGracefully(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", errColor(msg), err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func homeDir() string {
	usr, _ := user.Current()
	return usr.HomeDir
}
