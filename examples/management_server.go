package main

import (
	"github.com/golang/glog"
	"github.com/mungaij83/go-openvpn"
	"github.com/mungaij83/go-openvpn/core"
	"os"
	"time"
)

func main() {
	managemnt := go_openvpn.NewVpnManagement("127.0.0.1", "", 17505, "", core.ClientMode)
	err := managemnt.StartClient()
	if err != nil {
		glog.Errorf("Failed to connect: %v", err)
		os.Exit(-1)
	}
	// Query status once every 5 seconds
	count := 0
	for {
		if count > 5 {
			break
		}
		count += 1
		st, err := managemnt.Status()
		if err != nil {
			glog.Error(err)
		} else {
			glog.Infof("Status: %v", st)
		}
		time.Sleep(time.Second * 5)
	}
	managemnt.Shutdown()
}
