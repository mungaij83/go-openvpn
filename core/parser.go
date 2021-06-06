package core

import (
	"bytes"
	"errors"
	"github.com/golang/glog"
	"github.com/mungaij83/go-openvpn/utils"
	"regexp"
	"strconv"
	"strings"
)

const (
	ClientList = "(?ms)OpenVPN CLIENT LIST\n" +
		"Updated,([^\n]*)\n" +
		"(.*)\n" +
		"ROUTING TABLE\n" +
		"(.*)\n" +
		"GLOBAL STATS\n" +
		"(.*)" +
		"\nEND\n"
)

type EventData struct {
	Event     string
	EventType string
	Realtime  bool
	EventData string
}

func (ed EventData) EventName() string {
	tmp := ed.Event + "_" + ed.EventType
	return strings.TrimSuffix(tmp, "_")
}

type CommandParser struct {
	ClientListReg *regexp.Regexp
}

func NewCommandParser() CommandParser {
	reg, _ := regexp.Compile(ClientList)
	return CommandParser{
		ClientListReg: reg,
	}
}

func (cp CommandParser) Join(evt []string, start int, sep string) string {
	b := bytes.Buffer{}
	for i := start; i < len(evt); i++ {
		if i != start {
			b.WriteString(sep)
		}
		b.WriteString(evt[i])
	}
	return strings.TrimSuffix(b.String(), sep)
}

func (cp CommandParser) ParseEvent(evt string) EventData {
	dt := EventData{}
	el := strings.Split(evt, ":")
	if strings.HasPrefix(evt, ">") {
		dt.Realtime = true
		dt.Event = strings.TrimPrefix(el[0], ">")
		dt.EventData = cp.Join(el, 1, ":")
	} else if strings.HasPrefix(evt, "OpenVPN") {
		dt.Event = "CLIENT_LIST"
		dt.EventData = strings.TrimSuffix(evt, "OpenVPN CLIENT LIST")
	} else {
		dt.Event = el[0]
		dt.EventData = cp.Join(el, 1, ":")
	}

	// Add type
	switch dt.Event {
	case "CLIENT":
		typed := strings.Split(el[1], ",")
		dt.EventType = typed[0]
		dt.EventData = cp.Join(typed, 1, ",")
		break
	default:
		dt.EventType = ""
	}
	return dt
}

func (cp CommandParser) ParseClients(out string) ([]utils.Client, error) {
	match := cp.ClientListReg.FindAllStringSubmatch(out, -1)
	if len(match) == 0 {
		return nil, errors.New("no client found")
	}
	for ix, il := range match {
		glog.V(2).Infof("Index[%d]: %v", ix, il)
	}
	return cp.clientList(match[0][2])
}

func (cp *CommandParser) clientList(match string) ([]utils.Client, error) { // {{{
	if len(match) < 3 {
		glog.Errorf("Invalid client list, regexp failed: %v", match)
		return nil, errors.New("invalid client list")
	}
	clients := cp.makeCsvList(match)

	clientsList := make([]utils.Client, len(clients))
	for _, c := range clients {
		cc := utils.Client{
			CommonName:   c["Common Name"],
			PublicIP:     c["Real Address"],
			PrivateIP:    c["Virtual Address"],
			BytesRecived: cp.parseInt64(c["Bytes Received"]),
			BytesSent:    cp.parseInt64(c["Bytes Sent"]),
			LastRef:      c["Last Ref"],
			Env:          nil,
		}
		clientsList = append(clientsList, cc)
	}
	return clientsList, nil
}

func (CommandParser) parseInt64(d string) int64 {
	n, err := strconv.ParseInt(d, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func (CommandParser) makeCsvList(data string) (list []map[string]string) { // {{{
	list = make([]map[string]string, 0)

	rows := strings.Split(data, "\n")

	cols := strings.Split(rows[0], ",")

	for i, row := range rows[1:] {
		values := strings.Split(row, ",")

		list = append(list, make(map[string]string, 0))

		for c, col := range cols {
			list[i][col] = values[c]
		}
	}
	return
}
