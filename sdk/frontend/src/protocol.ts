/**
 * WG AI - 消息ID定义
 * 与 config/routes.json 保持同步
 */

/** 消息ID常量 */
export const MsgID = {
  // ============ 角色 (1001-1099) ============
  Role: {
    Login: 1001,
    Heartbeat: 1002,
    GetInfo: 1003,
  },

  // ============ 物品 (2001-2099) ============
  Item: {
    List: 2001,
    Use: 2002,
    Add: 2003,
  },

  // ============ 场景 (3001-3099) ============
  Scene: {
    Enter: 3001,
    Move: 3002,
    Leave: 3003,
    GetNearby: 3004,
    GetSceneInfo: 3005,
  },

  // ============ 行军 (4001-4099) ============
  March: {
    CreateArmy: 4001,
    DeleteArmy: 4002,
    GetArmies: 4003,
    StartMarch: 4004,
    CancelMarch: 4005,
    GetMarchInfo: 4006,
  },

  // ============ 士兵 (5001-5099) ============
  Soldier: {
    List: 5001,
    Get: 5002,
    Train: 5003,
    CancelTrain: 5004,
    TrainQueue: 5005,
    CompleteTrain: 5006,
    Heal: 5007,
    CancelHeal: 5008,
    HealQueue: 5009,
    CompleteHeal: 5010,
    Dismiss: 5011,
    Configs: 5012,
    Stats: 5013,
  },
} as const;

/** 推送消息ID (6001+) */
export const PushMsgID = {
  // 场景推送
  SceneEnter: 6001,     // 玩家进入视野
  SceneLeave: 6002,     // 玩家离开视野
  SceneUpdate: 6003,    // 场景对象更新

  // 行军推送
  MarchStart: 6101,     // 开始行军
  MarchArrive: 6102,    // 到达目的地
  MarchReturn: 6103,    // 返回

  // 战斗推送
  BattleStart: 6201,    // 战斗开始
  BattleEnd: 6202,      // 战斗结束

  // 士兵推送
  TrainComplete: 6301,  // 训练完成
  HealComplete: 6302,   // 治疗完成
} as const;

/** 模块名称映射 */
export const ModuleName: Record<number, string> = {
  1001: 'role', 1002: 'role', 1003: 'role',
  2001: 'item', 2002: 'item', 2003: 'item',
  3001: 'scene', 3002: 'scene', 3003: 'scene', 3004: 'scene', 3005: 'scene',
  4001: 'march', 4002: 'march', 4003: 'march', 4004: 'march', 4005: 'march', 4006: 'march',
  5001: 'soldier', 5002: 'soldier', 5003: 'soldier', 5004: 'soldier', 5005: 'soldier',
  5006: 'soldier', 5007: 'soldier', 5008: 'soldier', 5009: 'soldier', 5010: 'soldier',
  5011: 'soldier', 5012: 'soldier', 5013: 'soldier',
};

/** 方法名称映射 */
export const MethodName: Record<number, string> = {
  1001: 'login', 1002: 'heartbeat', 1003: 'get_info',
  2001: 'list', 2002: 'use', 2003: 'add',
  3001: 'enter', 3002: 'move', 3003: 'leave', 3004: 'get_nearby', 3005: 'get_scene_info',
  4001: 'create_army', 4002: 'delete_army', 4003: 'get_armies',
  4004: 'start_march', 4005: 'cancel_march', 4006: 'get_march_info',
  5001: 'list', 5002: 'get', 5003: 'train', 5004: 'cancel_train', 5005: 'train_queue',
  5006: 'complete_train', 5007: 'heal', 5008: 'cancel_heal', 5009: 'heal_queue',
  5010: 'complete_heal', 5011: 'dismiss', 5012: 'configs', 5013: 'stats',
};
