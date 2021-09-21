package ssh

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	client *ssh.Client
}

func NewClient(addr, rootPassword string) (*Client, error) {
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(rootPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	cli, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("ssh dial %s err: %s", addr, err.Error())
	}
	return &Client{cli}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// RemoteCMD ...
func (c *Client) RemoteCMD(cmd string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("new session err: %s", err.Error())
	}
	defer session.Close()

	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("run remote cmd %s err: %s", cmd, err.Error())
	}
	return nil
}

// RemoteCMDOut ...
func (c *Client) RemoteCMDOut(cmd string) ([]byte, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("new session err: %s", err.Error())
	}
	defer session.Close()
	return session.Output(cmd)
}

// 单独调用，每条命令建立一个ssh连接，适合执行单个命令

// SSHSession ...
func SSHSession(addr, rootPassword string) (*ssh.Session, error) {
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(rootPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	cli, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("ssh dial %s err: %s", addr, err.Error())
	}
	return cli.NewSession()
}

// RemoteCMD ...
func RemoteCMD(addr, rootPassword, cmd string) error {
	session, err := SSHSession(addr, rootPassword)
	if err != nil {
		return fmt.Errorf("SSHSession %s err: %s", addr, err.Error())
	}
	defer session.Close()
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("session run %s err: %s", cmd, err.Error())
	}
	return nil
}

// RemoteCMDOut ...
func RemoteCMDOut(addr, passwd, cmd string) ([]byte, error) {
	session, err := SSHSession(addr, passwd)
	if err != nil {
		return nil, fmt.Errorf("SSHSession %s err: %s", addr, err.Error())
	}
	defer session.Close()
	return session.Output(cmd)
}
