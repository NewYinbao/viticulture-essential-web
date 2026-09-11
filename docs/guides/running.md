# 运行与存档

在项目根目录使用 Go 1.25+ 执行：

```powershell
powershell -NoProfile -File scripts/build-continuation.ps1
.\Start-Viticulture-Continuation.cmd
```

续作服务使用端口 3015，存档目录 runtime/continuation-data。房主访问 http://localhost:3015，朋友使用房主的局域网 IP。建房、加入与恢复已有座位需要房间内对应的昵称和密码。

旧程序和旧对局不会自动迁移。升级前停止使用该存档的进程并备份；同一目录只能由一个服务写入。旧无密码座位的登记规则见[玩家密码记录](../versions/06-expansions-563db1b/records/玩家密码与续作记录-20260911.md)。

源码运行：

```sh
go run ./cmd/viticulture -addr 0.0.0.0:3015 -data ./runtime/continuation-data
```

退出终端或 Ctrl+C 停止自己启动的服务。EXE、存档和本地验收产物不提交到 Git。公网访问见[分享指南](cloudflare.md)。
