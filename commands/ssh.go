package commands

import (
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os/user"
	"github.com/urfave/cli"
	"fmt"
	"os"
	"golang.org/x/crypto/ssh/terminal"
)

func keyFile(path string) (key ssh.Signer, err error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	key, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		return
	}
	return
}

func newSession(user string, host string, port string, path string) (session *ssh.Session, err error) {
	singer, err := keyFile(path)
	if err != nil {
		return
	}
	auth := []ssh.AuthMethod{ssh.PublicKeys(singer)}

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: auth,
	}
	sshConfig.SetDefaults()

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), sshConfig)
	if err != nil {
		return
	}

	session, err = client.NewSession()
	if err != nil {
		return
	}

	log.Debug("[x] - new session created.")
	return
}

func runCommand(user string, host string, port string, key string, cmd string) {
	session, err := newSession(user, host, port, key)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	fd := int(os.Stdin.Fd())
	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		log.Panic(err)
	}
	defer terminal.Restore(fd, oldState)

	// execute commands
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		log.Panic(err)
	}

	// setup terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:              1, // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // set input speed to 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // set output speed to 14.4kbaud
	}

	// request pseudo terminal
	if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
		log.Fatal(err)
	}
	session.Run(cmd)
}

// SSHCommands returns ssh commands
func SSHCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "ssh",
			Usage:  "ssh to address",
			Action: func(c *cli.Context) {
				os.Exit(1) // not provided
				usr, _ := user.Current()
				file := usr.HomeDir + "/.ssh/zhulux-staging"
				runCommand("root", "192.168.3.2", "22", file, "docker exec -it 49012ca13d37 bash")
			},
		},
	}
}
