package commands

import (
	"io"
	"fmt"
	"os/user"
	"github.com/fatih/color"
	"os/exec"
	"os"
	"strings"
	"strconv"
)

var keyColor = color.New(color.FgRed).SprintfFunc()
var valueColor = color.New(color.FgGreen).SprintfFunc()
var runningColor = color.New(color.FgGreen).SprintfFunc()
var shutdownColor = color.New(color.FgRed).SprintfFunc()
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

func terminalWidth() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 60
	} else {
		widthStr := strings.Split(strings.TrimSpace(string(out)), " ")[1]
		width, _ := strconv.Atoi(widthStr)
		return width - 30
	}
}
