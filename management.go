package openvpn

import (
	log "github.com/cihub/seelog"
	"github.com/golang/glog"
	"github.com/mungaij83/go-openvpn/core"
	"github.com/mungaij83/go-openvpn/utils"
)

type Management struct {
	Conn          *Process
	connector     core.OpenVpnConnector
	Mode          int
	Events        chan utils.Event `json:"-"`
	OpenVpnEvents chan string
	clientEnv     map[string]string
	shutdown      chan bool
}

func NewManagement(conn *Process, connector core.OpenVpnConnector) *Management {
	return &Management{
		connector:     connector,
		Conn:          conn,
		Events:        make(chan utils.Event),
		OpenVpnEvents: make(chan string, 10),
		clientEnv:     make(map[string]string, 0),
		shutdown:      make(chan bool),
	}
}

func (m *Management) Start() error { // {{{
	err := m.connector.Connect()
	if err != nil {
		return err
	}
	m.connector.Listen(m.OpenVpnEvents)
	return nil
}

func (m *Management) Fire(name string, args ...string) {
	select {
	case m.Events <- utils.Event{
		Name: name,
		Args: args,
	}:
	default:
		glog.Warningf("Lost event: %v args: %v",  name,args)
	}
}

func (m *Management) Shutdown() {
	log.Info("Management: shutdown")
	close(m.shutdown)
	err := m.connector.Close()
	if err != nil {
		glog.Error(err)
	}
}
