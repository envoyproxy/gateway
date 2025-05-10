#!/usr/bin/env bash

source "test/standalone/suite.sh"

pre_run() {
    start_envoy_gateway "8888:8888"
    expect_eq $? 0 "failed to start envoy gateway"

    start_backend "local-server" "local-server.local" "3000"
    expect_eq $? 0 "failed to start backend"
}

run() {
    # copy config triggers a resources update
    copy_config "examples/standalone/quickstart-containers.yaml"

    # send requests until converge
    send_http_request "0.0.0.0:8888" "www.example.com"
    expect_eq $? 0 "fail to get respond from backend"
}

post_run() {
    stop_backend "local-server"
    stop_envoy_gateway
}
