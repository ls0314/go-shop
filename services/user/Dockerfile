# user-service 镜像。构建上下文 = **仓库根**。
#
# ============================================================
# 为什么上下文必须是仓库根
# ============================================================
#
# 各服务 import `demo-shop/api/...`(生成的 proto 代码),而它们之间
# **没有 replace 指令** —— 全靠仓库根的 go.work 把 ./api 与 ./services/*
# 并成一个模块图才编译得过。
#
# 所以以 services/user 为上下文是编译不过的(go.work 里的路径在镜像里不存在)。
# 代价是每次构建都要把仓库传进 daemon —— 用根级 .dockerignore 压掉 node_modules 等。
#
# ============================================================
# 构建期自检
# ============================================================
#
# 第二阶段的 `--check` 会真的执行一次入口脚本,确认"二进制可执行 +
# 默认配置存在"。这样 COPY 或配置写错会让**构建失败**,而不是等到起容器。

FROM golang:1.25-alpine AS builder

# GOTOOLCHAIN=auto:三个服务的 go.mod 要求 go 1.25.11,若基础镜像的补丁版
# 更低,Go 会自动下载对应工具链(需要网络);不低于它则直接用本地的。
ENV GOTOOLCHAIN=auto \
    CGO_ENABLED=0 \
    GOOS=linux

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

WORKDIR /src

# 先拷 go.work 与**所有**模块的依赖清单。
#
# 注意:不能只拷目标服务的 go.mod —— go.work 的模块图覆盖全部模块,
# 而各服务依赖 api 模块的**源码**(不是它的 go.mod),所以源码必须一起拷。
# 因此这里不做"依赖层单独缓存"的分层(那样需要只拷部分源码,在 workspace 下做不到)。
COPY go.work go.work.sum ./
COPY api/go.mod ./api/
COPY services/user/go.mod ./services/user/
COPY services/bff/go.mod ./services/bff/
COPY services/product/go.mod ./services/product/
COPY services/trade/go.mod ./services/trade/
COPY services/marketing/go.mod ./services/marketing/

RUN go mod download

# 全部源码:api 是共享的,其余服务只为了满足 go.work 的 use 列表
COPY api/ ./api/
COPY services/ ./services/

WORKDIR /src/services/user
# 一个镜像里放**两个**二进制:主程序,以及迁移工具。
#
# 为什么迁移也放进同一个镜像:两者共用同一份 internal/config 与依赖,
# 分成两个镜像会重复整个构建过程,而运行时又必须严格同版本 ——
# 合成一个镜像天然保证"迁移工具与它所迁移的服务是同一份代码"。
RUN go build -trimpath -ldflags="-s -w" -o /out/user-service . \
 && go build -trimpath -ldflags="-s -w" -o /out/user-service-migrate ./cmd/migrate

# ------------------------------------------------------------

FROM alpine:3.20

# ca-certificates:出网 TLS 需要
# tzdata:缺它时 time.Local 退化成 UTC(订单时间、券有效期都受影响)
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/user-service /app/bin/user-service
# 迁移二进制:compose 的 migrate-user 用它(该容器覆盖 ENTRYPOINT)
COPY --from=builder /out/user-service-migrate /app/bin/user-service-migrate

# 默认配置用本地开发那份,只为让镜像"能独立起来"。
# compose 会用 etc/user.docker.yaml 覆盖它 —— 容器内地址(etcd/DB/Redis
# 用服务名)与本地(127.0.0.1)不同,不覆盖就会连自己。
COPY services/user/etc/user.yaml /app/etc/user.yaml
COPY services/user/migrations/ /app/migrations/

COPY docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
# 符号链接:入口脚本靠 basename 判断"我是哪个服务"
RUN ln -sf /app/entrypoint.sh /app/user-service

RUN /app/user-service --check

ENTRYPOINT ["/app/user-service"]
