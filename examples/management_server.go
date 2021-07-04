package main

import (
	"github.com/golang/glog"
	"github.com/mungaij83/go-openvpn"
	"github.com/mungaij83/go-openvpn/core"
	openssl "github.com/mungaij83/go-openvpn/core/ssl"
	"log"
	"os"
	"time"
)

func main() {
	var err error
	var ca *openssl.CA
	var cert *openssl.Cert
	var dh *openssl.DH
	cfg := openvpn.NewConfig("")
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
	cfg.Set("proto", "tcp")
	cfg.Device("tun")
	cfg.Set("verify-client-cert", "none")
	cfg.Flag("management-client-auth")
	cfg.ServerMode(1196, ca, cert, dh, nil)
	cfg.Address("127.0.0.1", 17505)
	cfg.IpPool("10.255.255.0/24")
	cfg.Flag("management-hold")

	process := openvpn.NewProcess("", cfg)
	managemnt := openvpn.NewVpnManagement(cfg.InterfaceAddress(), "", cfg.Port(), "", core.ServerMode)
	err = managemnt.StartServer()
	if err != nil {
		glog.Errorf("Failed to connect: %v", err)
		os.Exit(-1)
	}
	err = process.Start()
	if err != nil {
		glog.Error(err)
		os.Exit(-1)
	}
	actioner := VpnDummyActions{}
	// Query status once every 5 seconds
	timmer := time.NewTicker(time.Second * 5)
	count := 0
	for {
		select {
		case _ = <-timmer.C:
			count += 1
			st, err := managemnt.Status()
			if err != nil {
				glog.Error(err)
			} else {
				glog.Infof("Status: %v", st)
			}
			break
		case event := <-managemnt.Events:
			glog.Infof("Event received: %+v", event)
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
				glog.Infof("Other Events[%s]: %v", event.EventName(), event)
			}
			break
		}
	}
	//managemnt.Shutdown()
}
