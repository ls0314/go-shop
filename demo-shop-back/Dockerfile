# ============ 构建 ============
FROM golang:1.25-alpine AS builder

WORKDIR /build


ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0:纯静态编译,产物不依赖 glibc,可跑在 alpine 上
# (本项目的驱动链 pgx/libpq 均为纯 Go,可以安全关闭;-s -w 去符号表瘦身)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /out/server .

# ============ 运行 ============
FROM alpine:3.20

# ca-certificates:访问 ES 等外部的根证书;tzdata + TZ:expire_at 等时间计算用本地时区
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /out/server ./server
# 迁移 SQL 必须进镜像:RunMigrations 用相对路径 file://db/migrations,依赖 WORKDIR
COPY --from=builder /build/db ./db
# yaml 现在只是"本地默认值"(Step 1 的成果),env 覆盖层在运行时注入,所以可以进镜像
COPY --from=builder /build/resource ./resource

# 上传文件是运行时数据,声明为卷,不进镜像层
VOLUME ["/app/uploads"]

EXPOSE 9001

# exec 形式(非 shell 形式):让 server 成为 PID 1,docker stop 的 SIGTERM 能直达进程
ENTRYPOINT ["./server"]