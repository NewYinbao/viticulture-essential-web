# 验证记录：扩展、玩家密码与上下文界面

- 提交：`563db1b1c6f7a08e6e79dcab09c7c7b659fcd077`
- 主要证据：[expansions-release.md](records/expansions-release.md)、[CONTINUATION_FINAL-20260912.md](records/audits/CONTINUATION_FINAL-20260912.md) 及两份最终功能复核。

## 最终本地检查

| 检查 | 结果与边界 |
| --- | --- |
| Go 默认测试、规则审计、`go vet` | 全包通过 |
| 前端单元检查 | 帮助 7 项、Tuscany 5 项、轮询 3 项通过；最终复审记录 15/15 |
| JavaScript/格式 | Prettier 与 56 个 JS/CJS 语法检查通过 |
| 完整验收入口 | 12/12 步通过，运行前后源码指纹一致 |
| 开局配置 | 24 种组合在两份独立浏览器中通过；非法修改和开局锁定通过 |
| 自然终局 | 24 种配置从正常随机牌库和本人 View 决策到正式终局 |
| 建筑实际 UI | 36 张均通过真实 HTTP/浏览器触发和资源检查；使用明确标记的 seeded 场景 |
| 复杂互动 | 工人、Moor、Rhine、Messenger、Planner 与 Farmer 代表流程通过，含私密选择和重启 |
| 密码与撤销 | 6 组浏览器流程通过；同名防冒领、旧会话清理、旧座位安全设密 |
| 模块显示 | 24 组合和 2 个跨房切换检查通过；桌面与 390px 视口已检查 |

此前本轮记录还包括本体体验 14 组、规则 UI 5 组、EE 访客初始选项 141 组和预约等特殊交互 19 组。最终格式修复后重新通过全包 Go/审计/vet/gofmt、前端单元/格式/语法，以及 Farmer 预约折扣的真实 Edge 重启流程。

## 独立复审

两份独立功能复核在冻结快照上确认并关闭：

1. 客户端伪造 `BonusOverride` 的公共边界问题；Go 与真实 HTTP 覆盖 EE/Tuscany、落盘和 409 重放。
2. Planner × Farmer 只有 3 金币时折扣预约被预检删除的问题；Go、私有 View 和真实 Edge 跨季重启流程通过。

密码专项另有独立复审记录。工程审查修正 10 个 Go 文件的 gofmt 输出、Node 模块声明、检查入口和过时说明；受测程序行为文件的格式后哈希与预期逐一匹配。

## 构建与 CI

- 最终源码 manifest digest：`5a60cb1aee782e5f4fa8ae15d9988d64367474d3d45d077dd1bda7ec31c7232e`。
- `Viticulture-Continuation.exe` SHA-256：`513f41b507991384ed93c7de1a8e1c3261847822d0f4de8021e380aeadaacd46`。
- GitHub Actions [run 34629523327](https://github.com/NewYinbao/viticulture-essential-web/actions/runs/34629523327) 对本提交记录为成功；该项只归属于 `563db1b`，不回填前五个版本。

## 覆盖边界

- 24 个自然局策略不主动打访客，因此不能证明所有访客效果在自然牌序中执行。
- 36 张建筑验证使用 seeded 场景；它不是 36 次自然抽牌覆盖。
- 浏览器尚未直接执行 7 类特殊工人、37 张 Moor 和 74 张 Rhine 的具体效果；这些有领域测试，但目录或渲染检查不能写成浏览器效果执行。
- Rhine 的 179 项主要表示初始分支注册/渲染，没有穷举全部嵌套、人数、牌序和模块组合。
- 本地 WSL 缺少 C 编译器，未在本机运行 race；远端 CI 的结果应与本地限制分开描述。
- 390px 检查是浏览器视口，不是实体手机。
- 测试使用临时数据和随机回环端口，没有读取真实对局、替换现用服务或覆盖旧 EXE。
