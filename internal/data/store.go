package data

import (
	"sync"
)

// DataStore 数据存储接口 - 供逻辑层调用
type DataStore interface {
    GetPlayer(rid int64) (*PlayerData, error)
    GetField(rid int64, key string) (any, error)
    SetField(rid int64, key string, value any) error
    GetArray(rid int64, key string) (any, error)
    SetArray(rid int64, key string, value any) error
    MarkDirty(rid int64)
    GetPlayers(rids []int64) ([]*PlayerData, error)
    ForEachLoadedPlayer(func(rid int64, p *PlayerData) error) error
}

// PlayerStore 玩家数据管理器
type PlayerStore struct {
    players map[int64]*PlayerData
    mutex   sync.RWMutex
}

// NewPlayerStore 创建玩家数据存储
func NewPlayerStore() *PlayerStore {
    return &PlayerStore{
        players: make(map[int64]*PlayerData),
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

// SetArray 设置数组字段
func (s *PlayerStore) SetArray(rid int64, key string, value any) error {
    p, err := s.GetPlayer(rid)
    if err != nil {
        return err
    }
    p.SetArray(key, value)
    return nil
}

// MarkDirty 标记脏数据
func (s *PlayerStore) MarkDirty(rid int64) {
    p, err := s.GetPlayer(rid)
    if err != nil {
        return
    }
    p.MarkDirty()
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

// ForEachLoadedPlayer 遍历已加载的玩家数据
func (s *PlayerStore) ForEachLoadedPlayer(fn func(rid int64, p *PlayerData) error) error {
    s.mutex.RLock()
    snapshot := make([]struct {
        rid int64
        p   *PlayerData
    }, 0, len(s.players))
    for rid, p := range s.players {
        snapshot = append(snapshot, struct {
            rid int64
            p   *PlayerData
        }{rid: rid, p: p})
    }
    s.mutex.RUnlock()

    for _, entry := range snapshot {
        if err := fn(entry.rid, entry.p); err != nil {
            return err
        }
    }
    return nil
}
