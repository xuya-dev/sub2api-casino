#!/usr/bin/env bash
# ============================================================================
# Sub2API 娱乐场 · 一键注入部署（Linux/macOS，零外部依赖）
#
# 依赖: bash, sed, awk, go, node, pnpm（全部系统自带或 sub2api 必需）
# 用法:
#   bash install.sh [sub2api源码根目录] [选项]
#   选项:
#     --revert      还原源码树（删除注入文件与补丁）
#     --skip-build  只注入源码，不构建
#     --run         注入构建后启动服务（默认行为）
#
# 远程一条命令:
#   curl -fsSL https://raw.githubusercontent.com/your-org/sub2api-casino/main/install.sh | bash -s -- /path/to/sub2api
# ============================================================================
set -euo pipefail

SUB2API_ROOT="${1:-${SUB2API_ROOT:-/path/to/sub2api}}"
MODE="run"
if [ "${2:-}" = "--revert" ]; then
  MODE="revert"
elif [ "${2:-}" = "--skip-build" ]; then
  MODE="skip"
fi

[ -f "$SUB2API_ROOT/backend/cmd/server/main.go" ] || {
  echo "[错误] $SUB2API_ROOT 不是 sub2api 源码根目录"
  exit 2
}

CASINO_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$SUB2API_ROOT/backend"
FRONTEND_DIR="$SUB2API_ROOT/frontend"
PATCHES_TXT="$CASINO_DIR/inject/patches-extracted.txt"

# ---- 依赖检查 ----
need() { command -v "$1" >/dev/null 2>&1 || { echo "[错误] 缺少 $1（$2）"; exit 1; }; }
need bash "系统自带"
need sed "系统自带"
need awk "系统自带"
need go "https://go.dev/"
need pnpm "npm i -g pnpm"

if [ "$MODE" = "revert" ]; then
  revert_all
  exit 0
fi

# ---- 注入 ----
inject_files() {
  echo
  echo "[1/5] 拷贝源码到 sub2api 树..."
  mkdir -p "$BACKEND_DIR/internal/casino"
  cp -r "$CASINO_DIR/inject/backend/internal/casino/." "$BACKEND_DIR/internal/casino/"
  cp "$CASINO_DIR/inject/backend-patch/casino.go" "$BACKEND_DIR/internal/server/routes/casino.go"

  mkdir -p "$FRONTEND_DIR/src/views/casino"
  cp -r "$CASINO_DIR/inject/frontend/src/views/casino/." "$FRONTEND_DIR/src/views/casino/"

  mkdir -p "$FRONTEND_DIR/src/components/casino"
  cp -r "$CASINO_DIR/inject/frontend/src/components/casino/." "$FRONTEND_DIR/src/components/casino/"

  mkdir -p "$FRONTEND_DIR/src/assets/casino"
  cp -r "$CASINO_DIR/inject/frontend/src/assets/casino/." "$FRONTEND_DIR/src/assets/casino/"

  cp "$CASINO_DIR/inject/frontend/src/api/casino.ts" "$FRONTEND_DIR/src/api/casino.ts"
  cp "$CASINO_DIR/inject/frontend/src/i18n/locales/zh/casino.ts" "$FRONTEND_DIR/src/i18n/locales/zh/casino.ts"
  cp "$CASINO_DIR/inject/frontend/src/i18n/locales/en/casino.ts" "$FRONTEND_DIR/src/i18n/locales/en/casino.ts"
  echo "    拷贝完成"
}

# ---- 补丁应用（纯 awk，无 jq 依赖） ----
# patches-extracted.txt 格式:
#   ########## name #N ##########
#   --- ANCHOR(前文末220字) ---
#   <anchor 多行>
#   --- BLOCK ---
#   <block 多行>
apply_patches() {
  echo "[2/5] 打标记补丁..."
  awk '
    /^##########/ {
      name = $2; n = $3
      sub(/#/, "", n); sub(/#/, "", n)
      mode = ""; in_anchor = 0; in_block = 0
      next
    }
    /^--- ANCHOR/ { in_anchor = 1; in_block = 0; anchor = ""; next }
    /^--- BLOCK/ { in_anchor = 0; in_block = 1; block = ""; next }
    /^##########/ { next }
    in_anchor { anchor = anchor $0 "\n" }
    in_block { block = block $0 "\n" }
    /^##########/ && anchor != "" {
      # 输出：name \t anchor \t block（用 \x01 分隔避免冲突）
      printf "%s\x01%s\x01%s\n", name, anchor, block
      anchor = ""; block = ""
    }
  ' "$PATCHES_TXT" > /tmp/casino-patches.tsv

  # 逐补丁应用
  while IFS=$'\x01' read -r name anchor block; do
    [ -n "$name" ] || continue
    file_key=$(file_key_for "$name")
    [ -n "$file_key" ] || continue
    local_file="$SUB2API_ROOT/$file_key"
    [ -f "$local_file" ] || continue

    # 剥离旧标记块（幂等）
    sed -i 's|// ===CASINO:BEGIN===.*// ===CASINO:END===||g' "$local_file"
    sed -i 's|// ===CASINO:BEGIN===||g; s|// ===CASINO:END===||g' "$local_file"
    sed -i 's|<!-- ===CASINO:BEGIN=== -->.*<!-- ===CASINO:END=== -->||g' "$local_file"
    sed -i 's|<!-- ===CASINO:BEGIN=== -->||g; s|<!-- ===CASINO:END=== -->||g' "$local_file"

    # 在锚点后插入 block（用 python 处理多行，fallback 到 awk）
    if command -v python3 >/dev/null 2>&1; then
      python3 -c "
import sys
with open('$local_file', 'r', encoding='utf-8') as f:
    s = f.read()
anchor = '''$anchor'''
block = '''$block'''
if anchor in s:
    s = s.replace(anchor, anchor + block, 1)
    with open('$local_file', 'w', encoding='utf-8') as f:
        f.write(s)
    print('ok')
else:
    sys.exit(1)
" 2>/dev/null && echo "  [patch] $file_key ($name)" || echo "  [patch] $file_key ($name) [锚点未找到，跳过]"
    else
      # 无 Python 时用 awk（仅处理单行锚点，多行锚点会跳过并警告）
      if [ $(echo "$anchor" | wc -l) -gt 2 ]; then
        echo "  [patch] $file_key ($name) [多行锚点需 Python，跳过]"
      else
        awk -v a="$anchor" -v b="$block" '{
          if (index($0, a)) { print $0 b; next }
          print
        }' "$local_file" > "$local_file.tmp" && mv "$local_file.tmp" "$local_file"
        echo "  [patch] $file_key ($name)"
      fi
    fi
  done < /tmp/casino-patches.tsv
  rm -f /tmp/casino-patches.tsv
  echo "    补丁完成"
}

file_key_for() {
  case "$1" in
    backend_router) echo "backend/internal/server/router.go" ;;
    fe_router) echo "frontend/src/router/index.ts" ;;
    sidebar) echo "frontend/src/components/layout/AppSidebar.vue" ;;
    sidebar_admin) echo "frontend/src/components/layout/AppSidebar.vue" ;;
    api_index) echo "frontend/src/api/index.ts" ;;
    i18n_zh) echo "frontend/src/i18n/locales/zh/index.ts" ;;
    i18n_en) echo "frontend/src/i18n/locales/en/index.ts" ;;
    common_zh) echo "frontend/src/i18n/locales/zh/common.ts" ;;
    common_en) echo "frontend/src/i18n/locales/en/common.ts" ;;
    system_settings) echo "frontend/src/views/admin/SettingsView.vue" ;;
    settings_i18n_zh) echo "frontend/src/i18n/locales/zh/admin/settings.ts" ;;
    settings_i18n_en) echo "frontend/src/i18n/locales/en/admin/settings.ts" ;;
    *) echo "" ;;
  esac
}

# ---- 构建 ----
build() {
  echo "[3/5] 构建前端（pnpm install + build）..."
  cd "$FRONTEND_DIR"
  [ -d node_modules ] || pnpm install
  pnpm build

  echo "[4/5] 构建后端（go mod tidy + go build -tags embed）..."
  cd "$BACKEND_DIR"
  go mod tidy
  go build -tags embed -o bin/server-embed ./cmd/server
  echo "    构建完成: backend/bin/server-embed"
}

# ---- 启动 ----
run() {
  echo "[5/5] 启动服务..."
  pkill -f server-embed 2>/dev/null || true
  pkill -f "cmd/server" 2>/dev/null || true
  cd "$BACKEND_DIR"
  export DATA_DIR="$BACKEND_DIR"
  nohup ./bin/server-embed > server-embed.log 2>&1 &
  sleep 5
  if curl -s http://localhost:8080/api/v1/casino/status 2>/dev/null | grep -q "enabled"; then
    echo
    echo "✓ 娱乐场已启动: http://localhost:8080"
    echo "  登录后点击侧边栏「娱乐场」即可游玩"
    echo "  卸载/还原: bash install.sh $SUB2API_ROOT --revert"
  else
    echo
    echo "⚠ 启动超时，查看日志: backend/server-embed.log"
  fi
}

# ---- 还原 ----
revert_all() {
  echo
  echo "还原 sub2api 源码树..."
  rm -rf "$BACKEND_DIR/internal/casino"
  rm -f "$BACKEND_DIR/internal/server/routes/casino.go"
  rm -rf "$FRONTEND_DIR/src/views/casino"
  rm -rf "$FRONTEND_DIR/src/components/casino"
  rm -rf "$FRONTEND_DIR/src/assets/casino"
  rm -f "$FRONTEND_DIR/src/api/casino.ts"
  rm -f "$FRONTEND_DIR/src/i18n/locales/zh/casino.ts"
  rm -f "$FRONTEND_DIR/src/i18n/locales/en/casino.ts"

  # 剥离标记补丁
  for rel in backend/internal/server/router.go \
             frontend/src/router/index.ts \
             frontend/src/components/layout/AppSidebar.vue \
             frontend/src/api/index.ts \
             frontend/src/i18n/locales/zh/index.ts \
             frontend/src/i18n/locales/en/index.ts \
             frontend/src/i18n/locales/zh/common.ts \
             frontend/src/i18n/locales/en/common.ts \
             frontend/src/views/admin/SettingsView.vue \
             frontend/src/i18n/locales/zh/admin/settings.ts \
             frontend/src/i18n/locales/en/admin/settings.ts; do
    local_file="$SUB2API_ROOT/$rel"
    [ -f "$local_file" ] || continue
    sed -i 's|// ===CASINO:BEGIN===.*// ===CASINO:END===||g' "$local_file"
    sed -i 's|// ===CASINO:BEGIN===||g; s|// ===CASINO:END===||g' "$local_file"
    sed -i 's|<!-- ===CASINO:BEGIN=== -->.*<!-- ===CASINO:END=== -->||g' "$local_file"
    sed -i 's|<!-- ===CASINO:BEGIN=== -->||g; s|<!-- ===CASINO:END=== -->||g' "$local_file"
  done
  echo "✓ 还原完成（重新构建主站即可）"
}

# ---- 主流程 ----
inject_files
apply_patches
if [ "$MODE" != "skip-build" ]; then
  build
fi
if [ "$MODE" = "run" ]; then
  run
fi

