package gate

import (
	"encoding/binary"
	"io"
)

const (
	PacketHeaderSize = 4 // length field size
	MsgTypeSize      = 1

	MsgTypeRequest   = 0x01
	MsgTypeResponse  = 0x02
	MsgTypePush      = 0x03
	MsgTypeHandshake = 0x04
)

// EncodePacket creates a packet: [Length(4B)][MsgType(1B)][Payload]
func EncodePacket(msgType byte, payload []byte) []byte {
	totalLen := MsgTypeSize + len(payload)
	buf := make([]byte, PacketHeaderSize+totalLen)

	binary.BigEndian.PutUint32(buf[0:4], uint32(totalLen))
	buf[4] = msgType
	copy(buf[5:], payload)

	return buf
}

// DecodePacket reads and decodes a packet from reader
func DecodePacket(reader io.Reader) (msgType byte, payload []byte, err error) {
	// Read length header
	header := make([]byte, PacketHeaderSize)
	if _, err = io.ReadFull(reader, header); err != nil {
		return
	}

	totalLen := binary.BigEndian.Uint32(header)
	if totalLen < MsgTypeSize {
		err = io.ErrUnexpectedEOF
		return
	}

	// Read message type + payload
	data := make([]byte, totalLen)
	if _, err = io.ReadFull(reader, data); err != nil {
		return
	}

	msgType = data[0]
	payload = data[1:]
	return
}
