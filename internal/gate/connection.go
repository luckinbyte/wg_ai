package gate

import (
	"net"
	"time"
)

type Connection struct {
	conn       net.Conn
	remoteAddr string
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn:       conn,
		remoteAddr: conn.RemoteAddr().String(),
	}
}

func (c *Connection) ReadMessage() (msgType byte, payload []byte, err error) {
	return DecodePacket(c.conn)
}

func (c *Connection) WriteMessage(msgType byte, payload []byte) error {
	packet := EncodePacket(msgType, payload)
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := c.conn.Write(packet)
	return err
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) RemoteAddr() string {
	return c.remoteAddr
}

func (c *Connection) SetReadTimeout(d time.Duration) error {
	return c.conn.SetReadDeadline(time.Now().Add(d))
}
