#!/usr/bin/env bash

readonly DOCKER_NETWORK="envoy-gateway-standalone-test"

readonly EG_NAME="envoy-gateway-standalone"
readonly EG_LOCAL_WORK_DIR="/tmp/envoy-gateway-test" # Work dir on local host
readonly EG_LOCAL_CONFIG_DIR="${EG_LOCAL_WORK_DIR}/config"
readonly EG_WORK_DIR="/tmp/envoy-gateway" # Work dir on mapped volume
readonly EG_CONFIG_DIR="${EG_WORK_DIR}/config"

start_envoy_gateway() {
    if [ "$#" -lt 1 ]; then
        echo "Error: At least one port should be published."
        return 1
    fi

    local ports=("$@")
    local publish_ports=""
    for port in "${ports[@]}"; do
        publish_ports+="--publish $port "
    done

    docker run \
        --name $EG_NAME \
        --network $DOCKER_NETWORK \
        $publish_ports \
        --volume $EG_LOCAL_WORK_DIR:$EG_WORK_DIR \
        --detach \
        ${EG_IMAGE_NAME}:${EG_IMAGE_TAG} \
        server --config-path $EG_WORK_DIR/standalone.yaml
    return $?
}

stop_envoy_gateway() {
    docker stop $EG_NAME > /dev/null
    docker rm $EG_NAME > /dev/null
}

start_backend() {
    local name="$1"
    local hostname="$2"
    local port="$3"

    docker run \
        --name $name \
        --hostname $hostname \
        --network $DOCKER_NETWORK \
        --detach \
        python:3 \
        python3 -m http.server $port
    return $?
}

stop_backend() {
    local name="$1"

    docker stop $name > /dev/null
    docker rm $name > /dev/null
}

send_http_request() {
    local addr="$1"
    local hostname="$2"
    local retry="${3:-10}"
    local retry_delay="${4:-3}"
    local timeout="${5:-60}"

    curl -I \
        --connect-timeout $timeout \
        --retry $retry \
        --retry-delay $retry_delay \
        --retry-max-time $timeout \
        --retry-all-errors \
        --header "Host: ${hostname}" \
        http://$addr/
    return $?
}

copy_config() {
    local path="$1"

    cp $1 $EG_LOCAL_CONFIG_DIR
}

expect_eq() {
    local got="$1"
    local want="$2"
    local msg="$3"

    if [[ "$want" != "$got" ]]; then
        echo "[ERROR]: $msg" >&2
        echo "  Expected: '$want'" >&2
        echo "  Got:      '$got'"  >&2
        return 1
    fi

    return 0
}
