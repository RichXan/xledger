# 生产部署前端镜像：先构建静态资源，再由 Nginx 提供服务并代理 /api 到后端。

ARG NODE_IMAGE=docker.m.daocloud.io/library/node:20-alpine
ARG NGINX_IMAGE=docker.m.daocloud.io/library/nginx:1.27-alpine
FROM ${NODE_IMAGE} AS builder

WORKDIR /app

COPY frontend/app/package.json frontend/app/pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile

COPY frontend/app/ ./
RUN pnpm run build

FROM ${NGINX_IMAGE}

COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
