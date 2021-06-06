package core

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"net"
	"net/textproto"
	"strings"
)

type SocketConnector struct {
	socket     string
	password   string
	shutdown   chan bool
	mode       int
	listener   net.Listener
	connection net.Conn
}

func NewSocketConnector(socket string, password string, mode int) OpenVpnConnector {
	return &SocketConnector{
		socket:   socket,
		password: password,
		mode:     mode,
		shutdown: make(chan bool),
	}
}

func (s *SocketConnector) Connect() error {
	if s.mode == ServerMode {
		l, err := net.Listen("unix", s.socket)
		if err != nil {
			glog.Error(err)
			return err
		}
		s.listener = l
	} else {
		c, err := net.Dial("unix", s.socket)
		if err != nil {
			return err
		}
		s.connection = c
	}
	return nil
}

func (s *SocketConnector) SendCommand(command string) (string, error) {
	if s.mode == ServerMode {
		return "", errors.New("OpenVPN server running in server mode")
	}
	cmdStr := fmt.Sprintf("%s\n", strings.TrimSpace(command))
	_, err := s.connection.Write([]byte(cmdStr))
	if err != nil {
		glog.Infof("Failed to write: %v", err)
		return "", err
	} else {
		glog.Infof("Sent status command")
		message, err := bufio.NewReader(s.connection).ReadString('\n')
		if err != nil {
			glog.Infof("Failed to read response")
			return "", err
		}
		glog.V(2).Infof("Client response: %v", message)
		return message, nil
	}
}

func (s *SocketConnector) Listen(events chan []string) {
	if s.mode == ServerMode {
		go func() {
			for {
				fd, err := s.listener.Accept()
				glog.Info("Management: openvpn management interface have connected")
				if err != nil {
					select {
					case <-s.shutdown:
						glog.Info("Management: closed")
					default:
						glog.Errorf("accept error: %v", err)
					}
					return
				}

				go s.serve(fd, events)
			}
		}()
	}
}

func (s *SocketConnector) serve(c net.Conn, events chan []string) {
	reader := bufio.NewReader(c)
	tp := textproto.NewReader(reader)
	bufer := bytes.Buffer{}
	for {
		line, err := tp.ReadLine()
		if err != nil {
			break
		}
		bufer.WriteString(line)
	}
	lines := bufer.String()
	events <- []string{lines}
}

func (s *SocketConnector) Close() error {
	close(s.shutdown)
	return s.connection.Close()
}
