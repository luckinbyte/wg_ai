package plugin

import "errors"

// ============ 错误定义 ============

var (
	 ErrModuleNotFound = errors.New("module not found")
    ErrMethodNotFound = errors.New("method not found")
    ErrInvalidModule   = errors.New("invalid module type")
    ErrPluginLoadFail  = errors.New("plugin load failed")
    ErrPlayerNotFound = errors.New("player not found")
)

// ============ 数据访问接口 (供插件使用) ============

// DataAccessor 数据访问接口 - 简化版，供插件调用
type DataAccessor interface {
    // 获取基础字段 (返回值可直接类型断言修改)
    GetField(key string) (any, error)

    // 设置基础字段
    SetField(key string, value any) error

    // 获取数组字段 (返回指针，可直接修改)
    GetArray(key string) (any, error)

    // 设置数组字段
    SetArray(key string, value any) error

    // 标记脏数据
    MarkDirty()
}

// DataAdapter 数据访问适配器接口
type DataAdapter interface {
    DataAccessor
    RID() int64
}

// ============ 会话推送接口 ============

// SessionPush 会话推送接口
type SessionPush interface {
    // 推送消息给客户端
    Push(msgID uint16, data []byte) error
}

// ============ 逻辑上下文 ============

// LogicContext 逻辑处理上下文
type LogicContext struct {
    RID     int64        // 玩家ID
    UID     int64        // 用户ID
    Data    DataAccessor // 数据访问接口
    Session SessionPush  // 会话推送接口
}

// ============ 逻辑结果 ============

// LogicResult 逻辑处理结果
type LogicResult struct {
    Code    int            `json:"code"`    // 错误码，0表示成功
    Message string         `json:"message"` // 错误信息
    Data    map[string]any `json:"data"`    // 返回数据
    Push    []PushData     `json:"push"`    // 推送数据列表
}

// PushData 推送数据
type PushData struct {
    MsgID uint16 `json:"msg_id"` // 消息ID
    Data  []byte `json:"data"`   // 消息内容
}

// WithPush 添加推送数据(链式调用)
func (r *LogicResult) WithPush(msgID uint16, data []byte) *LogicResult {
    r.Push = append(r.Push, PushData{MsgID: msgID, Data: data})
    return r
}

// ============ 逻辑模块接口 ============

// LogicModule 逻辑模块接口 - 插件必须实现
type LogicModule interface {
    // 模块名称
    Name() string

    // 处理请求
    Handle(ctx *LogicContext, method string, params map[string]any) (*LogicResult, error)
}

// ============ 辅助函数 ============

// Success 创建成功结果
func Success(data map[string]any) *LogicResult {
    return &LogicResult{
        Code:    0,
        Message: "success",
        Data:    data,
        Push:    nil,
    }
}

// Error 创建错误结果
func Error(code int, message string) *LogicResult {
    return &LogicResult{
        Code:    code,
        Message: message,
        Data:    nil,
        Push:    nil,
    }
}

// ErrorWithData 创建带数据的错误结果
func ErrorWithData(code int, message string, data map[string]any) *LogicResult {
    return &LogicResult{
        Code:    code,
        Message: message,
        Data:    data,
        Push:    nil,
    }
}
