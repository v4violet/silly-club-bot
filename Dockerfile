FROM golang:1.22.5-alpine AS builder

WORKDIR /app

RUN apk update 
RUN apk add --no-cache git ca-certificates
RUN update-ca-certificates

COPY go.mod go.sum ./

RUN go mod download

RUN go mod verify

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \ 
    CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -buildvcs -o program .

FROM scratch

COPY --from=builder /app/program /app/.env.default /
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

ENTRYPOINT ["/program"]