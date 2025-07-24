#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-setup}"  # setup | cleanup
DOCKER_MAC_NET_CONNECT=${DOCKER_MAC_NET_CONNECT:-}
HOMEBREW_GOPROXY=${HOMEBREW_GOPROXY:-}

is_macos() {
    [[ "$(uname -s)" == "Darwin" ]]
}

is_installed() {
    [[ "$(brew list --formula | grep -Fx "docker-mac-net-connect")" == "docker-mac-net-connect" ]]
}

is_running() {
    brew services info docker-mac-net-connect --json 2>/dev/null | \
        jq -e '.[] | .user != null' >/dev/null
}

setup() {
    if ! is_macos || [[ "$DOCKER_MAC_NET_CONNECT" == "false" ]]; then
        return
    fi

    if is_installed && is_running; then
        echo "Docker Mac Net Connect is already installed and running."
        return
    fi

    echo "Docker Mac Net Connect is recommended on macOS to ensure Docker networking works properly."
    read -rp "Install and start Docker Mac Net Connect? [y/N]: " input
    case "$(echo "$input" | tr '[:upper:]' '[:lower:]')" in
        y|yes) ;;
        *) return ;;
    esac

    if ! is_installed; then
        [ -n "$HOMEBREW_GOPROXY" ] && export HOMEBREW_GOPROXY="$HOMEBREW_GOPROXY"
        brew install chipmk/tap/docker-mac-net-connect
    fi

    if ! is_running; then
        sudo brew services start chipmk/tap/docker-mac-net-connect
        sleep 5
    fi
}

case "$MODE" in
    setup) setup ;;
    cleanup) cleanup ;;
    *) echo "Usage: $0 [setup]"; exit 1 ;;
esac
