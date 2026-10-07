# product-service 镜像。构建上下文 = **仓库根**(理由见 services/user/Dockerfile 的注释)。

FROM golang:1.25-alpine AS builder

ENV GOTOOLCHAIN=auto \
    CGO_ENABLED=0 \
    GOOS=linux

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

WORKDIR /src

# go.work 的模块图覆盖全部模块,故所有 go.mod 都要在
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

WORKDIR /src/services/product
# 主程序 + 迁移工具(迁移工具与服务同镜像,保证同版本)
RUN go build -trimpath -ldflags="-s -w" -o /out/product-service . \
 && go build -trimpath -ldflags="-s -w" -o /out/product-service-migrate ./cmd/migrate

# ------------------------------------------------------------

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/product-service /app/bin/product-service
COPY --from=builder /out/product-service-migrate /app/bin/product-service-migrate

COPY services/product/etc/product.yaml /app/etc/product.yaml
COPY services/product/migrations/ /app/migrations/

COPY docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh && ln -sf /app/entrypoint.sh /app/product-service

RUN /app/product-service --check

ENTRYPOINT ["/app/product-service"]
