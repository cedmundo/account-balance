FROM golang:1.23.2 AS builder
ARG CGO_ENABLED=0
WORKDIR /app

COPY . .
RUN go work sync
RUN go build -o proc-txns-csv cmd/proc-txns-csv/main.go

FROM scratch
COPY --from=builder /app/proc-txns-csv /proc-txns-csv
COPY .env .env
VOLUME ["/support/files"]
ENTRYPOINT ["/proc-txns-csv"]