#!/bin/bash
set -euo pipefail

# Set up directories
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
META_FIGS_DIR="$SCRIPT_DIR/../meta_figs"
CSV_DIR="$SCRIPT_DIR/../csv_files"
HOME_DIR="/home/finn.wimberly"
if command -v rclone >/dev/null 2>&1; then
    RCLONE="rclone"
elif [ -x "$HOME/bin/rclone" ]; then
    RCLONE="$HOME/bin/rclone"
else
    echo "Error: rclone not found in PATH or at $HOME/bin/rclone" >&2
    exit 1
fi


# Sync all whaling figures (exclude .ipynb_checkpoints)
"$RCLONE" sync \
  "$META_FIGS_DIR/" \
  whaling_logbooks_GDrive:project_status/figures \
  --exclude "**/.ipynb_checkpoints/**"

# Copy Tier4logbooks_meta.csv
"$RCLONE" copy \
  "$CSV_DIR/Tier4logbooks_meta.csv" \
  whaling_logbooks_GDrive:project_status/