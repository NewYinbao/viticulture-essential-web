# EE Rules Fix 收尾交付报告

## 结论与范围
这是规则修复候选，**不是完整规则签核**。沿用上一工作线程的实现，本次只复核、重新测试/构建和补齐交付文档，没有重新研究规则，也没有修改或弱化 Plant 审计。

修复范围：Harvest Expert 收获并抽藤及中文说明；年度唯一灰工人及 Organizer/跨年/读档；越级酒陈年跨酒窖边界；Producer/Manager 仅禁止取回当前触发工人、保留更早同格工人及外层结算；取回轭工人后恢复轭可用性但不解除田地年度收获限制；交换藤建筑限制处理；up-to 最低执行量和公共资源可行性（不窥视牌堆）；普通行动和 Planner 可放弃奖励（declineBonus）。

## 已知待定与兼容边界
- **TestAuditDiffPlantMustNotOfferUproot 仍失败**：主板 Plant 目前接受无轭拔藤。其规则适用性待权威 EE 证据裁定；既不把当前行为签核为正确，也不在证据不足时强改实现。audit_core_diff_test.go 与修复前备份逐字节一致，失败断言保留。
- audit_diff 其余测试通过；不是“全测试通过”。日志中 not-a-directory 是存档失败回滚测试的故障注入，与 Plant 失败不同。
- 原存档缺少 grayWorkerOwner / triggerSeat；旧局已错误获得的资源和半结算动作不自动修复、不迁移。建议新房间验证。
- Organizer 外层第二访客的结算顺序、极端嵌套等旧版明确边界仍保留，不以本次补丁声称已获官方裁定。
- 前一线程修改的审计夹具已复核：summer-32 提供可取回工人；灰工人夹具设置所有权；winter-12 允许在动作预检时就拒绝零收获。Plant 审计未改。本轮未修改任何测试或实现。

## 本轮实际验证
| 命令（Go 路径 /opt/viticulture-toolchain/go/bin/go） | 结果 |
|---|---|
| go test -count=1 ./... | PASS |
| go test -tags ee_rule_audit -count=1 ./... | PASS |
| go vet ./... | PASS |
| go test -tags audit_diff -count=1 ./... | FAIL，仅 Plant 上述测试 |
| GOOS=windows GOARCH=amd64 go build -trimpath -ldflags '-s -w' | PASS |

原始日志：rules-fix-final-verification.json；前一线程测试和专项回归日志保留于 rules-fix-verification.json。

沿用 rules-fix-browser-result.json 的五项真实 Chromium/HTTP 隔离场景：普通抽卡奖励放弃、卖田 VP 奖励放弃、Planner 读档后放弃奖励、当前触发座位隐藏而早先座位可取回、Harvest Expert 双效果及译文。completed=true、errors=[]，发现 visitor_sequence.go 和补充测试晚于原报告，因此仅重新执行原五项定向脚本（/opt/splendor-runtime/bin/node rules-fix-browser-test.cjs），再次通过；未重跑广泛浏览器测试。该验证使用夹具，不等同自然完整对局或 Windows 原生执行验证。

## 产物和保护措施
- Viticulture-EE-Rules-Fix.exe：Windows amd64，7,576,576 字节。
- SHA-256：4634663f76b0ee59e0461e851a20e5b99d4f6136fe21ffa177200662d8c8802b。
- 上一线程已生成该新命名 EXE。本轮在临时目录重新构建，输出与其逐字节相同，因此保留原文件，不覆盖任何既有 EXE。
- Start-EE-Rules-Fix.cmd 已存在且参数符合交付要求，保留原文件：-addr 0.0.0.0:3012 -data "%~dp0data-ee-allcards"。
- 本轮没有启动交付启动器、操作在线服务或读取/写入真实存档。旧版 EXE 和备份均保留。不要同时用两个进程访问 data-ee-allcards；启动交付版前人工确认旧服务已停止并自行备份。
- Windows 原生冒烟已由主代理独立执行通过，结果见 rules-fix-native-result.json。执行 rules-fix-native-smoke.py：仅使用自动分配的回环端口和临时数据目录，测试健康、静态页、建房、加入、开始、认证及重启恢复；不使用 3012 或真实存档。

## 原生 HTTP 接口（核对 main.go）
参数为 -addr 和 -data。GET /api/health 返回 ok=true、version=ee-rules-fix。POST /api/create 使用 name，返回 token/code；POST /api/join 使用 name/code。GET /api/state 和 POST /api/action 使用 Bearer token 认证，动作 JSON 还需携带当前 revision。接口已核对 main.go。

## 环境问题
当前 read_file/search_files 文件工具无法解析 WSL 项目路径，terminal 中 Python pathlib 能正确访问，因此本轮文件复核/文档生成使用该备用方式。terminal 末尾有 Hermes Windows 缓存路径提示，不影响已捕获的 Go 子进程退出码、校验和与产物结果。项目不含 Git 元数据，变更复核使用 rules-fix-changed-files.json 指向的原始备份。

## 主代理独立复验
Windows 原生修复版健康、内嵌页面、认证拒绝、建房、加入、开始和重启恢复均通过；测试进程已停止，使用临时存档，无线上存档改动。EXE SHA-256 与上述构建一致。默认测试、ee_rule_audit 和 go vet 再次通过。完整 audit_diff 仍只保留 Plant 未裁定失败，不宣称全绿。
