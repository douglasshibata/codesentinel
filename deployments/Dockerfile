FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=2.4.0" -o /codesentinel ./cmd/codesentinel

FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /codesentinel /codesentinel
COPY configs/codesentinel.default.yaml /etc/codesentinel/codesentinel.yaml

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/codesentinel"]
CMD ["server", "--config", "/etc/codesentinel/codesentinel.yaml"]
