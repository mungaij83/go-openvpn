package main

import (
	"github.com/mungaij83/go-openvpn"
)

// Short-hands for some basic openvpn operating modes


func NewStaticKeyServer(key string, socket string) *openvpn.Process {
	c := openvpn.NewConfig(socket)
	p := openvpn.NewProcess(socket,c)


	c.Device("tun")
	//c.IpConfig("10.8.0.1", "10.8.0.2")
	c.Secret(key)

	c.KeepAlive(10, 60)
	c.PingTimerRemote()
	c.PersistTun()
	c.PersistKey()

	p.SetConfig(c)
	return p
}

func NewStaticKeyClient(remote, key, socket string) *openvpn.Process {
	c := openvpn.NewConfig(socket)
	p := openvpn.NewProcess(socket,c)


	c.Remote(remote, 1194)
	c.Device("tun")
	//c.IpConfig("10.8.0.2", "10.8.0.1")
	c.Secret(key)

	c.KeepAlive(10, 60)
	c.PingTimerRemote()
	c.PersistTun()
	c.PersistKey()

	p.SetConfig(c)
	return p
}
func main() {
	// A custom config example

	c := openvpn.NewConfig("")
	c.Set("config", "myconfigfile.conf")

	// Create the openvpn instance
	p := openvpn.NewProcess("",c)
	p.SetConfig(c)

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
