package go_openvpn

import (
	log "github.com/cihub/seelog"
	"github.com/golang/glog"
	"github.com/mungaij83/go-openvpn/core"
)

type Management struct {
	Conn      *Process
	connector core.OpenVpnConnector
	Mode      int
	events    chan []string
	clientEnv map[string]string
	shutdown  chan bool
}

func NewManagement(conn *Process, connector core.OpenVpnConnector) *Management {
	return &Management{
		connector: connector,
		Conn:      conn,
		events:    make(chan []string),
		clientEnv: make(map[string]string, 0),
		shutdown:  make(chan bool),
	}
}

func (m *Management) Start() error { // {{{
	err := m.connector.Connect()
	if err != nil {
		return err
	}
	m.connector.Listen(m.events)
	return nil
}

func (m *Management) Shutdown() {
	log.Info("Management: shutdown")
	close(m.shutdown)
	err := m.connector.Close()
	if err!=nil{
		glog.Error(err)
	}
}
