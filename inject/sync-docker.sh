#!/usr/bin/env bash
set -e
SRC="/mnt/c/OpenSource/Sub2api/plugin/casino/inject/frontend/src"
DST="$HOME/services/sub2api-casino/frontend/src"
cp "$SRC/components/casino/CasinoResultModal.vue" "$DST/components/casino/"
cp "$SRC/assets/casino/scratch/coin-burst.png" "$DST/assets/casino/scratch/"
cp "$SRC/assets/casino/scratch/banner-bg.png" "$DST/assets/casino/scratch/"
cp "$SRC/i18n/locales/zh/casino.ts" "$DST/i18n/locales/zh/casino.ts"
cp "$SRC/i18n/locales/en/casino.ts" "$DST/i18n/locales/en/casino.ts"
echo "synced"
grep -c "coin-burst" "$DST/components/casino/CasinoResultModal.vue"
cd "$HOME/services/sub2api-casino"
docker compose build --no-cache sub2api 2>&1 | tail -2
docker compose up -d --force-recreate 2>&1 | tail -2
