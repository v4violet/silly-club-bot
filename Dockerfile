FROM golang:1.22.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \ 
    CGO_ENABLED=0 GOOS=linux go build -o program

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/program /

ENTRYPOINT ["/program"]