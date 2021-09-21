package ssh

import (
	"testing"
)

func TestNewSession(t *testing.T) {
	cli, err := NewClient("127.0.0.1:22", "xxxx")
	if err != nil {
		t.Logf("NewSession err: %s", err.Error())
		return
	}
	defer cli.Close()

	cli.RemoteCMD("touch /root/2.txt")
	res, err := cli.RemoteCMDOut("ls -l /root")
	if err != nil {
		t.Logf("session.RemoteCMDOut err: %s", err.Error())
		return
	}
	t.Logf("res=%s", string(res))
}

func TestSSHSession(t *testing.T) {
	res, err := RemoteCMDOut("127.0.0.1:22", "xxxx", "ls -l /root")
	if err != nil {
		t.Errorf("RemoteCMDOut err: %s", err.Error())
		return
	}
	t.Logf("res=%s", string(res))
}
