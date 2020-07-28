#!/bin/bash

set -o nounset -o pipefail -o errexit
trap "exit 1" INT

GRN=$'\e[32;1m'
OFF=$'\e[0m'

go get github.com/coinbase/rosetta-cli

printf "${GRN}### Running rosetta-cli tests${OFF}\n"

#Add all rosetta-cli checks here
rosetta-cli check --lookup-balance-by-block=false  --end 50

printf "${GRN}### Tests finished.${OFF}\n"