# 临时公网分享

默认续作启动器提供局域网访问。需要临时公网链接时，在项目根目录构建续作程序、安装 cloudflared，再显式启动：

```powershell
powershell -NoProfile -File scripts/build-continuation.ps1
powershell -NoProfile -File scripts/install-cloudflared.ps1
.\dist\Viticulture-Continuation.exe -addr 127.0.0.1:3015 -data runtime/continuation-data -public -cloudflared dist/tools/cloudflared.exe
```

启动前确保该端口和存档没有被另一个进程占用。使用程序输出的完整分享链接；每次启动的访问密钥和临时域名可能变化。分享入口密钥与玩家座位密码是两层不同的验证。公网模式使用轮询同步。

停止该服务会结束临时分享。此操作需要联网，旧启动脚本不等于最新 Continuation 程序。[该功能开发及验证原始记录](../versions/06-expansions-563db1b/records/cloudflare.md)。
