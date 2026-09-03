<div align="center">

# 🛍️ demo-shop

**前后端分离的电商系统 —— Go (Gin) + Vue3 全栈实践项目**

覆盖「用户端购买 + 管理端运营」完整业务链路的后端工程能力展示项目：RBAC 权限体系、高并发安全控制、订单状态机、消息队列、全文检索、支付抽象等在真实业务场景中的落地实现。

</div>

---

## 目录

- [技术亮点](#技术亮点)
- [功能特性](#功能特性)
- [技术栈](#技术栈)
- [系统架构](#系统架构)
- [项目结构](#项目结构)
- [环境配置](#环境配置)

---

## 技术亮点

### 1. 完整的 RBAC 权限体系 —— 六维数据模型 + 三级权限控制

- **数据模型**：用户-角色、角色-权限、角色-菜单、菜单-权限、用户-部门、数据权限(scope)六类关联
- **菜单驱动动态路由**：后端返回菜单树，前端据此动态生成路由与侧边栏，不同角色看到不同后台
- **接口级鉴权**：自定义 `PermissionMiddleware`，按「路由 + HTTP Method」匹配权限点，用户持有任一权限码即放行；全部管理接口统一收口
- **前端按钮级控制**：登录后通过 `GET /api/v1/user/perms` 拉取用户权限码，页面用 `v-if` 控制「新增/编辑/删除/分配」按钮显隐，实现前后端双重权限校验
- **性能优化**：权限码两级缓存 —— 用户权限码存 `Redis user:perm:{id}`，接口权限点存 `api:perm:v{version}:{method}:{path}`，避免每次请求 3 表 JOIN

### 2. 订单超时自动取消 —— MQ 延迟消息 + 幂等设计

- 下单成功即发送 **RabbitMQ 延迟消息**，到期消费检查订单状态，超时自动取消并释放库存、归还优惠券
- 库存释放 / 券归还均用**条件 UPDATE + 影响行数**保证幂等，天然防 MQ 重复消费
- 下单接口携带 `idempotent_key`，防止前端重复提交产生脏单

### 3. 优惠券领取并发控制 —— 悲观锁 + 乐观锁双防线

```
① 悲观锁: SELECT ... FOR UPDATE 锁模板行
          → 锁内校验「每人限领」并插入用户券，同模板并发领取串行化
② 乐观锁: UPDATE ... WHERE received_count < total_count
          → 影响行数为 0 即判定售罄，防止总量超发
```

- 核销 / 取消归还均走条件 UPDATE，同一张券在并发下只允许一个事务成功
- 完整覆盖电商优惠券「领取 → 核销 → 归还」状态机

### 4. Elasticsearch 商品搜索（数据同步 + 组合查询）

- 商品数据同步至 ES(增量 + 全量同步方案)，业务查询与 MySQL 解耦
- 封装 `SearchRequest` / `SearchResult`，支持关键词 + 筛选组合查询、排序、分页与聚合

### 5. 支付网关抽象 —— 面向接口编程

- 定义 `PayGateway` 接口 + Mock 实现，业务层只依赖接口；替换真实支付渠道(微信/支付宝)零业务改动

### 6. Redis 缓存体系

- 商品/权限等热点数据缓存，缓存 miss 回源 DB 并回写
- 缓存键带版本号，接口权限变更后可整体失效，避免脏缓存

### 7. 工程化与代码质量

- **分层架构**：routes → middleware → handler → service → repository 严格分层，22 个业务模块同构
- **版本化数据库迁移**：golang-migrate 管理 12 个版本(up/down 可回滚)，表结构、约束、索引、权限 seed 全部落在 SQL 中
- 全局统一错误码 + 统一响应包装；JWT + 刷新令牌；雪花算法分布式 ID；操作日志审计；后台定时任务(对账/状态修复兜底)

---

## 功能特性

### 🛒 用户端

| 模块 | 说明 |
|------|------|
| 商品 | SPU/SKU 多规格、类目树浏览、ES 全文搜索、商品上下架 |
| 购物车 | 多规格勾选、实时库存校验、结算预览 |
| 订单 | 幂等下单、完整状态机(待支付→已支付→已发货→已完成/已取消/已退款)、MQ 超时自动取消 |
| 支付 | 支付网关 Mock，下单后限时支付闭环 |
| 优惠券 | 领券中心、我的卡券、结算选券核销、取消订单自动归还 |
| 用户 | 登录注册、JWT 鉴权 + 刷新令牌、收货地址管理 |

### 🔧 管理端

| 模块 | 说明 |
|------|------|
| RBAC 管理 | 用户 / 角色 / 权限 / 菜单 / 部门 / 数据权限 六个管理页面 |
| 商品运营 | 类目树、SPU/SKU 管理、库存台账与日志 |
| 订单管理 | 全状态流转、发货、退款审核 |
| 优惠券运营 | 模板创建(满减/直减、有效期双模式)、发放统计 |
| 操作日志 | 关键管理操作全审计 |

---

## 技术栈

| 端 | 技术 |
|----|------|
| 后端 | [Go](https://go.dev) · [Gin](https://github.com/gin-gonic/gin) · [GORM](https://gorm.io) · [golang-migrate](https://github.com/golang-migrate/migrate) |
| 数据层 | PostgreSQL 14 · Redis · Elasticsearch 7.17 · RabbitMQ 3.13 |
| 中间件 | JWT 鉴权 · Redis 缓存 · MQ 延迟消息 · ES 全文检索 |
| 前端 | [Vue 3](https://vuejs.org) `<script setup>` · [TypeScript](https://www.typescriptlang.org) · [Vite](https://vitejs.dev) · [Element Plus](https://element-plus.org) · [Pinia](https://pinia.vuejs.org) · Tailwind CSS |
| 基础设施 | Docker Compose 一键启动全套中间件 |

---

## 系统架构

```
┌─────────────────────────────┐       ┌─────────────────────────────┐
│       用户端 (Vue3)          │       │       管理端 (Vue3)          │
│   商品 / 购物车 / 下单/支付    │       │  商品/订单运营 + RBAC 管理    │
│   优惠券 / 地址 / 个人中心     │       │  菜单驱动动态路由 + 按钮权限   │
└──────────────┬──────────────┘       └──────────────┬──────────────┘
               │            JWT · RESTful · 统一响应 / 错误码              │
               └───────────────────────┬──────────────────────┘
                                       ▼
┌────────────────────────────────────────────────────────────────┐
│                后端 (Go + Gin · 严格分层)                       │
│   routes → middleware → handler → service → repository → model │
├────────────────────────────────────────────────────────────────┤
│  Redis(缓存) │ RabbitMQ(订单超时/异步) │ ES(商品搜索)             │
│  PostgreSQL(业务库 · 25 张表 · 12 个版本化迁移)                   │
└────────────────────────────────────────────────────────────────┘
```

**请求链路**：

```
请求 → routes(路由注册 / 鉴权挂载)
     → middleware：Auth(登录校验) → Permission(接口权限点) → OperationLog(审计)
     → handler：参数绑定 / 从上下文取用户 / 统一响应
     → service：业务规则 / 事务编排 / 状态机流转 / 并发控制
     → repository：GORM 数据访问 / 行锁 / 条件更新(乐观锁)
```

---

## 项目结构

```
demo-shop/
├── demo-shop-back/                 # Go 后端
│   ├── main.go                     # 入口：自动迁移 → 初始化基建 → 注册路由
│   ├── db/migrations/              # 12 个 golang-migrate 版本化迁移(up/down)
│   ├── resource/application.yaml   # 配置(端口 / 数据库 / Redis / MQ / ES)
│   └── src/
│       ├── routes/                 # 22 个模块路由注册(统一挂载鉴权)
│       ├── middleware/             # Auth / Permission / OperationLog
│       ├── handler/                # 22 个 HTTP 处理器
│       ├── service/                # 22 个业务服务(事务 / 状态机 / 锁)
│       ├── repository/             # 21 个数据访问层(行锁 / 条件更新 / JOIN)
│       ├── model/                  # 实体 / 请求 / 响应 + 统一错误码
│       ├── infra/                  # 基建封装：cache / es / mq / pay
│       ├── task/                   # 后台定时任务
│       └── utils/                  # JWT / 雪花 ID / 响应包装
├── demo_shop_front/                # Vue3 前端
│   └── src/
│       ├── views/platform/         # 管理端业务页(商品/订单/库存/支付/优惠券)
│       ├── views/system/           # 管理端 RBAC 页(用户/角色/权限/菜单/部门/数据权限)
│       ├── views/shop/             # 用户端页(商品/购物车/结算/订单/优惠券)
│       ├── views/layout + userLayout # 管理端 / 用户端布局(动态侧边栏)
│       ├── pinia/modules/          # 状态管理(动态路由 / 用户权限码)
│       └── api/                    # 接口封装(110+ 接口)
└── docker/docker-init.yml          # 基础设施编排(一键启动)
```

---

## 环境配置

### 前置条件
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)(启动 PostgreSQL / Redis / RabbitMQ / Elasticsearch / Kibana)
- Go 1.22+（go.mod 声明 `go 1.25.0`，建议使用 1.25）
- Node.js 18+ & npm

### 第 1 步：启动基础设施(Docker)

```bash
cd docker
docker-compose -f docker-init.yml up -d    # 启动全部中间件
docker-compose -f docker-init.yml ps       # 查看状态
```

**服务访问地址：**

| 服务 | 地址 | 说明 |
|------|------|------|
| PostgreSQL | `localhost:5432` | 业务数据库 |
| Redis | `localhost:6379` | 缓存 |
| RabbitMQ 管理台 | http://localhost:15672 | 消息队列 |
| Elasticsearch | http://localhost:9200 | 商品搜索 |
| Kibana | http://localhost:5601 | ES 可视化 |

### 第 2 步：启动后端

```bash
cd demo-shop-back
go run .
```

> 首次启动自动执行数据库迁移(12 个版本)。后端运行于 `http://localhost:9001`，数据库连接等配置见 `resource/application.yaml`。

### 第 3 步：启动前端

```bash
cd demo_shop_front
npm install
npm run dev
```

浏览器访问 http://localhost:5173 。

### 停止服务

```bash
cd docker && docker-compose -f docker-init.yml down
```

> 本地体验账号见迁移 seed 数据(`db/migrations` 000003)；项目为本地学习/展示用途，未部署线上环境。

---

## License

[Apache License 2.0](LICENSE)

---

*本项目为学习与实践用途的 Demo，用于展示后端工程能力，非生产级实现。*
