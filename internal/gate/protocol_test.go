package gate

import (
	"bytes"
	"testing"
)

func TestPacketEncodeDecode(t *testing.T) {
	payload := []byte("hello world")

	// Encode
	packet := EncodePacket(MsgTypeRequest, payload)

	// Verify length field
	length := int(packet[0])<<24 | int(packet[1])<<16 | int(packet[2])<<8 | int(packet[3])
	if length != len(payload)+1 {
		t.Errorf("expected length %d, got %d", len(payload)+1, length)
	}

	// Decode
	reader := bytes.NewReader(packet)
	msgType, data, err := DecodePacket(reader)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if msgType != MsgTypeRequest {
		t.Errorf("expected msgType %d, got %d", MsgTypeRequest, msgType)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got %s", string(data))
	}
}

func TestDifferentMessageTypes(t *testing.T) {
	types := []byte{MsgTypeRequest, MsgTypeResponse, MsgTypePush, MsgTypeHandshake}
	for _, mt := range types {
		packet := EncodePacket(mt, []byte("test"))
		reader := bytes.NewReader(packet)
		gotType, _, err := DecodePacket(reader)
		if err != nil {
			t.Errorf("decode failed for type %d: %v", mt, err)
		}
		if gotType != mt {
			t.Errorf("expected type %d, got %d", mt, gotType)
		}
	}
}
