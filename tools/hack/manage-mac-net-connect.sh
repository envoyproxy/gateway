#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-setup}"  # setup | cleanup
DOCKER_MAC_NET_CONNECT=${DOCKER_MAC_NET_CONNECT:-}
HOMEBREW_GOPROXY=${HOMEBREW_GOPROXY:-}
FLAG_FILE="/tmp/.docker-mac-net-connect-started"

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
        echo "docker-mac-net-connect is already installed and running."
        return
    fi

    read -rp "Install and start docker-mac-net-connect? [y/N]: " input
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
        touch "$FLAG_FILE"
        sleep 5
    fi
}

cleanup() {
    if is_macos && [ -f "$FLAG_FILE" ]; then
        sudo brew services stop chipmk/tap/docker-mac-net-connect || true
        rm -f "$FLAG_FILE"
    fi
}

case "$MODE" in
    setup) setup ;;
    cleanup) cleanup ;;
    *) echo "Usage: $0 [setup|cleanup]"; exit 1 ;;
esac
