# 项目结构与边界

更新：2026-09-11。单个 Go module，运行时只需一个可执行文件。前端保留原生 ES modules；Node 仅用于格式化和浏览器测试。

```text
essential-web/
├── cmd/viticulture/main.go       # 参数、监听、关闭流程
├── internal/
│   ├── game/                     # 领域模型、规则、访客状态机及对应测试
│   │   ├── model.go              # Room / Player / Action / 棋盘配置
│   │   ├── engine.go             # Apply 事务与行动分派
│   │   ├── seasons.go            # 季节、起床顺序、年末与胜负
│   │   ├── setup.go              # Mama/Papa 初始化和赠礼
│   │   ├── actions.go            # 基础行动
│   │   ├── placement.go          # 格位、奖励、大工人、轭
│   │   ├── resources.go          # 葡萄、酒、陈年、交付
│   │   ├── workers.go            # 灰工人、取回限制
│   │   ├── choices.go            # 可持久化的待决选择与恢复
│   │   ├── view.go               # 玩家视角，隐藏对手牌面
│   │   ├── catalog.go / decks.go # 内嵌卡表和有限牌堆
│   │   ├── cards/ee_cards.json
│   │   └── visitor*.go           # 访客注册、效果、续步、可行性检查
│   ├── server/                   # HTTP、会话、revision、SSE、应用事务
│   └── store/                    # JSON 快照读写
├── web/
│   ├── embed.go                  # 将 static 内嵌到程序
│   └── static/
│       ├── index.html
│       ├── js/                   # 主视图、行动表单、访客、图形、翻译
│       ├── css/
│       └── art/                  # 原有本地 SVG
├── tests/e2e/                    # 真实 Chromium / HTTP / SSE 验证
├── tests/native/smoke.py         # Windows exe 冒烟与重启恢复
├── scripts/                     # PowerShell 构建与检查入口
├── docs/                        # 当前架构、API、规则及历史记录
├── artifacts/                   # 被忽略：测试证据、研究副本、旧产物
├── dist/                        # 被忽略：本次构建
└── runtime/                     # 被忽略：新启动器的存档
```

依赖方向：`cmd → server → game/store`，`server → web`。game 生产代码不依赖 HTTP、磁盘或 server；store 只序列化传入的快照，不决定游戏规则。领域持久化测试通过 store 验证选择队列能正确往返保存。

每个请求先验证会话和 revision。Room.Apply 在规则错误时恢复原状态；HTTP 层在候选状态成功落盘之后才提交并发送 SSE。视图只公开本人手牌、对手手牌类型数量和公共资源；SSE 也按玩家分别投影。

拆分保留原有存档文件名、JSON 字段和 API 路径。新增 handCounts 属于视图数据。不会修复旧存档中已经发生的错误资源；新启动器使用独立目录便于验收。

本次没有引入多 module、通用仓库接口、数据库或前端框架。今后更值得优先处理的是访客定义的统一数据约束和罕见组合裁定；只有出现真实部署／并发需求时，再考虑持久化或房间锁粒度的变化。
