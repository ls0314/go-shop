# ============ 阶段一:构建 ============
FROM node:22-alpine AS builder
WORKDIR /build

# npm 源同样用 ARG 保持环境中立(国内换 npmmirror,CI 用默认)
ARG NPM_REGISTRY=https://registry.npmjs.org
RUN npm config set registry ${NPM_REGISTRY}

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build        # vue-tsc -b && vite build → dist/

# ============ 阶段二:nginx 运行 ============
FROM nginx:1.27-alpine

# 官方 nginx 镜像特性:/etc/nginx/templates/*.template 会经 envsubst
# 替换环境变量后生成 /etc/nginx/conf.d/*.conf —— 后端地址因此可以运行时注入
COPY nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=builder /build/dist /usr/share/nginx/html

EXPOSE 80