import WebSocket from "ws";

const WS_URL = process.env.WS_URL || "ws://127.0.0.1:44446/ws";

// 消息类型常量
const MSG_REQUEST = 0x01;
const MSG_RESPONSE = 0x02;
const MSG_HANDSHAKE = 0x04;

// 消息ID
const MSG_ID = {
  ROLE_LOGIN: 1001,
  CITY_GET_INFO: 7001,
};

// 编码握手包: [MsgType(0x04)]
function encodeHandshake() {
  return Buffer.from([MSG_HANDSHAKE]);
}

// 编码请求包: [MsgType(0x01)][MsgID_H][MsgID_L][JSON payload]
function encodeRequest(msgId, payload = {}) {
  const json = Buffer.from(JSON.stringify(payload));
  const buf = Buffer.alloc(3 + json.length);
  buf[0] = MSG_REQUEST;
  buf.writeUInt16BE(msgId, 1);
  json.copy(buf, 3);
  return buf;
}

/**
 * 解码业务响应包: [Length(4B)][MsgType(0x02)][JSON payload]
 * 握手响应格式不同: [MsgType(0x02)][0x00 0x00 0x00 0x00]
 * 通过判断第 0 字节是否为 MSG_RESPONSE 来区分
 */
function decodeResponse(data) {
  // 握手响应: [0x02][...]，第 0 字节就是 msgType
  if (data[0] === MSG_RESPONSE) {
    return { msgType: data[0], payload: null, raw: data };
  }
  // 业务响应: [Length(4B)][0x02][JSON]
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

// 等待响应
function waitForResponse(ws, timeoutMs = 5000) {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error("response timeout"));
    }, timeoutMs);

    ws.once("message", (data) => {
      clearTimeout(timer);
      resolve(Buffer.from(data));
    });
  });
}

async function main() {
  console.log(`[1] 连接 WebSocket: ${WS_URL}`);
  const ws = new WebSocket(WS_URL);

  await new Promise((resolve, reject) => {
    ws.once("open", resolve);
    ws.once("error", reject);
  });
  console.log("    连接成功\n");

  // === Step 1: 握手 ===
  console.log("[2] 发送握手包");
  ws.send(encodeHandshake());
  const hsData = await waitForResponse(ws);
  const hs = decodeResponse(hsData);
  console.log(`    握手响应: msgType=0x${hs.msgType.toString(16)}, bytes=${hs.raw.toString("hex")}`);
  if (hs.msgType !== MSG_RESPONSE) {
    console.error("    握手失败!");
    ws.close();
    process.exit(1);
  }
  console.log("    握手成功\n");

  // === Step 2: 登录 ===
  console.log("[3] 发送登录请求 (msg_id=1001)");
  ws.send(encodeRequest(MSG_ID.ROLE_LOGIN));
  const loginData = await waitForResponse(ws);
  const login = decodeResponse(loginData);
  console.log(`    登录响应: ${JSON.stringify(login.payload, null, 2)}`);

  if (!login.payload || login.payload.code !== 0) {
    console.error("    登录失败!");
    ws.close();
    process.exit(1);
  }
  console.log("    登录成功\n");

  // === Step 3: 获取城池数据 ===
  console.log("[4] 发送获取城池数据请求 (msg_id=7001)");
  ws.send(encodeRequest(MSG_ID.CITY_GET_INFO));
  const cityData = await waitForResponse(ws);
  const city = decodeResponse(cityData);
  console.log(`    城池响应: ${JSON.stringify(city.payload, null, 2)}`);

  if (!city.payload || city.payload.code !== 0) {
    console.error("    获取城池数据失败!");
    ws.close();
    process.exit(1);
  }
  console.log("    获取城池数据成功\n");

  // === 完成 ===
  console.log("[5] 测试全部通过");
  ws.close();
  process.exit(0);
}

main().catch((err) => {
  console.error("测试失败:", err.message);
  process.exit(1);
});
