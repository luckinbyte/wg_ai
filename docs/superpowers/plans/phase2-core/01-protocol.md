# Task 8: Protocol Codec - 协议编解码

## 背景与目标

实现网络数据包的编解码器，定义二进制协议格式。

**协议格式：**
```
+----------------+----------------+------------------+
|  Length (4B)   |  MsgType (1B)  |  Protobuf Data   |
+----------------+----------------+------------------+
```

**为什么需要这个任务：**
- 客户端和服务器需要统一的通信协议
- 二进制协议比文本协议更高效
- 长度前缀解决 TCP 粘包问题

**输出：**
- `internal/gate/protocol.go` - 编解码实现

## 依赖

- Task 1: 项目结构已创建

## 步骤

### Step 1: Write the failing test

Create `internal/gate/protocol_test.go`:

```go
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
```

### Step 2: Run test to verify it fails

```bash
cd /root/ai_project/wg_ai
go test ./internal/gate/...
```

Expected: FAIL - undefined: EncodePacket, DecodePacket, MsgTypeRequest

### Step 3: Write minimal implementation

Create `internal/gate/protocol.go`:

```go
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
```

### Step 4: Run test to verify it passes

```bash
cd /root/ai_project/wg_ai
go test ./internal/gate/...
```

Expected: PASS

### Step 5: Commit

```bash
git add .
git commit -m "feat: add protocol encoder/decoder"
```

## 验证

```bash
cd /root/ai_project/wg_ai
go test ./internal/gate/... -v
```

Expected: PASS

## 完成标志

- [ ] 测试通过
- [ ] protocol.go 包含 EncodePacket 和 DecodePacket
- [ ] 定义了 MsgType 常量
- [ ] Commit 完成
