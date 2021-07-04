package openvpn

import (
	"github.com/mungaij83/go-openvpn/core"
	"sync"
	"testing"
)

var waitGroup sync.WaitGroup

func TestNewManagement(t *testing.T) {
	connector := core.NewTcpConnector("127.0.0.1", 7505, "", core.ClientMode)
	m := NewManagement(&Process{}, connector)

	if m == nil {
		t.Error("Return is nil")
	}
}


