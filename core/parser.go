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

var (
	ClientListReg, _ = regexp.Compile("(?ms)OpenVPN CLIENT LIST\n" +
		"Updated,([^\n]*)\n" +
		"(.*)\n" +
		"ROUTING TABLE\n" +
		"(.*)\n" +
		"GLOBAL STATS\n" +
		"(.*)" +
		"\nEND\n")
	ClientEnv, _ = regexp.Compile("([^=\r\n]+)=([^\r\n]*)")
)

type EventData struct {
	Event     string
	EventType string
	Completed bool
	HasEnd    bool
	Realtime  bool
	Invalid   bool
	Data      map[string]string
	EventData string
}

func (ed EventData) EventName() string {
	tmp := ed.Event + "_" + ed.EventType
	return strings.TrimSuffix(tmp, "_")
}

func (ed *EventData) Merge(data EventData) {
	ed.Completed = data.Completed
	ed.Invalid = data.Invalid
	for k, v := range data.Data {
		ed.Data[k] = v
	}
}

func (ed EventData) Get(k string) string {
	if len(ed.Data) > 0 {
		val, ok := ed.Data[k]
		if ok {
			return val
		}
	}
	return ""
}

type CommandParser struct {
	ClientListReg *regexp.Regexp
	buffer        *EventData
	dataBuffer    *bytes.Buffer
}

func NewCommandParser() CommandParser {
	return CommandParser{
		dataBuffer:    bytes.NewBufferString(""),
		ClientListReg: ClientListReg,
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

func (cp *CommandParser) ParseEvent(evt string) *EventData {
	// Handle listing events
	if cp.buffer != nil && cp.buffer.Event == "CLIENT_LIST" {
		cp.dataBuffer.WriteString(evt)
		if strings.Compare(strings.TrimSpace(evt), "END") == 0 {
			cp.buffer.EventData = cp.dataBuffer.String()
			cp.dataBuffer.Reset()
			dt := *cp.buffer
			cp.buffer = nil
			return &dt
		}
		return nil
	}
	// Handle other events
	dt := EventData{
		Data: make(map[string]string),
	}
	el := strings.Split(evt, ":")
	if strings.HasPrefix(evt, ">") {
		dt.Realtime = true
		dt.Event = strings.TrimPrefix(el[0], ">")
		dt.EventData = cp.Join(el, 1, ":")
	} else if strings.HasPrefix(evt, "OpenVPN") {
		dt.Event = "CLIENT_LIST"
		dt.HasEnd = true
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
		err := cp.ParseClient(&dt)
		if err != nil {
			glog.Error(err)
		}
		break
	case "BYTECOUNT":
		s := strings.Split(dt.EventData, ",")
		dt.Data["bytes_in"] = s[0]
		dt.Data["bytes_out"] = s[1]
		dt.Completed = true
		break
	case "BYTECOUNT_CLI":
		s := strings.Split(dt.EventData, ",")
		dt.Data["client_id"] = s[0]
		dt.Data["bytes_in"] = s[1]
		dt.Data["bytes_out"] = s[2]
		dt.Completed = true
		break
	case "CLIENT_LIST":
		cp.dataBuffer.WriteString(evt)
		dt.HasEnd = true
		break
	default:
		dt.Completed = true
		dt.EventType = ""
	}
	// Check for event
	if dt.Completed || dt.Invalid {
		glog.V(2).Infof("Returning event: %v", dt)
		if dt.Completed && cp.buffer != nil {
			cp.buffer.Merge(dt)
			dt = *cp.buffer
			cp.buffer = nil
		}
		return &dt
	} else {
		glog.V(2).Infof("Saving state: %v", dt)
		// Update state
		if cp.buffer != nil {
			cp.buffer.Merge(dt)
		} else {
			cp.buffer = &dt
		}
	}
	return nil
}
func (cp *CommandParser) ParseClient(data *EventData) error {
	switch data.EventType {
	case "ENV":
		if data.EventData == "END" {
			data.Completed = true
		} else {
			s := strings.Split(data.EventData, "=")
			data.Data[s[0]] = s[1]
		}
		break
	case "ADDRESS":
		s := strings.Split(data.EventData, ",")
		data.Data["client_id"] = s[0]
		data.Data["client_address"] = s[1]
		data.Data["primary_address"] = s[2]
		break
	case "DISCONNECT":
		data.HasEnd = true
		data.Data["client_id"] = data.EventData
		break
	case "ESTABLISHED":
		data.Data["client_id"] = data.EventData
		break
	case "REAUTH":
		s := strings.Split(data.EventData, ",")
		data.Data["client_id"] = s[0]
		data.Data["client_key_id"] = s[1]
		data.HasEnd = true
		break
	case "CONNECT":
		data.HasEnd = true
		s := strings.Split(data.EventData, ",")
		data.Data["client_id"] = s[0]
		data.Data["key_id"] = s[1]
		break
	default:
		data.Invalid = true
		glog.Warningf("Invalid client request: %v", data)
	}

	return nil
}

func (cp *CommandParser) ParseStatus(out string) ([]utils.Client, error) {
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
