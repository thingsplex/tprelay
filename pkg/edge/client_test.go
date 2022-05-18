package edge

import (
	"github.com/thingsplex/tprelay/pkg/proto/tunframe"
	"testing"
	"time"
)

func TestTunClient_Connect(t *testing.T) {
	tc := NewTunClient("ws://localhost:8090", "test-1", 1)

	if err := tc.Connect(); err != nil {
		t.Errorf("Connect() error = %v, ", err)
	}

	msg := tunframe.TunnelFrame{
		MsgType:   tunframe.TunnelFrame_WS_MSG,
		Headers:   nil,
		Params:    nil,
		ReqId:     0,
		CorrId:    0,
		ReqUrl:    "",
		ReqMethod: "",
		RespCode:  0,
		Payload:   []byte("hello world"),
	}

	tc.Send(&msg)
	tc.Close()

	time.Sleep(time.Second * 1)

}
