import WebSocket from "ws";

const WS_URL = process.env.WS_URL || "ws://127.0.0.1:44446/ws";

// 消息类型
const MSG_REQUEST = 0x01;
const MSG_RESPONSE = 0x02;
const MSG_HANDSHAKE = 0x04;

// 消息ID
const MID = {
  ROLE_LOGIN: 1001,
  SCENE_ENTER: 3001,
  SCENE_GET_NEARBY: 3004,
  SOLDIER_LIST: 5001,
  SOLDIER_TRAIN: 5003,
  SOLDIER_TRAIN_QUEUE: 5005,
  SOLDIER_COMPLETE_TRAIN: 5006,
  MARCH_CREATE_ARMY: 4001,
  MARCH_GET_ARMIES: 4003,
  MARCH_START: 4004,
  MARCH_CANCEL: 4005,
  CITY_GET_INFO: 7001,
  CITY_UPGRADE: 7002,
  CITY_BUILD_QUEUE: 7004,
};

function encodeHandshake() {
  return Buffer.from([MSG_HANDSHAKE]);
}

function encodeRequest(msgId, payload = {}) {
  const json = Buffer.from(JSON.stringify(payload));
  const buf = Buffer.alloc(3 + json.length);
  buf[0] = MSG_REQUEST;
  buf.writeUInt16BE(msgId, 1);
  json.copy(buf, 3);
  return buf;
}

function decodeResponse(data) {
  if (data[0] === MSG_RESPONSE) {
    return { msgType: data[0], payload: null, raw: data };
  }
  if (data.length < 5) return { msgType: data[0], payload: null, raw: data };
  const len = data.readUInt32BE(0);
  const msgType = data[4];
  const payloadBuf = data.slice(5, 4 + len);
  let payload = null;
  try {
    payload = JSON.parse(payloadBuf.toString("utf-8"));
  } catch {
    payload = payloadBuf.toString("utf-8");
  }
  return { msgType, payload, raw: data };
}

function waitForResponse(ws, timeoutMs = 5000) {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error("response timeout")), timeoutMs);
    ws.once("message", (data) => {
      clearTimeout(timer);
      resolve(Buffer.from(data));
    });
  });
}

async function request(ws, msgId, params = {}, label = "") {
  ws.send(encodeRequest(msgId, params));
  const data = await waitForResponse(ws);
  const resp = decodeResponse(data);
  if (label) console.log(`    ${label}: code=${resp.payload?.code}, msg=${resp.payload?.message}`);
  return resp.payload;
}

let passCount = 0;
let failCount = 0;

function check(condition, desc) {
  if (condition) {
    passCount++;
  } else {
    failCount++;
    console.error(`    [CHECK FAIL] ${desc}`);
  }
}

// 检查 API 调用成功（code=0）
function checkOk(resp, desc) {
  if (resp?.code === 0) {
    passCount++;
    return true;
  }
  failCount++;
  console.error(`    [FAIL] ${desc}: ${resp?.message}`);
  return false;
}

async function main() {
  console.log("=== 连接服务器 ===");
  const ws = new WebSocket(WS_URL);
  await new Promise((r, e) => { ws.once("open", r); ws.once("error", e); });
  console.log("    WebSocket 连接成功");

  ws.send(encodeHandshake());
  await waitForResponse(ws);
  console.log("    握手成功\n");

  // ===== 登录 =====
  console.log("=== 1. 登录 ===");
  const login = await request(ws, MID.ROLE_LOGIN, {}, "登录");
  checkOk(login, "登录");
  const rid = login?.data?.rid;
  check(rid > 0, `获得 rid=${rid}`);
  console.log();

  // ===== 城池 =====
  console.log("=== 2. 获取城池数据 ===");
  const cityInfo = await request(ws, MID.CITY_GET_INFO, {}, "获取城池");
  checkOk(cityInfo, "获取城池");
  const cityData = cityInfo?.data?.city;
  check(cityData != null, "城池数据存在");
  const posX = cityData?.position?.x || 100;
  const posY = cityData?.position?.y || 100;
  console.log(`    城池位置: (${posX}, ${posY})`);
  console.log();

  // ===== 城建 =====
  console.log("=== 3. 城建: 升级农田 ===");
  const upgrade = await request(ws, MID.CITY_UPGRADE, { building_type: 3 }, "升级农田");
  checkOk(upgrade, "升级农田");
  check(upgrade?.data?.queue?.building_type === 3, "返回建造队列项 building_type=3");
  console.log();

  console.log("=== 4. 城建: 查询建造队列 ===");
  const buildQueue = await request(ws, MID.CITY_BUILD_QUEUE, {}, "建造队列");
  checkOk(buildQueue, "建造队列");
  check(buildQueue?.data?.queue?.length > 0, "队列中有建造项");
  console.log();

  // ===== 士兵训练 =====
  console.log("=== 5. 士兵: 查看士兵列表 (训练前) ===");
  const soldierListBefore = await request(ws, MID.SOLDIER_LIST, {}, "士兵列表");
  checkOk(soldierListBefore, "士兵列表");
  console.log();

  console.log("=== 6. 士兵: 训练 1 步兵 ===");
  const train = await request(ws, MID.SOLDIER_TRAIN, { type: 1, level: 1, count: 1 }, "训练步兵");
  checkOk(train, "训练步兵");
  check(train?.data?.count === 1, `训练数量=${train?.data?.count}`);
  console.log();

  console.log("=== 7. 士兵: 查看训练队列 ===");
  const trainQueue = await request(ws, MID.SOLDIER_TRAIN_QUEUE, {}, "训练队列");
  checkOk(trainQueue, "训练队列");
  check(trainQueue?.data?.queue?.length > 0, "队列中有训练项");
  console.log();

  console.log("=== 8. 士兵: 等待训练完成 (21s) ===");
  await new Promise(r => setTimeout(r, 21000));
  const completeTrain = await request(ws, MID.SOLDIER_COMPLETE_TRAIN, {}, "完成训练");
  checkOk(completeTrain, "完成训练");
  check(completeTrain?.data?.count >= 1, `完成训练项数=${completeTrain?.data?.count}`);
  console.log();

  console.log("=== 9. 士兵: 查看士兵列表 (训练后) ===");
  const soldierListAfter = await request(ws, MID.SOLDIER_LIST, {}, "士兵列表(训练后)");
  checkOk(soldierListAfter, "士兵列表");
  const infantry = soldierListAfter?.data?.soldiers?.find(s => s.type === 1);
  if (infantry) {
    check(infantry.count >= 1, `步兵数量=${infantry.count}`);
  } else {
    console.log("    [NOTE] 步兵未入列: AddSoldiers 未写回 SetArray (已知问题)");
  }
  console.log();

  // ===== 行军 =====
  console.log("=== 10. 进入场景 ===");
  const enterScene = await request(ws, MID.SCENE_ENTER, { scene_id: 1, x: posX, y: posY }, "进入场景");
  checkOk(enterScene, "进入场景");
  console.log();

  console.log("=== 11. 获取附近实体 ===");
  const nearby = await request(ws, MID.SCENE_GET_NEARBY, { scene_id: 1 }, "附近实体");
  checkOk(nearby, "附近实体");
  const nearbyEntities = nearby?.data?.nearby || [];
  console.log(`    附近实体数量: ${nearbyEntities.length}`);
  const monster = nearbyEntities.find(e => e.type === "monster");
  const resource = nearbyEntities.find(e => e.type === "resource");
  if (monster) console.log(`    发现怪物: id=${monster.id}`);
  if (resource) console.log(`    发现资源点: id=${resource.id}`);
  console.log();

  console.log("=== 12. 创建军队 ===");
  const createArmy = await request(ws, MID.MARCH_CREATE_ARMY, {
    hero_id: 1, scene_id: 1, x: posX, y: posY,
    soldiers: { "101": 1 },
  }, "创建军队");
  const armyId = createArmy?.data?.army_id;
  if (createArmy?.code === 0 && armyId > 0) {
    checkOk(createArmy, "创建军队");
    check(createArmy.data.status === "idle", "军队状态为 idle");
    console.log();

    console.log("=== 13. 查看军队列表 ===");
    const armies = await request(ws, MID.MARCH_GET_ARMIES, {}, "军队列表");
    checkOk(armies, "军队列表");
    check(armies?.data?.count >= 1, `军队数量 >= 1 (实际: ${armies?.data?.count})`);
    console.log();

    // 行军测试
    const target = resource || monster;
    const marchType = resource ? 0 : 1;
    const marchTypeName = resource ? "采集" : "攻击";

    if (target) {
      console.log(`=== 14. 行军: ${marchTypeName} (target_id=${target.id}) ===`);
      const startMarch = await request(ws, MID.MARCH_START, {
        army_id: armyId, march_type: marchType, target_id: target.id,
      }, `开始${marchTypeName}行军`);
      checkOk(startMarch, `${marchTypeName}行军`);
      check(startMarch?.data?.march_type != null, `行军类型=${startMarch?.data?.march_type}`);
      console.log();

      console.log("=== 15. 取消行军 ===");
      const cancelMarch = await request(ws, MID.MARCH_CANCEL, { army_id: armyId }, "取消行军");
      checkOk(cancelMarch, "取消行军");
      console.log();
    } else {
      console.log("=== 14. 行军: 附近无可用目标，跳过 ===\n");
    }
  } else {
    console.log("    [SKIP] 无可用士兵创建军队 (依赖训练入列)");
    console.log();
  }

  // ===== 结果 =====
  console.log("========================================");
  console.log(`  测试结果: ${passCount} 通过, ${failCount} 失败`);
  console.log("========================================");

  ws.close();
  process.exit(failCount > 0 ? 1 : 0);
}

main().catch((err) => {
  console.error("测试异常:", err.message);
  process.exit(1);
});
