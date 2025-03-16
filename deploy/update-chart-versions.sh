#!/bin/bash
set -eou pipefail

# Note: using '-i.bak' to support different versions of sed when using in-place editing.

# Swap tag in for main for URLs if the version is vx.x.x*
if [[ "${APP_VERSION}" =~ ^v[0-9]+.[0-9]+.[0-9]+(.*)? ]]; then
  sed -i.bak -e "s%spinkube/runtime-class-manager/main%spinkube/runtime-class-manager/${APP_VERSION}%g" "${STAGING_DIR}/${CHART_NAME}-${CHART_VERSION}/README.md"
  sed -i.bak -e "s%spinkube/runtime-class-manager/main%spinkube/runtime-class-manager/${APP_VERSION}%g" "${STAGING_DIR}/${CHART_NAME}-${CHART_VERSION}/templates/NOTES.txt"
fi

## Update Chart.yaml with CHART_VERSION and APP_VERSION
yq -i '.version = env(CHART_VERSION)' "${STAGING_DIR}/${CHART_NAME}-${CHART_VERSION}/Chart.yaml"
yq -i '.appVersion = env(APP_VERSION)' "${STAGING_DIR}/${CHART_NAME}-${CHART_VERSION}/Chart.yaml"

## Update values.yaml tags
yq -i '.image.tag = env(APP_VERSION)' "${STAGING_DIR}/${CHART_NAME}-${CHART_VERSION}/values.yaml"
yq -i '.rcm.shimDownloaderImage.tag = env(APP_VERSION)' "${STAGING_DIR}/${CHART_NAME}-${CHART_VERSION}/values.yaml"
yq -i '.rcm.nodeInstallerImage.tag = env(APP_VERSION)' "${STAGING_DIR}/${CHART_NAME}-${CHART_VERSION}/values.yaml"

# Cleanup
find "${STAGING_DIR}/${CHART_NAME}-${CHART_VERSION}" -type f -name '*.bak' -print0 | xargs -0 rm -- || true