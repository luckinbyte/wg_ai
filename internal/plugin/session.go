package plugin

import (
    "encoding/binary"

    "github.com/luckinbyte/wg_ai/internal/session"
)

// SessionAdapter 会话推送适配器
type SessionAdapter struct {
    sess *session.Session
}

// NewSessionAdapter 创建会话适配器
func NewSessionAdapter(sess *session.Session) *SessionAdapter {
    return &SessionAdapter{sess: sess}
}

// Push 实现 SessionPush 接口 - 推送消息给客户端
func (a *SessionAdapter) Push(msgID uint16, data []byte) error {
    packet := makePacket(msgID, data)
    return a.sess.Send(packet)
}

// makePacket 构造消息包
// 格式: [总长度4字节][消息ID 2字节][消息体 N字节]
func makePacket(msgID uint16, body []byte) []byte {
    totalLen := 4 + 2 + len(body) // len + msgID + body
    packet := make([]byte, totalLen)

    // 写入总长度
    binary.BigEndian.PutUint32(packet[0:4], uint32(totalLen))
    // 写入消息ID
    binary.BigEndian.PutUint16(packet[4:6], msgID)
    // 写入消息体
    copy(packet[6:], body)

    return packet
}
