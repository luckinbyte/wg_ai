# 通信协议文档

> **维护说明**: 新增协议时请更新本文档
>
> - 添加新消息ID → 更新 [消息路由表](#消息路由表)
> - 添加新模块 → 更新 [模块接口](#模块接口)
> - 修改协议格式 → 更新 [协议格式](#协议格式)

---

## 目录

- [传输层](#传输层)
- [协议格式](#协议格式)
- [消息类型](#消息类型)
- [消息路由表](#消息路由表)
- [模块接口](#模块接口)
- [错误码](#错误码)
- [附录](#附录)

---

## 传输层

### 支持的传输协议

| 协议 | 端口 | 路径 | 适用场景 |
|------|------|------|----------|
| **TCP** | 44445 | - | 手游 App、Unity/UE 客户端 |
| **WebSocket** | 44446 | `/ws` | H5、网页游戏、小程序 |

### 连接流程

```
┌─────────┐                    ┌─────────┐
│ Client  │                    │ Server  │
└────┬────┘                    └────┬────┘
     │                              │
     │──── Handshake ──────────────►│
     │◄─── HandshakeResponse ───────│
     │                              │
     │──── Request(msg_id) ────────►│
     │◄─── Response(code, data) ────│
     │                              │
     │◄─── Push(msg_id, event) ─────│  (服务端推送)
     │                              │
```

---

## 协议格式

### TCP 数据包格式

```
┌────────────┬──────────────┬─────────────────┐
│ Length     │ MsgType      │ Payload         │
│ (4 bytes)  │ (1 byte)     │ (N bytes)       │
└────────────┴──────────────┴─────────────────┘
```

| 字段 | 大小 | 字节序 | 说明 |
|------|------|--------|------|
| Length | 4 bytes | BigEndian | MsgType + Payload 的总长度 |
| MsgType | 1 byte | - | 消息类型 |
| Payload | N bytes | - | Protobuf 序列化数据 |

### WebSocket 数据帧格式

```
┌──────────────┬─────────────────┐
│ MsgType      │ Payload         │
│ (1 byte)     │ (N bytes)       │
└──────────────┴─────────────────┘
```

> WebSocket 帧自带长度，无需 Length 前缀

### Payload 结构 (Protobuf)

```protobuf
// 请求
message Request {
    Header header = 1;    // msg_id + sequence
    bytes payload = 2;    // 业务数据
}

// 响应
message Response {
    Header header = 1;    // msg_id + sequence
    int32 code = 2;       // 错误码 (0=成功)
    bytes payload = 3;    // 业务数据
}

// 推送
message Push {
    uint32 msg_id = 1;
    bytes payload = 2;
}

// 消息头
message Header {
    uint32 msg_id = 1;
    uint32 sequence = 2;
}
```

---

## 消息类型

| MsgType | 值 | 说明 |
|---------|-----|------|
| `MsgTypeHandshake` | 0x04 | 握手 |
| `MsgTypeRequest` | 0x01 | 客户端请求 |
| `MsgTypeResponse` | 0x02 | 服务端响应 |
| `MsgTypePush` | 0x03 | 服务端推送 |

---

## 消息路由表

### 消息ID分配规则

| 范围 | 模块 | 说明 |
|------|------|------|
| 1-999 | 系统 | 握手、内部消息 |
| 1001-1999 | role | 角色模块 |
| 2001-2999 | item | 物品模块 |
| 3001-3999 | scene | 场景模块 |
| 4001-4999 | march | 行军模块 |
| 5001-5999 | battle | 战斗模块 (预留) |
| 6001-6999 | hero | 英雄模块 (预留) |
| 7001-7999 | quest | 任务模块 (预留) |
| 8001-8999 | mail | 邮件模块 (预留) |
| 9001-9999 | alliance | 联盟模块 (预留) |

### 路由配置

> 文件位置: `config/routes.json`

| MsgID | 模块 | 方法 | 说明 |
|-------|------|------|------|
| 1 | - | handshake | 握手 |
| **1001** | role | login | 登录 |
| **1002** | role | heartbeat | 心跳 |
| **1003** | role | get_info | 获取角色信息 |
| **1004** | role | update_name | 更新角色名 |
| **2001** | item | list | 物品列表 |
| **2002** | item | use | 使用物品 |
| **2003** | item | add | 添加物品 |
| **3001** | scene | enter | 进入场景 |
| **3002** | scene | move | 移动 |
| **3003** | scene | leave | 离开场景 |
| **3004** | scene | get_nearby | 获取附近实体 |
| **3005** | scene | get_scene_info | 获取场景信息 |
| **4001** | march | create_army | 创建军队 |
| **4002** | march | delete_army | 解散军队 |
| **4003** | march | get_armies | 获取军队列表 |
| **4004** | march | start_march | 开始行军 |
| **4005** | march | cancel_march | 取消行军 |
| **4006** | march | get_march_info | 获取行军信息 |

---

## 模块接口

### 1. 系统模块 (System)

#### 1.1 握手 (Handshake)

**MsgID**: 1

**请求**:
```json
{
    "token": "player_auth_token"
}
```

**响应**:
```json
{
    "code": 0,
    "message": "success"
}
```

---

### 2. 角色模块 (Role)

> 模块位置: `plugin/role/` (支持热更)

#### 2.1 登录 (login)

**MsgID**: 1001

**请求**:
```json
{
    "token": "auth_token"
}
```

**响应**:
```json
{
    "rid": 10001,
    "name": "Player001"
}
```

#### 2.2 心跳 (heartbeat)

**MsgID**: 1002

**请求**: `{}`

**响应**: `{}`

#### 2.3 获取角色信息 (get_info)

**MsgID**: 1003

**请求**:
```json
{
    "rid": 10001
}
```

**响应**:
```json
{
    "rid": 10001,
    "name": "Player001",
    "level": 10,
    "exp": 1000
}
```

#### 2.4 更新角色名 (update_name)

**MsgID**: 1004

**请求**:
```json
{
    "name": "NewName"
}
```

**响应**:
```json
{
    "rid": 10001,
    "name": "NewName"
}
```

---

### 3. 物品模块 (Item)

> 模块位置: `plugin/item/` (支持热更)

#### 3.1 物品列表 (list)

**MsgID**: 2001

**请求**:
```json
{
    "type": 1
}
```

**响应**:
```json
{
    "items": [
        {"item_id": 1001, "count": 10},
        {"item_id": 1002, "count": 5}
    ],
    "total": 2
}
```

#### 3.2 使用物品 (use)

**MsgID**: 2002

**请求**:
```json
{
    "item_id": 1001,
    "count": 1
}
```

**响应**:
```json
{
    "item_id": 1001,
    "remaining": 9
}
```

#### 3.3 添加物品 (add)

**MsgID**: 2003

**请求**:
```json
{
    "item_id": 1001,
    "count": 5
}
```

**响应**:
```json
{
    "item_id": 1001,
    "count": 14
}
```

---

### 4. 场景模块 (Scene)

> 模块位置: `internal/scene/` (核心模块)

#### 4.1 进入场景 (enter)

**MsgID**: 3001

**请求**:
```json
{
    "scene_id": 1,
    "x": 100.0,
    "y": 100.0
}
```

**响应**:
```json
{
    "scene_id": 1,
    "position": {"x": 100.0, "y": 100.0},
    "entities": []
}
```

#### 4.2 移动 (move)

**MsgID**: 3002

**请求**:
```json
{
    "x": 150.0,
    "y": 150.0
}
```

**响应**:
```json
{
    "position": {"x": 150.0, "y": 150.0}
}
```

#### 4.3 离开场景 (leave)

**MsgID**: 3003

**请求**: `{}`

**响应**:
```json
{
    "message": "left scene"
}
```

#### 4.4 获取附近实体 (get_nearby)

**MsgID**: 3004

**请求**:
```json
{
    "radius": 100.0
}
```

**响应**:
```json
{
    "entities": [
        {"id": 1, "type": "player", "position": {"x": 110, "y": 110}},
        {"id": 2, "type": "monster", "position": {"x": 120, "y": 120}}
    ],
    "count": 2
}
```

#### 4.5 获取场景信息 (get_scene_info)

**MsgID**: 3005

**请求**:
```json
{
    "scene_id": 1
}
```

**响应**:
```json
{
    "scene_id": 1,
    "width": 1000,
    "height": 1000,
    "grid_size": 50,
    "entity_count": 10
}
```

---

### 5. 行军模块 (March)

> 模块位置: `internal/march/` (核心模块)

#### 5.1 创建军队 (create_army)

**MsgID**: 4001

**请求**:
```json
{
    "hero_id": 100,
    "soldiers": 1000,
    "scene_id": 1,
    "x": 100.0,
    "y": 100.0
}
```

**响应**:
```json
{
    "army_id": 1,
    "owner_id": 10001,
    "hero_id": 100,
    "soldiers": 1000,
    "status": "idle"
}
```

#### 5.2 解散军队 (delete_army)

**MsgID**: 4002

**请求**:
```json
{
    "army_id": 1
}
```

**响应**:
```json
{
    "army_id": 1
}
```

#### 5.3 获取军队列表 (get_armies)

**MsgID**: 4003

**请求**: `{}`

**响应**:
```json
{
    "armies": [
        {
            "id": 1,
            "owner_id": 10001,
            "hero_id": 100,
            "soldiers": 1000,
            "status": "idle",
            "position": {"x": 100, "y": 100},
            "scene_id": 1,
            "power": 10000,
            "load_capacity": 100000
        }
    ],
    "count": 1
}
```

#### 5.4 开始行军 (start_march)

**MsgID**: 4004

**请求**:
```json
{
    "army_id": 1,
    "march_type": 1,
    "target_id": 100
}
```

**march_type 取值**:
| 值 | 类型 | 说明 |
|----|------|------|
| 1 | collect | 采集 |
| 2 | attack | 攻击 |
| 3 | reinforce | 支援 |
| 4 | return | 返回 |

**响应**:
```json
{
    "army_id": 1,
    "march_type": "collect",
    "target_id": 100,
    "arrival_time": 1700000000000,
    "progress": 0.0
}
```

#### 5.5 取消行军 (cancel_march)

**MsgID**: 4005

**请求**:
```json
{
    "army_id": 1
}
```

**响应**:
```json
{
    "army_id": 1,
    "message": "march cancelled, returning to base"
}
```

#### 5.6 获取行军信息 (get_march_info)

**MsgID**: 4006

**请求**:
```json
{
    "army_id": 1
}
```

**响应**:
```json
{
    "id": 1,
    "status": "marching",
    "position": {"x": 150, "y": 150},
    "march": {
        "type": "collect",
        "target_id": 100,
        "target_pos": {"x": 200, "y": 200},
        "arrival_time": 1700000000000,
        "progress": 0.5
    }
}
```

---

## 错误码

| Code | 说明 |
|------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权/未登录 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
| 503 | 服务不可用 |

---

## 附录

### A. 相关文件

| 文件 | 说明 |
|------|------|
| `proto/cs/protocol.proto` | Protobuf 消息定义 |
| `config/routes.json` | 消息路由配置 |
| `internal/gate/protocol.go` | 协议编解码 |
| `internal/gate/tcp_server.go` | TCP 服务器 |
| `internal/gate/ws_server.go` | WebSocket 服务器 |

### B. 前端连接示例

#### JavaScript (WebSocket)

```javascript
// 连接服务器
const ws = new WebSocket('ws://localhost:44446/ws');
ws.binaryType = 'arraybuffer';

// 发送请求
function sendRequest(msgId, payload) {
    const header = new Uint8Array([
        (msgId >> 8) & 0xFF,
        msgId & 0xFF,
        0, 0  // sequence
    ]);
    const data = new Uint8Array(1 + 2 + payload.length);
    data[0] = 0x01;  // MsgTypeRequest
    data.set(header, 1);
    data.set(payload, 3);
    ws.send(data);
}

// 登录
ws.onopen = () => {
    sendRequest(1001, new TextEncoder().encode('{"token":"test"}'));
};

// 接收消息
ws.onmessage = (event) => {
    const data = new Uint8Array(event.data);
    const msgType = data[0];
    console.log('Received:', msgType, data.slice(1));
};
```

### C. 添加新协议流程

1. **分配 MsgID** - 根据模块范围分配
2. **更新路由配置** - `config/routes.json`
3. **实现处理方法** - 在对应模块中添加
4. **更新本文档** - 添加接口说明

---

*最后更新: 2024-01-XX*
