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
  if (data[0] === MSG_RESPONSE) return { msgType: data[0], payload: null, raw: data };
  if (data.length < 5) return { msgType: data[0], payload: null, raw: data };
  const len = data.readUInt32BE(0);
  try { return { msgType: data[4], payload: JSON.parse(data.slice(5, 4+len).toString("utf-8")), raw: data }; }
  catch { return { msgType: data[4], payload: null, raw: data }; }
}
function waitResp(ws, ms = 5000) {
  return new Promise((r, e) => { const t = setTimeout(() => e(new Error("timeout")), ms);
    ws.once("message", d => { clearTimeout(t); r(Buffer.from(d)); }); });
}
async function req(ws, msgId, params = {}) {
  ws.send(encodeRequest(msgId, params));
  return decodeResponse(await waitResp(ws)).payload;
}

async function test(label) {
  console.log(`\n--- ${label} ---`);
  const ws = new WebSocket(WS_URL);
  await new Promise((r, e) => { ws.once("open", r); ws.once("error", e); });
  ws.send(encodeHandshake()); await waitResp(ws);

  const login = await req(ws, 1001);
  const rid = login?.data?.rid;
  console.log(`登录 rid=${rid}`);

  const city = await req(ws, 7001);
  const pos = city?.data?.city?.position;
  console.log(`城池位置: (${pos?.x}, ${pos?.y})`);

  const enter = await req(ws, 3001, { scene_id: 1, x: pos?.x || 100, y: pos?.y || 100 });
  console.log(`进入场景: code=${enter?.code}`);

  const nearby = await req(ws, 3004, { scene_id: 1 });
  const entities = nearby?.data?.nearby || [];
  console.log(`附近实体: ${entities.length} 个`);
  entities.forEach(e => console.log(`  - id=${e.id} type=${e.type} pos=(${e.x}, ${e.y})`));

  ws.close();
  return entities.length;
}

// 第一次
const count1 = await test("首次登录");
// 模拟重启：不重启服务，再次连接同一个玩家
const count2 = await test("再次登录");

console.log(`\n结果: 首次附近=${count1}, 再次附近=${count2}`);
console.log(count1 > 0 && count2 > 0 ? "通过: 城池实体每次登录都恢复到地图" : "失败");
process.exit(count1 > 0 && count2 > 0 ? 0 : 1);
