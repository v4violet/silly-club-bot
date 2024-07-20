FROM golang:1.22.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o program

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/program /

ENTRYPOINT ["/program"]