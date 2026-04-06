package plugin

import (
    "net"
    "testing"

    "github.com/luckinbyte/wg_ai/internal/session"
)

func TestSessionAdapterPush(t *testing.T) {
    // 创建 mock connection
    server, client := net.Pipe()
    defer server.Close()
    defer client.Close()

    sess := session.New(1, 1, server)
    adapter := NewSessionAdapter(sess)

    // 测试推送
    done := make(chan []byte, 1)
    go func() {
        buf := make([]byte, 1024)
        n, _ := client.Read(buf)
        done <- buf[:n]
    }()

    err := adapter.Push(1001, []byte("test"))
    if err != nil {
        t.Fatal(err)
    }

    data := <-done
    if len(data) == 0 {
        t.Error("expected data")
    }
}
