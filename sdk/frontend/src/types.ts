/**
 * WG AI - 前端类型定义
 * 从 protobuf 自动生成或手动维护
 */

// ============ 基础类型 ============

/** 玩家ID */
export type RoleID = bigint;

/** 资源类型 */
export interface Resources {
  food: number;
  wood: number;
  stone: number;
  gold: number;
}

/** 位置 */
export interface Position {
  x: number;
  y: number;
}

// ============ 角色 ============

/** 玩家信息 */
export interface RoleInfo {
  rid: RoleID;
  name: string;
  level: number;
  exp: number;
  vip: number;
  resources: Resources;
  position: Position;
}

// ============ 物品 ============

/** 物品数据 */
export interface ItemData {
  id: number;
  configId: number;
  count: number;
}

// ============ 士兵 ============

/** 士兵配置 */
export interface SoldierConfig {
  id: number;           // 士兵ID (type*100 + level)
  type: number;         // 兵种类型: 1=步兵, 2=骑兵, 3=弓兵, 4=攻城
  level: number;        // 等级 1-5
  name: string;         // 名称
  attack: number;       // 攻击力
  defense: number;      // 防御力
  hp: number;           // 生命值
  speed: number;        // 移动速度
  load: number;         // 负重
  power: number;        // 战力
  costFood: number;     // 训练消耗粮食
  costWood: number;     // 训练消耗木材
  costStone: number;    // 训练消耗石材
  costGold: number;     // 训练消耗金币
  trainTime: number;    // 训练时间(秒)
  healFood: number;     // 治疗消耗粮食
  healWood: number;     // 治疗消耗木材
  healTime: number;     // 治疗时间(秒)
}

/** 士兵数据 */
export interface SoldierData {
  id: number;           // 士兵ID
  type: number;         // 兵种类型
  level: number;        // 等级
  count: number;        // 健康数量
  wounded: number;      // 受伤数量
}

/** 训练队列项 */
export interface TrainQueueItem {
  id: number;           // 队列ID
  soldierId: number;    // 士兵ID
  soldierType: number;  // 兵种类型
  level: number;        // 等级
  count: number;        // 数量
  startTime: number;    // 开始时间(毫秒)
  finishTime: number;   // 完成时间(毫秒)
  isUpgrade: boolean;   // 是否晋升训练
}

/** 治疗队列项 */
export interface HealQueueItem {
  id: number;           // 队列ID
  soldiers: Record<number, number>; // soldierId -> count
  startTime: number;    // 开始时间(毫秒)
  finishTime: number;   // 完成时间(毫秒)
}

// ============ 军队/行军 ============

/** 军队状态 */
export enum ArmyStatus {
  Idle = 'idle',           // 空闲
  Marching = 'marching',   // 行军中
  Collecting = 'collecting', // 采集中
  Battle = 'battle',       // 战斗中
  Stationing = 'stationing' // 驻扎中
}

/** 行军类型 */
export enum MarchType {
  Collect = 'collect',     // 采集
  Attack = 'attack',       // 攻击
  Reinforce = 'reinforce', // 支援
  Return = 'return'        // 返回
}

/** 军队数据 */
export interface ArmyData {
  id: number;
  ownerId: RoleID;
  heroId: number;
  soldiers: Record<number, number>; // soldierId -> count
  status: ArmyStatus;
  position: Position;
  sceneId: number;
  march?: MarchData;
  load?: Resources;
}

/** 行军数据 */
export interface MarchData {
  type: MarchType;
  targetId: number;
  targetPos: Position;
  path: Position[];
  startTime: number;
  arrivalTime: number;
  speed: number;
  progress: number;
  collectEndTime?: number;
}

// ============ 场景 ============

/** 实体类型 */
export enum EntityType {
  Player = 'player',
  Resource = 'resource',
  Monster = 'monster',
  City = 'city',
  Building = 'building'
}

/** 资源类型 */
export enum ResourceType {
  Food = 1,
  Wood = 2,
  Stone = 3,
  Gold = 4
}

/** 场景对象 */
export interface SceneObject {
  id: number;
  type: EntityType;
  position: Position;
  ownerId?: RoleID;
  level?: number;
  resourceType?: ResourceType;
  resourceAmount?: number;
}

/** 场景信息 */
export interface SceneInfo {
  id: number;
  width: number;
  height: number;
  objects: SceneObject[];
}

// ============ 通用响应 ============

/** 通用 API 响应 */
export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data?: T;
}

/** 服务器推送消息 */
export interface PushMessage {
  msgId: number;
  data: unknown;
}
