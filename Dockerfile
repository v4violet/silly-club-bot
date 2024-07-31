FROM golang:1.22.5-alpine AS builder

WORKDIR /app

RUN apk update

RUN apk add --no-cache git ca-certificates curl

RUN update-ca-certificates

RUN curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | sh -s -- --to /bin

COPY go.mod go.sum justfile ./

RUN go mod download

RUN go mod verify

COPY . .

ENV CGO_ENABLED=0

ENV GOOS=linux

ENV BUILD_STATIC=true

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \ 
    just build

FROM scratch

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /app/dist/program /

ENTRYPOINT ["/program"]