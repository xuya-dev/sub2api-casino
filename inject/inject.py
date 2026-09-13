# -*- coding: utf-8 -*-
# Sub2API 娱乐场 - 一键源码注入脚本（本文件由 gen_inject.py 生成，勿手改）
# 用法:
#   python inject.py              # 注入(拷贝+补丁) + 构建前端/后端
#   python inject.py --run        # 注入构建后重启后端(端口8080)
#   python inject.py --revert     # 还原 sub2api 源码树到干净状态
#   python inject.py --skip-build # 只注入不构建
import io, json, os, re, shutil, subprocess, sys, time

# PyInstaller 打包后 __file__ 指向临时目录，staging 应从 exe 所在目录找；
# exe 在插件根目录时 staging 在其 inject/ 子目录，直接放 inject/ 内时就在当前目录。
if getattr(sys, 'frozen', False):
    _base = os.path.dirname(os.path.abspath(sys.executable))
    HERE = os.path.join(_base, 'inject') if os.path.isdir(os.path.join(_base, 'inject', 'backend')) else _base
else:
    HERE = os.path.dirname(os.path.abspath(__file__))
DEFAULT_ROOT = r'C:/OpenSource/Sub2api/sub2api'
BEGIN, END = '// ===CASINO:BEGIN===', '// ===CASINO:END==='
HTML_BEGIN, HTML_END = '<!-- ===CASINO:BEGIN=== -->', '<!-- ===CASINO:END=== -->'
GO_REQUIRE = '\tgithub.com/jackc/pgx/v5 v5.7.4\n'
COPIES = json.loads(r'''[["backend/internal/casino", "backend/internal/casino"], ["backend-patch/casino.go", "backend/internal/server/routes/casino.go"], ["frontend/src/api/casino.ts", "frontend/src/api/casino.ts"], ["frontend/src/assets/casino", "frontend/src/assets/casino"], ["frontend/src/components/casino", "frontend/src/components/casino"], ["frontend/src/views/casino", "frontend/src/views/casino"], ["frontend/src/i18n/locales/zh/casino.ts", "frontend/src/i18n/locales/zh/casino.ts"], ["frontend/src/i18n/locales/en/casino.ts", "frontend/src/i18n/locales/en/casino.ts"]]''')
PATCHES = json.loads(r'''[{"file": "backend_router", "anchor": "PaymentWebhook, h.Admin.Payment, jwtAuth, adminAuth, auditLog, settingService, panelRateLimiter)\n\n\thandler.RegisterPageRoutes(v1, cfg.Pricing.DataDir, gin.HandlerFunc(jwtAuth), gin.HandlerFunc(adminAuth), settingService)", "block": "\tif err := routes.RegisterCasinoRoutes(v1, jwtAuth, adminAuth, cfg, redisClient); err != nil {\n\t\tlog.Printf(\"casino routes register failed: %v\", err)\n\t}\n\n", "_mode": "b"}, {"file": "fe_router", "anchor": "> import('@/views/user/RedeemView.vue'),\n    meta: {\n      requiresAuth: true,\n      requiresAdmin: false,\n      title: 'Redeem Code',\n      titleKey: 'redeem.title',\n      descriptionKey: 'redeem.description'\n    }\n  },", "block": "  {\n    path: '/casino',\n    name: 'CasinoLobby',\n    component: () => import('@/views/casino/CasinoLobbyView.vue'),\n    meta: {\n      requiresAuth: true,\n      requiresAdmin: false,\n      title: 'Casino',\n      titleKey: 'casino.title',\n      descriptionKey: 'casino.description'\n    }\n  },\n  {\n    path: '/casino/wheel',\n    name: 'CasinoWheel',\n    component: () => import('@/views/casino/CasinoWheelView.vue'),\n    meta: {\n      requiresAuth: true,\n      requiresAdmin: false,\n      title: 'Lucky Wheel',\n      titleKey: 'casino.wheel.title',\n      descriptionKey: 'casino.wheel.description'\n    }\n  },\n  {\n    path: '/casino/slots',\n    name: 'CasinoSlots',\n    component: () => import('@/views/casino/CasinoSlotsView.vue'),\n    meta: {\n      requiresAuth: true,\n      requiresAdmin: false,\n      title: 'Slot Machine',\n      titleKey: 'casino.slots.title',\n      descriptionKey: 'casino.slots.description'\n    }\n  },\n  {\n    path: '/casino/blackjack',\n    name: 'CasinoBlackjack',\n    component: () => import('@/views/casino/CasinoBlackjackView.vue'),\n    meta: {\n      requiresAuth: true,\n      requiresAdmin: false,\n      title: 'Blackjack',\n      titleKey: 'casino.blackjack.title',\n      descriptionKey: 'casino.blackjack.description'\n    }\n  },\n  {\n    path: '/casino/history',\n    name: 'CasinoHistory',\n    component: () => import('@/views/casino/CasinoHistoryView.vue'),\n    meta: {\n      requiresAuth: true,\n      requiresAdmin: false,\n      title: 'Game Records',\n      titleKey: 'casino.history.title',\n      descriptionKey: 'casino.history.description'\n    }\n  },\n  {\n    path: '/casino/sicbo',\n    name: 'CasinoSicbo',\n    component: () => import('@/views/casino/CasinoSicboView.vue'),\n    meta: {\n      requiresAuth: true,\n      requiresAdmin: false,\n      title: 'Sic Bo',\n      titleKey: 'casino.sicbo.title',\n      descriptionKey: 'casino.sicbo.description'\n    }\n  },\n  {\n    path: '/casino/baccarat',\n    name: 'CasinoBaccarat',\n    component: () => import('@/views/casino/CasinoBaccaratView.vue'),\n    meta: {\n      requiresAuth: true,\n      requiresAdmin: false,\n      title: 'Baccarat',\n      titleKey: 'casino.baccarat.title',\n      descriptionKey: 'casino.baccarat.description'\n    }\n  },\n  {\n    path: '/casino/scratch',\n    name: 'CasinoScratch',\n    component: () => import('@/views/casino/CasinoScratchView.vue'),\n    meta: {\n      requiresAuth: true,\n      requiresAdmin: false,\n      title: 'Scratch Card',\n      titleKey: 'casino.scratch.title',\n      descriptionKey: 'casino.scratch.description'\n    }\n  },\n\n", "_mode": "b"}, {"file": "sidebar", "anchor": "      'stroke-linejoin': 'round',\n          d: 'M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z'\n        })\n      ]\n    )\n}", "block": "// 老虎机风格内联图标（娱乐场入口）\nconst CasinoIcon = {\n  render: () =>\n    h(\n      'svg',\n      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },\n      [\n        h('rect', {\n          x: '3',\n          y: '4.5',\n          width: '18',\n          height: '13',\n          rx: '2.5',\n          'stroke-linecap': 'round',\n          'stroke-linejoin': 'round'\n        }),\n        h('path', {\n          'stroke-linecap': 'round',\n          'stroke-linejoin': 'round',\n          d: 'M7 17.5V20m10-2.5V20M7.5 8v6m4.5-6v6m4.5-6v6M6 20h12'\n        }),\n        h('path', {\n          'stroke-linecap': 'round',\n          'stroke-linejoin': 'round',\n          d: 'M17.5 1.75l1.2 2.05 2.3.35-1.7 1.65.45 2.3-2.25-1.15-2.25 1.15.45-2.3-1.7-1.65 2.3-.35z'\n        })\n      ]\n    )\n}\n\n", "_mode": "b"}, {"file": "sidebar", "anchor": "ayment },\n    { path: '/orders', label: t('nav.myOrders'), icon: OrderListIcon, hideInSimpleMode: true, featureFlag: flagPayment },\n    { path: '/redeem', label: t('nav.redeem'), icon: GiftIcon, hideInSimpleMode: true },", "block": "    ...(casinoEnabled.value ? [{ path: '/casino', label: t('nav.casino'), icon: CasinoIcon }] : []),\n\n", "_mode": "b"}, {"file": "api_index", "anchor": "Required, type LoginResponse } from './auth'\n\n// User APIs\nexport { keysAPI } from './keys'\nexport { usageAPI } from './usage'\nexport { userAPI } from './user'\nexport { redeemAPI, type RedeemHistoryItem } from './redeem'", "block": "export {\n  casinoAPI,\n  type CasinoMeta,\n  type CasinoMe,\n  type CasinoBetRow,\n  type CasinoLeaderboardRow,\n  type CasinoWheelSegment,\n  type CasinoSlotSymbol,\n  type CasinoBlackjackConfig,\n  type CasinoBlackjackGame\n} from './casino'\n\n", "_mode": "b"}, {"file": "i18n_zh", "anchor": " './landing'\nimport common from './common'\nimport dashboard from './dashboard'\nimport channelMonitorV2 from './channelMonitorV2'\nimport batchImage from './batchImage'\nimport admin from './admin'\nimport misc from './misc'", "block": "import casino from './casino'\n\n", "_mode": "b"}, {"file": "i18n_zh", "anchor": "  ...batchImage,\n  admin,\n  ...misc,", "block": "  ...casino,\n\n", "_mode": "b"}, {"file": "i18n_en", "anchor": " './landing'\nimport common from './common'\nimport dashboard from './dashboard'\nimport channelMonitorV2 from './channelMonitorV2'\nimport batchImage from './batchImage'\nimport admin from './admin'\nimport misc from './misc'", "block": "import casino from './casino'\n\n", "_mode": "b"}, {"file": "i18n_en", "anchor": "  ...batchImage,\n  admin,\n  ...misc,", "block": "  ...casino,\n\n", "_mode": "b"}, {"file": "common_zh", "anchor": "',\n      registerRequiredWarning: '请先阅读并同意最新条款后再注册。'\n    }\n  },\n\n  // Navigation\n  nav: {\n    dashboard: '仪表盘',\n    announcements: '公告',\n    apiKeys: 'API 密钥',\n    batchImage: '批量生图',\n    usage: '使用记录',\n    redeem: '兑换',", "block": "    casino: '娱乐场',\n\n", "_mode": "b"}, {"file": "common_en", "anchor": "before registering.'\n    }\n  },\n\n  // Navigation\n  nav: {\n    dashboard: 'Dashboard',\n    announcements: 'Announcements',\n    apiKeys: 'API Keys',\n    batchImage: 'Batch Images',\n    usage: 'Usage',\n    redeem: 'Redeem',", "block": "    casino: 'Casino',\n\n", "_mode": "b"}, {"file": "system_settings", "anchor": "import { adminAPI } from \"@/api\";", "block": "import CasinoAdminPanel from \"@/components/casino/CasinoAdminPanel.vue\";\n\n", "_mode": "b"}, {"file": "system_settings", "anchor": "  | \"email\"\n  | \"backup\"", "block": "  | \"casino\"\n\n", "_mode": "b"}, {"file": "system_settings", "anchor": "  { key: \"backup\" as SettingsTab, icon: \"database\" as const },", "block": "  { key: \"casino\" as SettingsTab, icon: \"trophy\" as const },\n\n", "_mode": "b"}, {"file": "system_settings", "anchor": "        <!-- Tab: Security — Admin API Key -->", "block": "        <div v-show=\"activeTab === 'casino'\" class=\"space-y-6\">\n          <CasinoAdminPanel />\n        </div>\n\n", "_mode": "b", "html": true}, {"file": "settings_i18n_zh", "anchor": "        backup: '数据备份',", "block": "        casino: '游戏管理',\n\n", "_mode": "b"}, {"file": "settings_i18n_en", "anchor": "        backup: 'Backup',", "block": "        casino: 'Casino',\n\n", "_mode": "b"}, {"file": "sidebar_admin", "anchor": "const adminSettingsStore = useAdminSettingsStore()", "block": "const casinoEnabled = ref(false)\nonMounted(async () => {\n  try {\n    const r = await fetch('/api/v1/casino/status')\n    const d = await r.json()\n    casinoEnabled.value = !!d.data?.enabled\n  } catch {\n    casinoEnabled.value = false\n  }\n})\n", "_mode": "b"}]''')
FILE_KEYS = {'backend_router': 'backend/internal/server/router.go', 'fe_router': 'frontend/src/router/index.ts', 'sidebar': 'frontend/src/components/layout/AppSidebar.vue', 'sidebar_admin': 'frontend/src/components/layout/AppSidebar.vue', 'api_index': 'frontend/src/api/index.ts', 'i18n_zh': 'frontend/src/i18n/locales/zh/index.ts', 'i18n_en': 'frontend/src/i18n/locales/en/index.ts', 'common_zh': 'frontend/src/i18n/locales/zh/common.ts', 'common_en': 'frontend/src/i18n/locales/en/common.ts', 'system_settings': 'frontend/src/views/admin/SettingsView.vue', 'settings_i18n_zh': 'frontend/src/i18n/locales/zh/admin/settings.ts', 'settings_i18n_en': 'frontend/src/i18n/locales/en/admin/settings.ts'}
for _p in PATCHES:
    _p['file'] = FILE_KEYS[_p['file']]


def read(p):
    with io.open(p, 'r', encoding='utf-8') as f:
        return f.read()

def write(p, s):
    with io.open(p, 'w', encoding='utf-8', newline='') as f:
        f.write(s)

def strip_blocks(s):
    # 循环剥离（多轮注入可能产生嵌套/双层标记），再清掉残留的孤立标记行
    # 兼容两种风格：// 注释（脚本内）与 <!-- --> 注释（Vue 模板内）
    pat = re.escape(BEGIN) + '.*?' + re.escape(END)
    hpat = re.escape(HTML_BEGIN) + '.*?' + re.escape(HTML_END)
    while re.search(pat, s, flags=re.S):
        s = re.sub(pat, '', s, flags=re.S)
    s = re.sub('^[ \t]*' + re.escape(BEGIN) + '[ \t]*\n?', '', s, flags=re.M)
    s = re.sub('^[ \t]*' + re.escape(END) + '[ \t]*\n?', '', s, flags=re.M)
    s = re.sub(r'\n{3,}', '\n\n', s)
    return s

def apply_patches(root):
    # 按文件分组：先整份 strip 掉旧标记块，再依次按锚点插入，最后一次性写回
    by_file = {}
    for item in PATCHES:
        by_file.setdefault(item['file'], []).append(item)
    for rel, items in by_file.items():
        path = os.path.join(root, rel.replace('/', os.sep))
        s = strip_blocks(read(path))
        for item in items:
            anchor = item['anchor']
            if s.count(anchor) != 1:
                raise SystemExit('anchor not unique: %s :: %r' % (rel, anchor[-80:]))
            b, e = (HTML_BEGIN, HTML_END) if item.get('html') else (BEGIN, END)
            block = b + '\n' + item['block'].rstrip('\n') + '\n' + e + '\n'
            s = s.replace(anchor, anchor + '\n' + block, 1)
            print('  [patch] ' + rel)
        write(path, s)

def copy_tree(root):
    for src_rel, dst_rel in COPIES:
        src = os.path.join(HERE, src_rel.replace('/', os.sep))
        dst = os.path.join(root, dst_rel.replace('/', os.sep))
        if os.path.isdir(src):
            if os.path.exists(dst):
                shutil.rmtree(dst)
            shutil.copytree(src, dst)
        else:
            os.makedirs(os.path.dirname(dst), exist_ok=True)
            shutil.copy2(src, dst)
        print('  [copy ] ' + dst_rel)

def ensure_go_mod(root):
    p = os.path.join(root, 'backend', 'go.mod')
    s = read(p)
    if 'github.com/jackc/pgx/v5' in s:
        return
    m = re.search(r'require \(\n', s)
    if not m:
        raise SystemExit('go.mod missing require block')
    s = s[:m.end()] + GO_REQUIRE + s[m.end():]
    write(p, s)
    print('  [gomod] + pgx/v5')

def revert(root):
    print('== revert sub2api tree ==')
    for _s, dst_rel in COPIES:
        dst = os.path.join(root, dst_rel.replace('/', os.sep))
        if os.path.isdir(dst):
            shutil.rmtree(dst); print('  [rm -d] ' + dst_rel)
        elif os.path.exists(dst):
            os.remove(dst); print('  [rm  ] ' + dst_rel)
    seen = set()
    for item in PATCHES:
        if item['file'] in seen:
            continue
        seen.add(item['file'])
        path = os.path.join(root, item['file'])
        if os.path.exists(path):
            write(path, strip_blocks(read(path)))
            print('  [clean] ' + item['file'])
    p = os.path.join(root, 'backend', 'go.mod')
    s = read(p)
    if GO_REQUIRE in s:
        write(p, s.replace(GO_REQUIRE, ''))
        print('  [gomod] - pgx/v5')
    print('tree restored.')

def build(root):
    fe = os.path.join(root, 'frontend')
    be = os.path.join(root, 'backend')
    if not os.path.isdir(os.path.join(fe, 'node_modules')):
        print('== pnpm install ==')
        subprocess.check_call('pnpm install', cwd=fe, shell=True)
    print('== pnpm build ==')
    subprocess.check_call('pnpm build', cwd=fe, shell=True)
    print('== go mod tidy ==')
    subprocess.check_call('go mod tidy', cwd=be, shell=True)
    print('== go build -tags embed ==')
    subprocess.check_call('go build -tags embed -o bin/server-embed.exe ./cmd/server', cwd=be, shell=True)
    print('build ok: backend/bin/server-embed.exe')

def run(root):
    subprocess.call('taskkill /IM server-embed.exe /F 2>nul & taskkill /IM server.exe /F 2>nul', shell=True)
    time.sleep(1)
    be = os.path.join(root, 'backend')
    env = dict(os.environ, DATA_DIR='./data')
    exe = os.path.join(be, 'bin', 'server-embed.exe')
    log = open(os.path.join(be, 'server-embed.log'), 'ab')
    subprocess.Popen([exe], cwd=be, env=env,
                     stdout=log, stderr=subprocess.STDOUT,
                     creationflags=subprocess.CREATE_NEW_PROCESS_GROUP)
    import urllib.request
    for i in range(30):
        time.sleep(2)
        try:
            with urllib.request.urlopen('http://127.0.0.1:8080/health', timeout=3) as r:
                if r.status == 200:
                    print('backend up: http://localhost:8080')
                    return
        except Exception:
            pass
    print('health check timeout, see backend/server-embed.log')

def main():
    argv = sys.argv[1:]
    root = DEFAULT_ROOT
    if '--root' in argv:
        root = argv[argv.index('--root') + 1]
    if '--revert' in argv:
        revert(root); return
    print('== copy files =='); copy_tree(root)
    print('== apply patches ==')
    apply_patches(root)
    ensure_go_mod(root)
    if '--skip-build' not in argv:
        build(root)
    if '--run' in argv:
        run(root)
    print('done.')

main()
