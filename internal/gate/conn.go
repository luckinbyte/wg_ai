package gate

import "errors"

var (
	ErrInvalidMessage = errors.New("invalid message format")
)

// Conn 连接接口 (TCP和WS通用)
type Conn interface {
	ReadMessage() (msgType byte, payload []byte, err error)
	WriteMessage(msgType byte, payload []byte) error
	Close() error
	RemoteAddr() string
}
