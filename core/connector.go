package core

const (
	ServerMode = 1 // OpenVPN running in server mode
	ClientMode = 2 // OpenVPN running in client mode
)

type OpenVpnConnector interface {
	Connect() error
	SendCommand(string) (string, error)
	Listen(events chan string)
	Close() error
}
