# Multi-stage build for CodeSentinel
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/codesentinel ./cmd/codesentinel

FROM node:22-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm install
COPY web/ .
RUN npm run build

FROM alpine:3.20
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY --from=backend-builder /app/codesentinel /usr/local/bin/codesentinel
COPY --from=frontend-builder /app/web/dist /app/web/dist

EXPOSE 8080
ENTRYPOINT ["codesentinel", "server"]
