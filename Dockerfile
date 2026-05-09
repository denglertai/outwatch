FROM golang:1.26-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/outwatch ./cmd/outwatch

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /home/nonroot
COPY --from=builder /out/outwatch /usr/local/bin/outwatch
ENTRYPOINT ["/usr/local/bin/outwatch"]
