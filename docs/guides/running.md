# 运行与存档

在项目根目录使用 Go 1.25+ 执行：

```powershell
powershell -NoProfile -File scripts/build-continuation.ps1
.\Start-Viticulture.cmd
```

续作服务使用端口 3015，存档目录 runtime/continuation-data。房主访问 http://localhost:3015，朋友使用房主的局域网 IP。建房、加入与恢复已有座位需要房间内对应的昵称和密码。

旧程序和旧对局不会自动迁移。升级前停止使用该存档的进程并备份；同一目录只能由一个服务写入。旧无密码座位的登记规则见[玩家密码记录](../versions/06-expansions-563db1b/records/玩家密码与续作记录-20260911.md)。

源码运行：

```sh
go run ./cmd/viticulture -addr 0.0.0.0:3015 -data ./runtime/continuation-data
```

退出终端或 Ctrl+C 停止自己启动的服务。EXE、存档和本地验收产物不提交到 Git。公网访问见[分享指南](cloudflare.md)。

项目根目录只保留 Start-Viticulture.cmd。启动时可选择局域网或公网分享，扩展包在房间开局设置中选择，与启动脚本无关。默认仍使用原续作的 runtime/continuation-data；旧版本 EXE 与存档保留，但不再提供旧版本快捷启动入口。

## 同一浏览器登录多个账号

点击“另开账号”，每个新标签页分别输入自己的昵称和座位密码。登录信息按标签页保存，刷新保留；关闭后新开标签页可重新登录恢复座位。退出或更换账号只影响当前标签页；修改密码会撤销同一座位的其他会话，不影响其他玩家。浏览器自带的“复制标签页”可能复制当前身份，请使用游戏内按钮打开独立登录页。

升级后，旧的浏览器级登录信息仅迁移到首次打开的标签页，随后移除旧共享凭据。密码不会保存在浏览器存储中。
