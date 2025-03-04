FROM golang:1.24.0-alpine3.21 AS builder
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . /app/
RUN apk add --no-cache g++ binutils-gold \
    && CGO_ENABLED=1 GO111MODULE=on GOOS=linux go build -o main main.go


FROM alpine:3.21.3
RUN apk --no-cache upgrade \
    && apk add --no-cache wget
WORKDIR /app
RUN if [ `uname -m` = "x86_64" ]; then \
        wget https://github.com/mikefarah/yq/releases/download/v4.45.1/yq_linux_amd64 -O /usr/local/bin/yq;  \
    else \
        wget https://github.com/mikefarah/yq/releases/download/v4.45.1/yq_linux_arm64 -O /usr/local/bin/yq; \
    fi \
    && chmod +x /usr/local/bin/yq
COPY --from=builder --chown=1000 /app/main ./main
COPY --from=builder --chown=1000 /app/noProxy.yaml ./noProxy.yaml
COPY --from=builder --chown=1000 /app/entrypoint.sh /entrypoint.sh
VOLUME /app/data
ENV DOMAIN_NAME=www.example.com
EXPOSE 443
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/main"]
