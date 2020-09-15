#!/bin/bash
set -e

GRN='\e[32;1m'
RED='\033[0;31m'
BOLDW='\e[1m'
OFF='\e[0m'

error() {
  local message="$2"
  local code="${3:-1}"
  if [[ -n "$message" ]] ; then
    echo -e "${RED} Error: ${message}; exiting with status ${code} ${OFF}"
  else
    echo -e "${RED} Error; exiting with status ${code} ${OFF}"
  fi
  exit_func "${code}"
}

exit_func() {
  echo -e "${GRN}Exiting...${OFF}"
  trap - SIGINT SIGTERM
  kill -- -$$
  exit "$1"
}

echo -e "${GRN}Running command: ${OFF}${BOLDW}lotus daemon $1 $2${OFF}"

[ -z "$GOLOG_LOG_LEVEL" ] && export GOLOG_LOG_LEVEL=INFO
echo -e "${GRN}Using Lotus logger level:${OFF}${BOLDW} ${GOLOG_LOG_LEVEL} ${OFF}"

lotus daemon $1 $2 2>&1 &

trap 'error ${LINENO}' ERR
trap 'exit_func 0' INT SIGINT

LOTUS_CHAIN_INDEX_CACHE=32768
LOTUS_CHAIN_TIPSET_CACHE=8192
LOTUS_RPC_TOKEN=$( cat /data/node/token )

echo -e "${GRN}### Launching rosetta-filecoin-proxy${OFF}"
rosetta-filecoin-proxy 2>&1
