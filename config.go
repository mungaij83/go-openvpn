package openvpn

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	openssl "github.com/mungaij83/go-openvpn/core/ssl"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	remote     string
	port       int
	ipAddress  string
	socketPath string
	flags      map[string]bool
	values     map[string]string
	params     []string
}

func NewConfig(socket string) *Config {
	c := &Config{
		socketPath: socket,
		ipAddress:  "127.0.0.1",
		port:       7505,
		flags:      make(map[string]bool),
		values:     make(map[string]string),
		params:     make([]string, 0),
	}
	// Socket configuration
	c.Flag("management-signal")
	c.Flag("management-up-down")
	c.Flag("management-client")
	if socket != "" {
		c.Set("management", fmt.Sprintf("%s unix", socket))
		//c.Flag("management-hold")
		glog.Infof("Current config: %v", c)
	} else if c.ipAddress != "" {
		c.Set("management", fmt.Sprintf("%s %d", c.ipAddress, c.port))
	}
	return c
}

func (c *Config) Set(key, val string) {
	a := strings.Split("--"+key+" "+val, " ")
	for _, ar := range a {
		c.params = append(c.params, ar)
	}
}

func (c *Config) Flag(key string) {
	//c.params = append(c.params, "--"+key)
	a := strings.Split("--"+key, " ")
	for _, ar := range a {
		c.params = append(c.params, ar)
	}
}

func (c *Config) Validate() (config []string, err error) {
	return c.params, nil
}

func (c *Config) ServerMode(port int, ca *openssl.CA, cert *openssl.Cert, dh *openssl.DH, ta *openssl.TA) {
	c.Set("mode", "server")
	c.Set("port", strconv.Itoa(port))
	f := flag.Lookup("v")
	if f != nil {
		c.Set("verb", f.Value.String())
	} else {
		c.Set("verb", "3")
	}
	o, _ := os.Getwd()

	c.Set("cd", o)
	c.Set("ca", ca.GetFilePath())
	c.Set("crl-verify", ca.GetCRLPath())
	c.Set("cert", cert.GetFilePath())
	c.Set("key", cert.GetKeyPath())
	c.Set("dh", dh.GetFilePath())
	if ta != nil {
		c.Flag("tls-server")
		c.Set("tls-auth", ta.GetFilePath())
	}
}

func (c *Config) ClientMode(ca *openssl.CA, cert *openssl.Cert, dh *openssl.DH, ta *openssl.TA) {
	c.Flag("client")
	c.Flag("tls-client")

	c.Set("ca", ca.GetFilePath())
	c.Set("cert", cert.GetFilePath())
	c.Set("key", cert.GetKeyPath())
	c.Set("dh", dh.GetFilePath())
	c.Set("tls-auth", ta.GetFilePath())
}

func (c *Config) Remote(r string, port int) {
	c.Set("port", strconv.Itoa(port))
	c.Set("remote", r)
	c.remote = r
}
func (c *Config) Protocol(p string) {
	c.Set("proto", p)
}
func (c *Config) Device(t string) {
	c.Set("dev", t)
}

func (c *Config) IpPool(pool string) {

	ip, network, err := net.ParseCIDR(pool)
	if err != nil {
		glog.Error(err)
		return
	}

	c.Set("server", ip.String()+" "+strconv.Itoa(int(network.Mask[0]))+"."+strconv.Itoa(int(network.Mask[1]))+"."+strconv.Itoa(int(network.Mask[2]))+"."+strconv.Itoa(int(network.Mask[3])))
}

func (c *Config) Secret(key string) {
	c.Set("secret", key)
}
func (c *Config) Address(address string, port int) {
	c.ipAddress = address
	c.port = port
	c.Set("management", fmt.Sprintf("%s %d", c.ipAddress, c.port))
}

func (c *Config) Port() int {
	return c.port
}

func (c *Config) InterfaceAddress() string {
	return c.ipAddress
}

func (c *Config) KeepAlive(interval, timeout int) {
	c.Set("keepalive", strconv.Itoa(interval)+" "+strconv.Itoa(timeout))
}
func (c *Config) PingTimerRemote() {
	c.Flag("ping-timer-rem")
}
func (c *Config) PersistTun() {
	c.Flag("persist-tun")
}
func (c *Config) PersistKey() {
	c.Flag("persist-key")
}

func (c *Config) Compression() {
	//comp-lzo
}
func (c *Config) ClientToClient() {
	c.Flag("client-to-client")
}
