FROM golang:1.23.2 AS builder
ARG CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o proc-txns-csv cmd/proc-txns-csv/main.go

FROM scratch
COPY --from=builder /app/proc-txns-csv /proc-txns-csv
COPY .env .env
VOLUME ["/files"]
ENTRYPOINT ["/proc-txns-csv"]