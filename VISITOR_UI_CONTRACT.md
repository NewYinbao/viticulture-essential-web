# 访客UI集成契约
当前支持kind=visitor以及kind=planner。choiceId/revision及所有权由服务器检查，只有本人手牌可选。全部76访客的初始及后续表单已接通，不使用手填JSON。
Planner冬初通过公开view.planned[0]选择相应动作表单；该结构不含未来私有选牌。Cottage双颜色原子提交。Queen明确失1VP/交2卡/付3金币。Importer托管写入nil保护已修复。
142普通初始选项+19特殊场景真实浏览器测试通过，数据为隔离夹具；natural-full-game-test.cjs是独立非夹具自然对局。
手牌点选在原区域高亮上抬，SSE同choiceId保留有效草稿，失败不清空，响应者权限独立。
