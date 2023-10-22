package ssh

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

type Session struct {
	*ssh.Session
}

func NewSession(addr, username, passwd string) (*Session, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	cli, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("ssh dial %s err: %s", addr, err.Error())
	}
	session, err := cli.NewSession()
	if err != nil {
		return nil, err
	}
	return &Session{session}, nil
}
