import {renderVisitor} from './visitor-ui.js';
import {localCard} from './card-i18n.js';
const $=s=>document.querySelector(s);
const n=(tag,text,cls)=>{const e=document.createElement(tag);if(text!=null)e.textContent=text;if(cls)e.className=cls;return e};
const typeNames={vine:'葡萄藤',order:'订单',summer:'夏季访客',winter:'冬季访客',mama:'家族起始牌 A',papa:'家族起始牌 B',red:'红',white:'白',blush:'桃红',sparkling:'起泡'};
const gifts={trellis:'棚架',irrigation:'灌溉塔',yoke:'轭',medium_cellar:'中酒窖',cottage:'小屋',windmill:'磨坊',tasting_room:'品酒室',worker:'额外工人',vp:'1 胜利分'};
export function cardArt(c,cls=''){const e=n('img',null,cls);const k=parseInt(c.id?.split('-').pop()||'1',10)||1;e.src='/art/handdrawn/'+(['summer','winter','mama','papa'].includes(c.type)?'portrait-'+((k-1)%12):c.type==='vine'?'vine-'+((k-1)%4):'order')+'.svg';e.alt='原创手绘风格插画（部分卡牌共享主题图）';return e}
function panel(){let e=$('#ee-choice');if(!e){e=n('section',null,'panel ee-choice');e.id='ee-choice';$('#game').insertBefore(e,$('#lobby-panel'))}return e}
export function renderEE(v,act,online){renderDiscard(v,act,online);const box=panel();box.replaceChildren();const head=n('div',null,'ee-heading');head.append(n('span','ESSENTIAL EDITION / ALL-CARD RC','eyebrow'));const counts=n('div',null,'ee-deck-counts');for(const [type,c] of Object.entries(v.deckCounts||{}))counts.append(n('span',`${typeNames[type]||type} ${c.deck} / 弃 ${c.discard}`));head.append(counts);box.append(head);
 box.append(n('p','76 张访客已接入；本体全卡验收版，非官方认证。特殊组合约定见规则说明；原创主题插画部分共用。','ee-warning'));
 if(v.phase==='finished'&&v.winnerIds?.length>1)$('#turn-label').textContent='共同获胜：'+v.players.filter(p=>v.winnerIds.includes(p.id)).map(p=>p.name).join('、');
 const me=v.players.find(p=>p.id===v.youId);if(me?.mama?.id){const family=n('details',null,'ee-family');family.open=v.phase==='setup';family.append(n('summary','家族传承 · '+localCard(me.mama).name+' ＆ '+localCard(me.papa).name));const row=n('div',null,'ee-parent-row');for(const original of [me.mama,me.papa]){const c=localCard(original);const a=n('article',null,'ee-parent-card');a.append(cardArt(c),n('strong',c.name),n('small',c.englishName,'card-english'),n('small',c.description||'开局资源见当前选择'));row.append(a)}family.append(row);box.append(family)}
 if(v.phase==='finished'){const result=n('section',null,'final-results');result.append(n('h3','最终结算 · '+v.players.filter(p=>(v.winnerIds||[v.winnerId]).includes(p.id)).map(p=>p.name).join('、')+' 获胜'));const table=n('table');const head=n('tr');for(const text of ['庄主','胜利分','金币','酒总值','葡萄总值'])head.append(n('th',text));table.append(head);for(const p of v.players){const row=n('tr');for(const text of [p.name,p.vp,p.coins,(p.wines||[]).reduce((s,w)=>s+w.value,0),(p.grapes||[]).reduce((s,g)=>s+g.value,0)])row.append(n('td',text));table.append(row)}result.append(table,n('small','依次比较胜利分、金币、酒总值、葡萄总值；完全相同则共同获胜。'));box.append(result)}
 const c=v.pendingChoice;if(!c){renderVisitor(document.createElement('div'),v,act,online);return;}
 if(c.visitor||['visitor','planner'].includes(c.kind)){const visitor=n('section',null,'visitor-choice');box.append(visitor);renderVisitor(visitor,v,act,online);return;}
 renderVisitor(document.createElement('div'),{...v,pendingChoice:null},act,online);
 const your=c.playerId===v.youId;box.append(n('h3',your?'请完成当前选择':'等待 '+(v.players.find(p=>p.id===c.playerId)?.name||'玩家')+' 完成选择'));
 if(!your){box.append(n('p','其他玩家的手牌与选择内容不会向你公开。','muted'));return}
 const form=n('form',null,'ee-choice-form');const err=n('p',null,'ee-warning');box.append(form,err);let locked=false;
 const submit=async a=>{if(locked||!online)return;locked=true;for(const b of form.querySelectorAll('button'))b.disabled=true;const ok=await act({type:'choose',choiceId:c.id,revision:v.revision,...a});if(!ok){locked=false;for(const b of form.querySelectorAll('button'))b.disabled=!online;err.textContent='提交未生效，请检查资源或局面是否已经变化。'}};
 if(c.kind==='discard'){form.remove();err.remove();box.append(n('p','请在下方原手牌区选择弃牌。'));}
 else for(const option of c.options||[]){let text=c.labels?.[option]||({summer_summer:'夏季访客＋夏季访客',summer_winter:'夏季访客＋冬季访客',winter_winter:'冬季访客＋冬季访客','summer+summer':'夏季访客＋夏季访客','summer+winter':'夏季访客＋冬季访客','winter+winter':'冬季访客＋冬季访客'}[option])||typeNames[option]||option;if(c.kind==='papa'){const o=v.parentOptions;text=option==='gift'?'接受赠礼：'+(gifts[o?.gift]||o?.gift):'改为额外 '+o?.coins+' 金币'}const b=n('button',text,'primary');b.type='button';b.disabled=!online;b.dataset.choice=option;b.onclick=()=>{const pair=option.split(/[_+,]/);submit({option,...(pair.length===2&&pair.every(x=>['summer','winter'].includes(x))?{colors:pair}:{})})};form.append(b)}
}
function select(form,name,label,items){const l=n('label',label),s=n('select');s.name=name;s.required=true;for(const [value,text]of items){const o=n('option',text);o.value=value;s.append(o)}l.append(s);form.append(l);return s}
function checks(form,name,label,items){form.append(n('p',label));for(const [value,text]of items){const l=n('label',null,'check'),i=n('input');i.type='checkbox';i.name=name;i.value=value;l.append(i,document.createTextNode(text));form.append(l)}}
export function setupEE(space,v,act){let dialog=$('#ee-action-dialog');if(!dialog){dialog=n('dialog');dialog.id='ee-action-dialog';document.body.append(dialog)}dialog.replaceChildren();const wake=space==='wake',p=v.players.find(p=>p.id===v.youId);dialog.append(n('h2',wake?'起床奖励 · 选择访客':space.name));const form=n('form'),fields=n('div'),error=n('p',null,'ee-warning');dialog.append(form);form.append(fields,error);const rev=v.revision;
 if(wake){select(fields,'color','选择一种访客',[['summer','夏季访客'],['winter','冬季访客']])}else{
 select(fields,'declineBonus','行动格奖励',[['no','领取（默认）'],['yes','放弃奖励，仍执行行动']]);
 const modeItems=space.id==='sell_grapes'?[['sell_grapes','出售葡萄'],['sell_field','出售空田地'],['buy_field','买回田地']]:space.id==='yoke'?[['harvest','收获一块田地'],['uproot','拔藤回手牌']]:space.id==='plant'?[['plant','种植葡萄藤'],['uproot','拔藤回手牌']]:[];
 select(fields,'worker','工人',[...(p.workers>0?[['normal','普通工人']]:[]),...(p.largeWorker?[['large','大工人']]:[])]);
 if(space.capacity>1&&space.capacity<6)select(fields,'slot','行动格（第1格为奖励格）',[[0,'自动选择'],...Array.from({length:space.capacity},(_,i)=>[i+1,'第'+(i+1)+'格'+(space.occupied.some(o=>o.slot===i+1)?' · 已占用':'')])]);
 const detail=n('div');let mode;if(modeItems.length){mode=select(fields,'mode','操作',modeItems)}fields.append(detail);
 const rebuild=()=>{detail.replaceChildren();const kind=mode?.value||space.id;
 if(['buy_field','sell_field','uproot'].includes(kind)||space.id==='yoke')select(detail,'field','田地',p.fields.map(f=>[f.index,`田地${f.index+1} · ${f.capacity}金币${f.sold?' · 已出售':''}`]));
 if(kind==='uproot'){select(detail,'cardId','拔回手牌的葡萄藤',p.fields.flatMap(f=>f.vines.map(c=>[c.id,`田地${f.index+1} · ${c.name}`])))}
 else if(kind==='plant'){detail.append(n('p','选择1张藤；奖励格可选择2张。为每张藤指定田地。'));for(const c of v.hand.filter(c=>c.type==='vine')){const l=n('label',null,'check'),i=n('input');i.type='checkbox';i.name='plantCards';i.value=c.id;l.append(i,document.createTextNode(c.name+' · 红'+c.red+' 白'+c.white));detail.append(l);select(detail,'field-'+c.id,'种入',p.fields.filter(f=>!f.sold).map(f=>[f.index,'田地'+(f.index+1)]))}}
 else if(kind==='harvest'&&space.id!=='yoke')checks(detail,'fields','选择田地；奖励格可收获2块。',p.fields.filter(f=>!f.sold&&!f.harvested&&f.vines.length).map(f=>[f.index,'田地'+(f.index+1)]));
 else if(kind==='make_wine'){detail.append(n('p','将葡萄分配到第1/2瓶配方；奖励格可用第3瓶。同瓶红白各一为桃红，两红一白为起泡。'));p.grapes.forEach((g,i)=>select(detail,'recipe-'+i,(typeNames[g.color]||g.color)+'葡萄 · 品质'+g.value,[['','不使用'],['1','第1瓶'],['2','第2瓶'],['3','第3瓶']]));for(const s of detail.querySelectorAll('select'))s.required=false;}
 else if(kind==='sell_grapes')checks(detail,'grapes','选择要出售的葡萄',p.grapes.map((g,i)=>[i,typeNames[g.color]+'葡萄 · '+g.value]));};
 if(mode)mode.onchange=rebuild;rebuild();
 }
 const buttons=n('div',null,'ee-buttons'),cancel=n('button','取消'),send=n('button','确认','primary');cancel.type='button';cancel.onclick=()=>dialog.close();buttons.append(cancel,send);form.append(buttons);
 form.onsubmit=async e=>{e.preventDefault();send.disabled=true;try{const d=new FormData(form),a=wake?{type:'wake',slot:5,color:d.get('color')}:{type:'place',space:space.id,large:d.get('worker')==='large',declineBonus:d.get('declineBonus')==='yes'};a.revision=rev;if(!wake){if(d.has('slot'))a.slot=Number(d.get('slot'));const mode=d.get('mode')||space.id;if(d.has('mode'))a.mode=mode;if(d.has('field'))a.field=Number(d.get('field'));if(d.has('cardId'))a.cardId=d.get('cardId');if(mode==='plant'){a.cardIds=d.getAll('plantCards');a.fields=a.cardIds.map(id=>Number(d.get('field-'+id)));if(!a.cardIds.length)throw Error('至少选择一张葡萄藤')}if(mode==='harvest'&&space.id!=='yoke'){a.fields=d.getAll('fields').map(Number);if(!a.fields.length)throw Error('至少选择一块田地')}if(mode==='sell_grapes')a.grapes=d.getAll('grapes').map(Number);if(mode==='make_wine'){const groups={};p.grapes.forEach((g,i)=>{const r=d.get('recipe-'+i);if(r)(groups[r]??=[]).push(i)});a.recipes=Object.values(groups);if(!a.recipes.length)throw Error('至少选择一个酿酒配方')}}if(await act(a))dialog.close();else error.textContent='操作未生效，请根据提示检查选择。'}catch(err){error.textContent=err.message}finally{send.disabled=false}};dialog.showModal();
}

// Choice-scoped state survives SSE, not another room/player/choice.
let discard=null;
function renderDiscard(v,act,online){
 const c=v.pendingChoice,own=c?.kind==='discard'&&c.playerId===v.youId;
 const key=own?JSON.stringify([v.code,v.youId,v.year,c.id]):null;
 const fresh=key!==discard?.key;
 $('#discard-prompt')?.remove();$('#hand').classList.toggle('discard-hand',own);
 if(!own){discard=null;return}
 if(fresh)discard={key,ids:new Set(),busy:false,error:''};
 const d=discard,valid=new Set((v.hand||[]).map(c=>c.id));d.view=v;d.online=online;
 for(const id of d.ids)if(!valid.has(id))d.ids.delete(id);
 const prompt=n('section',null,'discard-prompt');prompt.id='discard-prompt';
 prompt.setAttribute('role','region');prompt.setAttribute('aria-labelledby','discard-title');
 const title=n('h3','年末需要弃牌');title.id='discard-title';
 const status=n('p');status.id='discard-status';status.setAttribute('role','status');
 const hint=n('p','点击下方原手牌选择，再点取消；键盘 Tab 定位，空格 / Enter 选择。','muted');
 const error=n('p',d.error,'ee-warning');error.setAttribute('role','alert');
 const send=n('button','确认弃牌','primary');send.id='confirm-discard';send.type='button';
 prompt.append(title,status,hint,send,error);$('#hand-panel').insertBefore(prompt,$('#hand'));
 const cards=[...document.querySelectorAll('#hand .card')];

 const update=()=>{status.textContent='需要弃置 '+c.count+' 张 · 已选 '+d.ids.size+' / '+c.count+' 张（保留不超过7张）';send.disabled=!online||d.busy||d.ids.size!==c.count;send.textContent=d.busy?'正在提交…':'确认弃牌';prompt.setAttribute('aria-busy',String(d.busy));for(const card of cards){const chosen=d.ids.has(card.dataset.cardId);card.classList.toggle('discard-selected',chosen);card.setAttribute('aria-pressed',String(chosen));card.setAttribute('aria-disabled',String(!online||d.busy))}};
 for(const card of cards){card.tabIndex=0;card.setAttribute('role','button');card.setAttribute('aria-describedby','discard-status');const toggle=()=>{if(!online||d.busy)return;const id=card.dataset.cardId;if(d.ids.has(id))d.ids.delete(id);else d.ids.add(id);update()};card.onclick=toggle;card.onkeydown=e=>{if(e.key===' '||e.key==='Enter'){e.preventDefault();if(!e.repeat)toggle()}}}
 send.onclick=async()=>{if(discard!==d||d.busy||!online||d.ids.size!==c.count)return;d.busy=true;d.error='';update();let ok=false;try{ok=await act({type:'choose',choiceId:c.id,revision:v.revision,cardIds:[...d.ids]})}catch(e){d.error=e.message}finally{d.busy=false;if(discard===d){if(ok)d.ids.clear();else d.error=d.error||'提交未生效，请检查连接或最新局面后重试。';renderDiscard(d.view,act,d.online)}}};
 update();if(fresh)prompt.scrollIntoView({block:'start'});
}
