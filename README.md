# 葡萄酒庄园 · Viticulture Web

**中文** · [English](README.en.md)

与 2–6 位朋友在浏览器里经营酒庄：种植葡萄、酿酒、招待访客、交付订单。支持 Essential Edition 及可选扩展，游戏界面以中文为主。

[快速开始](#快速开始) · [功能介绍](#功能介绍) · [版本记录](#版本记录) · [开发指南](#开发指南) · [Buy me a HEYTEA](#buy-me-a-heytea)

## 快速开始

### 下载与启动

1. 打开 [Release 下载页](https://github.com/NewYinbao/viticulture-essential-web/releases/latest)，下载 **Windows x64 ZIP**。
2. 解压整个压缩包，双击 **Start-Viticulture.cmd**。
3. 在浏览器打开 **http://localhost:3015**。

无需安装 Go 或 Node.js，局域网游玩无需外网。仓库目前为私有，下载需要仓库访问权限。

### 从源码构建

如果希望自己构建，先安装 **Git 和 Go 1.25+**，确保终端可以执行 git 和 go。构建游戏不需要 Node.js；私有仓库需先获得访问权限，并完成 GitHub 身份验证。

在 Windows PowerShell 中执行：

~~~powershell
git clone https://github.com/NewYinbao/viticulture-essential-web.git
cd viticulture-essential-web
powershell -NoProfile -File scripts/build-continuation.ps1
.\Start-Viticulture.cmd
~~~

构建产物为 dist/Viticulture-Continuation.exe。启动时选择局域网模式，浏览器打开 **http://localhost:3015**。源码启动的存档在 runtime/continuation-data，更新时保留该目录。

macOS / Linux 可在克隆并进入项目目录后直接构建运行：

~~~sh
go build -trimpath -o viticulture ./cmd/viticulture
./viticulture -addr 0.0.0.0:3015 -data ./runtime/continuation-data
~~~

### 和朋友一起玩

房主填写昵称和座位密码创建房间；朋友在同一局域网打开 `http://房主IP:3015`，用房间码、自己的昵称和密码加入。房主在开局前选择扩展包，无需更换启动脚本。

下载包的存档保存在解压目录的 `data` 文件夹。关闭启动窗口可停止服务；更新前停止旧程序并备份该文件夹。恢复已有座位需要原昵称和密码。

[更多运行与存档说明](docs/guides/running.md) · [临时公网联机](docs/guides/cloudflare.md)

## 功能介绍

| 内容 | 支持情况 |
| --- | --- |
| Essential Edition | 2–6 人、76 张访客、独立手牌与本地存档 |
| Tuscany Essential | 四季主板、影响力；可选 36 张建筑和 11 类特殊工人 |
| Moor Visitors | 40 张访客加入本体牌库 |
| Visit from the Rhine Valley | 80 张替换访客；本体主板排除 4 张依赖 Tuscany 的牌，不与 EE/Moor 混洗 |
| 联机与操作 | 座位密码、断线恢复、棋盘内卡片面板、行动条件提示、规则速查和新手引导 |

未启用的扩展不会显示对应操作、规则或引导。World、Bordeaux、旧 Tuscany 额外模块和 Automa 不在当前范围。

## 版本记录

以下按三个主要开发阶段介绍；详细变更及验证范围见对应文档。

| 阶段 | 主要变化 | 详细说明 |
| --- | --- | --- |
| 扩展与联机完善 | Tuscany、Moor、Rhine、玩家密码及随扩展配置变化的界面；已提供 Windows 下载版 v2026.09.12 | [版本说明](docs/versions/06-expansions-563db1b/release-notes.md) |
| 操作体验与新手引导 | 行动可用性提示、棋盘内卡片面板、种植条件检查和规则引导 | [版本说明](docs/versions/05-guidance-80e9f7c/release-notes.md) |
| 本体实现与工程重构 | EE 规则实现、规则修正，以及游戏、服务、存储和网页分层 | [版本说明](docs/versions/02-refactor-f9227db/release-notes.md) |

## 开发指南

### 项目架构

后端使用 Go，前端使用原生 HTML、CSS 和 JavaScript。网页与素材内嵌在单个可执行文件中，游戏规则由服务端处理，客户端接收各玩家可见的局面。

```text
cmd/viticulture/   程序入口
internal/game/    游戏规则、牌库与领域测试
internal/server/  HTTP、玩家会话与实时同步
internal/store/   存档读写
internal/sharing/ 可选公网分享
web/static/       前端界面与素材
tests/            浏览器与原生验证
scripts/          构建与检查脚本
docs/             详细指南与历史记录
```

[详细架构说明](docs/guides/architecture.md)

### 本地开发

按[快速开始中的源码构建步骤](#从源码构建)获取项目并运行。修改内嵌网页后需重新运行或构建 Go 程序。前端检查与浏览器测试另需 **Node.js 22+**，在项目目录执行 npm ci 安装开发依赖。

[完整构建与测试方法](docs/guides/testing.md)

## Buy me a HEYTEA

如果这个项目让你和朋友玩得开心，欢迎微信扫码，请我喝杯喜茶 🍵

<img src="docs/assets/buy-me-a-heytea-wechat.png" alt="微信赞助二维码" width="270" height="270">

扫码后可选择预设金额，感谢支持。

## 关于本项目

这是 Viticulture 的非官方实现，与游戏出版方无隶属关系。游戏及素材权利归相应权利人所有；仓库目前未添加开源许可证。
