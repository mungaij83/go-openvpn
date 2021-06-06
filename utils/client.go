package utils

type Client struct {
	CommonName       string
	PublicIP         string
	PrivateIP        string
	BytesRecived     int64
	BytesSent        int64
	LastRef          string
	missing          int
	Env              map[string]string
}
