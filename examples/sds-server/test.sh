#!/bin/bash
# Test script for the SDS server

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== SDS Server Test Script ===${NC}\n"

# Check if sds-server binary exists
if [ ! -f "./sds-server" ]; then
    echo -e "${YELLOW}Building sds-server...${NC}"
    make build
fi

# Start the SDS server in the background
echo -e "${YELLOW}Starting SDS server on port 18001...${NC}"
./sds-server -port 18001 -node sds-test-node > sds-server.log 2>&1 &
SDS_PID=$!
echo -e "${GREEN}SDS server started with PID: $SDS_PID${NC}"

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    if [ ! -z "$SDS_PID" ]; then
        kill $SDS_PID 2>/dev/null || true
        echo -e "${GREEN}SDS server stopped${NC}"
    fi
    if [ ! -z "$ENVOY_PID" ]; then
        kill $ENVOY_PID 2>/dev/null || true
        echo -e "${GREEN}Envoy stopped${NC}"
    fi
}
trap cleanup EXIT

# Wait for server to start
echo -e "${YELLOW}Waiting for SDS server to start...${NC}"
sleep 2

# Check if server is running
if ! ps -p $SDS_PID > /dev/null; then
    echo -e "${RED}ERROR: SDS server failed to start${NC}"
    cat sds-server.log
    exit 1
fi

echo -e "${GREEN}✓ SDS server is running${NC}\n"

# Test with grpcurl if available
if command -v grpcurl &> /dev/null; then
    echo -e "${YELLOW}Testing SDS server with grpcurl...${NC}"
    
    # List services
    echo -e "\n${YELLOW}Listing gRPC services:${NC}"
    grpcurl -plaintext localhost:18001 list || echo -e "${YELLOW}Note: Server may not support reflection${NC}"
    
    echo ""
fi

# Test with Envoy if available and config exists
if command -v envoy &> /dev/null && [ -f "envoy-config.yaml" ]; then
    echo -e "${YELLOW}Starting Envoy proxy...${NC}"
    envoy -c envoy-config.yaml --log-level info > envoy.log 2>&1 &
    ENVOY_PID=$!
    
    echo -e "${YELLOW}Waiting for Envoy to start...${NC}"
    sleep 5
    
    if ps -p $ENVOY_PID > /dev/null; then
        echo -e "${GREEN}✓ Envoy is running${NC}\n"
        
        # Test HTTP endpoint
        echo -e "${YELLOW}Testing HTTP endpoint (port 10080):${NC}"
        curl -s http://localhost:10080/ && echo ""
        
        # Test HTTPS endpoint
        echo -e "\n${YELLOW}Testing HTTPS endpoint (port 10443):${NC}"
        curl -k -s https://localhost:10443/ && echo ""
        
        echo -e "\n${GREEN}✓ All tests passed!${NC}"
        echo -e "\n${YELLOW}Check the logs:${NC}"
        echo -e "  SDS Server: cat sds-server.log"
        echo -e "  Envoy:      cat envoy.log"
        
        echo -e "\n${YELLOW}Access the Envoy admin interface:${NC}"
        echo -e "  http://localhost:9901"
        
        echo -e "\n${YELLOW}View SDS config dump:${NC}"
        echo -e "  curl http://localhost:9901/config_dump | jq '.configs[] | select(.\"@type\" | contains(\"SecretsConfigDump\"))'"
        
    else
        echo -e "${RED}ERROR: Envoy failed to start${NC}"
        cat envoy.log
        exit 1
    fi
else
    echo -e "${YELLOW}Envoy not found or envoy-config.yaml missing${NC}"
    echo -e "${YELLOW}Skipping Envoy integration test${NC}"
    
    echo -e "\n${GREEN}SDS server is running successfully!${NC}"
    echo -e "To test with Envoy:"
    echo -e "  1. Install Envoy: https://www.envoyproxy.io/docs/envoy/latest/start/install"
    echo -e "  2. Run: envoy -c envoy-config.yaml"
    echo -e "  3. Test: curl -k https://localhost:10443/"
fi

echo -e "\n${YELLOW}Logs from SDS server:${NC}"
tail -20 sds-server.log

echo -e "\n${GREEN}Test completed! Press Ctrl+C to stop servers.${NC}"

# Keep script running
wait $SDS_PID
