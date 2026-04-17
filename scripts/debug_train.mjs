import WebSocket from "ws";
const WS_URL = "ws://127.0.0.1:44446/ws";
const MSG_REQUEST = 0x01, MSG_RESPONSE = 0x02, MSG_HANDSHAKE = 0x04;
function encHS() { return Buffer.from([MSG_HANDSHAKE]); }
function encReq(id, p={}) { const j=Buffer.from(JSON.stringify(p)); const b=Buffer.alloc(3+j.length); b[0]=MSG_REQUEST; b.writeUInt16BE(id,1); j.copy(b,3); return b; }
function dec(data) { try { return JSON.parse(data.slice(5, 4+data.readUInt32BE(0)).toString()); } catch { return null; } }
function wait(ws,ms=3000) { return new Promise((r,e)=>{ const t=setTimeout(()=>e(new Error('timeout')),ms); ws.once('message',d=>{clearTimeout(t);r(Buffer.from(d));}); }); }
const ws=new WebSocket(WS_URL);
await new Promise(r=>ws.once('open',r)); ws.send(encHS()); await wait(ws);

ws.send(encReq(1001)); const login=dec(await wait(ws)); console.log('login:', JSON.stringify(login?.data));
ws.send(encReq(1003)); const info=dec(await wait(ws)); console.log('role info:', JSON.stringify(info?.data));
ws.send(encReq(5003, {type:1, level:1, count:1})); const train=dec(await wait(ws)); console.log('train:', JSON.stringify(train));
ws.send(encReq(5001)); const list=dec(await wait(ws)); console.log('soldiers:', JSON.stringify(list?.data));
ws.close();
