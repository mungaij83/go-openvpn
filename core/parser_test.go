package core

import (
	"flag"
	"os"
	"testing"
)

var parser CommandParser

//func init() {
//
//}

func TestMain(m *testing.M) {
	_ = flag.Set("v", "3")
	_ = flag.Set("logtostderr", "true")
	flag.Parse()
	parser = NewCommandParser()
	os.Exit(m.Run())
}

func TestParseBytes(t *testing.T) {
	out := ">BYTECOUNT:3,5"
	evt := parser.ParseEvent(out)
	t.Logf("%+v", evt)
}

func TestCLientLine(t *testing.T) {
	out:=">CLIENT:CONNECT,89,ioosd"
	evt := parser.ParseEvent(out)
	t.Logf("%+v", evt)
}

func TestParseClients(t *testing.T) {
	out:=[]string{
		"OpenVPN CLIENT LIST\n",
		"Updated, Thu Feb 13 23:39:20 2014\n",
		"Common Name,Real Address,Bytes Received,Bytes Sent,Connected Since\n",
		"VPN_client,10.13.156.4:1194,12563,14885,Thu Feb 13 23:39:20 2014\n",
		"ROUTING TABLE\n",
		"Virtual Address,Common Name,Real Address,Last Ref\n",
		"192.168.11.4,VPN_client,10.13.156.4:1194,Thu Feb 13 23:39:20 2014\n",
		"GLOBAL STATS\n",
		"Max bcast/mcast queue length,0\n",
		"END\n"}
	var evt *EventData
	for _,t:=range out {
		evt=parser.ParseEvent(t)
		if evt!=nil {
			break
		}
	}
	if evt==nil{
		t.Fatal("Failed to parse output")
	} else {
		t.Logf("EVENT: %+v", evt)
		clients, err := parser.ParseStatus(evt.EventData)
		if err != nil {
			t.Fatal(err)
		} else {
			t.Logf("Clients: %+v", clients)
		}
	}
}

//func TestParse(t *testing.T) {
//	parser:=core.NewCommandParser()
//	done := make(chan bool)
//	go func() {
//		waitGroup.Add(1)
//		defer waitGroup.Done()
//
//		select {
//		case result := <-m.events:
//			for index := range result {
//				t.Log("Result[", index, "]: \n", strconv.Quote(result[index]))
//			}
//
//			if len(result) != 6 {
//				t.Error("Wrong length on answer, should be 5, is ", len(result))
//			}
//
//			if result[0] != "client-list" {
//				t.Error("result[0] is invalid")
//			}
//			if result[1] !=  {
//				t.Error("result[1] is invalid")
//			}
//			if result[2] != " Thu Feb 13 23:39:20 2014" {
//				t.Error("result[2] is invalid")
//			}
//			if result[3] !=
//				"Common Name,Real Address,Bytes Received,Bytes Sent,Connected Since\n"+
//					"VPN_client,10.13.156.4:1194,12563,14885,Thu Feb 13 23:39:20 2014" {
//				t.Error("result[3] is invalid")
//			}
//			if result[4] !=
//				"Virtual Address,Common Name,Real Address,Last Ref\n"+
//					"192.168.11.4,VPN_client,10.13.156.4:1194,Thu Feb 13 23:39:20 2014" {
//				t.Error("result[4] is invalid")
//			}
//			if result[5] != "Max bcast/mcast queue length,0" {
//				t.Error("result[5] is invalid")
//			}
//
//			return
//		case <-done:
//			t.Error("Parse done without result")
//		}
//	}()
//
//	m.parse([]byte("OpenVPN CLIENT LIST"), false)
//	m.parse([]byte("Updated, Thu Feb 13 23:39:20 2014"), false)
//	m.parse([]byte("Common Name,Real Address,Bytes Received,Bytes Sent,Connected Since"), false)
//	m.parse([]byte("VPN_client,10.13.156.4:1194,12563,14885,Thu Feb 13 23:39:20 2014"), false)
//	m.parse([]byte(""), false)
//	m.parse([]byte("ROUTING TABLE"), false)
//	m.parse([]byte("Virtual Address,Common Name,Real Address,Last Ref"), false)
//	m.parse([]byte("192.168.11.4,VPN_client,10.13.156.4:1194,Thu Feb 13 23:39:20 2014"), false)
//	m.parse([]byte(""), false)
//	m.parse([]byte("GLOBAL STATS"), false)
//	m.parse([]byte("Max bcast/mcast queue length,0"), false)
//	m.parse([]byte("END"), false)
//
//	close(done)
//
//	waitGroup.Wait()
//}