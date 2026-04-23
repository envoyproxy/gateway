#!/usr/bin/env bash
set -euo pipefail

log() { echo "==> $*"; }

log "Initial disk usage:"
df -h || true

# Remove large, unused language/tool runtimes
# To figure out what to remove. We can SSH into the GitHub Actions runner using: https://github.com/mxschmitt/action-tmate and then run: df -h
TO_DELETE=(
  /usr/local/lib/android
  /usr/share/dotnet
  /opt/ghc
  /usr/local/.ghcup
  /usr/share/swift
)

for path in "${TO_DELETE[@]}"; do
  if [ -d "$path" ]; then
    log "Removing $path"
    sudo rm -rf "$path"
  fi
done

log "Removing large packages..."
EXTRA_PKGS="azure-cli google-chrome-stable firefox powershell mono-devel libgl1-mesa-dri google-cloud-sdk google-cloud-cli"

sudo apt-get remove -y "$EXTRA_PKGS" --fix-missing || true
sudo apt-get autoremove -y || true
sudo apt-get clean || true

# Swap removal
if [ -f /mnt/swapfile ]; then
  log "Disabling and removing swapfile"
  sudo swapoff -a || true
  sudo rm -f /mnt/swapfile || true
fi

log "Final disk usage:"
df -h || true

log "Completed disk space reclamation."
