package plugin

import (
    "testing"

    "github.com/yourorg/wg_ai/internal/data"
)

func TestDataAdapterGetField(t *testing.T) {
    playerData := data.NewPlayerData(1)
    playerData.SetField("name", "player1")
    playerData.SetField("level", int64(10))

    adapter := NewDataAdapter(1, playerData)

    val, err := adapter.GetField("name")
    if err != nil {
        t.Fatal(err)
    }
    if val != "player1" {
        t.Errorf("expected 'player1', got '%v'", val)
    }

    val, err = adapter.GetField("level")
    if err != nil {
        t.Fatal(err)
    }
    if val != int64(10) {
        t.Errorf("expected 10, got %v", val)
    }
}

 func TestDataAdapterSetField(t *testing.T) {
    playerData := data.NewPlayerData(1)
    adapter := NewDataAdapter(1, playerData)

    err := adapter.SetField("gold", int64(1000))
    if err != nil {
        t.Fatal(err)
    }

    val, _ := adapter.GetField("gold")
    if val != int64(1000) {
        t.Errorf("expected 1000, got %v", val)
    }

    // 验证脏标记
    if !playerData.Dirty {
        t.Error("expected dirty flag")
    }
}

 func TestDataAdapterGetArray(t *testing.T) {
    playerData := data.NewPlayerData(1)

    // 设置数组
    items := []map[string]any{{"id": 1}, {"id": 2}}
    playerData.Arrays["items"] = &items

    adapter := NewDataAdapter(1, playerData)

    arr, err := adapter.GetArray("items")
    if err != nil {
        t.Fatal(err)
    }

    itemsPtr, ok := arr.(*[]map[string]any)
    if !ok {
        t.Fatal("type assertion failed")
    }

    if len(*itemsPtr) != 2 {
        t.Errorf("expected 2 items, got %d", len(*itemsPtr))
    }
}

 func TestDataAdapterMarkDirty(t *testing.T) {
    playerData := data.NewPlayerData(1)
    playerData.Dirty = false

    adapter := NewDataAdapter(1, playerData)
    adapter.MarkDirty()

    if !playerData.Dirty {
        t.Error("expected dirty flag after MarkDirty")
    }
}
