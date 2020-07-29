# Create builder container
FROM golang:1.14 as builder

ARG BRANCH_FIL=master
ARG REPO_FIL=https://github.com/filecoin-project/lotus
ARG NODEPATH=/lotus

ARG BRANCH_PROXY=master
ARG REPO_PROXY=https://github.com/Zondax/rosetta-filecoin-proxy.git
ARG PROXYPATH=/rosetta-proxy

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get install -yy apt-utils && \
    apt-get install -yy gcc git bzr jq pkg-config mesa-opencl-icd ocl-icd-opencl-dev

RUN git clone --single-branch --branch ${BRANCH_FIL} ${REPO_FIL} ${NODEPATH}
RUN cd ${NODEPATH} && make build && make install

RUN git clone --single-branch --branch ${BRANCH_PROXY} ${REPO_PROXY} ${PROXYPATH}
RUN cd ${PROXYPATH} && make build

# Create final container
FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive
ARG ROSETTA_PORT=8080
ARG LOTUS_API_PORT=1234
ARG PROXYPATH=/rosetta-proxy

# Install filecoin node
RUN apt-get update && \
    apt-get install -yy apt-utils  && \
    apt-get install -yy bzr jq pkg-config mesa-opencl-icd ocl-icd-opencl-dev

COPY --from=builder /usr/local/bin/lotus* /usr/local/bin/

RUN mkdir -p /data/{node,storage}
ENV LOTUS_PATH=/data/node/
ENV LOTUS_STORAGE_PATH=/data/storage/

#Install rosetta proxy
COPY --from=builder ${PROXYPATH}/rosetta-filecoin-proxy /usr/local/bin
ENV LOTUS_RPC_URL=ws://127.0.0.1:1234/rpc/v0
ENV LOTUS_RPC_TOKEN=""

EXPOSE $ROSETTA_PORT
EXPOSE $LOTUS_API_PORT

RUN echo ' \n\
'#!/bin/bash' \n\
GRN=$"\e[32;1m" \n\
OFF=$"\e[0m" \n\
lotus daemon& \n\
sleep 30 \n\
lotus log set-level ERROR \n\
peers="lotus net peers | wc -l" \n\
while [ $(eval $peers) -eq 0 ] \n\
do \n\
echo "${GRN}### Waiting for peers...${OFF}\n" \n\
sleep 5 \n\
done \n\
LOTUS_RPC_TOKEN=$( cat /data/node/token ) \n\
echo "${GRN}### Launching rosetta-filecoin-proxy${OFF}\n" \n\
rosetta-filecoin-proxy  ' >> /start.sh

RUN chmod +x /start.sh

CMD /start.sh && bash
