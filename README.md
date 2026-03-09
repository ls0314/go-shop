# demo-shop
一个商城demo

## 运行环境搭建

### 前置条件
- 安装 Docker Desktop (Mac/Windows)
- 确保 Docker 服务正在运行

### 运行命令

#### Windows 系统 (PowerShell)
```powershell
# 进入 docker 目录
cd docker

# 运行 Docker Compose
docker-compose -f docker-init.yml up -d

# 查看服务状态
docker-compose -f docker-init.yml ps

# 停止服务
docker-compose -f docker-init.yml down
```

### 服务访问地址
- PostgreSQL: localhost:5432
- Redis: localhost:6379
- RabbitMQ 管理界面: http://localhost:15672
- Elasticsearch: http://localhost:9200
- Kibana: http://localhost:5601