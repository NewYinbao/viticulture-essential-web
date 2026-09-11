> 历史记录 · [所属版本](../release-notes.md)。保留当时的结论、测试范围及本地证据路径；后续状态请查版本索引。

## EE Rules Fix 补丁（非完整规则签核）

收尾复验：默认测试、ee_rule_audit 和 go vet 通过；audit_diff 仅 Plant 待裁定项失败，断言未修改。详见 rules-fix-final-verification.json。Windows 原生冒烟待执行 rules-fix-native-smoke.py（隔离临时存档，不使用线上目录）。

新程序：Viticulture-EE-Rules-Fix.exe；启动器：Start-EE-Rules-Fix.cmd。
端口仍为 3012，数据目录仍为 data-ee-allcards。**不要与旧程序同时访问同一存档目录。** 本次没有停止旧服务或读写线上存档。
修复收获专家双效果、灰工人唯一性、越级酒陈年、工人取回/轭、交换藤建筑限制处理、up-to 最低执行量及可放弃行动格奖励。
原存档缺少本次新增的 grayWorkerOwner / triggerSeat，既有错误资源和待决动作不会自动修复；建议新房间验证。主板 Plant 拔藤规则仍待权威 EE 证据，原失败审计保留，不宣称全规则通过。
详见 RULE_FIX_REPORT.md；隔离真实浏览器复验：rules-fix-browser-test.cjs。API 行动增加 declineBonus:true（默认 false），普通放置及 Planner 执行均支持；Manager 本身不获夏季格奖励，外层奖励独立保留。

---

# 葡萄酒庄园 EE 本体全卡验收版（RC1）

## 范围
Go 单进程、本机服务器、2–6 位真人各自浏览器登录；全部资源本地内嵌，无需 Java、Node 或互联网。固定190条卡记录：42藤、36订单、38夏访客、38冬访客、18 Mama、18 Papa。76张访客真实效果、多人响应和延后行动已实现，不使用通用演示效果。

本版本是全卡验收候选，非官方认证，不承诺所有罕见组合已穷尽验证。不是 Tuscany/其他扩展，不含 Automa/单人。

## 启动
双击 `Start-EE-AllCards.cmd`。电脑访问 http://localhost:3012 。同一局域网其他设备访问 http://服务器局域网IP:3012 ，各用自己的浏览器昵称加入房间。防火墙如询问，只允许可信的专用网络。不要开放公网。

服务器存档目录 `data-ee-allcards`，与旧版隔离；原开发版exe、旧启动器和旧存档保留，未被覆盖。旧浏览器3011上的身份不会自动搬到3012；本版请新建房间。关闭服务器控制台或 Ctrl+C 停止，重新启动同一目录后原浏览器刷新恢复。不要删除浏览器本地存储。不要同时启动两个本版进程。

## 已执行验收
- Go 全量测试、独立基础规则审计测试及 go vet。
- 76卡初始效果/选项测试、费用/回滚/权限与持久化专项。
- 真浏览器142个普通初始选项场景及19个特殊场景（含Queen三分支、Organizer各行、Planner六类冬初行动），真实HTTP/SSE；这是资源夹具测试，不是142局自然对局。
- 不注入资源的自然完整对局：UI建房/加入/开始，合法API推进生产/访客/得分直到达到20分后的年末终局；断线追帧和服务器重启恢复。详细证据位于 `full-game-evidence/result.json`，随机发牌使每次结果不同。
- 原手牌区点选、高亮上抬、取消选择、手机视口、Cottage先确定两色再抽牌。
- Windows可执行验证详见 `release-verification.json`（生成后存在）。手机视口测试不是第二台物理设备实测。

## 明确的特殊组合约定和边界
1. 多个Planner预约按预约先后执行；未来冬初执行时才给行动奖励，不再次扣工人，无法执行不退工人。
2. Organizer先结算外层奖励格第二访客，再结束本季。该罕见组合优先级尚缺独立官方裁定；不是声称已官方签核。
3. Importer采用对手依次私有提交/托管协议；恰好3张转交，否则原样退还并抽3冬卡。
4. Manager任意嵌套、多张卡极端组合未穷举；发现边界错误应保留存档并报告，不以“测试通过”代替全规则证明。
5. 不自动修复旧开发版已经卡住的半结算存档，也不迁移旧简化规则预览存档。

## 美术与来源
原创本地SVG主题插画，部分卡牌共用；不是商业原图、不是逐卡AI生成的190张独立画作。卡牌中文为本项目译文，保留英文名用于对照。
规则来源和Queen卡面核实见 EE_RULE_AUDIT.md、EE_SOURCES.md。代码与美术授权链不同；只做本地交付，未完成公开发行许可证审计，不应再分发商业卡面或参考源。

## 开发复验
`go test -count=1 -tags ee_rule_audit ./...`
`go vet ./...`
`node visitor-browser-test.cjs`（脚本使用隔离WSL测试工具链，见脚本顶部）
`QA_SPECIAL=1 node visitor-browser-test.cjs`
`node natural-full-game-test.cjs`
可执行文件内嵌web；改前端后必须重新构建。
