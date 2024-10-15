FROM golang:1.22 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -v -a -o ./service cmd/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/service /app/service

ENTRYPOINT ["/app/service"]