package go_openvpn

import (
	"flag"
	"github.com/golang/glog"
	"github.com/mungaij83/go-openvpn/core"
)

func init() {
	_ = flag.Set("alsologtostderr", "1")
	flag.Parse()
}

const (
	TcpSocket  = 0
	UnixSocket = 1
)

type OpenVpnManagement struct {
	connectionType int
	events         chan []string
	shutdown       chan bool
	mode           int
	connection     core.OpenVpnConnector
}

func NewVpnManagement(ip string, socket string, port int, password string, mode int) OpenVpnManagement {
	vpn := OpenVpnManagement{
		events:   make(chan []string),
		shutdown: make(chan bool),
		mode:     mode,
	}
	// Initialize socket
	if len(socket) > 0 {
		vpn.connectionType = UnixSocket
		vpn.connection = core.NewSocketConnector(socket, password, mode)
	} else {
		vpn.connectionType = TcpSocket
		vpn.connection = core.NewTcpConnector(ip, port, password, mode)
	}
	return vpn
}

func (vm *OpenVpnManagement) StartClient() error {
	err := vm.connection.Connect()
	if err != nil {
		return err
	}
	go vm.connection.Listen(vm.events)
	return nil
}

func (vm *OpenVpnManagement) StartServer() error {
	err := vm.connection.Connect()
	if err != nil {
		return err
	}
	go vm.connection.Listen(vm.events)
	return nil
}

func (vm OpenVpnManagement) Status() (string, error) {
	return vm.connection.SendCommand("status 1")
}

func (vm OpenVpnManagement) Shutdown() {
	close(vm.shutdown)
	err := vm.connection.Close()
	if err != nil {
		glog.V(2).Infof("Failed to close connection: %v", err)
	}
}
