#!/usr/bin/env bash
set -euo pipefail

declare -A levels=([DEBUG]=0 [INFO]=1 [WARN]=2 [ERROR]=3)
script_logging_level="INFO"

log() {
    local log_message=$1
    local log_priority=$2

    #check if level exists
    [[ ${levels[$log_priority]} ]] || return 1

    #check if level is enough
    (( ${levels[$log_priority]} < ${levels[$script_logging_level]} )) && return 2

    #log here
    d=$(date '+%Y-%m-%dT%H:%M:%S')
    echo -e "${d}\t${log_priority}\t${log_message}"
}

log "start downloading shim from  ${SHIM_LOCATION}..." "INFO"

mkdir -p /assets

# overwrite default name of shim binary; use the name of shim resource instead
# to enable installing multiple versions of the same shim
curl -sLo "containerd-shim-${SHIM_NAME}" "${SHIM_LOCATION}"
ls -lah "containerd-shim-${SHIM_NAME}"

log "$(curl --version)" "INFO"
log "$(tar --version)" "INFO"
log "md5sum: $(md5sum containerd-shim-${SHIM_NAME})" "INFO"
log "sha256sum: $(sha256sum containerd-shim-${SHIM_NAME})" "INFO"

tar -xzf "containerd-shim-${SHIM_NAME}" -C /assets
log "download successful:" "INFO"

ls -lah /assets
