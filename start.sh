#!/bin/bash
set -e

GRN='\e[32;1m'
RED='\033[0;31m'
OFF='\e[0m'

error() {
  local message="$2"
  local code="${3:-1}"
  if [[ -n "$message" ]] ; then
    echo -e "${RED} Error: ${message}; exiting with status ${code} ${OFF}"
  else
    echo -e "${RED} Error; exiting with status ${code} ${OFF}"
  fi
  kill 0
  exit "${code}"
}

trap 'error ${LINENO}' ERR INT SIGINT

lotus daemon 2>&1 &
sleep 30

lotus log set-level ERROR

while [ "$(lotus net peers | wc -l)" -eq 0 ]
do
  echo -e "${GRN}### Waiting for peers...${OFF}"
  sleep 5
done

LOTUS_CHAIN_INDEX_CACHE=32768
LOTUS_CHAIN_TIPSET_CACHE=8192
LOTUS_RPC_TOKEN=$( cat /data/node/token )

echo -e "${GRN}### Launching rosetta-filecoin-proxy${OFF}"
rosetta-filecoin-proxy 2>&1
