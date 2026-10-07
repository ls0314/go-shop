# BFF 镜像(HTTP,不是 gRPC)。构建上下文 = **仓库根**。
#
# 与前四个服务的两点差别:
#   1. 它要挂 **公钥**(验签),user-service 挂的是私钥(签发);
#   2. 它没有 migrations 目录(不拥有任何表)。

FROM golang:1.25-alpine AS builder

ENV GOTOOLCHAIN=auto \
    CGO_ENABLED=0 \
    GOOS=linux

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

WORKDIR /src

COPY go.work go.work.sum ./
COPY api/go.mod ./api/
COPY services/user/go.mod ./services/user/
COPY services/bff/go.mod ./services/bff/
COPY services/product/go.mod ./services/product/
COPY services/trade/go.mod ./services/trade/
COPY services/marketing/go.mod ./services/marketing/

RUN go mod download

COPY api/ ./api/
COPY services/ ./services/

WORKDIR /src/services/bff
# 主程序是包根下的 bff.go(不是 cmd/ 子目录)
RUN go build -trimpath -ldflags="-s -w" -o /out/bff .

# ------------------------------------------------------------

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/bff /app/bin/bff

# 默认配置用本地那份只为让镜像能独立起;compose 会用 bff.docker.yaml 覆盖
# (etcd 地址与 S3 的 Endpoint 必须换,见 etc/bff.docker.yaml)
COPY services/bff/etc/bff.yaml /app/etc/bff.yaml

COPY docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh && ln -sf /app/entrypoint.sh /app/bff

RUN /app/bff --check

ENTRYPOINT ["/app/bff"]
