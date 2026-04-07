/**
 * WG AI - WebSocket API 封装
 * 用于 Three.js + TypeScript 前端
 */

import { MsgID } from './protocol';
import type {
  ApiResponse,
  BuildQueueResponse,
  CancelBuildResponse,
  CityInfoResponse,
  CityProductionResponse,
  UpgradeBuildingResponse,
} from './types';

/** 请求参数类型 */
type RequestParams = Record<string, unknown>;

/** WebSocket 配置 */
export interface WSConfig {
  url: string;
  reconnect?: boolean;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  heartbeatInterval?: number;
}

/** 事件回调类型 */
export type EventCallback<T = unknown> = (data: T) => void;

/** 请求回调类型 */
interface RequestCallback {
  resolve: (response: ApiResponse) => void;
  reject: (error: Error) => void;
  timeout: ReturnType<typeof setTimeout>;
}

/**
 * WebSocket 客户端
 */
export class WSClient {
  private ws: WebSocket | null = null;
  private config: Required<WSConfig>;
  private requestId = 0;
  private pendingRequests = new Map<number, RequestCallback>();
  private eventHandlers = new Map<number, Set<EventCallback>>();
  private reconnectAttempts = 0;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private isConnected = false;

  // 事件监听器
  public onConnected?: () => void;
  public onDisconnected?: () => void;
  public onError?: (error: Event) => void;

  constructor(config: WSConfig) {
    this.config = {
      url: config.url,
      reconnect: config.reconnect ?? true,
      reconnectInterval: config.reconnectInterval ?? 3000,
      maxReconnectAttempts: config.maxReconnectAttempts ?? 5,
      heartbeatInterval: config.heartbeatInterval ?? 30000,
    };
  }

  /**
   * 连接服务器
   */
  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.config.url);

        this.ws.onopen = () => {
          this.isConnected = true;
          this.reconnectAttempts = 0;
          this.startHeartbeat();
          this.onConnected?.();
          resolve();
        };

        this.ws.onclose = () => {
          this.isConnected = false;
          this.stopHeartbeat();
          this.onDisconnected?.();
          this.handleReconnect();
        };

        this.ws.onerror = (error) => {
          this.onError?.(error);
          if (!this.isConnected) {
            reject(new Error('Connection failed'));
          }
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event.data);
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * 断开连接
   */
  disconnect(): void {
    this.stopHeartbeat();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.isConnected = false;
  }

  /**
   * 发送请求
   */
  request<T = unknown>(msgId: number, params: RequestParams = {}): Promise<ApiResponse<T>> {
    return new Promise((resolve, reject) => {
      if (!this.ws || !this.isConnected) {
        reject(new Error('Not connected'));
        return;
      }

      const requestId = ++this.requestId;
      const message = {
        msg_id: msgId,
        rid: requestId,
        ...params,
      };

      // 设置超时
      const timeout = setTimeout(() => {
        this.pendingRequests.delete(requestId);
        reject(new Error('Request timeout'));
      }, 10000);

      this.pendingRequests.set(requestId, {
        resolve: resolve as (r: ApiResponse) => void,
        reject,
        timeout,
      });

      this.ws.send(JSON.stringify(message));
    });
  }

  /**
   * 监听推送消息
   */
  on<T = unknown>(msgId: number, callback: EventCallback<T>): () => void {
    if (!this.eventHandlers.has(msgId)) {
      this.eventHandlers.set(msgId, new Set());
    }
    this.eventHandlers.get(msgId)!.add(callback as EventCallback);

    // 返回取消监听函数
    return () => {
      this.eventHandlers.get(msgId)?.delete(callback as EventCallback);
    };
  }

  /**
   * 取消监听
   */
  off(msgId: number, callback: EventCallback): void {
    this.eventHandlers.get(msgId)?.delete(callback);
  }

  /**
   * 处理收到的消息
   */
  private handleMessage(data: string): void {
    try {
      const response = JSON.parse(data) as ApiResponse & { rid?: number; msg_id?: number };

      // 如果有 rid，则是请求响应
      if (response.rid && this.pendingRequests.has(response.rid)) {
        const callback = this.pendingRequests.get(response.rid)!;
        clearTimeout(callback.timeout);
        this.pendingRequests.delete(response.rid);
        callback.resolve(response);
        return;
      }

      // 如果有 msg_id，则是推送消息
      if (response.msg_id) {
        const handlers = this.eventHandlers.get(response.msg_id);
        if (handlers) {
          handlers.forEach((handler) => handler(response.data ?? response));
        }
      }
    } catch (error) {
      console.error('Failed to parse message:', error);
    }
  }

  /**
   * 开始心跳
   */
  private startHeartbeat(): void {
    this.heartbeatTimer = setInterval(() => {
      this.request(MsgID.Role.Heartbeat).catch(console.error);
    }, this.config.heartbeatInterval);
  }

  /**
   * 停止心跳
   */
  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  /**
   * 处理重连
   */
  private handleReconnect(): void {
    if (!this.config.reconnect) return;
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) return;

    this.reconnectAttempts++;
    setTimeout(() => {
      this.connect().catch(console.error);
    }, this.config.reconnectInterval);
  }
}

// ============ API 封装 ============

/**
 * 游戏 API
 */
export class GameAPI {
  constructor(private client: WSClient) {}

  // ============ 角色 ============

  /** 登录 */
  login(token: string) {
    return this.client.request(MsgID.Role.Login, { token });
  }

  /** 获取玩家信息 */
  getRoleInfo() {
    return this.client.request(MsgID.Role.GetInfo);
  }

  // ============ 物品 ============

  /** 获取物品列表 */
  getItems() {
    return this.client.request(MsgID.Item.List);
  }

  /** 使用物品 */
  useItem(itemId: number, count: number = 1) {
    return this.client.request(MsgID.Item.Use, { item_id: itemId, count });
  }

  // ============ 场景 ============

  /** 进入场景 */
  enterScene(sceneId: number) {
    return this.client.request(MsgID.Scene.Enter, { scene_id: sceneId });
  }

  /** 移动 */
  move(x: number, y: number) {
    return this.client.request(MsgID.Scene.Move, { x, y });
  }

  /** 离开场景 */
  leaveScene() {
    return this.client.request(MsgID.Scene.Leave);
  }

  /** 获取附近对象 */
  getNearby() {
    return this.client.request(MsgID.Scene.GetNearby);
  }

  /** 获取场景信息 */
  getSceneInfo(sceneId: number) {
    return this.client.request(MsgID.Scene.GetSceneInfo, { scene_id: sceneId });
  }

  // ============ 行军 ============

  /** 创建军队 */
  createArmy(heroId: number, soldiers: Record<number, number>, sceneId: number, x: number, y: number) {
    return this.client.request(MsgID.March.CreateArmy, {
      hero_id: heroId,
      soldiers,
      scene_id: sceneId,
      x,
      y,
    });
  }

  /** 解散军队 */
  deleteArmy(armyId: number) {
    return this.client.request(MsgID.March.DeleteArmy, { army_id: armyId });
  }

  /** 获取军队列表 */
  getArmies() {
    return this.client.request(MsgID.March.GetArmies);
  }

  /** 开始行军 */
  startMarch(armyId: number, marchType: number, targetId: number) {
    return this.client.request(MsgID.March.StartMarch, {
      army_id: armyId,
      march_type: marchType,
      target_id: targetId,
    });
  }

  /** 取消行军 */
  cancelMarch(armyId: number) {
    return this.client.request(MsgID.March.CancelMarch, { army_id: armyId });
  }

  // ============ 士兵 ============

  /** 获取士兵列表 */
  getSoldiers() {
    return this.client.request(MsgID.Soldier.List);
  }

  /** 获取士兵配置 */
  getSoldierConfigs() {
    return this.client.request(MsgID.Soldier.Configs);
  }

  /** 训练士兵 */
  trainSoldier(type: number, level: number, count: number, isUpgrade = false) {
    return this.client.request(MsgID.Soldier.Train, {
      type,
      level,
      count,
      is_upgrade: isUpgrade,
    });
  }

  /** 取消训练 */
  cancelTrain(queueId: number) {
    return this.client.request(MsgID.Soldier.CancelTrain, { queue_id: queueId });
  }

  /** 获取训练队列 */
  getTrainQueue() {
    return this.client.request(MsgID.Soldier.TrainQueue);
  }

  /** 完成训练 */
  completeTrain() {
    return this.client.request(MsgID.Soldier.CompleteTrain);
  }

  /** 治疗伤兵 */
  healSoldiers(soldiers: Array<{ soldier_id: number; count: number }>) {
    return this.client.request(MsgID.Soldier.Heal, { soldiers });
  }

  /** 获取治疗队列 */
  getHealQueue() {
    return this.client.request(MsgID.Soldier.HealQueue);
  }

  /** 完成治疗 */
  completeHeal() {
    return this.client.request(MsgID.Soldier.CompleteHeal);
  }

  /** 解散士兵 */
  dismissSoldier(soldierId: number, count: number) {
    return this.client.request(MsgID.Soldier.Dismiss, {
      soldier_id: soldierId,
      count,
    });
  }

  /** 获取士兵统计 */
  getSoldierStats() {
    return this.client.request(MsgID.Soldier.Stats);
  }

  // ============ 城池 ============

  /** 获取城池信息 */
  getCityInfo() {
    return this.client.request<CityInfoResponse>(MsgID.City.GetInfo);
  }

  /** 升级建筑 */
  upgradeBuilding(buildingType: number) {
    return this.client.request<UpgradeBuildingResponse>(MsgID.City.Upgrade, {
      building_type: buildingType,
    });
  }

  /** 取消建造 */
  cancelBuild(queueId: number) {
    return this.client.request<CancelBuildResponse>(MsgID.City.CancelBuild, {
      queue_id: queueId,
    });
  }

  /** 获取建造队列 */
  getBuildQueue() {
    return this.client.request<BuildQueueResponse>(MsgID.City.BuildQueue);
  }

  /** 获取资源产出 */
  getCityProduction() {
    return this.client.request<CityProductionResponse>(MsgID.City.Production);
  }
}
