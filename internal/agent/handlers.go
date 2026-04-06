package agent

import (
	"google.golang.org/protobuf/proto"

	cs "github.com/luckinbyte/wg_ai/proto/cs"
)

// RegisterDefaultHandlers registers built-in message handlers
func RegisterDefaultHandlers(a *Agent) {
	a.RegisterHandler(1002, handleHeartbeat) // MSG_ID_HEARTBEAT
}

func handleHeartbeat(a *Agent, msg *Message) ([]byte, error) {
	resp := &cs.HeartbeatResponse{}
	return proto.Marshal(resp)
}
