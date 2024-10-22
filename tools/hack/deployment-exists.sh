#!/usr/bin/env bash

DEPLOYMENT_LABEL_SELECTOR=$1
DEPLOYMENT_NAMESPACE=$2

# Timeout for deployment to exist (in seconds)
exist_timeout=300
end=$((SECONDS+exist_timeout))

while true; do
      deployment=$(kubectl get deployment -l "$DEPLOYMENT_LABEL_SELECTOR" -o name -n "$DEPLOYMENT_NAMESPACE")
      if [ -n "$deployment" ]; then
        echo "$deployment exists"
        break
    else
        echo "Waiting for deployment with label selectors = \"$DEPLOYMENT_LABEL_SELECTOR\" to exists in namespace: \"$DEPLOYMENT_NAMESPACE\""
    fi
    if [ $SECONDS -gt $end ]; then
        echo "The timeout for waiting for a deployment to exists has passed."
        exit 1
    fi
    sleep 5
done