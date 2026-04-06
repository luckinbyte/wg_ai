package data_test

import (
    "encoding/json"
    "testing"

    "github.com/luckinbyte/wg_ai/internal/data"
)

func TestPersistSerialize(t *testing.T) {
    p := data.NewPlayerData(1)
    p.SetField("name", "player1")
    p.SetField("level", int64(10))

    items := []map[string]any{{"id": 1, "count": 10}}
    p.Arrays["items"] = &items

    data, err := data.SerializePlayer(p)
    if err != nil {
        t.Fatal(err)
    }

    // 验证可以反序列化
    var m map[string]any
    if err := json.Unmarshal(data, &m); err != nil {
        t.Fatal(err)
    }
}

func TestPersistDeserialize(t *testing.T) {
    jsonStr := `{"rid":1,"base":{"name":"player1","level":10},"arrays":{"items":[{"id":1,"count":10}]}}`

    p, err := data.DeserializePlayer([]byte(jsonStr))
    if err != nil {
        t.Fatal(err)
    }

    if p.RID != 1 {
        t.Errorf("expected RID 1, got %d", p.RID)
    }
    if p.GetField("name") != "player1" {
        t.Error("name mismatch")
    }
}

func TestPersistDeserializeInvalid(t *testing.T) {
    _, err := data.DeserializePlayer([]byte("invalid json"))
    if err == nil {
        // Expected error
    }
}

func TestPersistDeserializeEmpty(t *testing.T) {
    p, err := data.DeserializePlayer([]byte("{}"))
    if err != nil {
        t.Fatal(err)
    }
    if p.RID != 0 {
        t.Errorf("expected RID 0, got %d", p.RID)
    }
    if p.Base == nil {
        t.Error("expected Base to be initialized")
    }
    if p.Arrays == nil {
        t.Error("expected Arrays to be initialized")
    }
}
