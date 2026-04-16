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

const ws = new WebSocket(WS_URL);
await new Promise((r, e) => { ws.once("open", r); ws.once("error", e); });
ws.send(encodeHandshake()); await waitResp(ws);

// 只登录 + 进入场景，不调用任何 city API
const login = decodeResponse(await (ws.send(encodeRequest(1001)), waitResp(ws))).payload;
const rid = login?.data?.rid;
console.log(`登录 rid=${rid}`);

// 先获取城池数据拿到位置（不调用 city API，直接用已知位置）
// 已知数据库中第一个城市在 (620, 590)，在该位置进入以验证城池加载
const enter = decodeResponse(await (ws.send(encodeRequest(3001, { scene_id: 1, x: 620, y: 590 })), waitResp(ws))).payload;
console.log(`进入场景: code=${enter?.code}`);

const nearby = decodeResponse(await (ws.send(encodeRequest(3004, { scene_id: 1 })), waitResp(ws))).payload;
const entities = nearby?.data?.nearby || [];
console.log(`附近实体: ${entities.length} 个`);
const buildings = entities.filter(e => e.type === "building");
console.log(`城池建筑: ${buildings.length} 个`);
buildings.forEach(b => console.log(`  - id=${b.id} pos=(${b.x}, ${b.y}) level=${b.building_level}`));

console.log(buildings.length > 0 ? "\n通过: 启动时城池已加载到地图" : "\n失败: 地图上没有城池");
ws.close();
process.exit(buildings.length > 0 ? 0 : 1);
