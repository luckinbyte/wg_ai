# WG AI Frontend SDK

## 目录结构

```
sdk/frontend/
├── src/
│   ├── index.ts      # 入口文件，导出所有模块
│   ├── types.ts      # 类型定义
│   ├── protocol.ts   # 消息ID常量
│   └── api.ts        # WebSocket 客户端和 API 封装
├── config/
│   ├── routes.json   # 路由配置
│   └── soldier.json  # 士兵配置
├── docs/
│   └── README.md     # 本文档
└── package.json
```

## 安装

将 `sdk/frontend/` 目录复制到你的前端项目中。

## 使用示例

### 1. 连接服务器

```typescript
import { WSClient, GameAPI } from './sdk/frontend/src';

// 创建 WebSocket 客户端
const client = new WSClient({
  url: 'ws://localhost:8080/ws',
  reconnect: true,
  heartbeatInterval: 30000,
});

// 创建 API 实例
const api = new GameAPI(client);

// 连接事件
client.onConnected = () => {
  console.log('已连接到服务器');
};

client.onDisconnected = () => {
  console.log('与服务器断开连接');
};

// 连接服务器
await client.connect();
```

### 2. 登录

```typescript
// 使用 token 登录
const response = await api.login('your-token-here');

if (response.code === 0) {
  console.log('登录成功', response.data);
}
```

### 3. 获取玩家信息

```typescript
const info = await api.getRoleInfo();
console.log('玩家信息:', info.data);
```

### 4. 场景操作

```typescript
// 进入场景
await api.enterScene(1);

// 移动
await api.move(100, 200);

// 获取附近对象
const nearby = await api.getNearby();
```

### 5. 士兵操作

```typescript
// 获取士兵配置
const configs = await api.getSoldierConfigs();

// 训练士兵 (类型1=步兵, 等级1, 数量100)
await api.trainSoldier(1, 1, 100);

// 查看训练队列
const queue = await api.getTrainQueue();

// 完成训练
await api.completeTrain();
```

### 6. 监听推送消息

```typescript
import { PushMsgID } from './sdk/frontend/src';

// 监听场景进入事件
client.on(PushMsgID.SceneEnter, (data) => {
  console.log('有玩家进入视野:', data);
});

// 监听战斗结束
client.on(PushMsgID.BattleEnd, (report) => {
  console.log('战斗结束:', report);
});
```

## 消息ID参考

| 模块 | 消息ID范围 | 说明 |
|------|-----------|------|
| Role | 1001-1099 | 角色相关 |
| Item | 2001-2099 | 物品相关 |
| Scene | 3001-3099 | 场景相关 |
| March | 4001-4099 | 行军相关 |
| Soldier | 5001-5099 | 士兵相关 |
| Push | 6001+ | 服务器推送 |

## 士兵ID规则

士兵ID = 兵种类型 × 100 + 等级

| 类型 | ID范围 | 说明 |
|------|--------|------|
| 步兵 | 101-105 | 1-5级 |
| 骑兵 | 201-205 | 1-5级 |
| 弓兵 | 301-305 | 1-5级 |
| 攻城 | 401-405 | 1-5级 |

## 与 Three.js 集成

```typescript
import * as THREE from 'three';
import { WSClient, GameAPI, PushMsgID } from './sdk/frontend/src';

class GameScene {
  private scene: THREE.Scene;
  private api: GameAPI;
  private playerMeshes: Map<number, THREE.Mesh> = new Map();

  constructor() {
    this.scene = new THREE.Scene();
    // ... 初始化 Three.js 场景

    const client = new WSClient({ url: 'ws://localhost:8080/ws' });
    this.api = new GameAPI(client);

    this.setupEventListeners();
  }

  private setupEventListeners() {
    // 监听玩家进入视野
    this.api['client'].on(PushMsgID.SceneEnter, (data: any) => {
      this.addPlayer(data);
    });

    // 监听玩家离开视野
    this.api['client'].on(PushMsgID.SceneLeave, (data: any) => {
      this.removePlayer(data.rid);
    });
  }

  private addPlayer(data: any) {
    const geometry = new THREE.BoxGeometry(1, 2, 1);
    const material = new THREE.MeshBasicMaterial({ color: 0x00ff00 });
    const mesh = new THREE.Mesh(geometry, material);
    mesh.position.set(data.x, 0, data.y);
    this.scene.add(mesh);
    this.playerMeshes.set(data.rid, mesh);
  }

  private removePlayer(rid: number) {
    const mesh = this.playerMeshes.get(rid);
    if (mesh) {
      this.scene.remove(mesh);
      this.playerMeshes.delete(rid);
    }
  }
}
```
