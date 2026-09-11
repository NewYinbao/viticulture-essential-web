# 葡萄酒庄园 · Essential Edition Web

Go + 原生 HTML/CSS/JavaScript 的局域网多人桌游实现。支持 EE 本体 2–6 位真人、固定卡库、76 张访客、独立手牌、SSE 同步及本地 JSON 存档。单个可执行文件内嵌网页与 SVG，不需要 Java、Node 或外网即可游玩。

本轮续作加入玩家座位密码与可选扩展，交付状态、测试及规则依据见 [最新交付记录](docs/expansions-release.md) 和 [玩家密码说明](docs/玩家密码与续作记录-20260911.md)。新程序使用 **Start-Viticulture-Continuation.cmd**，地址 `http://localhost:3015`、独立存档 `runtime/continuation-data`。它不会升级已运行的旧程序或迁移旧对局；以下原启动入口仍保留。

## 启动

本机已有构建时，双击 **Start-Viticulture.cmd**。主机打开 http://localhost:3013；朋友打开 `http://主机局域网IP:3013`，输入房间码加入。

新启动器使用 `runtime/ee-data`，与旧版目录隔离。返回入口后可点“回到上次的房间”；刷新或服务重启后同一浏览器会恢复原座位。关闭控制台或 Ctrl+C 停止服务。

从源码构建需要 Go 1.25+：

```powershell
powershell -File scripts/build.ps1
.\Start-Viticulture.cmd
```

也可直接运行：

```powershell
go run ./cmd/viticulture -addr 127.0.0.1:3013 -data ./runtime/ee-data
```

局域网监听使用 `-addr 0.0.0.0:3013`。同一个存档目录只能由一个进程使用。运行数据包含手牌和会话凭据，不纳入 Git；该程序面向可信局域网。

## 工程导航

| 路径 | 职责 |
| --- | --- |
| cmd/viticulture | 程序入口 |
| internal/game | 游戏规则、卡库、访客状态机及领域测试 |
| internal/server | HTTP、会话、SSE 与应用事务 |
| internal/store | JSON 快照存储 |
| web/static | 原生前端与本地 SVG；通过 web/embed.go 内嵌 |
| tests | 浏览器回归与 Windows 原生验证 |
| scripts | 构建、检查脚本 |
| docs | 架构、API、规则核对、UI 检查及历史记录 |
| dist / artifacts / runtime | 被忽略的构建、证据和存档目录 |

阅读 [架构说明](docs/architecture.md)、[当前 API](docs/api.md)、[规则核对与未决边界](docs/rules-review.md)、[UI 改动](docs/ui-review.md)、[构建与测试](docs/testing.md)。

## 验证

```powershell
powershell -File scripts/check.ps1
npm ci
npx playwright install chromium
npm run test:ui
npm run test:game
```

Node 依赖仅用于开发。其他专项、截图位置和本次测试结果见 [testing.md](docs/testing.md)。GitHub Actions 配置了 Go（含 race）和浏览器检查。

## 实现范围

非官方游戏实现。续作范围为 EE、Tuscany Essential 四季主板与影响力、36 张建筑卡、11 种特殊工人、Moor 40 张访客与 Rhine 80 张替换访客；不含 World、Bordeaux、旧版额外模块或 Automa。具体可用状态以最新交付记录为准。未启用的模块不出现在游戏内操作、规则与引导中。

旧版本的记录保留在 docs/history，旧启动器保留兼容用途。旧 exe 和研究副本不随源码仓库分发。参考来源见 [规则核对](docs/rules-review.md)；未擅自为已有项目内容添加新的开源授权。

## 可选临时公网分享

默认仍为可信局域网；Cloudflare Quick Tunnel 需明确启用。双击启动器，回车默认局域网，输入 `2` 或 `y` 开启公网。安装、口令、安全边界与测试结果见 [Cloudflare 分享说明](docs/cloudflare.md)。已用真实公网及 Windows Edge 双浏览器验证登录、加入、开局、轮询与重连，并验证 Windows 父进程退出时回收隧道；不是新一轮完整游戏规则验收。
