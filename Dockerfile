FROM golang:1.24.0-alpine3.21 AS builder
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . /app/
RUN CGO_ENABLED=1 GO111MODULE=on GOOS=linux go build -o main main.go

FROM alpine:3.21.3
RUN apk --no-cache upgrade \
    && adduser -D -h /app -u 1000 app
WORKDIR /app
COPY --from=builder --chown=1000 /app/main ./main
VOLUME /app/data
EXPOSE 8080
USER 1000
CMD ["/app/main"]
