#!/bin/sh
# 所有服务镜像共用的入口脚本。镜像里放的是 bin/<名>,本脚本按参数把它跑起来。
#
# 为什么要这一层(而不是 compose 里直接写全路径):
#   1. 可以在启动前**自检二进制与配置存在**,并把缺失项一次性说清楚 ——
#      否则容器的表现是 exec 失败后立刻退出,日志只有一行 "no such file",
#      而真正的原因多半是"COPY 少了一个目录"或"挂载路径写错";
#   2. 默认配置路径只有这一份,5 个服务的 compose 写法才能统一;
#   3. 支持 `docker run <镜像> --check` 只做自检不起服务 ——
#      CI 或人工都能用它在**不启动全栈**的前提下验证镜像是否可用。
#
# 用法:
#   <镜像>                     → 用 /app/etc/<镜像名>.yaml
#   <镜像> -f etc/other.yaml   → 指定配置(相对 /app)
#   <镜像> --check             → 只自检,不起服务

set -e

BIN_DIR=/app/bin
CONFIG_NAME="$(basename "$0")"

CHECK_ONLY=0
CONFIG_PATH="etc/${CONFIG_NAME}.yaml"

# 只处理本脚本真正需要识别的参数;其余参数直接忽略而不是透传 ——
# 透传需要正确的参数重排,而当前没有那个需求,写错了反而更难查。
while [ $# -gt 0 ]; do
  case "$1" in
    --check)
      CHECK_ONLY=1
      ;;
    -f)
      CONFIG_PATH="${2:-}"
      if [ -z "$CONFIG_PATH" ]; then
        printf '[entrypoint] 致命:-f 后面缺少配置路径\n' >&2
        exit 1
      fi
      shift
      ;;
  esac
  shift
done

BIN_PATH="${BIN_DIR}/${CONFIG_NAME}"
FATAL=0

printf '[entrypoint] 服务=%s\n' "$CONFIG_NAME"

if [ ! -x "$BIN_PATH" ]; then
  printf '[entrypoint] 致命:二进制不存在或不可执行 %s\n' "$BIN_PATH" >&2
  printf '[entrypoint]   镜像内 %s 的实际内容:\n' "$BIN_DIR" >&2
  ls -l "$BIN_DIR" >&2 || true
  FATAL=1
fi

if [ ! -f "/app/${CONFIG_PATH}" ]; then
  printf '[entrypoint] 致命:配置文件不存在 /app/%s\n' "$CONFIG_PATH" >&2
  printf '[entrypoint]   /app/etc 的实际内容:\n' >&2
  ls -l /app/etc >&2 || true
  printf '[entrypoint]   提示:compose 里应把 etc/<名>.docker.yaml 挂到 /app/etc/<名>.yaml\n' >&2
  FATAL=1
fi

# 密钥:只有 user-service(私钥)与 bff(公钥)需要。
# 缺了服务本身也会失败,但它的报错只说"读取公钥文件失败",不说挂载错了。
case "$CONFIG_NAME" in
  user-service)
    [ -f /app/jwt_keys/dev_private.pem ] || {
      printf '[entrypoint] 致命:缺 /app/jwt_keys/dev_private.pem\n' >&2
      printf '[entrypoint]   提示:compose 里应挂 ./jwt_keys:/app/jwt_keys:ro\n' >&2
      FATAL=1
    }
    ;;
  bff)
    [ -f /app/jwt_keys/dev_public.pem ] || {
      printf '[entrypoint] 致命:缺 /app/jwt_keys/dev_public.pem\n' >&2
      printf '[entrypoint]   提示:compose 里应挂 ./jwt_keys:/app/jwt_keys:ro\n' >&2
      FATAL=1
    }
    ;;
esac

# 配置里出现 127.0.0.1,最常见的成因是"挂成了本地开发那份配置" ——
# 它在容器里指向自己而不是别的服务,症状是连不上 etcd/DB 而报错却不提配置。
if [ -f "/app/${CONFIG_PATH}" ] && grep -q '127\.0\.0\.1' "/app/${CONFIG_PATH}"; then
  printf '[entrypoint] 警告:配置里出现 127.0.0.1,容器内通常应指向服务名\n' >&2
  grep -n '127\.0\.0\.1' "/app/${CONFIG_PATH}" >&2 || true
  printf '[entrypoint]   (若确认是刻意为之可忽略)\n' >&2
fi

if [ "$FATAL" -eq 1 ]; then
  exit 1
fi

if [ "$CHECK_ONLY" -eq 1 ]; then
  printf '[entrypoint] 自检通过(未启动服务)\n'
  exit 0
fi

printf '[entrypoint] 启动 %s -f %s\n' "$BIN_PATH" "$CONFIG_PATH"
exec "$BIN_PATH" -f "$CONFIG_PATH"
