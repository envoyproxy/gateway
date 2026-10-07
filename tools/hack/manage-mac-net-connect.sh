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
    pgrep -f '(^|/)docker-mac-net-connect([[:space:]]|$)' >/dev/null 2>&1
}

setup() {
    if ! is_macos || [[ "$DOCKER_MAC_NET_CONNECT" == "false" ]]; then
        return
    fi

    if ! command -v brew >/dev/null 2>&1; then
        echo "Homebrew is required to install Docker Mac Net Connect." >&2
        return 1
    fi

    if is_installed && is_running; then
        echo "Docker Mac Net Connect is already installed and running."
        return
    fi

    echo "Docker Mac Net Connect is recommended on macOS to ensure Docker networking works properly."
    read -rp "Install and start Docker Mac Net Connect? [y/N]: " input
    case "$(echo "$input" | tr '[:upper:]' '[:lower:]')" in
        y|yes) ;;
        *)
            echo "Docker Mac Net Connect is required; set DOCKER_MAC_NET_CONNECT=false to skip setup." >&2
            return 1
            ;;
    esac

    if ! is_installed; then
        [ -n "$HOMEBREW_GOPROXY" ] && export HOMEBREW_GOPROXY="$HOMEBREW_GOPROXY"
        brew install chipmk/tap/docker-mac-net-connect
    fi

    if ! is_running; then
        sudo brew services start chipmk/tap/docker-mac-net-connect
    fi

    for ((attempt = 0; attempt < 10; attempt++)); do
        if is_running; then
            return
        fi
        sleep 1
    done

    echo "Docker Mac Net Connect did not start." >&2
    return 1
}

case "$MODE" in
    setup) setup ;;
    *) echo "Usage: $0 [setup]"; exit 1 ;;
esac
