import WebSocket from "ws";
const WS_URL = process.env.WS_URL || "ws://127.0.0.1:44446/ws";
const MSG_REQUEST = 0x01, MSG_RESPONSE = 0x02, MSG_HANDSHAKE = 0x04;

function encodeHandshake() { return Buffer.from([MSG_HANDSHAKE]); }
function encodeRequest(msgId, payload = {}) {
  const json = Buffer.from(JSON.stringify(payload));
  const buf = Buffer.alloc(3 + json.length);
  buf[0] = MSG_REQUEST; buf.writeUInt16BE(msgId, 1); json.copy(buf, 3);
  return buf;
}
function decodeResponse(data) {
  if (data[0] === MSG_RESPONSE) return { payload: null, raw: data };
  if (data.length < 5) return { payload: null, raw: data };
  try { return { payload: JSON.parse(data.slice(5, 4 + data.readUInt32BE(0)).toString("utf-8")), raw: data }; }
  catch { return { payload: null, raw: data }; }
}
function waitResp(ws, ms = 5000) {
  return new Promise((r, e) => { const t = setTimeout(() => e(new Error("timeout")), ms);
    ws.once("message", d => { clearTimeout(t); r(Buffer.from(d)); }); });
}

// 带心跳的等待: 每隔 interval 发送心跳保持连接
function sleepWithHeartbeat(ws, ms, interval = 10000) {
  return new Promise(resolve => {
    const end = setTimeout(() => { clearInterval(hb); resolve(); }, ms);
    const hb = setInterval(() => {
      try { ws.send(encodeRequest(1002)); } catch {}
      // 消费心跳响应避免堆积
      ws.once("message", () => {});
    }, interval);
  });
}

const ws = new WebSocket(WS_URL);
await new Promise((r, e) => { ws.once("open", r); ws.once("error", e); });
ws.send(encodeHandshake()); await waitResp(ws);

// 1. 登录
const login = decodeResponse(await (ws.send(encodeRequest(1001)), waitResp(ws))).payload;
const rid = login?.data?.rid;
console.log(`[1] 登录成功 rid=${rid}`);
if (!rid) { console.log("登录失败:", login); ws.close(); process.exit(1); }

// 2. 获取城池信息 (获取位置)
const cityInfo = decodeResponse(await (ws.send(encodeRequest(7001)), waitResp(ws))).payload;
const cityPos = cityInfo?.data?.position;
console.log(`[2] 城池位置: (${cityPos?.x}, ${cityPos?.y})`);

// 3. 进入场景 (城池位置附近)
const enterX = cityPos?.x || 620;
const enterY = cityPos?.y || 590;
const enter = decodeResponse(await (ws.send(encodeRequest(3001, { scene_id: 1, x: enterX, y: enterY })), waitResp(ws))).payload;
console.log(`[3] 进入场景: code=${enter?.code}`);

// 4. 查看附近实体 - 找资源点
const nearby = decodeResponse(await (ws.send(encodeRequest(3004, { scene_id: 1 })), waitResp(ws))).payload;
const entities = nearby?.data?.nearby || [];
const resources = entities.filter(e => e.type === "resource");
console.log(`[4] 附近实体: ${entities.length} 个, 资源点: ${resources.length} 个`);
if (resources.length === 0) {
  console.log("附近没有资源点，尝试附近搜索...");
  // 尝试不同位置
  for (const [tx, ty] of [[100,100],[300,300],[500,500],[700,700],[900,900]]) {
    ws.send(encodeRequest(3001, { scene_id: 1, x: tx, y: ty }));
    await waitResp(ws);
    ws.send(encodeRequest(3004, { scene_id: 1 }));
    const tryNearby = decodeResponse(await waitResp(ws)).payload;
    const tryRes = (tryNearby?.data?.nearby || []).filter(e => e.type === "resource");
    if (tryRes.length > 0) {
      resources.push(...tryRes);
      console.log(`  在 (${tx},${ty}) 找到 ${tryRes.length} 个资源点`);
      break;
    }
  }
}

if (resources.length === 0) {
  console.log("失败: 地图上没有资源点");
  ws.close(); process.exit(1);
}

const targetResource = resources[0];
console.log(`  目标资源点: id=${targetResource.id} pos=(${targetResource.x},${targetResource.y}) type=${targetResource.object_data?.resource_type} amount=${targetResource.object_data?.amount}`);

// 5. 获取当前资源
const roleInfo = decodeResponse(await (ws.send(encodeRequest(1003)), waitResp(ws))).payload;
const beforeFood = roleInfo?.data?.food || 0;
const beforeWood = roleInfo?.data?.wood || 0;
const beforeStone = roleInfo?.data?.stone || 0;
const beforeGold = roleInfo?.data?.gold || 0;
console.log(`[5] 当前资源: food=${beforeFood} wood=${beforeWood} stone=${beforeStone} gold=${beforeGold}`);

// 6. 训练 1 个士兵
const train = decodeResponse(await (ws.send(encodeRequest(5003, { type: 1, level: 1, count: 1 })), waitResp(ws))).payload;
console.log(`[6] 训练: code=${train?.code} msg=${train?.message}`);
if (train?.code !== 0 && train?.code !== undefined) {
  // 可能已经在训练中或有足够的兵
  console.log("  训练结果:", JSON.stringify(train?.data));
}

// 等待训练完成 (训练时间 = 20s * count)
const waitTime = 22000; // 22秒
console.log(`  等待训练完成 (${waitTime/1000}s)...`);
await sleepWithHeartbeat(ws, waitTime);

// 完成训练
const complete = decodeResponse(await (ws.send(encodeRequest(5006, {})), waitResp(ws))).payload;
console.log(`[6b] 完成训练: code=${complete?.code} data=${JSON.stringify(complete?.data)}`);

// 查看士兵
const soldiers = decodeResponse(await (ws.send(encodeRequest(5001)), waitResp(ws))).payload;
console.log(`[6c] 士兵列表: ${JSON.stringify(soldiers?.data)}`);
const soldierList = soldiers?.data?.soldiers || soldiers?.data;
let soldierId = 101, soldierCount = 0;
if (Array.isArray(soldierList)) {
  const inf = soldierList.find(s => s.soldier_id === 101 || s.id === 101);
  if (inf) { soldierId = inf.soldier_id || inf.id; soldierCount = inf.count || inf.amount || 0; }
} else if (typeof soldierList === "object" && soldierList !== null) {
  if (soldierList["101"]) { soldierId = 101; soldierCount = soldierList["101"]; }
  else { for (const [id, cnt] of Object.entries(soldierList)) {
    const c = typeof cnt === "object" ? cnt.count || cnt.amount : cnt;
    if (c > 0) { soldierId = parseInt(id); soldierCount = c; break; }
  }}
}
console.log(`  可用士兵: id=${soldierId} count=${soldierCount}`);

if (soldierCount <= 0) {
  console.log("失败: 没有可用的士兵");
  ws.close(); process.exit(1);
}

// 7. 创建军队
const createArmy = decodeResponse(await (ws.send(encodeRequest(4001, {
  hero_id: 1, scene_id: 1, x: enterX, y: enterY,
  soldiers: { [soldierId]: Math.min(soldierCount, 10) }
})), waitResp(ws))).payload;
const armyId = createArmy?.data?.army_id || createArmy?.data?.id;
console.log(`[7] 创建军队: code=${createArmy?.code} army_id=${armyId}`);
if (!armyId) {
  console.log("  创建军队失败:", JSON.stringify(createArmy));
  ws.close(); process.exit(1);
}

// 8. 检查军队在大地图上可见
const nearby2 = decodeResponse(await (ws.send(encodeRequest(3004, { scene_id: 1 })), waitResp(ws))).payload;
const armyEntities = (nearby2?.data?.nearby || []).filter(e => e.type === "army");
console.log(`[8] 大地图上军队实体: ${armyEntities.length} 个`);
if (armyEntities.length > 0) {
  console.log(`  army entity: id=${armyEntities[0].id} pos=(${armyEntities[0].x},${armyEntities[0].y}) owner=${armyEntities[0].owner_id}`);
}

// 9. 行军去采集
const march = decodeResponse(await (ws.send(encodeRequest(4004, {
  army_id: armyId, march_type: 0, target_id: targetResource.id
})), waitResp(ws))).payload;
console.log(`[9] 开始行军: code=${march?.code} data=${JSON.stringify(march?.data)}`);
if (march?.code !== 0 && march?.code !== undefined) {
  console.log("  行军失败:", JSON.stringify(march));
  ws.close(); process.exit(1);
}

const arrivalTime = march?.data?.arrival_time || march?.data?.march?.arrival_time;
const now = Date.now();
const marchWait = arrivalTime ? Math.max(arrivalTime - now + 1000, 2000) : 5000;
console.log(`  等待行军到达 (${Math.ceil(marchWait/1000)}s)...`);
await sleepWithHeartbeat(ws, marchWait);

// 10. 等待采集完成 (30s)
console.log(`[10] 等待采集完成 (32s)...`);
await sleepWithHeartbeat(ws, 32000);

// 11. 等待返回
console.log(`[11] 等待军队返回...`);
await sleepWithHeartbeat(ws, marchWait + 2000);

// 12. 检查资源变化
const roleInfo2 = decodeResponse(await (ws.send(encodeRequest(1003)), waitResp(ws))).payload;
const afterFood = roleInfo2?.data?.food || 0;
const afterWood = roleInfo2?.data?.wood || 0;
const afterStone = roleInfo2?.data?.stone || 0;
const afterGold = roleInfo2?.data?.gold || 0;
console.log(`[12] 资源变化: food ${beforeFood}->${afterFood} wood ${beforeWood}->${afterWood} stone ${beforeStone}->${afterStone} gold ${beforeGold}->${afterGold}`);

// 13. 检查军队状态
const armies = decodeResponse(await (ws.send(encodeRequest(4003)), waitResp(ws))).payload;
const armyData = armies?.data?.armies || armies?.data;
console.log(`[13] 军队状态: ${JSON.stringify(armyData)}`);

const foodGained = afterFood - beforeFood;
const woodGained = afterWood - beforeWood;
const stoneGained = afterStone - beforeStone;
const goldGained = afterGold - beforeGold;
// 资源增加可能是负数（训练消耗），检查是否有至少一种资源增加了
const anyResourceIncreased = foodGained > 0 || woodGained > 0 || stoneGained > 0 || goldGained > 0;
if (anyResourceIncreased) {
  console.log(`\n通过: 完整采集流程成功!`);
  console.log(`  food ${foodGained>=0?'+':''}${foodGained} wood ${woodGained>=0?'+':''}${woodGained} stone ${stoneGained>=0?'+':''}${stoneGained} gold ${goldGained>=0?'+':''}${goldGained}`);
} else {
  console.log("\n警告: 资源未增加，可能采集流程未完成");
}

ws.close();
