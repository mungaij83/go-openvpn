package openvpn

import (
	"flag"
	"github.com/golang/glog"
	"github.com/mungaij83/go-openvpn/core"
	"github.com/mungaij83/go-openvpn/utils"
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
	events         chan string
	shutdown       chan bool
	Events         chan core.EventData
	parser         core.CommandParser
	mode           int
	connection     core.OpenVpnConnector
}

func NewVpnManagement(ip string, socket string, port int, password string, mode int) OpenVpnManagement {
	vpn := OpenVpnManagement{
		events:   make(chan string),
		Events:   make(chan core.EventData),
		shutdown: make(chan bool),
		mode:     mode,
		parser:   core.NewCommandParser(),
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

// Starts OpenVPN management interface in client mode
// Events are sent by OpenVPN
func (vm *OpenVpnManagement) StartClient() error {
	err := vm.connection.Connect()
	if err != nil {
		return err
	}
	go vm.connection.Listen(vm.events)

	return nil
}

// Starts OpenVPN management interface in server mode
// Events are poled from the server
func (vm *OpenVpnManagement) StartServer() error {
	// Start Connectors
	err := vm.connection.Connect()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case e := <-vm.events:
				evt := vm.parser.ParseEvent(e)
				if evt != nil {
					glog.V(3).Infof("EVENT: %+v", evt)
					vm.Events <- *evt
				}
				break
			case <-vm.shutdown:
				glog.Infof("Shutdown server")
				return
			}
		}
	}()
	// Start server if supported
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
func (vm OpenVpnManagement) GetClients(data core.EventData) ([]utils.Client, error) {
	return vm.parser.ParseStatus(data.EventData)
}
func (vm *OpenVpnManagement) Exec(cmd string) {
	_, err := vm.connection.SendCommand(cmd)
	if err != nil {
		glog.V(2).Infof("Failed to send command: %v", err)
	}
}
