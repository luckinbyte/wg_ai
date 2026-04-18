import WebSocket from "ws";
const ws = new WebSocket("ws://127.0.0.1:44446/ws");
const HS=0x04,REQ=0x01;
function enc(id,p={}){const j=Buffer.from(JSON.stringify(p));const b=Buffer.alloc(3+j.length);b[0]=REQ;b.writeUInt16BE(id,1);j.copy(b,3);return b;}
function dec(d){try{return JSON.parse(d.slice(5,4+d.readUInt32BE(0)).toString());}catch{return null;}}
function wait(ws,ms=3000){return new Promise((r,e)=>{const t=setTimeout(()=>e('timeout'),ms);ws.once('message',d=>{clearTimeout(t);r(Buffer.from(d));});});}
await new Promise(r=>ws.once('open',r));ws.send(Buffer.from([HS]));await wait(ws);

ws.send(enc(1001));const login=dec(await wait(ws));console.log('1. login rid:',login?.data?.rid);

// 查看初始资源
ws.send(enc(1003));const info=dec(await wait(ws));
console.log('2. resources:',JSON.stringify({food:info?.data?.food,wood:info?.data?.wood,gold:info?.data?.gold}));

// 训练
ws.send(enc(5003,{type:1,level:1,count:1}));
// 收集response + push
const msgs=[];
await new Promise(r=>{const t=setTimeout(r,2000);const h=d=>{msgs.push(Buffer.from(d));if(msgs.length>=3){clearTimeout(t);ws.off('message',h);r();}};ws.on('message',h);});
msgs.forEach((m,i)=>{try{const p=JSON.parse(m.slice(5,4+m.readUInt32BE(0)).toString());console.log('3. train msg['+i+']:',JSON.stringify(p).substring(0,200));}catch{}});

// 立即查看士兵 (训练时间未到，应该为空)
ws.send(enc(5001));const list=dec(await wait(ws));console.log('4. soldiers before complete:',JSON.stringify(list?.data));

// 立即尝试完成训练 (时间未到)
ws.send(enc(5006,{}));const comp=dec(await wait(ws));console.log('5. complete_train (too early):',JSON.stringify(comp?.data));

// 等待训练完成 (20秒)
console.log('6. waiting 22s for training...');
await new Promise(r=>setTimeout(r,22000));

// 再次完成训练
ws.send(enc(5006,{}));const comp2=dec(await wait(ws));console.log('7. complete_train (after wait):',JSON.stringify(comp2?.data));

// 查看士兵
ws.send(enc(5001));const list2=dec(await wait(ws));console.log('8. soldiers after complete:',JSON.stringify(list2?.data));

ws.close();
