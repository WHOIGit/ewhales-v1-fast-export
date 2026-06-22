#!/bin/bash
set -euo pipefail

# extract export base dir
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BASE="$(basename "$SCRIPT_DIR")"
STAMP="${BASE%%_*}"

# extract YMD parts
YEAR="${STAMP:2:2}"
MONTH="${STAMP:4:2}"
DAY="${STAMP:6:2}"

# local paths
HOME_DIR="/home/finn.wimberly"
CSV_DIR="$SCRIPT_DIR/../csv_files"
PKL_DIR="$SCRIPT_DIR/../pkl_files"

# build folder structure for drive
REMOTE_ROOT="whaling_logbooks_GDrive:data/exports"
REMOTE="${REMOTE_ROOT}/${MONTH}-${DAY}-${YEAR}_export"

# sync with rclone
"$HOME_DIR/bin/rclone" copy "$CSV_DIR" "$REMOTE/csv_files" \
  --filter "- .ipynb_checkpoints/**" \
  --filter "+ Tier*.csv" \
  --filter "- *"

"$HOME_DIR/bin/rclone" copy "$PKL_DIR" "$REMOTE/pkl_files" \
  --filter "- .ipynb_checkpoints/**" \
  --filter "+ Tier*.pkl" \
  --filter "- *"