FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /build/bin cmd/main.go

FROM golang:1.20-alpine AS runner

WORKDIR /

COPY --from=builder /build/bin /app

EXPOSE 8888

USER nonroot:nonroot

ENTRYPOINT ["/app"]