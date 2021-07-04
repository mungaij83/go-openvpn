package main

import (
	"fmt"
	"github.com/mungaij83/go-openvpn/core"
	"time"
)

type VpnDummyActions struct {
}

func (vd VpnDummyActions) HoldRelease() string {
	return "hold release"
}

func (vd VpnDummyActions) Usage(duration time.Duration) string {
	if duration.Seconds() > 0 {
		return fmt.Sprintf("bytecount %d", int64(duration.Seconds()))
	}
	return "bytecount 0"
}

func (vd VpnDummyActions) UnAuthenticated(data core.EventData) string {
	cid := data.Get("client_id")
	ckid := data.Get("key_id")
	return fmt.Sprintf("client-deny %s %s \"Invalid username or password\" [\"Invalid username or password\"]", cid, ckid)
}

func (vd VpnDummyActions) Authenticated(data core.EventData) string {
	usr := data.Get("client_id")
	ckid := data.Get("key_id")
	return fmt.Sprintf("client-auth-nt %s %s", usr, ckid)
}
