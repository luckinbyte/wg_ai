package data_test

import (
	"sync"
	"testing"

    "github.com/yourorg/wg_ai/internal/data"
)

func TestNewPlayerData(t *testing.T) {
    p := data.NewPlayerData(12345)
    if p.RID != 12345 {
        t.Errorf("expected RID 12345, got %d", p.RID)
    }
    if p.Base == nil {
        t.Error("expected Base to be initialized")
    }
    if p.Arrays == nil {
        t.Error("expected Arrays to be initialized")
    }
}

func TestPlayerDataGetSetField(t *testing.T) {
    p := data.NewPlayerData(1)

    // Test Set
    p.SetField("name", "player1")
    p.SetField("level", int64(10))

    // Test Get
    if p.GetField("name") != "player1" {
        t.Error("name mismatch")
    }
    if p.GetField("level") != int64(10) {
        t.Error("level mismatch")
    }
}

func TestPlayerDataGetArray(t *testing.T) {
    p := data.NewPlayerData(1)

    items := []map[string]any{{"id": 1}, {"id": 2}}
    p.Arrays["items"] = &items

    arr := p.GetArray("items")
    if arr == nil {
        t.Error("expected items array")
    }
}

func TestPlayerDataConcurrency(t *testing.T) {
    p := data.NewPlayerData(1)

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(2)
        go func(i int) {
            defer wg.Done()
            p.SetField("field", i)
        }(i)
        go func() {
            defer wg.Done()
            p.GetField("field")
        }()
    }
    wg.Wait()
}
