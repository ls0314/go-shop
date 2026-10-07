# marketing-service 镜像。构建上下文 = **仓库根**(理由见 services/user/Dockerfile 的注释)。

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

WORKDIR /src/services/marketing
# 主程序 + 迁移工具(迁移工具与服务同镜像,保证同版本)
RUN go build -trimpath -ldflags="-s -w" -o /out/marketing-service . \
 && go build -trimpath -ldflags="-s -w" -o /out/marketing-service-migrate ./cmd/migrate

# ------------------------------------------------------------

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/marketing-service /app/bin/marketing-service
COPY --from=builder /out/marketing-service-migrate /app/bin/marketing-service-migrate

COPY services/marketing/etc/marketing.yaml /app/etc/marketing.yaml
COPY services/marketing/migrations/ /app/migrations/

COPY docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh && ln -sf /app/entrypoint.sh /app/marketing-service

RUN /app/marketing-service --check

ENTRYPOINT ["/app/marketing-service"]
