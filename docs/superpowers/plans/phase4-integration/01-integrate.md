# Task 22: Integrate Components - 组件集成

## 背景

将 TCP Server、Session Manager、Agent Manager 完整集成，实现完整的消息处理流程。

## 步骤

### Step 1: Update TCPServer handleConnection

确保 `internal/gate/tcp_server.go` 中的 handleConnection 正确集成：

```go
func (s *TCPServer) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	c := NewConnection(conn)
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	// Read handshake
	msgType, data, err := c.ReadMessage()
	if err != nil {
		return
	}
	if msgType != MsgTypeHandshake {
		return
	}

	// Parse handshake token (simplified)
	// In production, validate via login service

	// Create session
	sess := s.sessionMgr.Create(1, conn) // UID=1 for now

	// Bind to agent
	agent := s.agentMgr.Assign()
	agent.BindSession(sess)

	// Send handshake response
	resp := []byte{0, 0, 0, 0}
	c.WriteMessage(MsgTypeResponse, resp)

	// Message loop
	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		msgType, data, err := c.ReadMessage()
		if err != nil {
			break
		}

		if msgType == MsgTypeRequest && len(data) >= 2 {
			msgID := uint16(data[0])<<8 | uint16(data[1])
			agent.Push(&agent.Message{
				MsgID:   msgID,
				Payload: data[2:],
				Sess:    sess,
			})
		}
	}

	// Cleanup
	agent.UnbindSession(sess.RID)
	s.sessionMgr.Remove(sess.RID)
}
```

### Step 2: Commit

```bash
git add .
git commit -m "feat: integrate TCP server with session and agent"
```

## 完成标志

- [ ] handleConnection 完整实现
- [ ] 测试通过
