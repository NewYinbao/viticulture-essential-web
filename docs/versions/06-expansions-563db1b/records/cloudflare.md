> 历史记录 · [所属版本](../release-notes.md)。保留当时的结论、测试范围及本地证据路径；后续状态请查版本索引。

# Cloudflare 临时公网分享

## 当前状态与边界

已实现并发布到 `dist/Viticulture.exe`：可选 Quick Tunnel、每次启动随机访问口令和公网模式状态轮询。旧可执行文件备份为 `dist/Viticulture-before-cloudflare-881671f5142f.exe`。

实测通过：Go 全套测试、Windows 原生 sharing 测试、3 项 JS 轮询测试；最终发布程序经真实 Cloudflare 链接，用 Windows Edge 两个隔离浏览器完成口令登录、创建/加入、开局、轮询同步、离线恢复和刷新恢复。确认公网模式没有发起 SSE，浏览器无未捕获脚本错误。Windows 强制结束测试主进程后，其 cloudflared 子进程随 Job Object 退出。另测回车默认 LAN 和端口占用时不创建存档。

这些是连接功能验收，不是重新完成一局全游戏或卡牌规则签核；没有第二台物理设备实测，也未手动点击控制台关闭按钮。所有公网测试使用独立测试存档，测试隧道现已停止；未修改现有对局。

默认启动器选择 LAN，不自动开启公网；仅明确选择 Public 才启动隧道。程序直接运行默认监听 127.0.0.1:3010，启动器默认 0.0.0.0:3013。LAN 不启用共享口令门禁，只应在可信网络使用。公网模式会将流量交由 Cloudflare 处理，不提供 SLA，不是加固后的互联网托管服务。

## 安装与启动（Windows PowerShell）

先由发布/构建流程准备 dist/Viticulture.exe，再显式安装固定版本：

    powershell -NoProfile -ExecutionPolicy Bypass -File scripts/install-cloudflared.ps1
    powershell -NoProfile -ExecutionPolicy Bypass -File scripts/start-viticulture.ps1 -Mode Public

也可双击 Start-Viticulture.cmd，在提示时明确输入 2 或 y/yes；回车、1 或其他输入使用 LAN。脚本不会修改 PATH、防火墙或安装系统服务。缺少 cloudflared 时直接报错，不启动服务器或隧道。安装器先检查 SHA256，再替换 dist/tools/cloudflared.exe，已有文件会备份。

将控制台打印的临时 HTTPS 地址和本次访问口令私下发送给可信朋友。打开链接，在入口表单提交口令，然后创建或加入房间。口令不是房间码，也不替代玩家会话凭据；所有获知口令的人均可进入应用。不要把口令放入 URL、公开聊天、截图或日志附件。重启后口令变化，旧 Cookie 不能再次认证，临时地址也不保证保持。

启动器存档默认 runtime/ee-data，不迁移旧存档。重用目录前须自行停止旧服务器；程序不会杀掉占用端口的其他进程。关闭窗口或 Ctrl+C 结束本次服务与隧道。Windows 使用 Job Object 管理子进程，已验证强制终止父进程会回收子进程；特殊已有 Job 环境仍可能导致明确的启动失败。

## 轮询、门禁与故障

官方明确 Quick Tunnel 不支持 SSE，因此公网门禁认证后 /api/transport 返回 {"poll":true}，/api/events 返回 501；前端改用 /api/state 轮询。默认每次请求完成后等待 1500ms，下次再发；请求 10 秒时触发 AbortController。关闭轮询会取消请求和后续重试。LAN 仍可使用 SSE。官方还说明最多 200 个在途请求，超限返回 429；本方案不适合大量公开访客。

门禁覆盖 API、健康检查与静态资源。登录设置 HttpOnly、SameSite=Lax Cookie；TLS 或可信的 HTTPS 转发标记下设置 Secure。它只是共享秘密门禁，没有完整账户系统、限速或细粒度权限，不应暴露给不可信人群。

隧道使用独立临时目录及显式空 config.yml，不重命名或读取用户 Cloudflare 凭据。启动取得 trycloudflare 地址后，携带门禁 Cookie 检查公网 /api/transport，只有 HTTP 200 且正文完全符合预期才打印地址。进程提前退出、上下文取消、启动失败或就绪超时均报错，停止此次服务；隧道后续退出也结束服务。不自动退回无门禁公网模式。网络/代理故障可先检查网络和固定版本安装；不要关闭门禁来绕过就绪检查。

入口页面使用严格 Content-Security-Policy；认证通过后，应用 handler 会覆盖为允许同源 JS/CSS 的策略。已通过真实浏览器资源加载与交互验证，不存在此前静态审查疑似的策略阻断。邀请房间码会在口令提交后保留。

## 固定版本与官方依据

官方 GitHub Release API 实测 tag 为 **2026.9.0**（口语称 v2026.9.0，但带 v 的 API tag 返回 404），Windows amd64 资产是 cloudflared-windows-amd64.exe：

- API：https://api.github.com/repos/cloudflare/cloudflared/releases/tags/2026.9.0
- 资产：https://github.com/cloudflare/cloudflared/releases/download/2026.9.0/cloudflared-windows-amd64.exe
- API digest：sha256:547057326266f0e1c7d50d102dbd22ff283d740c055bd61e94f10e2c606f89af

已从官方地址安装上述版本，并独立计算本地二进制 SHA256，与官方 digest 和安装器固定值一致。

Quick Tunnel 限制官方页面：https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/do-more-with-tunnels/trycloudflare/

本次页面直取返回 403，改为读取同一官方文档仓库源码，确认原文 “Quick Tunnels do not support Server-Sent Events (SSE).”：
https://raw.githubusercontent.com/cloudflare/cloudflare-docs/production/src/content/docs/cloudflare-one/networks/connectors/cloudflare-tunnel/do-more-with-tunnels/trycloudflare.mdx

## 本地复验

    go test ./...
    node --test tests/polling-unit-test.cjs

Go 测试覆盖门禁、口令与 Cookie、拒绝 SSE、缺失/不可执行程序、假子进程提前退出、取消、模拟公网就绪与拒绝错误响应、重复 Stop 和临时配置清理。JS 测试不需 Playwright 或新增依赖，使用 Node 内置测试器及隔离 VM/假计时器，覆盖单请求、鉴权、错误重试、超时取消和关闭后的迟到响应。它不覆盖 app.js 的浏览器集成。

## 发布与复现

启动入口：`Start-Viticulture.cmd`，默认地址 `http://localhost:3013`，默认存档 `runtime/ee-data`。高级非交互参数：

```powershell
.\Start-Viticulture.cmd -Mode LAN
.\Start-Viticulture.cmd -Mode Public
.\Start-Viticulture.cmd -Mode Public -Address 127.0.0.1:3013 -Data runtime/ee-data
```

公网模式同时要求本地/LAN访问者输入口令。临时 HTTPS 域名每次可能改变，浏览器的原玩家会话按域名隔离，所以不保证更换域名后自动恢复原座位。建议一局期间保持服务开启；固定域名/命名隧道与跨域名身份恢复不在本次实现范围。

测试结果见 `artifacts/cloudflare/browser-result.json`。浏览器脚本 `tests/cloudflare-browser-test.cjs` 需要运行中的隔离公网测试服务，以及自行创建的 `artifacts/cloudflare/test-private.json`（字段 `url` 和 `key`）；运行 `node tests/cloudflare-browser-test.cjs`。测试凭据文件不可提交、分享，结束后应删除并停止测试进程。已有实际测试凭据和日志已清理。该脚本会创建两名测试玩家，因此必须使用独立数据目录，不要对当前对局运行。

最终 Windows 发布 SHA256：`763e208d19050f3b1315b4473b49f850ff8eef95172f36f137cc9e96800070c2`。
