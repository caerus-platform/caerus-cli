package commands

import (
	"io"
	"fmt"
	"os/user"
)

func closeGracefully(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func homeDir() string {
	usr, _ := user.Current()
	return usr.HomeDir
}
