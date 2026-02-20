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

tar -xzf "containerd-shim-${SHIM_NAME}" -C /tmp
# there may be multiple files in the archive; only copy the shim binary to /assets
cp /tmp/containerd-shim-${SHIM_NAME} /assets
log "download successful:" "INFO"

# Verify SHA-256 if provided
if [ -n "${SHIM_SHA256:-}" ]; then
    log "verifying SHA-256 digest..." "INFO"
    if echo "${SHIM_SHA256} containerd-shim-${SHIM_NAME}" | sha256sum -c -; then
        log "SHA-256 verification passed" "INFO"
    else
        log "SHA-256 verification FAILED: expected ${SHIM_SHA256} for containerd-shim-${SHIM_NAME}" "ERROR"
        exit 1
    fi
fi

ls -lah /assets
