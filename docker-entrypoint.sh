#!/bin/bash
set -e

Windpay_CONFIG=${Windpay_HOME}/config/config.toml
Windpay_GENESIS=${Windpay_HOME}/config/genesis.json

# Init genesis state if geth not exist
DATA_DIR=$(cat ${Windpay_CONFIG} | grep -A1 '\[Node\]' | grep -oP '\"\K.*?(?=\")')

GETH_DIR=${DATA_DIR}/geth
if [ ! -d "$GETH_DIR" ]; then
  geth --datadir ${DATA_DIR} init ${Windpay_GENESIS}
fi

exec "geth" "--config" ${Windpay_CONFIG} "$@"
