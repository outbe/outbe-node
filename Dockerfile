FROM golang:1.23.6-alpine3.20 AS build-env

SHELL ["/bin/sh", "-ecuxo", "pipefail"]

RUN set -eux; apk add --no-cache \
    ca-certificates \
    build-base \
    git \
    linux-headers \
    bash \
    binutils-gold

WORKDIR /code

ADD go.mod go.sum ./
RUN set -eux; \
    export ARCH=$(uname -m); \
    WASM_VERSION=$(go list -m all | grep github.com/CosmWasm/wasmvm || true); \
    if [ ! -z "${WASM_VERSION}" ]; then \
      WASMVM_REPO=$(echo $WASM_VERSION | awk '{print $1}');\
      WASMVM_VERS=$(echo $WASM_VERSION | awk '{print $2}');\
      wget -O /lib/libwasmvm_muslc.a https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERS}/libwasmvm_muslc.$(uname -m).a;\
    fi; \
    go mod download;

# Copy over code
COPY . /code

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
# then log output of file /code/bin/outbe-noded
# then ensure static linking
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true make build \
  && file /code/build/outbe-noded \
  && echo "Ensuring binary is statically linked ..." \
  && (file /code/build/outbe-noded | grep "statically linked")

# --------------------------------------------------------
FROM cosmwasm/optimizer:0.16.0 AS optimizer

RUN apk add jq tar bash

COPY --from=build-env /code/build/outbe-noded /usr/bin/outbe-noded

# Unset entrypoint for being able to use it in CI
ENTRYPOINT []
  
# --------------------------------------------------------
FROM alpine:3.21

COPY --from=build-env /code/build/outbe-noded /usr/bin/outbe-noded

RUN apk add --no-cache ca-certificates curl make bash jq sed

WORKDIR /opt

# rest server, tendermint p2p, tendermint rpc
EXPOSE 1317 26656 26657 8545 8546

CMD ["/usr/bin/outbe-noded", "version"]
