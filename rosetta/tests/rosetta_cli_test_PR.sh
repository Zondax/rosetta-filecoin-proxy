#!/bin/bash

set -o nounset -o pipefail -o errexit
set +u
trap "exit 1" INT

GRN=$'\e[32;1m'
YLW=$'\e[33;1m'
OFF=$'\e[0m'

NETWORK=$1
if [ -z $NETWORK ]; then
    NETWORK="mainnet"
    echo "${YLW}Using default mainnet value for network${OFF}"
fi

# The checked range is taken relative to the node's current head. A pinned
# height stops working once the node is restored from a newer snapshot, because
# it no longer holds the state needed to compute traces for older tipsets.
# CHECK_START_INDEX pins the start explicitly (e.g. to reproduce a past range).
CHECK_HEAD_OFFSET=${CHECK_HEAD_OFFSET:-200}
CHECK_BLOCKS=${CHECK_BLOCKS:-60}

CONFIG=./rosetta-config-PR-${NETWORK}.json
ONLINE_URL=$(jq -r '.online_url' "$CONFIG")
NETWORK_ID=$(jq -c '{network_identifier: .network}' "$CONFIG")

go install github.com/coinbase/rosetta-cli@v0.10.0

if [ -z "$CHECK_START_INDEX" ]; then
    HEAD=""
    # The proxy is started in the background just before this script runs.
    for _ in $(seq 1 60); do
        HEAD=$(curl -sf -X POST -H 'Content-Type: application/json' -d "$NETWORK_ID" \
            "$ONLINE_URL/network/status" | jq -r '.current_block_identifier.index // empty') && [ -n "$HEAD" ] && break
        sleep 5
    done
    if [ -z "$HEAD" ]; then
        echo "could not read the current head from $ONLINE_URL/network/status" >&2
        exit 1
    fi
    CHECK_START_INDEX=$((HEAD - CHECK_HEAD_OFFSET))
fi
CHECK_END_INDEX=$((CHECK_START_INDEX + CHECK_BLOCKS))

RUN_CONFIG=$(mktemp)
trap 'rm -f "$RUN_CONFIG"' EXIT
jq --argjson start "$CHECK_START_INDEX" --argjson end "$CHECK_END_INDEX" \
    '.data.start_index = $start | .data.end_conditions.index = $end' "$CONFIG" > "$RUN_CONFIG"

rm -rf /tmp/rosetta-cli-test/*

printf "${GRN}### Running rosetta-cli tests ($NETWORK, blocks $CHECK_START_INDEX-$CHECK_END_INDEX)${OFF}\n"

#Add all rosetta-cli checks here
rosetta-cli check:data --configuration-file "$RUN_CONFIG"

printf "${GRN}### Tests finished.${OFF}\n"
