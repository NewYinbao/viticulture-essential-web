const { chromium } = require('playwright');
const http = require('node:http');
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');

const root = path.resolve(__dirname, '..', 'web/static');
const fixtures = JSON.parse(fs.readFileSync(process.env.RHINE_BROWSER_FIXTURES, 'utf8'));
const server = http.createServer((req, res) => {
  if (req.url === '/') { res.setHeader('Content-Type', 'text/html');return res.end('<div id="hand"></div><div id="choice"></div>'); }
  const filename = path.resolve(root, '.' + new URL(req.url, 'http://localhost').pathname);
  if (path.relative(root, filename).startsWith('..')) { res.statusCode=404;return res.end(); }
  try { res.setHeader('Content-Type', filename.endsWith('.js')?'text/javascript':'text/plain');res.end(fs.readFileSync(filename)); }
  catch {res.statusCode=404;res.end();}
});
(async()=>{
 let browser;
 try {
  await new Promise(r=>server.listen(0,'127.0.0.1',r));
  browser=await chromium.launch({channel:'msedge',headless:true});
  const page=await browser.newPage();const errors=[];page.on('pageerror',e=>errors.push(e.message));
  await page.goto('http://127.0.0.1:'+server.address().port);
  const results=await page.evaluate(async views=>{
   const {renderVisitor}=await import('/js/visitor-ui.js');
   const box=document.querySelector('#choice');const results=[];
   for(const view of views){
    const key=view.pendingChoice.visitor.cardId+':'+view.pendingChoice.options[0];
    try {const shown=renderVisitor(box,view,()=>{},true);if(!shown||!box.querySelector('h3'))throw Error('missing visitor title');results.push({key});}
    catch(e){results.push({key,error:e.stack});}
   }
   const hidden=structuredClone(views[0]);hidden.config.visitors='ee';box.replaceChildren();
   if(renderVisitor(box,hidden,()=>{},true)||box.children.length)results.push({error:'Rhine UI leaked into EE visitors'});
   return results;
  },fixtures);
  assert.deepEqual(errors,[]);
  const failures=results.filter(r=>r.error);assert.deepEqual(failures,[]);
  console.log(JSON.stringify({passed:results.length,errors,expansionVisibility:'passed'}));
 } finally {await browser?.close();await new Promise(r=>server.close(r));}
})().catch(e=>{console.error(e);process.exitCode=1;});
