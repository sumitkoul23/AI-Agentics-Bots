# Multi-stage Dockerfile for the `agenticd` validator binary.
#
# Build:   docker build -t ghcr.io/sumitkoul23/agenticd:dev -f deploy/docker/Dockerfile chain
# Run:     docker run --rm -p 26657:26657 -p 1317:1317 ghcr.io/sumitkoul23/agenticd:dev

FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git make gcc musl-dev linux-headers
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN make build

FROM alpine:3.20
RUN apk add --no-cache ca-certificates jq curl bash
RUN addgroup -S agentic && adduser -S agentic -G agentic
COPY --from=builder /src/build/agenticd /usr/local/bin/agenticd
COPY scripts/ /home/agentic/scripts/
RUN chown -R agentic:agentic /home/agentic
USER agentic
WORKDIR /home/agentic
EXPOSE 26656 26657 1317 9090
ENTRYPOINT ["/usr/local/bin/agenticd"]
CMD ["start", "--minimum-gas-prices", "0.0001ugen"]
