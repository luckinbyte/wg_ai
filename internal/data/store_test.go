package data_test

import (
    "errors"
    "testing"

    "github.com/luckinbyte/wg_ai/internal/data"
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

func TestPlayerStoreForEachLoadedPlayer(t *testing.T) {
    store := data.NewPlayerStore()

    p1, err := store.GetPlayer(1)
    if err != nil {
        t.Fatal(err)
    }
    p2, err := store.GetPlayer(2)
    if err != nil {
        t.Fatal(err)
    }

    seen := make(map[int64]*data.PlayerData)
    err = store.ForEachLoadedPlayer(func(rid int64, p *data.PlayerData) error {
        seen[rid] = p
        return nil
    })
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(seen) != 2 {
        t.Fatalf("expected 2 loaded players, got %d", len(seen))
    }
    if seen[1] != p1 {
        t.Fatalf("expected player 1 instance to match snapshot")
    }
    if seen[2] != p2 {
        t.Fatalf("expected player 2 instance to match snapshot")
    }
}

func TestPlayerStoreForEachLoadedPlayerStopsOnError(t *testing.T) {
    store := data.NewPlayerStore()

    _, err := store.GetPlayer(1)
    if err != nil {
        t.Fatal(err)
    }
    _, err = store.GetPlayer(2)
    if err != nil {
        t.Fatal(err)
    }

    wantErr := errors.New("stop")
    calls := 0
    err = store.ForEachLoadedPlayer(func(rid int64, p *data.PlayerData) error {
        calls++
        return wantErr
    })
    if !errors.Is(err, wantErr) {
        t.Fatalf("expected %v, got %v", wantErr, err)
    }
    if calls != 1 {
        t.Fatalf("expected callback to stop after first error, got %d calls", calls)
    }
}
