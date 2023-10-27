package scp

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/sftp"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	*sftp.Client
}

func NewClient(addr, username, passwd string) (*Client, error) {
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
	client, err := sftp.NewClient(cli)
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

// Scp 从本地推送到远端
func (c *Client) Scp(local, remote string) error {
	localInfo, err := os.Stat(local)
	if err != nil {
		return err
	}
	localFile, err := os.Open(local)
	if err != nil {
		return err
	}
	defer localFile.Close()

	remoteFile, err := c.Create(remote)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	bar := progressbar.DefaultBytes(
		localInfo.Size(),
		fmt.Sprintf("scp %s to %s", local, remote),
	)

	_, err = io.Copy(io.MultiWriter(remoteFile, bar), localFile)
	if err != nil {
		return err
	}
	return nil
}

// Pull 从远端拉取到本地
func (c *Client) Pull(remote, local string) error {
	remoteFile, err := c.Open(remote)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	fileInfo, err := remoteFile.Stat()
	if err != nil {
		return err
	}

	localFile, err := os.OpenFile(local, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer localFile.Close()

	bar := progressbar.DefaultBytes(
		fileInfo.Size(),
		fmt.Sprintf("pull %s to %s", remote, local),
	)

	_, err = io.Copy(io.MultiWriter(localFile, bar), remoteFile)
	if err != nil {
		return err
	}
	return nil
}
