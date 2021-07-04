package main

import (
	"github.com/mungaij83/go-openvpn"
	openssl "github.com/mungaij83/go-openvpn/core/ssl"
	"log"
)

func main() {
	// This example first tries to load and if not found creates all the components needed for a TLS tunnel

	var err error
	var ca *openssl.CA
	var cert *openssl.Cert
	var dh *openssl.DH
	var ta *openssl.TA

	ssl := openssl.Openssl{
		Path: "certs", // A storage folder, where to store all certs

		Country:      "SE",
		Province:     "Example provice",
		City:         "Example city",
		Organization: "Example organization",
		CommonName:   "Example client",
		Email:        "Example email",
	}

	if ca, err = ssl.LoadOrCreateCA("ca.crt", "ca.key"); err != nil {
		log.Println("LoadOrCreateCA failed: ", err)
		return
	}
	// Note the last bool parameter! This is important beacuse it will generate a "client"-cert
	if cert, err = ssl.LoadOrCreateCert("clients/client1.crt", "clients/client1.key", "client1", ca, false); err != nil {
		log.Println("LoadOrCreateCert failed: ", err)
		return
	}
	if dh, err = ssl.LoadOrCreateDH("DH1024.pem", 1024); err != nil {
		log.Println("LoadOrCreateDH failed: ", err)
		return
	}
	if ta, err = ssl.LoadOrCreateTA("TA.key"); err != nil {
		log.Println("LoadOrCreateTA failed: ", err)
		return
	}
	c := openvpn.NewConfig("")

	c.ClientMode(ca, cert, dh, ta)
	c.Remote("remote", 1194)
	c.Device("tun")

	c.KeepAlive(10, 60)
	c.PingTimerRemote()
	c.PersistTun()
	c.PersistKey()
	// Create the openvpn instance
	p := openvpn.NewProcess("localhost", c)

	// Start the process
	p.Start()

	// Listen for events
	//for {
	//	select {
	//	case event := <-p.Events:
	//		log.Println("Event: ", event.Name, "(", event.Args, ")")
	//	}
	//}
}
