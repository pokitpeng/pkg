package scp

import (
	"testing"
)

func TestScp(t *testing.T) {
	client, err := NewClient("117.50.194.154:22", "root", "ucloud@pengpeng123")
	if err != nil {
		t.Error(err)
		return
	}
	defer client.Close()
	err = client.Scp("/Users/user/Downloads/zhetian.txt", "/tmp/zhetian.txt")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestPull(t *testing.T) {
	client, err := NewClient("117.50.194.154:22", "root", "ucloud@pengpeng123")
	if err != nil {
		t.Error(err)
		return
	}
	defer client.Close()

	err = client.Pull("/tmp/zhetian.txt", "/tmp/zhetian.txt")
	if err != nil {
		t.Error(err)
		return
	}
}
