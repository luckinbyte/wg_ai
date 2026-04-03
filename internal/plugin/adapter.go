package plugin

import (
    "github.com/yourorg/wg_ai/internal/data"
)

// DataAdapter 数据访问适配器 - 连接 PlayerData 和 DataAccessor 接口
type DataAdapter struct {
    rid  int64
    data *data.PlayerData
}

// NewDataAdapter 创建数据适配器
func NewDataAdapter(rid int64, playerData *data.PlayerData) *DataAdapter {
    return &DataAdapter{
        rid:  rid,
        data: playerData,
    }
}

// GetField 实现 DataAccessor 接口
func (a *DataAdapter) GetField(key string) (any, error) {
    return a.data.GetField(key), nil
}

// SetField 实现 DataAccessor 接口
func (a *DataAdapter) SetField(key string, value any) error {
    a.data.SetField(key, value)
    a.data.Lock()
    a.data.Dirty = true
    a.data.Unlock()
    return nil
}

// GetArray 实现 DataAccessor 接口 - 返回指针，可直接修改
func (a *DataAdapter) GetArray(key string) (any, error) {
    return a.data.GetArray(key), nil
}

// MarkDirty 实现 DataAccessor 接口
func (a *DataAdapter) MarkDirty() {
    a.data.Lock()
    a.data.Dirty = true
    a.data.Unlock()
}

 // RID 获取玩家ID
func (a *DataAdapter) RID() int64 {
    return a.rid
}
