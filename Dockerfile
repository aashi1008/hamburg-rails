FROM golang:1.23.7 AS builder
WORKDIR /app
COPY . .
RUN go build -o service ./cmd/main.go

FROM gcr.io/distroless/base-debian12
COPY --from=builder /app/service /service
CMD ["/service"]
