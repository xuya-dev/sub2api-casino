# -*- coding: utf-8 -*-
"""从 patches-extracted.txt 生成自包含的 inject.py（补丁数据 JSON 内嵌）。"""
import io, json, re

src = io.open(r'C:/OpenSource/Sub2api/plugin/casino/inject/patches-extracted.txt', encoding='utf-8').read()
patches = []
cur = None
for line in src.splitlines(True):
    m = re.match(r'########## (\S+) #(\d+) ##########', line)
    if m:
        cur = {'file': m.group(1), 'anchor': '', 'block': '', '_mode': ''}
        patches.append(cur)
        continue
    if cur is None:
        continue
    if line.startswith('--- ANCHOR'):
        cur['_mode'] = 'a'; continue
    if line.startswith('--- BLOCK'):
        cur['_mode'] = 'b'; continue
    if cur['_mode'] == 'a':
        cur['anchor'] += line
    elif cur['_mode'] == 'b':
        cur['block'] += line

for p in patches:
    p['anchor'] = p['anchor'].rstrip('\n')
# i18n index 的第二个补丁：锚点截取自已打补丁的文件（混入标记文本），改用纯原文短锚
for p in patches:
    if p['file'] in ('i18n_zh', 'i18n_en') and '...casino,' in p['block']:
        p['anchor'] = '  ...batchImage,\n  admin,\n  ...misc,'
for p in patches:
    if p['file'] == 'system_settings' and 'v-show' in p['block']:
        p['html'] = True

FILE_KEYS = {
    'backend_router': 'backend/internal/server/router.go',
    'fe_router': 'frontend/src/router/index.ts',
    'sidebar': 'frontend/src/components/layout/AppSidebar.vue',
    'sidebar_admin': 'frontend/src/components/layout/AppSidebar.vue',
    'api_index': 'frontend/src/api/index.ts',
    'i18n_zh': 'frontend/src/i18n/locales/zh/index.ts',
    'i18n_en': 'frontend/src/i18n/locales/en/index.ts',
    'common_zh': 'frontend/src/i18n/locales/zh/common.ts',
    'common_en': 'frontend/src/i18n/locales/en/common.ts',
    'system_settings': 'frontend/src/views/admin/SettingsView.vue',
    'settings_i18n_zh': 'frontend/src/i18n/locales/zh/admin/settings.ts',
    'settings_i18n_en': 'frontend/src/i18n/locales/en/admin/settings.ts',
}

COPIES = [
    ['backend/internal/casino', 'backend/internal/casino'],
    ['backend-patch/casino.go', 'backend/internal/server/routes/casino.go'],
    ['frontend/src/api/casino.ts', 'frontend/src/api/casino.ts'],
    ['frontend/src/assets/casino', 'frontend/src/assets/casino'],
    ['frontend/src/components/casino', 'frontend/src/components/casino'],
    ['frontend/src/views/casino', 'frontend/src/views/casino'],
    ['frontend/src/i18n/locales/zh/casino.ts', 'frontend/src/i18n/locales/zh/casino.ts'],
    ['frontend/src/i18n/locales/en/casino.ts', 'frontend/src/i18n/locales/en/casino.ts'],
]

head = (
    "# -*- coding: utf-8 -*-\n"
    "# Sub2API 娱乐场 - 一键源码注入脚本（本文件由 gen_inject.py 生成，勿手改）\n"
    "# 用法:\n"
    "#   python inject.py              # 注入(拷贝+补丁) + 构建前端/后端\n"
    "#   python inject.py --run        # 注入构建后重启后端(端口8080)\n"
    "#   python inject.py --revert     # 还原 sub2api 源码树到干净状态\n"
    "#   python inject.py --skip-build # 只注入不构建\n"
    "import io, json, os, re, shutil, subprocess, sys, time\n\n"
    "HERE = os.path.dirname(os.path.abspath(__file__))\n"
    "DEFAULT_ROOT = r'C:/OpenSource/Sub2api/sub2api'\n"
    "BEGIN, END = '// ===CASINO:BEGIN===', '// ===CASINO:END==='\n"
    "HTML_BEGIN, HTML_END = '<!-- ===CASINO:BEGIN=== -->', '<!-- ===CASINO:END=== -->'\n"
    "GO_REQUIRE = '\\tgithub.com/jackc/pgx/v5 v5.7.4\\n'\n"
    "COPIES = json.loads(r'''__COPIES__''')\n"
    "PATCHES = json.loads(r'''__PATCHES__''')\n"
    "FILE_KEYS = " + repr(FILE_KEYS) + "\n"
    "for _p in PATCHES:\n"
    "    _p['file'] = FILE_KEYS[_p['file']]\n\n"
)

body = '''
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
    s = re.sub('^[ \\t]*' + re.escape(BEGIN) + '[ \\t]*\\n?', '', s, flags=re.M)
    s = re.sub('^[ \\t]*' + re.escape(END) + '[ \\t]*\\n?', '', s, flags=re.M)
    s = re.sub(r'\\n{3,}', '\\n\\n', s)
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
            block = b + '\\n' + item['block'].rstrip('\\n') + '\\n' + e + '\\n'
            s = s.replace(anchor, anchor + '\\n' + block, 1)
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
    m = re.search(r'require \\(\\n', s)
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
'''

out = head + body
out = out.replace('__COPIES__', json.dumps(COPIES)).replace('__PATCHES__', json.dumps(patches, ensure_ascii=False))
io.open(r'C:/OpenSource/Sub2api/plugin/casino/inject/inject.py', 'w', encoding='utf-8').write(out)
print('inject.py generated,', len(out), 'chars,', len(patches), 'patches')
