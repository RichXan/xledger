# Xledger Deploy

本目录提供面向部署的完整启动入口，使用 [docker-compose.yaml](/home/xan/paramita/xledger/deploy/docker-compose.yaml) 编排以下服务：

- `postgres`
- `redis`
- `xledger-backend`
- `xledger-frontend`

## 启动

在本目录执行：

```bash
docker compose up -d --build
```

启动后默认暴露以下端口：

- 前端：`http://127.0.0.1:4173`
- 后端：`http://127.0.0.1:8080`
- Postgres：`127.0.0.1:5432`
- Redis：`127.0.0.1:6379`

## 常用命令

```bash
docker compose ps
docker compose logs -f
docker compose down
docker compose down -v
```

## 关键文件

- [docker-compose.yaml](/home/xan/paramita/xledger/deploy/docker-compose.yaml)：完整部署编排入口
- [backend.config.yaml](/home/xan/paramita/xledger/deploy/backend.config.yaml)：后端容器运行配置
- [frontend.Dockerfile](/home/xan/paramita/xledger/deploy/frontend.Dockerfile)：前端生产部署镜像
- [nginx.conf](/home/xan/paramita/xledger/deploy/nginx.conf)：前端静态资源服务与 `/api` 反向代理配置

## 注意事项

- 后端当前通过挂载 [backend.config.yaml](/home/xan/paramita/xledger/deploy/backend.config.yaml) 启动，部署前请替换其中的 `auth.code_pepper` 和 `auth.token_secret`
- 当前默认前端通过 Nginx 提供静态文件，并将 `/api/` 请求转发到 `xledger-backend:8080`
- 如果本机已经占用了 `5432`、`6379`、`8080` 或 `4173`，启动前需要先释放这些端口
