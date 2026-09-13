# Sub2API 娱乐场（Casino for Sub2API）

Sub2API 的源码注入式娱乐玩法插件：**大转盘 · 老虎机 · 21 点 · 骰宝 · 百家乐 · 刮刮乐**，使用主站账户余额下注。
注入后娱乐场成为主站的一部分——原生 Vue 页面、主站登录态、同源 API、主站侧边栏入口，
不改变 sub2api 任何既有代码（全部改动通过幂等标记补丁注入，可一键还原）。

> ⚠️ 免责声明：本项目仅供学习研究与娱乐用途。游戏结果由服务端随机产生，长期期望回报为负。
> 请遵守所在地区法律法规，未满 18 岁或当地禁止此类内容请勿使用，请理性游戏。

---

## 特性

- **六款游戏**（全部服务端结算，理论回报率 92%~97%）：
  - **大转盘**：12 扇区权重抽签，canvas 旋转动画
  - **老虎机**：3 卷轴单线结算（5×3 视觉机台，三连/两连派彩）
  - **21 点**：标准规则，Blackjack 赔 3:2，支持要牌/停牌/加倍；牌局可刷新恢复，闲置 12 小时自动按停牌结算
  - **骰宝**：大小/单双四注项，围骰通杀，倍数含本金
  - **百家乐**：单副牌标准补牌表，闲 2 / 庄 1.95（5% 佣金）/ 和 9，和局退本
  - **刮刮乐**：「面值 × 数量」模型 × **三种玩法**（见下）
- **原生集成**：页面为主站 Vue 路由（`/casino/*`），鉴权走主站 JWT，余额与主站实时同步
- **资金安全**：单条 SQL 原子结算（`balance >= bet` 守卫）、独立账本表 `casino_bets`、Redis 余额缓存失效
- **管理能力**：主站「系统设置 → 游戏管理」标签页内可开关娱乐模式、配置下注限额与各游戏赔率（含 RTP 校验），并支持盈亏统计
- **娱乐模式开关**：关闭后用户侧隐藏娱乐场入口、游戏接口返回 403，管理员可随时重新开启
- **干净注入**：`inject.py --revert` 一键把 sub2api 源码树还原到未注入状态

## 刮刮乐三种玩法

「选面值（2~50 元）× 购买数量（1~10 张）」一次性购入并逐张独立结算，所有玩法涂层均为真实可刮图层（支持指针拖刮 / 自动刮奖 / 一键刮开动画）：

| 玩法 | 规则 | 默认 RTP |
|------|------|----------|
| 经典刮奖 | 单格涂层直接开出倍数，中奖率 29%，最高 88.888 倍 | ≈93.9% |
| 幸运7 | 7 格涂层寻找幸运7，每格独立 8.5% 命中，命中格倍数累加，单卡理论上限 622 倍 | ≈94.9% |
| 幸运连线 | 3×3 宫格 8 条线（3 横 3 竖 2 斜），三连相同符号即中该符号倍数；出奖时服务端保证恰好一条连线，其余格子防误中奖填充 | ≈92.7% |

## 环境要求

| 依赖 | 版本 | 用途 |
|------|------|------|
| sub2api 源码 | 与插件版本匹配的 checkout | 注入目标 |
| Go | ≥ 1.22 | 后端编译 |
| Node.js + pnpm | Node ≥ 18 | 前端编译（`npm i -g pnpm`）|
| PostgreSQL | sub2api 所需版本 | 账本与余额（共用主站库）|
| Redis | 任意近期版本 | 主站登录限流 / 余额缓存失效（必需）|

## 一键部署

### 方式一：预编译可执行文件（推荐，零依赖）

从 [Releases](https://github.com/your-org/sub2api-casino/releases) 下载对应平台的单文件：

| 平台 | 文件 | 用法 |
|------|------|------|
| Windows | `inject-casino-windows-amd64.exe` | `inject-casino-windows-amd64.exe --root C:\path\to\sub2api --run` |
| Linux | `inject-casino-linux-amd64` | `./inject-casino-linux-amd64 --root /path/to/sub2api --run` |
| macOS | `inject-casino-darwin-amd64` | `./inject-casino-darwin-amd64 --root /path/to/sub2api --run` |

> 无需 Python / Node 环境（构建工具链由可执行文件内嵌）。

### 方式二：Python 脚本（需 Python ≥ 3.9）

```powershell
# Windows
python inject\inject.py --root C:\path\to\sub2api --run

# Linux / macOS
python3 inject/inject.py --root /path/to/sub2api --run
```

### 方式三：Windows 批处理（无需 Python，但补丁应用需 Python）

```powershell
install.bat C:\path\to\sub2api
```

> 批处理版只负责拷贝与构建；幂等补丁应用仍依赖 Python（系统无 Python 时会提示手动打补丁）。

### 远程一键（Linux/macOS）

```bash
# 把仓库地址换成你实际存放本插件的地址
curl -fsSL https://raw.githubusercontent.com/your-org/sub2api-casino/main/install.sh | bash -s -- /path/to/sub2api
```

所有方式都会自动：拷贝源码 → 打幂等标记补丁 → `pnpm build` → `go build -tags embed` → 启动后端（:8080）。
完成登录后，点击侧边栏「**娱乐场**」即可。

### 卸载 / 还原

```powershell
# Windows（exe）
inject-casino.exe --revert --root C:\path\to\sub2api

# Windows（Python）
python inject\inject.py --revert --root C:\path\to\sub2api

# Linux / macOS
python3 inject/inject.py --revert --root /path/to/sub2api
```

还原后重新构建一次主站即可。

## 游戏规则速览

| 游戏 | 规则 |
|------|------|
| 大转盘 | 下注后旋转，中奖金额 = 下注 × 扇区倍数（含本金）；扇区按服务端权重抽取 |
| 老虎机 | 视觉为 5×3 机台，实际**中间三列的中间一行**为结算区（三连/两连派奖）；不支持多赢线 |
| 21 点 | 目标不超过 21 点且大于庄家；Blackjack 赔 3:2；支持要牌/停牌/加倍；**不支持**分牌/保险/投降 |
| 骰宝 | 三颗骰子点数和：大（11-17）/ 小（4-10）/ 单 / 双，倍数含本金默认 2；围骰（三同）通杀全部注项 |
| 百家乐 | 闲/庄/和三向下注；A 计 1 点、J/Q/K 计 0，按标准补牌表要第三张；和局退本 |
| 刮刮乐 | 见上文「刮刮乐三种玩法」；刮开涂层约 60% 自动判定翻开，奖品即时入账 |

所有概率与派彩由服务端配置驱动，玩家端接口不暴露权重；管理员可在「系统设置 → 游戏管理」调整，
或直接调用 `GET/PUT /api/v1/casino/admin/config`。

## 架构

```
注入后 sub2api 源码树新增：
backend/internal/casino/            Go 包：pgx store + 各游戏数学引擎 + Gin handlers
backend/internal/server/routes/     casino 路由薄封装（/api/v1/casino/*）
frontend/src/views/casino/          八个原生 Vue 页面（大厅/六游戏/记录）
frontend/src/components/casino/     轮盘 canvas / 老虎机列带 / 可刮涂层 / 扑克牌渲染器
frontend/src/api/casino.ts          主站 axios 客户端封装（自动带 token / 401 刷新）
frontend/src/i18n/locales/*/casino.ts  中英文案
```

- 路由鉴权：复用主站 `JWTAuthMiddleware` / `AdminAuthMiddleware`，无独立登录体系
- 响应格式：主站信封 `{code:0, message, data}`，前端 axios 拦截器统一解包
- 结算：`UPDATE users SET balance = balance - bet + payout WHERE ... AND balance >= bet RETURNING balance`

详细契约见 [`inject/CONTRACT.md`](inject/CONTRACT.md)。

## 目录说明（本仓库）

```
inject/
├── inject.py               一键注入/构建/启动（--revert 还原，由 gen_inject.py 生成）
├── gen_inject.py           inject.py 的生成器（修改补丁后重新生成）
├── patches-extracted.txt   全部幂等标记补丁的源记录
├── CONTRACT.md             架构与 API 契约
├── backend/                Go 包源码（注入到 backend/internal/casino）
├── backend-patch/          路由薄封装（注入到 backend/internal/server/routes）
└── frontend/               Vue 页面/组件/API/i18n/素材（注入到 frontend/src）
install.sh                  Linux/macOS 一键拉取部署
install.bat                 Windows 批处理（无需 Python，补丁应用仍需 Python）
inject-casino.exe           预编译 Windows 单 exe（零 Python 依赖，8MB）
LICENSE                     MIT
```

## 常见问题

- **为什么要"注入"而不是独立服务？** sub2api 对页面施加了严格 CSP（`connect-src`/`frame-src`
  仅允许自身与少量白名单），独立进程的跨源 API 与 iframe 会被浏览器拦截；注入后全部同源。
- **余额单位/充值提现？** 余额即主站账户余额，充值提现均在主站完成，插件不持有资金入口。
- **限流？** 游戏接口按用户 5 次/秒滑动窗口限流，超限返回"操作过于频繁"。
- **长时间未操作的 21 点牌局？** 闲置 12 小时后由后台巡检按停牌规则自动结算。
- **刮刮乐配置兼容？** 旧版持久化配置缺 lucky7/lines 键时，加载/保存自动补默认值，无需手动迁移。
- **Windows 下启动报"First run detected"？** 这是 sub2api 的已知行为：它默认找 `/app/data/config.yaml`
  （Docker 路径），Windows 下会落空。解决：启动前设 `DATA_DIR=backend目录`，或在 `backend/config.yaml`
  里写好数据库/Redis 连接信息。
- **没有 Python 能用吗？** 能。下载对应平台的预编译可执行文件（Releases），一条命令搞定：
  - Windows：`inject-casino-windows-amd64.exe --root C:\path\to\sub2api --run`
  - Linux/macOS：`./inject-casino-linux-amd64 --root /path/to/sub2api --run`
- **怎么发布新版本？** 打 tag 推送到仓库即可触发 GitHub Actions 自动构建三平台可执行文件并发布到 Releases：
  `git tag v1.0.0 && git push --tags`

## Roadmap

- [ ] 骰宝/百家乐派彩管理端编辑 UI（后端配置已支持）
- [ ] 音效与更多动画档位
- [ ] 刮刮乐玩法配置化管理端 UI
