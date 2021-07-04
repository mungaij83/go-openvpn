package core

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"net"
	"net/textproto"
	"strings"
	"time"
)

type SocketConnector struct {
	socket     string
	password   string
	shutdown   chan bool
	events     chan string
	mode       int
	listener   net.Listener
	connection net.Conn
}

func NewSocketConnector(socket string, password string, mode int) OpenVpnConnector {
	return &SocketConnector{
		socket:   socket,
		events:   make(chan string, 10),
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
		glog.V(2).Infof("CMD IN: %v", command)
		s.events <- command
		return "", errors.New("OpenVPN server running in server mode, queued")
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

func (s *SocketConnector) Listen(events chan string) {
	if s.mode == ServerMode {
		go func() {
			for {
				fd, err := s.listener.Accept()
				glog.Info("Management: openvpn management interface have connected")
				if err != nil {
					select {
					case <-s.shutdown:
						glog.Info("Management: closed")
						break
					default:
						glog.Errorf("accept error: %v", err)
					}
					continue
				}

				go s.serve(fd, events)
			}
		}()
	} else {
		glog.Warningf("Socket Server running in client mode")
	}
}

func (s *SocketConnector) serve(c net.Conn, events chan string) {
	glog.V(2).Infof("Serving client: %v", c.RemoteAddr().String())
	reader := bufio.NewReader(c)
	tp := textproto.NewReader(reader)
	go func() {
		timer := time.NewTimer(time.Second * 5)
		for {
			select {
			case e := <-s.events:
				glog.V(4).Infof("CMD: %v", e)
				s.Write(c, e)
				break
			case <-timer.C:
				//s.Write(c, "status")
				break
			case <-s.shutdown:
				return
			}
		}
	}()
	for {
		line, err := tp.ReadLine()
		if err != nil {
			break
		}
		glog.V(8).Infof("Received MSG:%v", line)
		events <- line
	}
}

func (s *SocketConnector) Write(c net.Conn, cmd string) {
	glog.V(4).Infof("Received: %v", cmd)
	data := fmt.Sprintf("%s\n", strings.TrimSpace(cmd))
	d, err := c.Write([]byte(data))
	if err != nil {
		glog.Errorf("Failed to send command: %v", err)
	} else {
		glog.V(2).Infof("Command sent to the server[%d]: %s", d, strings.TrimSpace(cmd))
	}
}

func (s *SocketConnector) Close() error {
	close(s.shutdown)
	return s.connection.Close()
}
