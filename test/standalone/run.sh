#!/usr/bin/env bash

set -o nounset

readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m' # No Color

readonly TEST_DIR="test/standalone/tests"

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

source "test/standalone/suite.sh"

setup_env() {
    mkdir -p $EG_LOCAL_CONFIG_DIR
    chmod -R 777 $EG_LOCAL_WORK_DIR

    # Write envoy-gateway standalone config.
    cat > $EG_LOCAL_WORK_DIR/standalone.yaml <<EOF
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
gateway:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
provider:
  type: Custom
  custom:
    resource:
      type: File
      file:
        paths: ["${EG_CONFIG_DIR}"]
    infrastructure:
      type: Host
      host: {}
logging:
  level:
    default: info
extensionApis:
  enableBackend: true
EOF

    # Pull the test image in advance.
    docker pull python:3

    docker network create $DOCKER_NETWORK

    # Create TLS certificates locally.
    docker run --rm \
        --volume $EG_LOCAL_WORK_DIR:$EG_WORK_DIR \
        ${EG_IMAGE_NAME}:${EG_IMAGE_TAG} \
        certgen --local
}

cleanup_env() {
    docker network rm $DOCKER_NETWORK > /dev/null

    rm -rf $EG_LOCAL_WORK_DIR

    # Clean global env set during test.
    unset EG_IMAGE_NAME EG_IMAGE_TAG
}

execute_test() {
    local test_file="$1"
    local test_name=$(basename "$test_file")

    echo -e "\n${YELLOW}=== Running test: ${test_name} ===${NC}"

    source "$test_file"
    set -x

    if declare -f pre_run > /dev/null; then
        echo "Executing pre_run..."
        if ! pre_run; then
            echo -e "${YELLOW}pre_run completed with errors${NC}"
        fi
    fi

    echo "Executing run..."
    local start_time=$(date +%s)
    run
    local run_exit_code=$?
    local end_time=$(date +%s)
    local elapsed_time=$(echo "$end_time - $start_time" | bc)
    if [ $run_exit_code -eq 0 ]; then
        echo -e "${GREEN}run passed${NC} (${elapsed_time}s)"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}run failed${NC} (${elapsed_time}s)"
        ((FAILED_TESTS++))
    fi

    if declare -f post_run > /dev/null; then
        echo "Executing post_run..."
        if ! post_run; then
            echo -e "${YELLOW}post_run completed with errors${NC}"
        fi
    fi

    set +x
    unset -f pre_run run post_run
}

main() {
    # Set these var in global env.
    EG_IMAGE_NAME="$1"
    EG_IMAGE_TAG="$2"

    echo "Starting standalone e2e test"
    echo "Looking for test cases in: $TEST_DIR"

    if [[ ! -d "$TEST_DIR" ]]; then
        echo -e "${RED}Error: Test directory '$TEST_DIR' not found${NC}"
        exit 1
    fi

    set -e
    local test_files=()
    while IFS= read -r -d $'\0' file; do
        test_files+=("$file")
    done < <(find "$TEST_DIR" -name "*.sh" -print0)

    TOTAL_TESTS=${#test_files[@]}

    if [[ $TOTAL_TESTS -eq 0 ]]; then
        echo -e "${YELLOW}No test cases found in $TEST_DIR${NC}"
        exit 0
    fi

    echo "Found $TOTAL_TESTS test case(s)"

    setup_env
    set +e

    for test_file in "${test_files[@]}"; do
        execute_test "$test_file"
    done

    cleanup_env

    echo -e "\n${YELLOW}=== Test Summary ===${NC}"
    echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
    echo -e "Total: $TOTAL_TESTS"

    if [[ $FAILED_TESTS -gt 0 ]]; then
        exit 1
    fi
}

main "$@"
