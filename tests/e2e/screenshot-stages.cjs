const {ROOT,GO,temp,executable,sourceHashes}=require('./runtime.cjs');
const {chromium}=require('playwright');
const fs=require('fs'),path=require('path'),net=require('net'),crypto=require('crypto');
const {execFileSync,spawn}=require('child_process');
const root=ROOT,out=path.join(root,'artifacts','stage-screenshots');fs.mkdirSync(out,{recursive:true});
const required=['entry','lobby','setup','wake','summer','fall','winter','year_end','finished','hand'];
const m={startedAt:new Date().toISOString(),viewport:{width:1440,height:1000},screenshots:[],actions:[],browserErrors:[],consoleErrors:[],requestFailures:[],missingStages:required,notes:['Real UI and legal authenticated API only; no source/state/DOM/rules modifications.','76 visitor effects and expansions remain unimplemented.'],maxSteps:1500};
const persist=()=>{m.missingStages=required.filter(x=>!m.screenshots.some(s=>s.stage===x));fs.writeFileSync(path.join(out,'manifest.json'),JSON.stringify(m,null,2))};
const hashes=sourceHashes;
const delay=ms=>new Promise(r=>setTimeout(r,ms));
(async()=>{let browser,server;const tmp=temp('viticulture-stages-');m.temporaryDirectory=tmp;m.sourceHashesBefore=hashes();persist();
try{
execFileSync(GO,['build','-o',path.join(tmp,executable('ee-server')),'./cmd/viticulture'],{cwd:root,timeout:120000});
const port=await new Promise((resolve,reject)=>{const s=net.createServer();s.on('error',reject);s.listen(0,'127.0.0.1',()=>{const p=s.address().port;s.close(()=>resolve(p))})});const base='http://127.0.0.1:'+port;m.base=base;m.dataDirectory=path.join(tmp,'data');
server=spawn(path.join(tmp,executable('ee-server')),['-addr','127.0.0.1:'+port,'-data',m.dataDirectory],{cwd:root,stdio:['ignore','pipe','pipe']});m.serverPID=server.pid;let logs='';server.stdout.on('data',d=>logs+=d);server.stderr.on('data',d=>logs+=d);
for(let i=0;i<100;i++){try{const r=await fetch(base+'/api/health');if(r.ok){m.health=await r.json();break}}catch{}if(i===99)throw Error('server not ready '+logs);await delay(100)}
browser=await chromium.launch({headless:true,args:['--no-sandbox']});const pages=[],revs=[0,0,0],ids=[];
for(let i=0;i<3;i++){const ctx=await browser.newContext({viewport:m.viewport});const p=await ctx.newPage();p.on('pageerror',e=>m.browserErrors.push({player:i,error:e.message}));p.on('console',e=>{if(e.type()==='error')m.consoleErrors.push({player:i,text:e.text()})});p.on('requestfailed',r=>m.requestFailures.push({player:i,url:r.url().split('?')[0],failure:r.failure()}));const cdp=await ctx.newCDPSession(p);await cdp.send('Network.enable');cdp.on('Network.eventSourceMessageReceived',e=>{try{revs[i]=JSON.parse(e.data).revision}catch{}});pages.push(p)}
const snap=async p=>p.evaluate(async()=>{const t=localStorage.getItem('vineyard-ee-token');return(await fetch('/api/state?token='+t)).json()});
const sync=async revision=>{for(let n=0;n<150;n++){if(revs.every(r=>r>=revision))return;await delay(40)}throw Error('SSE revision timeout '+revision+' actual '+revs)};
async function capture(stage,file,p,s,selector){if(s)await sync(s.revision);await p.evaluate(async()=>{await document.fonts.ready;await Promise.all([...document.images].map(i=>i.decode().catch(()=>{})));await new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r)))});if(selector)await p.locator(selector).scrollIntoViewIfNeeded();else await p.evaluate(()=>window.scrollTo(0,0));await delay(150);const metrics=await p.evaluate(()=>({scrollY:window.scrollY,brokenImages:[...document.images].filter(i=>!i.complete||!i.naturalWidth).map(i=>i.src),title:document.querySelector('#season-title')?.textContent,handCards:document.querySelectorAll('#hand .card').length}));const dest=path.join(out,file);if(stage==='hand')await p.locator(selector).screenshot({path:dest});else await p.screenshot({path:dest,fullPage:false});m.screenshots.push({stage,path:dest,phase:s?.phase||'entry',year:s?.year??null,revision:s?.revision??null,sseRevisions:s?[...revs]:null,player:s?.youId,handCount:s?.hand?.length,choice:s?.pendingChoice,players:s?.players?.map(p=>({id:p.id,name:p.name,vp:p.vp,coins:p.coins,handCount:p.handCount})),occupied:s?.spaces?.filter(x=>x.occupied?.length).map(x=>({space:x.id,occupied:x.occupied})),winnerIds:s?.winnerIds,actionIndex:m.actions.length,source:'Actual browser after legal UI/API actions',metrics});persist();console.log('CAPTURE',stage,s?.year,s?.revision)}
await pages[0].goto(base);await capture('entry','01-entry.png',pages[0]);await pages[0].locator('#nickname').fill('山丘庄主');await pages[0].locator('#create-room').click();await pages[0].locator('#code-label').waitFor({state:'visible'});let s=await snap(pages[0]);m.roomCode=s.code;
for(let i=1;i<3;i++){await pages[i].goto(base+'/?room='+s.code);await pages[i].locator('#nickname').fill(['','河谷庄主','林间庄主'][i]);await pages[i].locator('#join-room').click();await pages[i].locator('#code-label').waitFor({state:'visible'})}for(const p of pages)ids.push((await snap(p)).youId);s=await snap(pages[0]);await capture('lobby','02-lobby.png',pages[0],s);m.actions.push({source:'UI create/join',players:3,revision:s.revision});
async function act(p,s,a){const payload={...a,revision:s.revision};const result=await p.evaluate(async a=>{const r=await fetch('/api/action',{method:'POST',headers:{'Content-Type':'application/json',Authorization:'Bearer '+localStorage.getItem('vineyard-ee-token')},body:JSON.stringify(a)});return {status:r.status,state:await r.json()}},payload);m.actions.push({actor:s.youId,before:{phase:s.phase,year:s.year,revision:s.revision,handCount:s.hand?.length},action:payload,status:result.status,after:{phase:result.state.phase,year:result.state.year,revision:result.state.revision,handCount:result.state.hand?.length},error:result.state.error});if(result.status!==200){persist();throw Error(JSON.stringify(result))}return result.state}
await act(pages[0],s,{type:'start'});let summerDraws=0,winterDraws=0;
for(let step=0;step<m.maxSteps;step++){
s=await snap(pages[0]);const index=ids.indexOf(s.turnId);const p=pages[index<0?0:index];s=await snap(p);const has=x=>m.screenshots.some(z=>z.stage===x);
if(s.phase==='setup'&&!has('setup'))await capture('setup','03-papa-setup.png',p,s,'#ee-choice');
if(s.phase==='wake'&&!has('wake'))await capture('wake','04-spring-wake.png',p,s,'#wake-panel');
if(s.phase==='summer'&&summerDraws>=2&&!has('summer'))await capture('summer','05-summer-occupied.png',p,s,'#board');
if(s.phase==='fall'&&!has('fall'))await capture('fall','06-fall-choice.png',p,s,'#ee-choice');
if(s.phase==='winter'&&winterDraws>0&&!has('winter'))await capture('winter','07-winter-board.png',p,s,'#board');
if(s.phase==='year_end'&&!has('year_end')){await capture('year_end','08-year-end-discard.png',p,s,'#ee-choice');await capture('hand','10-hand-closeup.png',p,s,'#hand-panel')}
if(s.phase==='finished'){await capture('finished','09-finished.png',pages[0],await snap(pages[0]));m.completed=true;break}
let a;if(s.pendingChoice){const c=s.pendingChoice;a={type:'choose',choiceId:c.id,...(c.kind==='discard'?{cardIds:s.hand.slice(0,c.count).map(c=>c.id)}:{option:c.kind==='papa'?'coins':'winter'})}}
else if(s.phase==='wake'){const slot=index===0?6:([1,2].find(n=>!s.wakeSlots.find(w=>w.slot===n).playerId));a={type:'wake',slot}}
else if(s.phase==='summer'||s.phase==='winter'){const me=s.players.find(x=>x.id===s.youId);if(index===0&&s.year===1&&((s.phase==='summer'&&summerDraws<2)||(s.phase==='winter'&&winterDraws<1))){a={type:'place',space:s.phase==='summer'?'draw_vine':'draw_order',large:me.workers===0,slot:0};if(s.phase==='summer')summerDraws++;else winterDraws++}else a={type:'pass'}}else throw Error('Unexpected phase '+s.phase);
await act(p,s,a);if(step%30===0)persist();if(step===m.maxSteps-1)throw Error('maximum steps reached');
}
m.serverLog=logs;
}catch(e){m.error=e.stack;console.error(e);process.exitCode=1}
finally{if(browser)await browser.close();if(server){server.kill('SIGTERM');await Promise.race([new Promise(r=>server.once('exit',r)),delay(5000)]);if(server.exitCode===null)server.kill('SIGKILL');m.serverStopped=server.exitCode!==null||server.signalCode!==null}m.sourceHashesAfter=hashes();m.sourceUnchanged=JSON.stringify(m.sourceHashesBefore)===JSON.stringify(m.sourceHashesAfter);m.finishedAt=new Date().toISOString();persist();}
})()
