package core

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"net"
	"net/textproto"
	"strings"
)

type TcpConnector struct {
	port       int
	mode       int
	ipAddress  string
	password   string
	shutdown   chan bool
	events     chan string
	connection net.Conn
	listener   net.Listener
}

func NewTcpConnector(ipAddress string, port int, password string, mode int) OpenVpnConnector {
	return &TcpConnector{
		shutdown:  make(chan bool),
		events:    make(chan string, 10),
		mode:      mode,
		port:      port,
		ipAddress: ipAddress,
		password:  password,
	}
}
func (s TcpConnector) GetManagementAddress() string {
	return fmt.Sprintf("%s:%d", s.ipAddress, s.port)
}

func (s *TcpConnector) Connect() error {
	var err error
	if s.mode == ServerMode {
		glog.V(2).Infof("OpenVPN is started in server mode")
		s.listener, err = net.Listen("tcp", s.GetManagementAddress())
		if err != nil {
			glog.Error(err)
			return err
		}
		glog.V(2).Infof("OpenVPN started on: %v", s.GetManagementAddress())
	} else {
		glog.V(2).Infof("OpenVPN is started in client mode")
		c, err := net.Dial("tcp", "127.0.0.1:17505")
		if err != nil {
			glog.Error(err)
			return err
		}
		s.connection = c
		glog.V(2).Infof("Client connected: %v", c.RemoteAddr().String())
	}
	return nil
}

func (s *TcpConnector) SendCommand(command string) (string, error) {
	if s.mode == ServerMode {
		s.events <- command
		return "", errors.New("OpenVPN server running in server mode")
	}
	cmdStr := fmt.Sprintf("%s\n", strings.TrimSpace(command))
	_, err := s.connection.Write([]byte(cmdStr))
	if err != nil {
		glog.V(2).Infof("Failed to write: %v", err)
		return "", err
	} else {
		glog.V(2).Infof("TCP: Sent status command")
		message, err := bufio.NewReader(s.connection).ReadString('\n')
		if err != nil {
			glog.Infof("Failed to read response")
			return "", err
		}
		glog.V(2).Infof("TCP response: %v", message)
		return message, nil
	}
}

func (s *TcpConnector) serve(c net.Conn, events chan string) {
	glog.V(2).Infof("Serving client: %v", c.RemoteAddr().String())
	reader := bufio.NewReader(c)
	go func() {
		for {
			select {
			case <-s.shutdown:
				return
			case e := <-s.events:
				command := fmt.Sprintf("%s\n", strings.TrimSpace(e))
				_, err := c.Write([]byte(command))
				if err != nil {
					glog.Error(err)
				} else {
					glog.Infof(">>CMD OUT: %v", command)
				}

			}
		}
	}()
	tp := textproto.NewReader(reader)
	for {
		line, err := tp.ReadLine()
		if err != nil {
			break
		}
		events <- line
	}
}

func (s *TcpConnector) Listen(events chan string) {
	if s.mode == ServerMode {
		go func() {
			for {
				conn, err2 := s.listener.Accept()
				if err2 != nil {
					glog.Error(err2)
				} else {
					go s.serve(conn, events)
				}
			}
		}()
	} else {
		glog.Warningf("TCP Server running in client mode")
	}
}

func (s TcpConnector) Close() error {
	close(s.shutdown)

	if s.connection != nil {
		return s.connection.Close()
	}

	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
