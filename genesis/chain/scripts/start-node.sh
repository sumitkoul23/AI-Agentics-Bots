#!/usr/bin/env bash
# start-node.sh — run the devnet validator in the foreground.
#
# Useful flags to tweak via env:
#   MIN_GAS_PRICES   default "0.0001ugen"
#   AGENTICD_HOME    default $HOME/.skymetricd
#   LOG_LEVEL        default "info"
set -euo pipefail

AGENTICD_HOME="${AGENTICD_HOME:-$HOME/.skymetricd}"
MIN_GAS_PRICES="${MIN_GAS_PRICES:-0.0001ugen}"
LOG_LEVEL="${LOG_LEVEL:-info}"

echo "▶ skymetricd start (home=${AGENTICD_HOME})"
exec skymetricd start \
    --home "${AGENTICD_HOME}" \
    --minimum-gas-prices "${MIN_GAS_PRICES}" \
    --log_level "${LOG_LEVEL}" \
    --rpc.laddr "tcp://0.0.0.0:26657" \
    --grpc.address "0.0.0.0:9090" \
    --api.enable=true \
    --api.address "tcp://0.0.0.0:1317"
