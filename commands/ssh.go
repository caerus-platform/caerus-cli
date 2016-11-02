package commands

import (
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os/user"
)

func KeyFile() (key ssh.Signer, err error) {
	usr, _ := user.Current()
	file := usr.HomeDir + "/.ssh/zhulux-staging"
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	key, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		return
	}
	return
}

func newSession(host string) (session *ssh.Session, err error) {
	singer, err := KeyFile()
	if err != nil {
		return
	}
	auth := []ssh.AuthMethod{ssh.PublicKeys(singer)}

	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: auth,
	}
	sshConfig.SetDefaults()

	client, err := ssh.Dial("tcp", "192.168.3.3:22", sshConfig)
	if err != nil {
		return
	}

	session, err = client.NewSession()
	if err != nil {
		return
	}

	log.Println("[x] - new session created.")
	return
}
