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
    defer p.mutex.RUnlock()

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
    if raw.Base != nil {
        p.Base = raw.Base
    }
    if raw.Arrays != nil {
        p.Arrays = raw.Arrays
    }
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
