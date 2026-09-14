#!/usr/bin/env bash
DIST="$HOME/services/sub2api-casino/backend/internal/web/dist"
echo "=== css files ==="
ls "$DIST/assets/" | grep "\.css" | head -5
echo "=== dark rule in css ==="
grep -l "dark .crm-card" "$DIST/assets/"*.css 2>/dev/null | head -2 || echo "NOT FOUND in dist css"
echo "=== index.html 引用 ==="
grep -o "assets/index-[A-Za-z0-9_-]*\.css" "$DIST/index.html"
grep -o "index-[A-Za-z0-9_-]*\.js" "$DIST/index.html" | head -2
echo "=== 服务端返回 index 引用 ==="
curl -s http://localhost:8080/ | grep -o "index-[A-Za-z0-9_-]*\.js" | head -2
