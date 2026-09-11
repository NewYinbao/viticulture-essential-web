> 历史记录 · [所属版本](../release-notes.md)。保留当时的结论、测试范围及本地证据路径；后续状态请查版本索引。

用户要求在当前葡萄酒庄园 Go/原生JS工程实现真实可玩的多扩展并开局可选。已确认范围：Tuscany Essential 的扩展主板/影响力、36建筑卡、11特殊工人，Moor新增40访客，Rhine80访客（按模块过滤4张Tuscany专用）；不做World、Bordeaux、旧Tuscany额外模块/促销。不能只做开关/元数据/演示效果冒充完成。

先读 docs/expansions-scope.md、README、现有AGENTS（若有）、internal/game、web/static。出版社规则/卡面及适用勘误为标准。参考资料在相邻 ../boardspace.net，可只读交叉核对，不能复制其源码或商业美术进本项目。PDF/官网真实证据不足的牌效必须明确未核实，不要猜规则。文档中官网链接已核对，各牌效果尚未审计，务必实际获取规则。原创本地SVG和中文说明，无新增在线运行依赖。

当前引擎EE本体：model Room/Player/Action/Choice；engine.Apply交易回滚；server候选存档后提交；seasons是全桌夏冬本体流程；visitorDefs注册+visitor_special/summer/winter效果分派；decks.Catalog/InitDecks有限牌堆；View显式公私投影；JS app/ee-ui/action-panel各维护映射。Tuscany个人跨季/奖励/特殊工人/建筑可能须严谨扩展，绝不能仅把全桌季节改四季伪装真实Tuscany。所有扩展与原有访客/Planner/Organizer恢复链必须正确。原本体不开扩展完全保留。公共配置持久化、View/API/SSE/轮询同样可见。只有房主在lobby能配置，revision检查/非法组合/开局锁定/旧存档缺省本体。主板EE或Tuscany单选；建筑/特殊工人分别开关；访客EE、EE+Moor、Rhine三选一。Rhine替换而非混洗，并按依赖滤牌。

实施方式：尽快落地真实代码，按规则核对=>小批实现+对应测试=>前后端接入=>组合验收推进，持续更新 docs/expansions-progress.md 的已实现/未实现/未核实清单。不要长时间只调查不写任何成果，不要把本轮末尾一句计划当完成。可以工作到完整验收；如果工具/规则资料客观阻塞，保留可验证工作并明确阻塞，不要虚称完整。不得删减默认回归断言凑通过。

保护边界（必须）：当前git已有未提交的Cloudflare功能，原封保留其语义；已备份artifacts/expansions/source-before-expansions-*.zip。用户PUBLIC服务在3013+runtime/ee-data，另有本体EXE实例；绝不能停止/重启这些用户进程，不能读/更改对局以测试、不能覆盖正在用的dist/Viticulture.exe，不要打开公网测试、不改Windows防火墙/服务/全局PATH/账号配置、不推送/提交Git。新build用dist/Viticulture-Expansions.exe，独立Start-Viticulture-Expansions.cmd/新端口3014/runtime/expansions-data；发布不得自动启动永久服务器。测试必须新临时存档随机回环端口，结束只停自己测试进程。

环境：Windows原生Codex shell需自己验证。项目 C:/Users/admin/Games/Viticulture/essential-web；可通过wsl.exe进入WSL，路径/mnt/c/Users/admin/Games/Viticulture/essential-web。Go在WSL /opt/viticulture-toolchain/go/bin/go，Node WSL /opt/splendor-runtime/bin/node，Windows Node C:/Program Files/nodejs/node.exe。已有node_modules/playwright，用Windows Edge headless channel msedge可测试；此前CLI是否有computer_use不证明实际后端。Windows Python是可用启动/HTTP验证路线；Linux交叉编译GOOS=windows GOARCH=amd64。不要因默认Go不在PATH就停止。

验收：新增规则精确边界+各种卡分支、模块组合（默认关闭/独立启用/合法混合/非法混合）、非房主/旧revision/开局后修改拒绝且状态不变、deck隔离/有限抽弃牌/恢复、隐藏手牌HTTP+SSE/轮询、原有go test ./...和适用audit标签回归；原生Windows发布程序测试；两个独立浏览器真实create/config/join/start和扩展代表动作、保存重启恢复、桌面/手机viewport截图并检查明显溢出。Tuscany/Moor/Rhine任一未完成都不能称三扩展完整，不得把未实现效果在UI标为可用。若不得已留中间版本，开关必须明确禁用阻止开局，文档明确还不可玩；目标依然完成而不是主动降级。

最终输出中文，给绝对路径、具体实现/测试结果、缺失或未核实事项、是否可用的精确结论。写 docs/expansions-release.md 和 artifacts/expansions/implementation-result.json（至少status=complete|partial|blocked、summary、implemented列表、remaining列表、tests列表、binary绝对路径或null）；不能只给一行DONE。整个任务不改Hermes配置。

用户追加必做需求：阅读 docs/expansions-help-requirements.md，规则说明与新手引导必须适配每个启用模块、牌组及实际待决选择，纳入验收；不完成不得标全任务complete。
