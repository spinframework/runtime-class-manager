#!/bin/bash
set -euo pipefail

# Currently used in .github/workflows/helm-chart-smoketest.yml
# Iterates through all sample shims and applies their labels to all nodes
# using the current kubernetes context.
# Waits for all corresponding Shim resources to be ready, else fails.

for shim_file in $(ls config/samples/sample_shim*); do
  label="$(cat $shim_file | yq '.spec.nodeSelector' | tr -d '"' | tr -d '[:space:]' | sed s/:/=/g)"
  kubectl label node --all $label

  shim_name="$(cat $shim_file | yq '.metadata.name')"
  timeout=300
  SECONDS=0 # Reset the internal bash timer to 0
  success=false

  echo "Waiting for the $shim_name shim to be ready/installed..."

  while [[ $SECONDS -lt $timeout ]]; do
    # Fetch both nodes and nodesReady
    read -r nodes nodesReady <<< $(kubectl get shim "$shim_name" \
      -o jsonpath='{.status.nodes} {.status.nodesReady}' 2>/dev/null)

    # Check to see if all nodes are ready
    if [[ -n "$nodes" ]] && [[ -n "$nodesReady" ]] && [[ "$nodes" -eq "$nodesReady" ]]; then
      echo "Success: all nodes have the $shim_name shim installed."
      success=true
      break
    fi

    sleep 2
  done

  if [[ "${success}" != "true" ]]; then
    echo "Error: Timed out after ${timeout}s waiting for the $shim_name shim to be ready."
    exit 1
  fi
done