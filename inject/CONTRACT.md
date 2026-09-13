# 娱乐场 · 源码注入形态 总纲（v1）

目标：娱乐场不再是独立服务/外部页面，而是通过一个注入脚本，把代码打入 sub2api
源码树并构建启动——后端成为主站内部包、前端成为主站原生 Vue 路由页面，
鉴权/主题/部署与主站完全一致。本文件是所有实现方（子代理）的共同契约。

## 注入布局（staging → 目标路径）

staging 目录：本仓库 `inject/`（与 inject.py 同级的 backend/、frontend/、backend-patch/）。
注入脚本 `inject.py` 把 staging 拷入 sub2api 源码树，并在标记处打补丁。

- `backend/internal/casino/**`        → `sub2api/backend/internal/casino/**`
  （casino 包：store/games/gamemgr/handlers + Register 挂载入口）
- `backend-patch/casino.go`           → `sub2api/backend/internal/server/routes/casino.go`
- `frontend/src/views/casino/**`      → `sub2api/frontend/src/views/casino/**`
- `frontend/src/components/casino/**` → `sub2api/frontend/src/components/casino/**`
- `frontend/src/api/casino.ts`        → `sub2api/frontend/src/api/casino.ts`
- `frontend/src/assets/casino/**`     → `sub2api/frontend/src/assets/casino/**`
  （reference-*.jpg、symbols/、scratch/ 素材 + game-scenes.css）
- 补丁片段（标记之间替换，幂等，源记录见 `patches-extracted.txt`）：
  - 后端路由注册：`// ===CASINO:BEGIN=== ... // ===CASINO:END===`
  - 前端路由：`router/index.ts` 内同样标记
  - 前端侧边栏菜单 / api 导出 / i18n 注册 / 系统设置页：同样标记

## 后端 API 契约（同源，全部挂主站用户鉴权；admin 路由要求管理员）

统一前缀 `/api/v1/casino`。响应为主站标准信封 `{code:0, message:"success", data:{…}}`，
失败走主站 response 辅助函数（HTTP 4xx/5xx + 中文 message）；前端 axios 拦截器自动解包。
以下仅描述 data 载荷：

- GET  /api/v1/casino/meta
  → {min_bet, max_bet, daily_loss_limit,
     wheel:{segments:[{label,multiplier}]},
     slots:{symbols:[{id,emoji}]},
     blackjack:{blackjack_pays,dealer_stands_soft17,double_allowed},
     sicbo:{big,small,odd,even},
     baccarat:{player,banker,tie},
     scratch:{win_rate,max_multiplier,
              lucky7:{cells,hit_rate,max_cells_mult},
              lines:{win_rate,symbols:[{id,multiplier}]}}}
- GET  /api/v1/casino/me → {user:{id,email,role,balance}, today_profit, today_rounds}
- GET  /api/v1/casino/history?limit=&offset= → {items:[BetRow...]}
  BetRow: {id,game,bet,payout,delta,balance_after,status,created_at}
- GET  /api/v1/casino/leaderboard → {items:[{rank,player,profit}]}（近7天前5）
- POST /api/v1/casino/games/wheel/spin {bet} → {segment,multiplier,payout,bet,balance}
- POST /api/v1/casino/games/slots/spin {bet}
  → {reels:[emoji×3],reel_ids:[id×3],multiplier,payout,bet,balance}
- POST /api/v1/casino/games/sicbo/roll {bet,bet_type:big|small|odd|even}
  → {dice:[n,n,n],sum,result:big|small|triple,multiplier,payout,bet,balance}
  （大 11-17 / 小 4-10 / 单 奇和 / 双 偶和，倍数含本金默认 2；围骰通杀全部判负，
   game 记 "sicbo"；配置键 sicbo:{big,small,odd,even}，赔率须 ≥1）
- POST /api/v1/casino/games/baccarat/deal {bet,side:player|banker|tie}
  → {player_cards:[{r,s}],banker_cards,player_points,banker_points,
     outcome,multiplier,payout,bet,balance}
  （单副牌，A=1/10JQK=0，标准天牌与要牌表；倍数含本金默认 player 2、
   banker 1.95（5% 佣金）、tie 9；和局只对 tie 注赔付，game 记 "baccarat"；
   配置键 baccarat:{player,banker,tie}，赔率须 ≥1）
- POST /api/v1/casino/games/scratch/reveal {mode:classic|lucky7|lines, face_value, count:1~10}
  → {mode,face_value,count,cards,win_count,total_payout,total_bet,balance}
  cards[i]: {multiplier,prize,cells,win_lines?}
    - classic：cells 为单格 {kind:"mult"|"dud", multiplier?}
    - lucky7 ：cells 为 7 格 {kind:"seven"|"dud", multiplier?}，命中格倍数累加
    - lines  ：cells 为 9 格（行优先）{kind:"symbol", symbol, multiplier}，
               win_lines 为中奖线（每条 3 个下标，构造保证至多一条）
  一次性原子结算 total = face_value × count（game 记 "scratch"）
- POST /api/v1/casino/blackjack/deal {bet} → {game}
- POST /api/v1/casino/blackjack/hit|stand|double {game_id} → {game}
- GET  /api/v1/casino/blackjack/current → {game|null}
  game: {game_id,bet,doubled,player_cards:[{r,s}],dealer_cards,player_total,
         player_soft,status:'active'|'settled',result,payout,can_double,balance}
- GET/PUT /api/v1/casino/admin/config（admin）→ {config, rtp}
  （config 含 sicbo/baccarat/scratch 键；rtp.scratch={classic,lucky7,lines}、
   rtp.sicbo/baccarat 为各注项理论回报；旧配置缺省键时加载/保存自动补默认值）
- GET  /api/v1/casino/admin/stats（admin）→ {summary, games, recent}

行为约束（与旧插件一致，不得改变）：
- 服务端权重抽签 + 单条 SQL 原子结算
  （UPDATE users SET balance=balance-$1+$2 WHERE id=$3 AND deleted_at IS NULL
    AND status='active' AND balance >= $1 RETURNING balance）
- 自有账本表 casino_bets / casino_blackjack_games / casino_settings
- 不支持分牌/保险/多赢线；slots 三卷轴单线（视觉 5×3 仅装饰）
- 21点闲置 12h 自动按停牌结算；double_allowed=false 时前后端都禁加倍

## 相对旧插件的差异

- 不再有独立登录/会话/cookie：身份取自主站鉴权上下文（中间件注入的 user id/role）。
- 旧 store 中所有 `token_version` 引用删除（真实 sub2api users 表无此列）。
- 结算后若主站有余额缓存则尽力失效（redis best-effort，失败不影响结算）。
- 菜单注入：改为前端侧边栏静态菜单补丁（不再写 settings.custom_menu_items）。

## 前端页面（原生 Vue，sub2api 组件/主题）

- 路由（登录保护，与主站一致）：
  /casino            游戏大厅（CasinoLobbyView）
  /casino/wheel      大转盘（CasinoWheelView）
  /casino/slots      老虎机（CasinoSlotsView）
  /casino/blackjack  21点（CasinoBlackjackView）
  /casino/sicbo      骰宝（CasinoSicboView）
  /casino/baccarat   百家乐（CasinoBaccaratView）
  /casino/scratch    刮刮乐（CasinoScratchView，三玩法 + 可刮涂层）
  /casino/history    游戏记录（CasinoHistoryView）
- 布局沿用主站默认（侧边栏+顶栏），内容区放游戏舞台；
  各游戏舞台视觉以 assets/casino 下的生成素材（reference-*.jpg / scratch/*）为蓝本，
  颜色/排版尽量用主站 token 与组件（Card/Button/Badge/统计条）。
- 管理面板：主站「系统设置 → 游戏管理」标签页（SettingsView 标记补丁挂 CasinoAdminPanel）。
- API 全走主站 api 客户端（自动带 token、统一错误提示）。

## inject.py（一键注入）

1. 把 staging 拷入 sub2api 对应路径（覆盖）。
2. 对 `backend` 路由总入口、`router/index.ts`、侧边栏菜单文件做标记间替换（幂等）。
3. `pnpm build`（frontend）→ 产物在 backend/internal/web/dist。
4. `go build -tags embed -o bin/server-embed.exe ./cmd/server`。
5. `--run` 时停止旧进程并启动新后端（默认仅打印启动命令）。

## 源码树干净原则（硬性）

- sub2api 源码树**默认保持原封不动**：所有新增文件只存放在 staging，
  所有标记补丁（router.go / router/index.ts / AppSidebar.vue / api 导出 /
  i18n index / 系统设置页 / go.mod 的 pgx 依赖）都以数据形式保存在
  gen_inject.py 与 patches-extracted.txt 里。
- **只有执行 inject.py 时**才把文件拷入源码树并打上标记补丁；
  脚本提供 `--revert`：删除注入的新增文件、移除所有 CASINO 标记块、
  还原 go.mod，使源码树回到干净状态。
- 开发期为了编译验证临时打上的补丁，验收完成后必须 `--revert` 还原再由
  inject.py 重新注入；最终交付 = 干净源码树 + inject.py 一键（注入→构建→启动）。
