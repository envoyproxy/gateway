#!/usr/bin/env bash

DEPLOYMENT_LABEL_SELECTOR=$1
DEPLOYMENT_NAMESPACE=$2


# Timeout for deployment to exist (in seconds)
exist_timeout=25
end=$((SECONDS+exist_timeout))

while true; do
      deployment=$(kubectl get deployment -l "$DEPLOYMENT_LABEL_SELECTOR" -o "$DEPLOYMENT_NAMESPACE" -n)
      if [ -n "$deployment" ]; then
        echo "Deployment exists."
        break
    else
        echo "Waiting for deployment to exist..."
    fi
    if [ $SECONDS -gt $end ]; then
        echo "Timeout waiting for deployment to exist."
        exit 1
    fi
    sleep 5
done