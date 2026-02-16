#!/usr/bin/env bash

GITHUB_REPO_URL="https://github.com/envoyproxy/gateway"

LATEST_RELEASE="$GITHUB_REPO_URL/releases/tag/latest"

LATEST_TAG="$GITHUB_REPO_URL/tree/latest"

RELEASE_HTTP_STATUS_CODE=$(curl -i -m 10 -o /dev/null -w %\{http_code\} $LATEST_RELEASE)

if [ "$RELEASE_HTTP_STATUS_CODE" != "200" ]; then
    echo "\033[0;31mLatest release of eg is published with failures.\033[0m"
    exit 1
else 
    echo "\033[36mLatest release of eg is published normally\033[0m"
fi

TAG_HTTP_STATUS_CODE=$(curl -i -m 10 -o /dev/null -w %\{http_code\} $LATEST_TAG)

if [ "$TAG_HTTP_STATUS_CODE" != "200" ]; then
    echo "\033[0;31mLatest tag of eg is published with failures.\033[0m"
    exit 1
else 
    echo "\033[36mLatest tag of eg is published normally\033[0m"
fi

LATEST_INTALL_MANIFEST="$GITHUB_REPO_URL/releases/download/latest/install.yaml"

INSTALL_HTTP_STATUS_CODE=$(curl -i -m 10 -o /dev/null -w %\{http_code\} $LATEST_INTALL_MANIFEST)

if [ "$INSTALL_HTTP_STATUS_CODE" = "404" ]; then
    echo "\033[0;31mLatest install.yaml of eg is published with failures.\033[0m"
    exit 1
else 
    echo "\033[36mLatest install.yaml of eg is published normally\033[0m"
fi

LATEST_QUICKSTART_MANIFEST="$GITHUB_REPO_URL/releases/download/latest/quickstart.yaml"

QUICKSTART_HTTP_STATUS_CODE=$(curl -i -m 10 -o /dev/null -w %\{http_code\} $LATEST_QUICKSTART_MANIFEST)

if [ "$QUICKSTART_HTTP_STATUS_CODE" = "404" ]; then
    echo "\033[0;31mLatest quickstart.yaml of eg is published with failures.\033[0m"
    exit 1
else 
    echo "\033[36mLatest quickstart.yaml of eg is published normally\033[0m"
fi
