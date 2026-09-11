# 验证记录：种植检查、规则速查与新手引导

- 提交：`80e9f7c4234e9e10bd19dabc01122e429436c242`
- 主要证据：[ui-review.md](records/ui-review.md)、[testing.md](records/testing.md) 及修改后的 `tests/e2e/usability-test.cjs`。

## 本提交新增的三组浏览器场景

1. 不可种藤的完整缺项和只读检查：核对设施、容量、田地出售状态，并确认不会选中无效藤或发送行动。
2. 四步导览、键盘操作、偏好保存及规则页面往返：核对打开规则不会丢失行动草稿，Esc 层级正确。
3. 390px 布局和上下文提示切换：核对 SSE 更新、回合变化和待决选择时提示定位正确。

当版 UI 记录将体验检查累计数更新为 14 组，并列出 `desktop-plant-conditions.png`、`mobile-plant-conditions.png`、`desktop-rules.png`、`mobile-rules.png`、`mobile-beginner-guide.png` 等本地截图证据。

## 当版累计检查

截至本提交，默认 Go、`ee_rule_audit`、`audit_diff`、vet 及此前本体浏览器/原生验证仍记录为通过。新增功能主要由真实服务器上的 usability 场景验证；查看规则和切换新手模式不调用游戏行动 API。

## 覆盖边界

- 14 组是累计体验场景，不是 14 局完整自然游戏。
- 390px 是浏览器 viewport，不是实体手机。
- 规则说明只覆盖本版 EE 范围；不能用本页证明后续扩展规则、配置隔离或密码功能。
- 没有单独记录本提交的远端 CI 运行，因此不作 CI 通过声明。
