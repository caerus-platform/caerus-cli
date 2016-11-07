package commands

import "io"

// Close gracefully close io.Closer
func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}
