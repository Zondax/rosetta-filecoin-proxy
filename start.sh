#!/bin/bash
set -e

GRN=$"\e[32;1m"
OFF=$"\e[0m"

lotus daemon 2>&1 &
sleep 30

lotus log set-level ERROR

while [ "$(lotus net peers | wc -l)" -eq 0 ]
do
  echo "${GRN}### Waiting for peers...${OFF}"
  sleep 5
done

LOTUS_CHAIN_INDEX_CACHE=32768
LOTUS_CHAIN_TIPSET_CACHE=8192
LOTUS_RPC_TOKEN=$( cat /data/node/token )

echo "${GRN}### Launching rosetta-filecoin-proxy${OFF}"
rosetta-filecoin-proxy


