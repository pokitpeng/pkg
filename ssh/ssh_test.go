package ssh

import (
	"fmt"
	"testing"
)

func TestNewSession(t *testing.T) {
	session, err := NewSession("117.50.194.154:22", "root", "ucloud@pengpeng123")
	if err != nil {
		return
	}
	defer session.Close()
	out, err := session.CombinedOutput("ls /root")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(string(out))
}
