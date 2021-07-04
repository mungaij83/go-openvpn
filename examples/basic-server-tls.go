package main

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/mungaij83/go-openvpn"
	"github.com/mungaij83/go-openvpn/core"
	openssl "github.com/mungaij83/go-openvpn/core/ssl"
	"log"
	"os"
	"time"
)


func main() {
	// This example first tries to load and if not found creates all the components needed for a TLS tunnel
	var err error
	var ca *openssl.CA
	var cert *openssl.Cert
	var dh *openssl.DH
	//var ta *openssl.TA

	ssl := openssl.Openssl{
		Path:         "certs", // A storage folder, where to store all certs
		Country:      "SE",
		Province:     "Example provice",
		City:         "Example city",
		Organization: "Example organization",
		CommonName:   "Example commonname",
		Email:        "Example email",
	}

	if ca, err = ssl.LoadOrCreateCA("ca.crt", "ca.key"); err != nil {
		log.Println("LoadOrCreateCA failed: ", err)
		return
	}
	// Note the last bool parameter! This is important beacuse it will generate a "server"-cert
	if cert, err = ssl.LoadOrCreateCert("server/server.crt", "server/server.key", "server", ca, true); err != nil {
		log.Println("LoadOrCreateCert failed: ", err)
		return
	}
	if dh, err = ssl.LoadOrCreateDH("DH2048.pem", 2048); err != nil {
		log.Println("LoadOrCreateDH failed: ", err)
		return
	}
	//if ta, err = ssl.LoadOrCreateTA("TA.key"); err != nil {
	//	log.Println("LoadOrCreateTA failed: ", err)
	//	return
	//}
	processFile := fmt.Sprintf("/tmp/management-server-%d.sock", os.Getpid())
	c := openvpn.NewConfig(processFile)
	c.Set("proto", "tcp")
	c.Device("tun")
	c.Flag("client-cert-not-required")
	c.Flag("management-client-auth")
	c.ServerMode(1195, ca, cert, dh, nil)
	c.IpPool("10.255.255.0/24")

	c.KeepAlive(10, 60)
	c.PingTimerRemote()
	c.PersistTun()
	c.PersistKey()
	// Create the openvpn instance
	p := openvpn.NewProcess(processFile, c)

	managemnt := openvpn.NewVpnManagement("127.0.0.1", processFile, 17505, "", core.ServerMode)
	err = managemnt.StartServer()
	if err != nil {
		glog.Error(err)
		os.Exit(-1)
	}
	// Wait for management interface to startup
	time.Sleep(time.Second * 5)
	// Start process
	err = p.Start()
	if err != nil {
		glog.Error(err)
		os.Exit(-3)
	}
	actioner := VpnDummyActions{}
	// Turn on real time Events
	managemnt.Exec("status on")
	managemnt.Exec("echo on")
	managemnt.Exec(actioner.Usage(time.Second * 4))
	// Listen for events
	for {
		select {
		case event := <-managemnt.Events:
			nm := event.EventName()
			switch nm {
			case "CLIENT_CONNECT", "CLIENT_REAUTH":
				glog.Infof("Authenticating Event: %s (%v)", event.Event, event.EventName())
				pwd := event.Get("password")
				user := event.Get("username")
				if pwd == "test" && user == "test" {
					glog.Infof("Authenticated Event: %s (%v) -> %s", event.Event, event.EventName(), user)
					managemnt.Exec(actioner.Authenticated(event))
				} else {
					glog.Infof("Authentication Failed Event: %s (%v) -> %s", event.Event, event.EventName(), user)
					managemnt.Exec(actioner.UnAuthenticated(event))
				}
				break
			case "CLIENT_DISCONNECTED":
				glog.V(1).Infof("Client closed session: %b", event.Data)
				break
			case "HOLD":
				managemnt.Exec(actioner.HoldRelease())
				break
			case "CLIENT_LIST":
				clients, err := managemnt.GetClients(event)
				if err != nil {
					glog.Error(err)
				} else {
					glog.Infof("Clients: %+v", clients)
				}
				break
			default:
				glog.Infof("Other Events[%s]: %v", event.EventName(),event )
			}

		}
	}
}
