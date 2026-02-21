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

num_retry=${NUM_RETRY:-3}
sleep_duration=${SLEEP_DURATION:-2}

for (( i=0; i < $num_retry+1; i++ ))
do
    if curl -sLo "containerd-shim-${SHIM_NAME}" "${SHIM_LOCATION}"
    then
        ls -lah "containerd-shim-${SHIM_NAME}"
        break
    elif [ $i -eq $num_retry ]
    then
        log "number of failed downloads reached max retry ${num_retry}" "ERROR"
        exit 1
    else
        log "download failed. retry after sleep... (${i}/${num_retry})" "ERROR"
        sleep $sleep_duration
    fi
done

# overwrite default name of shim binary; use the name of shim resource instead
# to enable installing multiple versions of the same shim

log "$(curl --version)" "INFO"
log "$(tar --version)" "INFO"
log "md5sum: $(md5sum containerd-shim-${SHIM_NAME})" "INFO"
log "sha256sum: $(sha256sum containerd-shim-${SHIM_NAME})" "INFO"

tar -xzf "containerd-shim-${SHIM_NAME}" -C /assets
log "download successful:" "INFO"

# Verify SHA-256 if provided
if [ -n "${SHIM_SHA256:-}" ]; then
    log "verifying SHA-256 digest..." "INFO"
    if echo "${SHIM_SHA256}  containerd-shim-${SHIM_NAME}" | sha256sum -c -; then
        log "SHA-256 verification passed" "INFO"
    else
        log "SHA-256 verification FAILED: expected ${SHIM_SHA256} for containerd-shim-${SHIM_NAME}" "ERROR"
        exit 1
    fi
fi

ls -lah /assets
