> 历史记录 · [所属版本](../release-notes.md)。保留当时的结论、测试范围及本地证据路径；后续状态请查版本索引。

# Essential Edition research sources

Scope: Essential Edition base multiplayer rules, not Tuscany Essential or Viticulture Plus. Original handdrawn-style artwork only; commercial card art is not included.

- Publisher EE overview: https://stonemaiergames.com/games/viticulture/essential-edition/
- Publisher FAQ: https://stonemaiergames.com/games/viticulture/faq/
- EE rulebook mirror: https://boardspace.net/viticulture/english/VitiRulebook_EssEd_2nd_r6.pdf
- Publisher-linked card inventory: https://docs.google.com/spreadsheets/d/1mCRAuf99t6tuiPRWHbaffc07waRcMcDtx7FVwrSqcbw/edit#gid=0
- Boardspace data reference: https://github.com/ddyer0/boardspace.net/tree/main/client/boardspace-java/boardspace-games/viticulture
- Parent visual revision (mechanics unchanged): https://stonemaiergames.com/pride-in-the-world-of-viticulture/

## Important differences
Boardspace defaults to Tuscany Essential and deliberately omits Queen from the winter deck. Its default or Plus settings must not be treated as the base EE standard. EE requires 38 summer and 38 winter visitor records including Queen. Boardspace replay revision numbers do not identify physical game editions.

The publisher-linked card inventory is not a complete machine-readable effect database. This project has not completed independent card-by-card source verification. Java source files are GPL-3.0-or-later; commercial images have no verified reuse authorization for this separate project. No commercial card image is bundled. Final source/distribution license audit is still required before public distribution; this is an isolated local development tree, not a release package.

## 当前状态及来源核实
76张访客已实现，有逐初始分支及HTTP/持久化/浏览器测试；不是官方规则认证。细节见EE_RULE_AUDIT.md和README.md。
Queen卡面已由主代理直接目视核实：右邻座选择失1 VP、交任意2手牌或付3金币。图片只作规则参考，未加入游戏资源。
Planner预约次序和Organizer双访客pass优先级仍为公开的实现约定，未声称获得独立官方裁定。
