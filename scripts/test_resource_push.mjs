import WebSocket from "ws";
const WS_URL = "ws://127.0.0.1:44446/ws";
const MSG_REQUEST = 0x01, MSG_RESPONSE = 0x02, MSG_PUSH = 0x03, MSG_HANDSHAKE = 0x04;
function encHS() { return Buffer.from([MSG_HANDSHAKE]); }
function encReq(id, p={}) { const j=Buffer.from(JSON.stringify(p)); const b=Buffer.alloc(3+j.length); b[0]=MSG_REQUEST; b.writeUInt16BE(id,1); j.copy(b,3); return b; }
function dec(data) { try { return JSON.parse(data.slice(5, 4+data.readUInt32BE(0)).toString()); } catch { return null; } }
function wait(ws,ms=3000) { return new Promise((r,e)=>{ const t=setTimeout(()=>e(new Error('timeout')),ms); ws.once("message",d=>{clearTimeout(t);r(Buffer.from(d));}); }); }

const ws = new WebSocket(WS_URL);
await new Promise(r=>ws.once("open",r)); ws.send(encHS()); await wait(ws);

// 登录
ws.send(encReq(1001)); const login = dec(await wait(ws));
const rid = login?.data?.rid;
console.log(`[1] 登录 rid=${rid}`);

// 获取初始资源
ws.send(encReq(1003)); const info1 = dec(await wait(ws));
console.log(`[2] 初始资源: food=${info1?.data?.food} wood=${info1?.data?.wood} stone=${info1?.data?.stone} gold=${info1?.data?.gold}`);

// 训练 1 个士兵 (会触发推送)
ws.send(encReq(5003, {type:1, level:1, count:1}));
// 应收到两个消息: 1个response + 1个push
const msgs = [];
const collector = new Promise(resolve => {
  const timer = setTimeout(resolve, 3000);
  const handler = d => {
    msgs.push(Buffer.from(d));
    if (msgs.length >= 2) { clearTimeout(timer); ws.off("message", handler); resolve(); }
  };
  ws.on("message", handler);
});
await collector;

for (let i = 0; i < msgs.length; i++) {
  const m = msgs[i];
  if (m.length >= 5) {
    try {
      const parsed = JSON.parse(m.slice(5, 4+m.readUInt32BE(0)).toString());
      console.log(`[3] msg[${i}]: code=${parsed.code} data=${JSON.stringify(parsed.data)?.substring(0,120)}`);
    } catch {}
  }
}

// 再次查询资源确认变动
ws.send(encReq(1003)); const info2 = dec(await wait(ws));
console.log(`[4] 训练后资源: food=${info2?.data?.food} wood=${info2?.data?.wood} stone=${info2?.data?.stone} gold=${info2?.data?.gold}`);

const foodDiff = (info2?.data?.food||0) - (info1?.data?.food||0);
const woodDiff = (info2?.data?.wood||0) - (info1?.data?.wood||0);
const goldDiff = (info2?.data?.gold||0) - (info1?.data?.gold||0);
console.log(`  变动: food=${foodDiff} wood=${woodDiff} gold=${goldDiff}`);

ws.close();
