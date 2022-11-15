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

go install github.com/coinbase/rosetta-cli@v0.10.0

rm -rf /tmp/rosetta-cli-test/*

printf "${GRN}### Running rosetta-cli tests ($NETWORK)${OFF}\n"

#Add all rosetta-cli checks here
rosetta-cli check:data --configuration-file ./rosetta-config-PR-${NETWORK}.json

printf "${GRN}### Tests finished.${OFF}\n"
