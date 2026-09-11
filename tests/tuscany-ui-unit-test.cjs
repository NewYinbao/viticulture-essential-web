const test = require('node:test');
const assert = require('node:assert/strict');
const {pathToFileURL} = require('node:url');
const path = require('node:path');
const load = n => import(pathToFileURL(path.join(__dirname,'../web/static/js/'+n+'.js')).href);
const view = () => ({config:{board:'tuscany'},youId:'a',hand:[],players:[{id:'a',name:'甲',coins:3,vp:0,influence:{},grapes:[]},{id:'b',name:'乙',influence:{}}],influenceRegions:[{id:'pisa',name:'比萨',reward:'coins2',points:1},{id:'siena',name:'锡耶纳',reward:'coin',points:2}]});
test('Tuscany physical slot rewards do not inherit EE first slot or overflow rewards',async()=>{
 const {bonusKey}=await load('action-options');
 assert.equal(bonusKey({id:'draw_vine',capacity:1,occupied:[],bonusSlots:{1:'draw_vine'}},1),'draw_vine');
 assert.equal(bonusKey({id:'tour',capacity:2,occupied:[],bonusSlots:{2:'coin'}},1),'');
 assert.equal(bonusKey({id:'tour',capacity:2,occupied:[],bonusSlots:{2:'coin'}},2),'coin');
 assert.equal(bonusKey({id:'tour',capacity:2,occupied:[],bonusSlots:{2:'coin'}},-1),'');
 assert.equal(bonusKey({id:'tour',capacity:2,occupied:[]},1),'coin');
});
test('Influence switches from place to move at six own stars, never borrows opponent stars',async()=>{
 const {tuscanyInputState:f}=await load('tuscany-inputs');const v=view();v.players[1].influence.pisa=6;
 assert.deepEqual(f('influence',v,{to:'siena'}).payload,{influence:[{to:'siena'}]});
 v.players[0].influence={pisa:6};
 assert(f('influence',v,{from:'pisa',to:'pisa'}).reason);
 assert.equal(f('influence',v,{from:'pisa',to:'siena'}).reason,'');
 assert(f('influence',{...v,config:{structures:true}},{to:'siena'}).reason);
});
test('Trading exact own cards, preselected colors, and authoritative VP floor',async()=>{
 const {tuscanyInputState:f}=await load('tuscany-inputs');const v=view();
 v.hand=[{id:'a',type:'vine',name:'A'},{id:'b',type:'vine',name:'B'}];
 assert(f('trade',v,{give:'cards',receive:'vp',card1:'a',card2:'a'}).reason);
 assert(f('trade',v,{give:'cards',receive:'vp',card1:'a',card2:'foreign'}).reason);
 assert.equal(f('trade',v,{give:'cards',receive:'vp',card1:'a',card2:'b'}).reason,'');
 assert(f('trade',v,{give:'coins',receive:'cards',color1:'vine'}).reason);
 assert.equal(f('trade',v,{give:'coins',receive:'cards',color1:'vine',color2:'order'}).reason,'');
 v.players[0].vp=-5;assert(f('trade',v,{give:'vp',receive:'coins'}).reason);
 v.players[0].vp=-4;assert.equal(f('trade',v,{give:'vp',receive:'coins'}).reason,'');
});
test('Influence ties and unique leader labels derive only from public region counts',async()=>{
 const {regionStanding:f}=await load('tuscany-board');const v=view();
 assert.match(f(v,v.influenceRegions[0]).label,/暂无星/);
 v.players[0].influence.pisa=2;assert.match(f(v,v.influenceRegions[0]).label,/甲.*领先/);
 v.players[1].influence.pisa=2;assert.match(f(v,v.influenceRegions[0]).label,/并列.*不得分/);
});
test('Tuscany visitor lesson distinguishes coin ordering from two-card slot',async()=>{
 const {actionLessonForView:f}=await load('rules-content');
 assert.match(f(view(),'summer_visitor'),/金币奖励.*访客前或后/);
 assert.doesNotMatch(f({config:{board:'ee'}},'summer_visitor'),/金币奖励/);
});
