# Phase 4: 示例插件

> **Goal:** 创建可热更的业务逻辑插件 (RoleLogic, ItemLogic)

---

## 4.1 创建 RoleLogic 插件

**Files:**
- Create: `plugin/role/role.go`
- Create: `plugin/role/role_test.go`

- [ ] **Step 1: Write the failing test**

```go
// plugin/role/role_test.go
package main

import (
    "testing"
    
    "github.com/yourorg/wg_ai/internal/data"
    "github.com/yourorg/wg_ai/internal/plugin"
    p "github.com/yourorg/wg_ai/plugin"
)

func TestRoleLogicName(t *testing.T) {
    logic := &RoleLogic{}
    if logic.Name() != "role" {
        t.Errorf("expected 'role', got '%s'", logic.Name())
    }
}

func TestRoleLogicLogin(t *testing.T) {
    logic := &RoleLogic{}
    
    playerData := data.NewPlayerData(1)
    playerData.SetField("name", "player1")
    playerData.SetField("level", int64(10))
    playerData.SetField("exp", int64(5000))
    
    ctx := &p.LogicContext{
        RID:  1,
        UID:  100,
        Data: plugin.NewDataAdapter(1, playerData),
    }
    
    result, err := logic.Handle(ctx, "login", nil)
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }
    if result.Data["name"] != "player1" {
        t.Error("name mismatch")
    }
    if result.Data["level"] != int64(10) {
        t.Error("level mismatch")
    }
}

func TestRoleLogicGetInfo(t *testing.T) {
    logic := &RoleLogic{}
    
    playerData := data.NewPlayerData(1)
    playerData.SetField("name", "test")
    playerData.SetField("level", int64(5))
    playerData.SetField("exp", int64(100))
    playerData.SetField("gold", int64(1000))
    playerData.SetField("vip", int64(0))
    
    ctx := &p.LogicContext{
        RID:  1,
        Data: plugin.NewDataAdapter(1, playerData),
    }
    
    result, err := logic.Handle(ctx, "get_info", nil)
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }
    if result.Data["rid"] != int64(1) {
        t.Error("rid mismatch")
    }
}

func TestRoleLogicUpdateName(t *testing.T) {
    logic := &RoleLogic{}
    
    playerData := data.NewPlayerData(1)
    playerData.SetField("name", "oldname")
    
    ctx := &p.LogicContext{
        RID:  1,
        Data: plugin.NewDataAdapter(1, playerData),
    }
    
    result, err := logic.Handle(ctx, "update_name", map[string]any{
        "name": "newname",
    })
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }
    
    // 验证数据已更新
    name, _ := ctx.Data.GetField("name")
    if name != "newname" {
        t.Errorf("expected 'newname', got '%v'", name)
    }
}

func TestRoleLogicMethodNotFound(t *testing.T) {
    logic := &RoleLogic{}
    
    playerData := data.NewPlayerData(1)
    ctx := &p.LogicContext{
        RID:  1,
        Data: plugin.NewDataAdapter(1, playerData),
    }
    
    _, err := logic.Handle(ctx, "unknown_method", nil)
    if err != p.ErrMethodNotFound {
        t.Errorf("expected ErrMethodNotFound, got %v", err)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./plugin/role/... -v`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```go
// plugin/role/role.go
package main

import (
    p "github.com/yourorg/wg_ai/plugin"
)

// RoleLogic 角色逻辑模块
type RoleLogic struct{}

// Name 实现LogicModule接口
func (l *RoleLogic) Name() string {
    return "role"
}

// Handle 处理请求
func (l *RoleLogic) Handle(ctx *p.LogicContext, method string, params map[string]any) (*p.LogicResult, error) {
    switch method {
    case "login":
        return l.handleLogin(ctx, params)
    case "heartbeat":
        return l.handleHeartbeat(ctx, params)
    case "get_info":
        return l.handleGetInfo(ctx, params)
    case "update_name":
        return l.handleUpdateName(ctx, params)
    default:
        return nil, p.ErrMethodNotFound
    }
}

// handleLogin 处理登录
func (l *RoleLogic) handleLogin(ctx *p.LogicContext, params map[string]any) (*p.LogicResult, error) {
    name, _ := ctx.Data.GetField("name")
    level, _ := ctx.Data.GetField("level")
    exp, _ := ctx.Data.GetField("exp")
    
    return p.Success(map[string]any{
        "rid":   ctx.RID,
        "name":  name,
        "level": level,
        "exp":   exp,
    }), nil
}

// handleHeartbeat 处理心跳
func (l *RoleLogic) handleHeartbeat(ctx *p.LogicContext, params map[string]any) (*p.LogicResult, error) {
    return p.Success(nil), nil
}

// handleGetInfo 获取玩家信息
func (l *RoleLogic) handleGetInfo(ctx *p.LogicContext, params map[string]any) (*p.LogicResult, error) {
    name, _ := ctx.Data.GetField("name")
    level, _ := ctx.Data.GetField("level")
    exp, _ := ctx.Data.GetField("exp")
    gold, _ := ctx.Data.GetField("gold")
    vip, _ := ctx.Data.GetField("vip")
    
    return p.Success(map[string]any{
        "rid":   ctx.RID,
        "name":  name,
        "level": level,
        "exp":   exp,
        "gold":  gold,
        "vip":   vip,
    }), nil
}

// handleUpdateName 更新名字
func (l *RoleLogic) handleUpdateName(ctx *p.LogicContext, params map[string]any) (*p.LogicResult, error) {
    name, ok := params["name"].(string)
    if !ok || name == "" {
        return p.Error(2, "invalid name"), nil
    }
    
    // 更新数据
    if err := ctx.Data.SetField("name", name); err != nil {
        return nil, err
    }
    
    return p.Success(map[string]any{"name": name}), nil
}

// 导出符号 - 必须命名为 "Role" + "Module"
var RoleModule p.LogicModule = &RoleLogic{}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./plugin/role/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add plugin/role/role.go plugin/role/role_test.go
git commit -m "feat(plugin): add RoleLogic plugin with login/get_info/update_name"
```

---

## 4.2 创建 ItemLogic 插件

**Files:**
- Create: `plugin/item/item.go`
- Create: `plugin/item/item_test.go`

- [ ] **Step 1: Write the failing test**

```go
// plugin/item/item_test.go
package main

import (
    "testing"
    
    "github.com/yourorg/wg_ai/internal/data"
    "github.com/yourorg/wg_ai/internal/plugin"
    p "github.com/yourorg/wg_ai/plugin"
)

func TestItemLogicName(t *testing.T) {
    logic := &ItemLogic{}
    if logic.Name() != "item" {
        t.Errorf("expected 'item', got '%s'", logic.Name())
    }
}

func TestItemLogicList(t *testing.T) {
    logic := &ItemLogic{}
    
    playerData := data.NewPlayerData(1)
    items := &[]ItemData{
        {ID: 1, CfgID: 100, Count: 10},
        {ID: 2, CfgID: 101, Count: 5},
    }
    playerData.Arrays["items"] = items
    
    ctx := &p.LogicContext{
        RID:  1,
        Data: plugin.NewDataAdapter(1, playerData),
    }
    
    result, err := logic.Handle(ctx, "list", nil)
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }
    
    itemsResult, ok := result.Data["items"].(*[]ItemData)
    if !ok {
        t.Fatal("items type error")
    }
    if len(*itemsResult) != 2 {
        t.Errorf("expected 2 items, got %d", len(*itemsResult))
    }
}

func TestItemLogicUse(t *testing.T) {
    logic := &ItemLogic{}
    
    playerData := data.NewPlayerData(1)
    items := &[]ItemData{
        {ID: 1, CfgID: 100, Count: 10},
    }
    playerData.Arrays["items"] = items
    
    ctx := &p.LogicContext{
        RID:  1,
        Data: plugin.NewDataAdapter(1, playerData),
    }
    
    // 使用 3 个
    result, err := logic.Handle(ctx, "use", map[string]any{
        "item_id": int64(1),
        "count":   int64(3),
    })
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }
    
    // 验证剩余数量
    itemsResult := *items
    if itemsResult[0].Count != 7 {
        t.Errorf("expected 7 remaining, got %d", itemsResult[0].Count)
    }
}

func TestItemLogicUseNotEnough(t *testing.T) {
    logic := &ItemLogic{}
    
    playerData := data.NewPlayerData(1)
    items := &[]ItemData{
        {ID: 1, CfgID: 100, Count: 5},
    }
    playerData.Arrays["items"] = items
    
    ctx := &p.LogicContext{
        RID:  1,
        Data: plugin.NewDataAdapter(1, playerData),
    }
    
    // 尝试使用 10 个 (只有 5 个)
    result, err := logic.Handle(ctx, "use", map[string]any{
        "item_id": int64(1),
        "count":   int64(10),
    })
    if err != nil {
        t.Fatal(err)
    }
    if result.Code == 0 {
        t.Error("expected error code for not enough items")
    }
}

func TestItemLogicAdd(t *testing.T) {
    logic := &ItemLogic{}
    
    playerData := data.NewPlayerData(1)
    items := &[]ItemData{}
    playerData.Arrays["items"] = items
    
    ctx := &p.LogicContext{
        RID:  1,
        Data: plugin.NewDataAdapter(1, playerData),
    }
    
    result, err := logic.Handle(ctx, "add", map[string]any{
        "cfg_id": int64(100),
        "count":  int64(5),
    })
    if err != nil {
        t.Fatal(err)
    }
    if result.Code != 0 {
        t.Errorf("expected code 0, got %d", result.Code)
    }
    
    // 验证物品已添加
    itemsResult := *items
    if len(itemsResult) != 1 {
        t.Errorf("expected 1 item, got %d", len(itemsResult))
    }
    if itemsResult[0].CfgID != 100 {
        t.Error("cfg_id mismatch")
    }
    if itemsResult[0].Count != 5 {
        t.Error("count mismatch")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./plugin/item/... -v`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```go
// plugin/item/item.go
package main

import (
    "time"
    
    p "github.com/yourorg/wg_ai/plugin"
)

// ItemData 物品数据结构
type ItemData struct {
    ID    int64 `json:"id"`     // 唯一ID
    CfgID int64 `json:"cfg_id"` // 配置ID
    Count int64 `json:"count"`  // 数量
}

// ItemLogic 物品逻辑模块
type ItemLogic struct{}

// Name 实现LogicModule接口
func (l *ItemLogic) Name() string {
    return "item"
}

// Handle 处理请求
func (l *ItemLogic) Handle(ctx *p.LogicContext, method string, params map[string]any) (*p.LogicResult, error) {
    switch method {
    case "list":
        return l.handleList(ctx, params)
    case "use":
        return l.handleUse(ctx, params)
    case "add":
        return l.handleAdd(ctx, params)
    default:
        return nil, p.ErrMethodNotFound
    }
}

// handleList 获取背包列表
func (l *ItemLogic) handleList(ctx *p.LogicContext, params map[string]any) (*p.LogicResult, error) {
    items, err := ctx.Data.GetArray("items")
    if err != nil {
        return nil, err
    }
    
    // 如果背包为空，返回空列表
    if items == nil {
        return p.Success(map[string]any{
            "items": &[]ItemData{},
        }), nil
    }
    
    return p.Success(map[string]any{
        "items": items,
    }), nil
}

// handleUse 使用物品
func (l *ItemLogic) handleUse(ctx *p.LogicContext, params map[string]any) (*p.LogicResult, error) {
    itemID, _ := params["item_id"].(int64)
    count, _ := params["count"].(int64)
    if count <= 0 {
        count = 1
    }
    
    // 获取背包数据指针
    itemsAny, err := ctx.Data.GetArray("items")
    if err != nil {
        return nil, err
    }
    
    if itemsAny == nil {
        return p.Error(202, "item not found"), nil
    }
    
    items, ok := itemsAny.(*[]ItemData)
    if !ok {
        return p.Error(201, "invalid items data"), nil
    }
    
    // 查找并扣除
    for i, item := range *items {
        if item.ID == itemID {
            if item.Count < count {
                return p.Error(200, "not enough items"), nil
            }
            
            (*items)[i].Count -= count
            
            // 删除空物品
            if (*items)[i].Count <= 0 {
                *items = append((*items)[:i], (*items)[i+1:]...)
            }
            
            ctx.Data.MarkDirty()
            
            remaining := int64(0)
            if i < len(*items) {
                remaining = (*items)[i].Count
            }
            
            return p.Success(map[string]any{
                "remaining": remaining,
            }), nil
        }
    }
    
    return p.Error(202, "item not found"), nil
}

// handleAdd 添加物品
func (l *ItemLogic) handleAdd(ctx *p.LogicContext, params map[string]any) (*p.LogicResult, error) {
    cfgID, _ := params["cfg_id"].(int64)
    count, _ := params["count"].(int64)
    if count <= 0 {
        count = 1
    }
    
    // 获取背包数据指针
    itemsAny, err := ctx.Data.GetArray("items")
    if err != nil {
        return nil, err
    }
    
    var items *[]ItemData
    if itemsAny == nil {
        // 初始化背包
        newItems := make([]ItemData, 0)
        items = &newItems
    } else {
        items, _ = itemsAny.(*[]ItemData)
    }
    
    // 添加物品
    newItem := ItemData{
        ID:    generateItemID(),
        CfgID: cfgID,
        Count: count,
    }
    *items = append(*items, newItem)
    
    // 更新回 Arrays (如果是新创建的)
    if itemsAny == nil {
        // 需要设置回 Arrays (这里通过 MarkDirty 间接触发)
        ctx.Data.MarkDirty()
    }
    
    ctx.Data.MarkDirty()
    
    return p.Success(map[string]any{
        "item": newItem,
    }), nil
}

// generateItemID 生成物品唯一ID
func generateItemID() int64 {
    return time.Now().UnixNano()
}

// 导出符号
var ItemModule p.LogicModule = &ItemLogic{}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./plugin/item/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add plugin/item/item.go plugin/item/item_test.go
git commit -m "feat(plugin): add ItemLogic plugin with list/use/add"
```

---

## 4.3 编译脚本 Makefile

**Files:**
- Create: `Makefile` (如果不存在则创建，否则追加)

- [ ] **Step 1: Write Makefile**

```makefile
# Makefile
.PHONY: build plugins all clean test plugin-role plugin-item

# 主程序
build:
	go build -o bin/game ./cmd/game

# 编译所有插件
plugins: plugin-role plugin-item

# 编译单个插件
plugin-role:
	go build -buildmode=plugin -o plugins/role.so ./plugin/role

plugin-item:
	go build -buildmode=plugin -o plugins/item.so ./plugin/item

# 全量编译
all: build plugins

# 清理
clean:
	rm -rf bin/*
	rm -rf plugins/*.so

# 运行测试
test:
	go test ./...

# 运行测试 (带覆盖率)
test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# 热更示例 (编译 + 调用API)
hotreload-role: plugin-role
	curl -X POST http://localhost:8081/admin/hotreload \
		-H "Content-Type: application/json" \
		-d '{"module": "role", "path": "./plugins/role.so"}'

hotreload-item: plugin-item
	curl -X POST http://localhost:8081/admin/hotreload \
		-H "Content-Type: application/json" \
		-d '{"module": "item", "path": "./plugins/item.so"}'
```

- [ ] **Step 2: Create plugins directory**

```bash
mkdir -p plugins
```

- [ ] **Step 3: Test compilation**

Run: `make plugins`
Expected: 编译成功，生成 plugins/role.so, plugins/item.so

- [ ] **Step 4: Commit**

```bash
git add Makefile plugins/.gitkeep
git commit -m "build: add Makefile with plugin compilation support"
```

---

## Phase 4 Summary

| File | Description |
|------|-------------|
| `plugin/role/role.go` | 角色逻辑插件 |
| `plugin/role/role_test.go` | 单元测试 |
| `plugin/item/item.go` | 物品逻辑插件 |
| `plugin/item/item_test.go` | 单元测试 |
| `Makefile` | 编译脚本 |
| `plugins/` | 编译输出目录 |
