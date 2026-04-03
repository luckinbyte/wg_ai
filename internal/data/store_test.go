package data_test

import (
    "testing"

    "github.com/yourorg/wg_ai/internal/data"
)

func TestNewPlayerStore(t *testing.T) {
    store := data.NewPlayerStore()
    if store == nil {
        t.Error("expected PlayerStore")
    }
}

func TestPlayerStoreGetPlayer(t *testing.T) {
    store := data.NewPlayerStore()

    // 首次获取会创建
    p, err := store.GetPlayer(1)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if p.RID != 1 {
        t.Errorf("expected RID 1, got %d", p.RID)
    }

    // 再次获取应返回同一个实例
    p2, _ := store.GetPlayer(1)
    if p != p2 {
        t.Error("expected same instance")
    }
}

func TestPlayerStoreGetField(t *testing.T) {
    store := data.NewPlayerStore()

    _, err := store.GetPlayer(1)
    if err != nil {
        t.Fatal(err)
    }

    store.SetField(1, "level", int64(10))

    val, err := store.GetField(1, "level")
    if err != nil {
        t.Fatal(err)
    }
    if val != int64(10) {
        t.Errorf("expected level 10, got %v", val)
    }
}

func TestPlayerStoreMarkDirty(t *testing.T) {
    store := data.NewPlayerStore()

    p, _ := store.GetPlayer(1)
    if p.Dirty {
        t.Error("should not be dirty initially")
    }

    store.MarkDirty(1)
    if !p.Dirty {
        t.Error("should be dirty after MarkDirty")
    }
}

func TestPlayerStoreGetPlayers(t *testing.T) {
    store := data.NewPlayerStore()

    store.GetPlayer(1)
    store.GetPlayer(2)
    store.GetPlayer(3)

    players, err := store.GetPlayers([]int64{1, 2, 3})
    if err != nil {
        t.Fatal(err)
    }
    if len(players) != 3 {
        t.Errorf("expected 3 players, got %d", len(players))
    }
}
