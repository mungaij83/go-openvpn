package core

import "time"

type VpnActions interface {
	Authenticated(data EventData) string
	UnAuthenticated(data EventData) string
	HoldRelease(data EventData) string
	Usage(duration time.Duration) string
}
