# Phase 1: 数据层基础

> **Goal:** 建立玩家数据存储结构和管理器，使用 map[string]any + slice 存储数据

---

## 1.1 定义 PlayerData 数据结构

**Files:**
- Create: `internal/data/types.go`
- Create: `internal/data/types_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/data/types_test.go
package data

import (
    "sync"
    "testing"
)

func TestNewPlayerData(t *testing.T) {
    p := NewPlayerData(12345)
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
    p := NewPlayerData(1)
    
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
    p := NewPlayerData(1)
    
    items := []map[string]any{{"id": 1}, {"id": 2}}
    p.Arrays["items"] = &items
    
    arr := p.GetArray("items")
    if arr == nil {
        t.Error("expected items array")
    }
}

func TestPlayerDataConcurrency(t *testing.T) {
    p := NewPlayerData(1)
    
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/data/... -v`
Expected: FAIL - package data not found

- [ ] **Step 3: Write minimal implementation**

```go
// internal/data/types.go
package data

import "sync"

// PlayerData 玩家数据容器
type PlayerData struct {
    RID    int64          `json:"rid"`     // 角色ID
    Base   map[string]any `json:"base"`    // 基础字段: name, level, exp, gold...
    Arrays map[string]any `json:"arrays"`  // 数组字段: items, heroes, tasks...
    Dirty  bool           `json:"-"`       // 脏标记
    mutex  sync.RWMutex   `json:"-"`       // 读写锁
}

// NewPlayerData 创建空的玩家数据
func NewPlayerData(rid int64) *PlayerData {
    return &PlayerData{
        RID:    rid,
        Base:   make(map[string]any),
        Arrays: make(map[string]any),
        Dirty:  false,
    }
}

// Lock 加写锁
func (p *PlayerData) Lock() {
    p.mutex.Lock()
}

// Unlock 解写锁
func (p *PlayerData) Unlock() {
    p.mutex.Unlock()
}

// RLock 加读锁
func (p *PlayerData) RLock() {
    p.mutex.RLock()
}

// RUnlock 解读锁
func (p *PlayerData) RUnlock() {
    p.mutex.RUnlock()
}

// GetField 获取基础字段
func (p *PlayerData) GetField(key string) any {
    p.mutex.RLock()
    defer p.mutex.RUnlock()
    return p.Base[key]
}

// SetField 设置基础字段
func (p *PlayerData) SetField(key string, value any) {
    p.mutex.Lock()
    defer p.mutex.Unlock()
    p.Base[key] = value
    p.Dirty = true
}

// GetArray 获取数组字段
func (p *PlayerData) GetArray(key string) any {
    p.mutex.RLock()
    defer p.mutex.RUnlock()
    return p.Arrays[key]
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/data/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/data/types.go internal/data/types_test.go
git commit -m "feat(data): add PlayerData structure with map[string]any storage"
```

---

## 1.2 实现 DataStore 接口和 PlayerStore

**Files:**
- Create: `internal/data/store.go`
- Create: `internal/data/store_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/data/store_test.go
package data

import (
    "testing"
)

func TestNewPlayerStore(t *testing.T) {
    store := NewPlayerStore(nil)
    if store == nil {
        t.Error("expected PlayerStore")
    }
}

func TestPlayerStoreGetPlayer(t *testing.T) {
    store := NewPlayerStore(nil)
    
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
    store := NewPlayerStore(nil)
    
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
    store := NewPlayerStore(nil)
    
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
    store := NewPlayerStore(nil)
    
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/data/... -v -run TestPlayerStore`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```go
// internal/data/store.go
package data

import (
    "database/sql"
    "sync"
)

// DataStore 数据存储接口 - 供逻辑层调用
type DataStore interface {
    GetPlayer(rid int64) (*PlayerData, error)
    GetField(rid int64, key string) (any, error)
    SetField(rid int64, key string, value any) error
    GetArray(rid int64, key string) (any, error)
    MarkDirty(rid int64)
    GetPlayers(rids []int64) ([]*PlayerData, error)
}

// PlayerStore 玩家数据管理器
type PlayerStore struct {
    players map[int64]*PlayerData
    mutex   sync.RWMutex
    db      *sql.DB
}

// NewPlayerStore 创建玩家数据存储
func NewPlayerStore(db *sql.DB) *PlayerStore {
    return &PlayerStore{
        players: make(map[int64]*PlayerData),
        db:      db,
    }
}

// GetPlayer 获取玩家数据 (不存在则创建)
func (s *PlayerStore) GetPlayer(rid int64) (*PlayerData, error) {
    s.mutex.RLock()
    p, exists := s.players[rid]
    s.mutex.RUnlock()
    
    if exists {
        return p, nil
    }
    
    // 创建新数据
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    // double check
    if p, exists = s.players[rid]; exists {
        return p, nil
    }
    
    p = NewPlayerData(rid)
    s.players[rid] = p
    return p, nil
}

// GetField 获取基础字段
func (s *PlayerStore) GetField(rid int64, key string) (any, error) {
    p, err := s.GetPlayer(rid)
    if err != nil {
        return nil, err
    }
    return p.GetField(key), nil
}

// SetField 设置基础字段
func (s *PlayerStore) SetField(rid int64, key string, value any) error {
    p, err := s.GetPlayer(rid)
    if err != nil {
        return err
    }
    p.SetField(key, value)
    return nil
}

// GetArray 获取数组字段
func (s *PlayerStore) GetArray(rid int64, key string) (any, error) {
    p, err := s.GetPlayer(rid)
    if err != nil {
        return nil, err
    }
    return p.GetArray(key), nil
}

// MarkDirty 标记脏数据
func (s *PlayerStore) MarkDirty(rid int64) {
    p, err := s.GetPlayer(rid)
    if err != nil {
        return
    }
    p.Lock()
    p.Dirty = true
    p.Unlock()
}

// GetPlayers 批量获取玩家数据
func (s *PlayerStore) GetPlayers(rids []int64) ([]*PlayerData, error) {
    players := make([]*PlayerData, 0, len(rids))
    for _, rid := range rids {
        p, err := s.GetPlayer(rid)
        if err != nil {
            return nil, err
        }
        players = append(players, p)
    }
    return players, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/data/... -v -run TestPlayerStore`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/data/store.go internal/data/store_test.go
git commit -m "feat(data): add DataStore interface and PlayerStore implementation"
```

---

## 1.3 实现数据持久化

**Files:**
- Create: `internal/data/persist.go`
- Create: `internal/data/persist_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/data/persist_test.go
package data

import (
    "encoding/json"
    "testing"
)

func TestPersistSerialize(t *testing.T) {
    p := NewPlayerData(1)
    p.SetField("name", "player1")
    p.SetField("level", int64(10))
    
    items := []map[string]any{{"id": 1, "count": 10}}
    p.Arrays["items"] = &items
    
    data, err := SerializePlayer(p)
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
    
    p, err := DeserializePlayer([]byte(jsonStr))
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/data/... -v -run TestPersist`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```go
// internal/data/persist.go
package data

import (
    "database/sql"
    "encoding/json"
    "time"
)

// Persist 持久化器
type Persist struct {
    db *sql.DB
}

// NewPersist 创建持久化器
func NewPersist(db *sql.DB) *Persist {
    return &Persist{db: db}
}

// SerializePlayer 序列化玩家数据
func SerializePlayer(p *PlayerData) ([]byte, error) {
    p.RLock()
    defer p.RUnlock()
    
    return json.Marshal(struct {
        RID    int64          `json:"rid"`
        Base   map[string]any `json:"base"`
        Arrays map[string]any `json:"arrays"`
    }{
        RID:    p.RID,
        Base:   p.Base,
        Arrays: p.Arrays,
    })
}

// DeserializePlayer 反序列化玩家数据
func DeserializePlayer(data []byte) (*PlayerData, error) {
    var raw struct {
        RID    int64          `json:"rid"`
        Base   map[string]any `json:"base"`
        Arrays map[string]any `json:"arrays"`
    }
    
    if err := json.Unmarshal(data, &raw); err != nil {
        return nil, err
    }
    
    p := NewPlayerData(raw.RID)
    p.Base = raw.Base
    p.Arrays = raw.Arrays
    return p, nil
}

// Load 从数据库加载玩家数据
func (p *Persist) Load(rid int64) (*PlayerData, error) {
    var baseJSON, arraysJSON []byte
    err := p.db.QueryRow(
        "SELECT base, arrays FROM player_data WHERE rid = ?",
        rid,
    ).Scan(&baseJSON, &arraysJSON)
    
    if err == sql.ErrNoRows {
        return NewPlayerData(rid), nil
    }
    if err != nil {
        return nil, err
    }
    
    player := NewPlayerData(rid)
    
    if err := json.Unmarshal(baseJSON, &player.Base); err != nil {
        return nil, err
    }
    if err := json.Unmarshal(arraysJSON, &player.Arrays); err != nil {
        return nil, err
    }
    
    return player, nil
}

// Save 保存玩家数据到数据库
func (p *Persist) Save(data *PlayerData) error {
    baseJSON, err := json.Marshal(data.Base)
    if err != nil {
        return err
    }
    
    arraysJSON, err := json.Marshal(data.Arrays)
    if err != nil {
        return err
    }
    
    _, err = p.db.Exec(`
        INSERT INTO player_data (rid, base, arrays, updated_at)
        VALUES (?, ?, ?, ?)
        ON DUPLICATE KEY UPDATE base = ?, arrays = ?, updated_at = ?
    `, data.RID, baseJSON, arraysJSON, time.Now(), baseJSON, arraysJSON, time.Now())
    
    return err
}

// SaveAll 批量保存脏数据
func (p *Persist) SaveAll(players map[int64]*PlayerData) error {
    for _, player := range players {
        player.RLock()
        dirty := player.Dirty
        player.RUnlock()
        
        if !dirty {
            continue
        }
        
        if err := p.Save(player); err != nil {
            return err
        }
        
        player.Lock()
        player.Dirty = false
        player.Unlock()
    }
    return nil
}

// CreateTable 创建数据表 (用于初始化)
func (p *Persist) CreateTable() error {
    _, err := p.db.Exec(`
        CREATE TABLE IF NOT EXISTS player_data (
            rid BIGINT PRIMARY KEY,
            base JSON,
            arrays JSON,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
        )
    `)
    return err
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/data/... -v -run TestPersist`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/data/persist.go internal/data/persist_test.go
git commit -m "feat(data): add persist layer with JSON serialization"
```

---

## Phase 1 Summary

| File | Description |
|------|-------------|
| `internal/data/types.go` | PlayerData 结构定义 |
| `internal/data/types_test.go` | 单元测试 |
| `internal/data/store.go` | DataStore 接口和 PlayerStore |
| `internal/data/store_test.go` | 单元测试 |
| `internal/data/persist.go` | 持久化逻辑 |
| `internal/data/persist_test.go` | 单元测试 |
