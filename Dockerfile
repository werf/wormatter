ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache curl && \
    sh -c "$(curl -fsSL https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

WORKDIR /app
COPY go.mod go.sum Taskfile.yaml ./
RUN go mod download
COPY . .
RUN task build

FROM alpine:3.21

COPY --from=builder /app/bin/wormatter /usr/local/bin/wormatter

ENTRYPOINT ["wormatter"]

