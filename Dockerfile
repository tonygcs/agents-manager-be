FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git openssh-client

WORKDIR /app

RUN mkdir -p /root/.ssh && \
    ssh-keyscan github.com >> /root/.ssh/known_hosts

ENV GOPRIVATE=github.com/algorath-software/*

COPY go.mod go.sum ./
RUN --mount=type=ssh git config --global url."git@github.com:".insteadOf "https://github.com/" && \
    go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o agents-manager-be .


FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/agents-manager-be .
COPY etc/ etc/

ENTRYPOINT ["./agents-manager-be"]
