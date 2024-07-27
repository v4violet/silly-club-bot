FROM golang:1.22.5-alpine AS builder

WORKDIR /app

RUN apk update && apk add git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \ 
    CGO_ENABLED=0 go build -buildvcs -o program .

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/program /app/.env.default /

ENTRYPOINT ["/program"]