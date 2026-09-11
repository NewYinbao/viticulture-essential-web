# 葡萄酒庄园 · Viticulture Web

**中文** · [English](README.en.md)

把葡萄酒庄园搬到浏览器：2–6 位朋友各自经营酒庄，在同一个房间里种植、酿酒、招待访客和交付订单。

Go + 原生 HTML/CSS/JavaScript 实现。网页、中文卡牌说明和 SVG 随单个程序内嵌；局域网游玩无需 Node.js 或外网。游戏界面目前以中文为主。

[快速开始](#快速开始) · [版本记录](docs/README.md) · [开发指南](docs/guides/architecture.md) · [Buy me a HEYTEA](#buy-me-a-heytea)

## 下载 Windows 版

打开 [Release 下载页](https://github.com/NewYinbao/viticulture-essential-web/releases/latest)，下载 Windows x64 ZIP，解压后双击包内的 Start-Viticulture.cmd。无需安装 Go 或 Node.js。仓库目前为私有，下载需要仓库访问权限。

## 可以玩什么

| 内容 | 支持情况 |
| --- | --- |
| Essential Edition | 2–6 人、76 张访客、有限牌库、独立手牌及本地存档 |
| Tuscany Essential | 四季主板、个人过季、影响力；可选 36 张建筑、11 类特殊工人 |
| Moor Visitors | 40 张访客加入本体牌库 |
| Visit from the Rhine Valley | 80 张替换访客；EE 主板过滤 4 张依赖 Tuscany 的牌，不与 EE/Moor 混洗 |
| 玩家身份 | 房间座位密码、改密、退出和会话撤销，不能仅凭同名认领他人座位 |
| 操作与帮助 | 棋盘内卡片操作面板、悬停/聚焦提示、种植条件检查、规则速查和可关闭的新手引导 |

未选扩展不会出现在游戏内操作、规则和引导中。World、Bordeaux、旧 Tuscany 额外模块和 Automa 不在当前范围。

## 快速开始

### Windows

从源码构建需要 **Go 1.25+**。在项目根目录执行：

```powershell
powershell -NoProfile -File scripts/build-continuation.ps1
.\Start-Viticulture-Continuation.cmd
```

已有本轮构建时，直接双击 `Start-Viticulture-Continuation.cmd`。

1. 房主打开 `http://localhost:3015`，填写昵称和座位密码，创建房间。
2. 朋友打开 `http://房主局域网IP:3015`，输入房间码、自己的昵称和密码加入。
3. 房主选择扩展后开始；配置在开局后锁定。回到已有座位需提供该房间的原昵称与密码。

存档位于 `runtime/continuation-data`。关闭启动窗口或按 Ctrl+C 停止服务器；同一存档目录只能由一个进程使用。Git 仓库不包含 EXE、真实存档或会话凭据。

### 其他平台 / 直接运行源码

```sh
go run ./cmd/viticulture -addr 0.0.0.0:3015 -data ./runtime/continuation-data
```

旧启动器保留兼容用途，不会自动升级旧 EXE 或迁移旧对局。版本切换与旧座位首次设密见 [运行指南](docs/guides/running.md)。需要临时公网链接时，按 [Cloudflare 分享指南](docs/guides/cloudflare.md) 显式启用。

## 文档与版本

版本按真实 Git 提交归档，每版都有 **开发记录、验证记录、Release Notes**。目录编号仅用于排序，不代表新增的发行标签。

- [文档与版本索引](docs/README.md)
- [当前功能版本：扩展、密码与界面完善 · 563db1b](docs/versions/06-expansions-563db1b/release-notes.md)
- [架构与职责](docs/guides/architecture.md) · [HTTP API](docs/guides/api.md) · [构建与测试](docs/guides/testing.md)

`563db1b` 的完整验收为 12/12 步通过，含 24 配置、24 自然终局、36 建筑夹具和代表性复杂交互；两份独立功能复审及工程审查已完成，GitHub CI（含 race）通过。自然局策略不打访客，不能当作所有卡牌组合的证明，具体范围见 [该版验证记录](docs/versions/06-expansions-563db1b/validation.md)。

## 开发

```sh
go test -tags "ee_rule_audit audit_diff" ./...
go vet ./...
npm ci
npx playwright install chromium
npm run test:units
npm run format:check
npm run test:ui
```

Node.js 22+ 仅用于开发验证。Windows 也可用 `scripts/check.ps1` 执行 Go 审计、vet 和 gofmt 检查；完整扩展浏览器验收需 Windows Edge，见 [测试指南](docs/guides/testing.md)。

```text
cmd/viticulture/   程序入口
internal/game/    游戏规则、牌库与领域测试
internal/server/  密码、HTTP、玩家视角和实时同步
internal/store/   存档读写
internal/sharing/ 可选公网分享
web/static/       原生前端和本地素材
tests/            浏览器与原生验证
docs/             当前指南、版本记录和文档素材
```

## Buy me a HEYTEA

如果这个项目让你和朋友玩得开心，欢迎微信扫码，请我喝杯喜茶 🍵

<img src="docs/assets/buy-me-a-heytea-wechat.png" alt="微信赞助收款二维码" width="270" height="270">

扫码后可在微信中选择预设金额，感谢支持。[赞助码与预设金额的维护说明](docs/guides/support.md)

## 关于本项目

这是 Viticulture 的非官方实现，与游戏出版方无隶属关系。规则核对来源和解释边界保留在各版记录中；参考实现未作为商业卡图的再分发授权。仓库目前没有添加开源许可证，公开可读不等于取得再分发许可。
