FROM golang:1.23.2 AS builder
ARG CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o gen-txns-csv cmd/gen-txns-csv/main.go

FROM scratch
COPY --from=builder /app/gen-txns-csv /gen-txns-csv
VOLUME ["/files"]
ENTRYPOINT ["/gen-txns-csv"]