<div align="center">

# 🛍️ demo-shop

**前后端分离电商系统 —— Go (Gin) + Vue 3 全栈实践项目**

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14-4169E1?logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-6-DC382D?logo=redis&logoColor=white)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3.13-FF6600?logo=rabbitmq&logoColor=white)
![Elasticsearch](https://img.shields.io/badge/Elasticsearch-7.17-005571?logo=elasticsearch&logoColor=white)
![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vuedotjs&logoColor=white)
![License](https://img.shields.io/badge/License-Apache_2.0-blue)
[![CI](https://github.com/Zhaokun-2026/demo-shop/actions/workflows/ci.yml/badge.svg)](https://github.com/Zhaokun-2026/demo-shop/actions/workflows/ci.yml)

覆盖「用户端购买 + 管理端运营」完整业务链路：RBAC 权限体系、高并发安全控制、订单状态机、消息队列、全文检索、支付抽象在真实业务场景中的工程化落地。每个核心设计均可在代码中定位实现。

</div>

---

## 目录

- [界面速览](#-界面速览)
- [核心技术亮点](#-核心技术亮点)
- [核心链路设计](#-核心链路设计)
- [功能特性](#-功能特性)
- [技术栈](#-技术栈)
- [系统架构](#-系统架构)
- [项目结构](#-项目结构)
- [快速开始](#-快速开始)

---

##  界面速览

| 用户端 · 首页 | 用户端 · 商品详情 | 用户端 · 购物车/结算 |
|:---:|:---:|:---:|
| ![首页](docs/screenshots/shop-home.png) | ![商品详情](docs/screenshots/shop-product.png) | ![购物车结算](docs/screenshots/shop-checkout.png) |
| **用户端 · 订单详情** | **管理端 · 订单管理** | **管理端 · RBAC 权限** |
| ![订单详情](docs/screenshots/shop-order.png) | ![订单管理](docs/screenshots/admin-order.png) | ![RBAC](docs/screenshots/admin-rbac.png) |

---

##  核心技术亮点

### 1. 优惠券领取并发控制 —— 悲观锁 + 乐观锁「双防线」

高并发领券要同时解决两个问题：**单人超领**（绕过限领次数）与**总量超发**（库存卖超）。两条防线分别拦截：

```sql
-- 第一道防线：悲观锁，同模板的并发领取在此串行化
SELECT * FROM coupon_template WHERE template_id = ? FOR UPDATE;
-- 锁内完成「每人限领」校验 + 插入用户券，保证「读-判断-插入」原子性

-- 第二道防线：乐观锁条件更新，数据库侧原子自增
UPDATE coupon_template SET received_count = received_count + 1
WHERE template_id = ? AND received_count < total_count;
-- 影响行数为 0 → 券已售罄 → 事务回滚，防止总量超发
```

- 整个领取过程包在单个事务中，**锁的生命周期 = 事务生命周期**，避免锁提前释放留下并发窗口
- 核销 / 归还均走条件 UPDATE（`WHERE status = 'unused'` / `WHERE status = 'used'`），并发下同一张券只有一个事务能成功，**天然幂等防 MQ 重复消费**
- 完整覆盖优惠券「领取 → 核销 → 归还」状态机

 实现：[coupon_service.go](demo-shop-back/src/service/coupon_service.go) · [coupon_repo.go](demo-shop-back/src/repository/coupon_repo.go)

### 2. 库存一致性 —— 两段式模型 + 流水台账 + 操作幂等

- **三字段库存模型**：`stock`（可售）/ `lock_stock`（锁定）/ `sold_count`（销量）分离 —— 下单只锁定、支付才真扣、取消即释放、退款再回补
- 每次变更都是 `SELECT ... FOR UPDATE` 锁 SKU 行 → 校验可用量 → 条件 UPDATE 挪锁存量，配合 `CHECK (stock >= 0 AND lock_stock >= 0)` 数据库兜底防负数
- **库存流水台账**（`sys_product_stock_log`）：每次变更写一条流水，CHECK 约束枚举变更类型；写前检查「订单 + 变更类型」是否已存在，**同一订单同一操作只执行一次**，防 MQ 重复消费 / 回调重放导致库存重复变动

 实现：[inventory_service.go](demo-shop-back/src/service/inventory_service.go)

### 3. 订单超时自动取消 —— RabbitMQ TTL + 死信队列模拟延迟消息

RabbitMQ 无原生延迟消息，采用 **TTL + DLX（死信交换机）** 方案：

```
下单事务提交成功
      │ 发布持久化消息(DeliveryMode: Persistent)
      ▼
order.delay.queue (无消费者, x-message-ttl = 15min)
      │ TTL 到期, 死信路由
      ▼
order.dead.exchange ──▶ order.dead.queue ──▶ 消费者 goroutine
                                                │
                                          CancelOrder:
                                          仅 pending_pay 可取消(状态机幂等)
                                          已支付/已取消 → 直接 Ack 跳过
                                                │
                                    条件 UPDATE 取消订单 + 释放库存 + 归还优惠券
```

- 生产侧：下单事务 **commit 成功后**才发消息，消息持久化防 Broker 宕机丢失
- 消费侧：**订单状态机前置校验天然幂等**，重复投递不会重复取消；解析失败 Ack 丢弃防无限重投递
- 解耦设计：MQ 包不依赖业务层，通过 `OrderCanceller` 接口由订单服务反向注册，接口倒置

 实现：[rabbitmq.go](demo-shop-back/src/infra/mq/rabbitmq.go) · [order_delay.go](demo-shop-back/src/infra/mq/order_delay.go) · [consumer.go](demo-shop-back/src/infra/mq/consumer.go)

### 4. 幂等下单 + 雪花算法分布式 ID

- 下单接口携带 `idempotent_key`，`user_order_master` 上建**唯一索引**，前端重复提交直接返回已有订单，杜绝脏单
- 自研**雪花算法**：标准 64 位划分（41 时间戳 / 10 workerId / 12 序列号），处理**同毫秒序列溢出自旋**与**时钟回拨回退**；订单号 = `"DS" + 雪花 ID 转 36 进制`
- 最大事务将「核销优惠券 → 写订单主表 → 写明细 → 逐项锁库存 → 删购物车 → 写订单日志」六类操作原子完成，repo 层通过 `WithTx(tx)` 绑定事务实例

 实现：[order_service.go](demo-shop-back/src/service/order_service.go) · [snowflake.go](demo-shop-back/src/utils/snowflake.go)

### 5. Redis 缓存体系 —— 一致性策略 + 版本号失效 + 弱依赖降级

| 缓存键 | TTL | 说明 |
|---|---|---|
| `product:detail:{user/admin}:{id}` | 10 min | 商品详情（用户端/管理端视图分离） |
| `sku:stock:{id}` | 30 s | SKU 实时库存，短 TTL 兜底 |
| `category:{id}` | 1 h | 类目树 |
| `user:perm:{id}` | 30 min | 用户权限码，省去每次请求 3 表 JOIN |
| `api:perm:v{version}:{method}:{path}` | 30 min | 接口权限点 |

- **一致性策略**：Cache-Aside 读模式（miss 回源回写）+ **写库后删除** + 短 TTL 兜底，失效失败最多 30 秒旧值
- **版本号批量失效**：接口权限缓存键带版本号，权限变更时 `INCR api:perm:version` 一次失效全部接口缓存，避免逐 key 删除
- **弱依赖降级**：Redis 不可用时打 WARN 后业务直查 DB，缓存层全部判空处理，**不阻塞主流程**

 实现：[cache.go](demo-shop-back/src/infra/cache/cache.go) · [product_service.go](demo-shop-back/src/service/product_service.go) · [permission_service.go](demo-shop-back/src/service/permission_service.go)

### 6. Elasticsearch 商品搜索 —— 增量 + 全量双周期对账

- 业务写路径同步写 ES，失败仅 WARN 降级不阻塞主流程，**靠对账任务保证最终一致**
- **增量对账（5 分钟）**：基于 Redis 时间水位（RFC3339）拉取变更商品重新索引，**水位在写入成功后才推进**（否则漏数据）
- **全量对账（12 小时）**：全量比对 DB 与 ES 的 SPU ID 差集，重建缺失文档、清理孤儿文档

 实现：[es.go](demo-shop-back/src/infra/es/es.go) · [reconcile.go](demo-shop-back/src/task/reconcile.go)

### 7. 支付网关抽象 —— 策略模式 + 回调多重校验

- 定义 `PayGateway` 接口（创建支付 / 解析回调 / 主动查询）+ 注册中心，业务层只依赖接口；当前 Mock 实现已跑通支付闭环，接入微信/支付宝零业务改动
- 支付回调安全链：**签名验证（真实网关）→ 幂等检查（已 success 直接返回）→ 金额双重校验（回调金额 = 支付单金额 = 订单应付）→ 状态机校验（仅待支付可支付）→ 单事务落库**（更新支付单 → 更新订单 → 写日志 → 逐项扣减库存）

 实现：[gateway.go](demo-shop-back/src/infra/pay/gateway.go) · [mock_gateway.go](demo-shop-back/src/infra/pay/mock_gateway.go) · [payment_service.go](demo-shop-back/src/service/payment_service.go)

### 8. RBAC 权限体系 —— 六维数据模型 + 三级权限控制

- **数据模型**：用户-角色、角色-权限、角色-菜单、菜单-权限、用户-部门、数据权限（scope）六类关联
- **接口级鉴权**：自定义 `PermissionMiddleware`，按「路由 + HTTP Method」匹配权限点，用户权限码构建 map 后 O(1) 匹配，全部管理接口统一收口
- **菜单驱动动态路由**：后端返回菜单树，前端动态生成路由与侧边栏；页面按权限码控制按钮显隐，前后端双重校验
- **两级缓存**：用户权限码与接口权限点分别缓存，权限变更走版本号批量失效（见亮点 5）

 实现：[auth.go](demo-shop-back/src/middleware/auth.go) · [permission_service.go](demo-shop-back/src/service/permission_service.go)

### 9. 工程化与代码质量

- **版本化数据库迁移**：golang-migrate 管理 12 个版本（up/down 可回滚），表结构、约束、索引、权限 seed 全部落在 SQL 中，seed 幂等（`ON CONFLICT DO NOTHING`）
- **PostgreSQL 特性利用**：部分唯一索引（`WHERE is_deleted = FALSE`，软删除不占唯一约束）、CHECK 约束防脏数据、幂等键唯一索引
- **操作日志异步审计**：请求 Body 读后回填复用；密码/token 等敏感字段**递归脱敏**；按 rune 截断防切坏中文；经有界 channel（容量 1024）由 2 个 goroutine 异步入库，队列满丢弃告警不阻塞请求
- **文件分片上传**：1MB 分片 + MD5 标识 + 断点续传（分片存在性检查），`sync.Map` 按文件粒度互斥防并发合并
- 分层架构 routes → middleware → handler → service → repository，22 个业务模块同构；全局统一响应 / 错误码；JWT 双令牌（access 30min / refresh 24h）；后台定时任务（ES 对账）

 实现：[operation_log.go](demo-shop-back/src/middleware/operation_log.go) · [file_upload.go](demo-shop-back/src/utils/file_upload.go) · [db/migrations](demo-shop-back/db/migrations)

---

##  核心链路设计

**领券并发控制**（详见[亮点 1](#1-优惠券领取并发控制--悲观锁--乐观锁双防线)）：

```
领取请求 ──▶ BEGIN TRANSACTION
              │
              ├─① SELECT ... FOR UPDATE 锁模板行   ← 悲观锁: 同模板并发在此串行化
              │     锁内校验「每人限领」→ 插入用户券
              │
              ├─② UPDATE ... SET received_count = received_count + 1
              │     WHERE received_count < total_count
              │     └─ RowsAffected == 0 → 售罄 → ROLLBACK   ← 乐观锁: 防总量超发
              ▼
            COMMIT (锁的生命周期 = 事务生命周期)
```

**下单 → 超时取消 → 资源释放**：

```
幂等检查(idempotent_key) ──▶ 事务: 核销券/锁库存/写订单 ──▶ COMMIT ──▶ 发送 15min 延迟消息
                                                                            │ TTL 到期死信
                                                                            ▼
                                            消费 ──▶ 状态机校验(仅 pending_pay) ──▶ 取消订单
                                                                                      ├─ 释放库存(条件更新+流水, 幂等)
                                                                                      └─ 归还优惠券(条件更新, 幂等)
```

**支付回调**：

```
回调 ──▶ 签名验证 ──▶ 幂等检查(已 success 直接返回) ──▶ 金额双重校验 ──▶ 状态机校验
      ──▶ 单事务: 更新支付单 → 更新订单 → 写日志 → 逐项扣减库存(行锁+流水)
```

---

## 功能特性

###  用户端

| 模块 | 说明 |
|------|------|
| 商品 | SPU/SKU 多规格、类目树浏览、ES 全文搜索、商品上下架 |
| 购物车 | 多规格勾选、实时库存校验、结算预览 |
| 订单 | 幂等下单、完整状态机(待支付→已支付→已发货→已完成/已取消)、MQ 超时自动取消 |
| 支付 | 支付网关 Mock，下单后限时支付闭环 |
| 优惠券 | 领券中心、我的卡券、结算选券核销、取消订单自动归还 |
| 用户 | 登录注册、JWT 鉴权 + 刷新令牌、收货地址管理 |

###  管理端

| 模块 | 说明 |
|------|------|
| RBAC 管理 | 用户 / 角色 / 权限 / 菜单 / 部门 / 数据权限 六个管理页面 |
| 商品运营 | 类目树、SPU/SKU 管理、库存台账与流水日志 |
| 订单管理 | 全状态流转、发货、退款处理(库存/优惠券自动归还) |
| 优惠券运营 | 模板创建(满减/直减、有效期双模式)、发放统计 |
| 操作日志 | 关键管理操作全审计(敏感字段脱敏) |

---

## 技术栈

| 端 | 技术 |
|----|------|
| 后端 | [Go 1.25](https://go.dev) · [Gin](https://github.com/gin-gonic/gin) · [GORM](https://gorm.io) · [golang-migrate](https://github.com/golang-migrate/migrate) |
| 数据层 | PostgreSQL 14（26 张业务表）· Redis 6 · Elasticsearch 7.17 · RabbitMQ 3.13 |
| 中间件 | JWT 鉴权 · Redis 缓存 · MQ 延迟消息 · ES 全文检索 |
| 前端 | [Vue 3](https://vuejs.org) `<script setup>` · [TypeScript](https://www.typescriptlang.org) · [Vite](https://vitejs.dev) · [Element Plus](https://element-plus.org) · [Pinia](https://pinia.vuejs.org) · Tailwind CSS |
| 基础设施 | Docker Compose 一键启动全套中间件 |

---

## 系统架构

```
┌─────────────────────────────┐               ┌─────────────────────────────┐
│       用户端 (Vue3)          │               │       管理端 (Vue3)         │
│   商品 / 购物车 / 下单/支付   │               │  商品/订单运营 + RBAC 管理   │
│   优惠券 / 地址 / 个人中心     │              │  菜单驱动动态路由 + 按钮权限  │
└──────────────┬──────────────┘                └──────────────┬──────────────┘
               │            JWT · RESTful · 统一响应 / 错误码  │
               └───────────────────────┬──────────────────────┘
                                       ▼
┌────────────────────────────────────────────────────────────────┐
│                后端 (Go + Gin · 严格分层)                       │
│   routes → middleware → handler → service → repository → model │
├────────────────────────────────────────────────────────────────┤
│  Redis(缓存) │ RabbitMQ(订单超时/异步) │ ES(商品搜索)            │
│  PostgreSQL(业务库 · 26 张表 · 12 个版本化迁移)                  │
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
│       ├── repository/             # 数据访问层(行锁 / 条件更新 / JOIN)
│       ├── model/                  # 实体 / 请求 / 响应 + 统一错误码
│       ├── infra/                  # 基建封装：cache / es / mq / pay
│       ├── task/                   # 后台定时任务(ES 双周期对账)
│       └── utils/                  # JWT / 雪花 ID / 响应包装 / 分片上传
├── demo_shop_front/                # Vue3 前端
│   └── src/
│       ├── views/platform/         # 管理端业务页(商品/订单/库存/支付/优惠券)
│       ├── views/system/           # 管理端 RBAC 页(用户/角色/权限/菜单/部门/数据权限)
│       ├── views/shop/             # 用户端页(商品/购物车/结算/订单/优惠券)
│       ├── views/layout + userLayout # 管理端 / 用户端布局(动态侧边栏)
│       ├── pinia/modules/          # 状态管理(动态路由 / 用户权限码)
│       └── api/                    # 接口封装(12 个模块，与后端一一对应)
├── docs/screenshots/               # 界面截图
└── docker/docker-init.yml          # 基础设施编排(一键启动)
```

---

## 快速开始

### 前置条件
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)（启动 PostgreSQL / Redis / RabbitMQ / Elasticsearch / Kibana）
- Go 1.25（go.mod 声明 `go 1.25.0`）
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

> 本地体验账号见迁移 seed 数据（`db/migrations` 000003）。

---

## License

[Apache License 2.0](LICENSE)
