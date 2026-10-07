<div align="center">

# 🛍️ demo-shop

**前后端分离电商系统 —— Go (go-zero) 微服务 + Vue 3 全栈实践项目**

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![go-zero](https://img.shields.io/badge/go--zero-1.10-0A7EBB?logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14-4169E1?logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-6-DC382D?logo=redis&logoColor=white)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3.13-FF6600?logo=rabbitmq&logoColor=white)
![Elasticsearch](https://img.shields.io/badge/Elasticsearch-7.17-005571?logo=elasticsearch&logoColor=white)
![MinIO](https://img.shields.io/badge/MinIO-S3-C72E49?logo=minio&logoColor=white)
![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vuedotjs&logoColor=white)
![License](https://img.shields.io/badge/License-Apache_2.0-blue)
[![CI](https://github.com/ls0314/go-shop/actions/workflows/ci.yml/badge.svg)](https://github.com/ls0314/go-shop/actions/workflows/ci.yml)

覆盖「用户端购买 + 管理端运营」完整业务链路：**BFF + 四个微服务**、RBAC 权限体系、高并发安全控制、订单状态机、消息队列、全文检索、支付抽象在真实业务场景中的工程化落地。每个核心设计均可在代码中定位实现。

</div>

---

## 目录

- [界面速览](#界面速览)
- [核心技术亮点](#核心技术亮点)
- [核心链路设计](#核心链路设计)
- [功能特性](#功能特性)
- [技术栈](#技术栈)
- [系统架构](#系统架构)
- [项目结构](#项目结构)
- [快速开始](#快速开始)

---

##  界面速览

| 用户端 · 首页 | 用户端 · 商品详情 | 用户端 · 购物车/结算 |
|:---:|:---:|:---:|
| ![首页](docs/screenshots/shop-home.png) | ![商品详情](docs/screenshots/shop-product.png) | ![购物车结算](docs/screenshots/shop-checkout.png) |
| **用户端 · 订单详情** | **管理端 · 订单管理** | **管理端 · RBAC 权限** |
| ![订单详情](docs/screenshots/shop-order.png) | ![订单管理](docs/screenshots/admin-order.png) | ![RBAC](docs/screenshots/admin-rbac.png) |

---

##  核心技术亮点

> **演进说明**：项目原为 Gin 单体，现已完成微服务拆分 —— 单体代码已归档至 `archive/monolith` 分支（`monolith-final` tag）并从主分支移除。

### 1. 服务拆分与 BFF 网关 —— 111 条路由 × 一次 RPC 一次域

拆分按**数据所有权**切分，而不是按页面或按层。每个服务独占自己的库，跨域只经 BFF：

| 服务 | 端口 | 数据库 | 表 | 职责 |
|---|---|---|---|---|
| **BFF** | 9003 (HTTP) | — | — | 唯一对外入口，聚合与协议转换 |
| user-service | 9004 | `user_db` | 15 | 用户 / 角色 / 权限 / 菜单 / 部门 / 数据权限 / 地址 |
| product-service | 9005 | `product_db` | 5 | 类目 / SPU / SKU / 图片 / 库存流水（含 ES 搜索与对账任务） |
| trade-service | 9007 | `trade_db` | 5 | 购物车 / 订单 / 支付 / outbox |
| marketing-service | 9009 | `marketing_db` | 3 | 券模板 / 用户券（含 Redis 预扣闸门与对账任务） |

- **BFF 聚合 111 条路由**（25 个前缀块），对前端保持**原单体时代的 URL 契约不变** —— 前端只改了代理端口，业务代码零改动
- **协议转换集中在 BFF**：HTTP（请求多为 snake_case，响应各域在 camelCase / snake_case 间并不一致）+ gRPC（proto 的 `error_msg` 业务错误通道）。转换函数按域集中，不散落在 handler
- **一次请求一次 RPC**：BFF 不做 N+1 聚合，需要合并的接口由服务侧提供（如库存的 SPU 维度聚合）
- **错误分类统一出口**：把下游结果分成「成功 / 业务失败 / 基础设施故障」三态，映射到 200 / 400 / 503；凭据与权限问题分别用哨兵标成 401 / 403

### 2. 接口级鉴权 —— 权限码**在生成期**固化，不查运行时表

单体的鉴权是"运行时按 `路由 + Method` 查权限表"。微服务下这条链不能照搬：BFF 无状态、也不该为了鉴权每次多打一次 RPC。

- **权限码在生成期烘焙进路由表**：一个代码生成器读取 `.api` 的 111 条路由，对其中 78 条管理端路由**逐条经 RPC 查 `sys_permission`**，把结果写成编译期字面量注入路由文件
- **运行时只做一件事**：取当前用户的权限码集合（一次 RPC，1~2ms），与路由上固化的码求交集。**不查 api→权限 映射、不引入 BFF 侧缓存、不用版本号失效**
- **fail-closed**：路由未配权限码 → 403 拒绝；用户无匹配码 → 403；取不到身份 → 500（**刻意不用 401**，见亮点 15）
- **改权限后必须重跑生成器**：这正是"固化"的代价，也是对的选择 —— 权限点是低频变更的配置，而鉴权是每个请求都要走的路径
- 5 条路由例外（登录后立即需要的自助接口），以显式清单声明「不判权」，从宽需留下痕迹

### 3. 库存一致性 —— 两段式模型 + 流水台账 + 操作幂等

- **三字段库存模型**：`stock`(可售) / `lock_stock`(锁定) / `sold_count`(销量) 分离 —— 下单只锁定、支付才真扣、取消即释放、退款再回补
- 每次变更都是 `SELECT ... FOR UPDATE` 锁 SKU 行 → 校验可用量 → 条件 UPDATE 挪锁存量，配合 `CHECK (stock >= 0 AND lock_stock >= 0)` 数据库兜底防负数
- **库存流水台账**：每次变更写一条流水，CHECK 约束枚举变更类型
- **幂等由数据库保证**：`(idempotency_key, sku_id, change_type)` 三元组**部分唯一索引** —— 重复调用撞唯一约束而非静默重复扣减。这比"先查后做"可靠：后者在并发下有窗口
- **一次调用 = 一次 RPC**：拆分后四个库存操作（锁定 / 扣减 / 释放 / 回补）各自成为一次幂等 RPC，由 trade-service 的 Saga 按序调用

### 4. 优惠券领取并发控制 —— 悲观锁 + 乐观锁「双防线」

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

### 5. 订单超时自动取消 —— RabbitMQ TTL + 死信队列模拟延迟消息

RabbitMQ 无原生延迟消息，采用 **TTL + DLX（死信交换机）** 方案：

```
下单事务内写 outbox 行(与订单同事务提交)      ← 见亮点 6：不在此刻直接发 MQ
      │ 投递器轮询取 pending 行
      │ 发布持久化消息(DeliveryMode: Persistent)
      ▼
order.delay.queue (无消费者, x-message-ttl = 15min)
      │ TTL 到期, 死信路由
      ▼
order.dead.exchange ──▶ order.dead.queue ──▶ 消费者 goroutine
                                                │
                                          CancelOrderBySystem:
                                          仅 pending_pay 可取消(状态机幂等)
                                          已支付/已取消 → 跳过并计 skipped
                                                │
                                    解除订单取消 + 释放库存 + 归还优惠券
```

- **延迟时长单一事实源**：队列 TTL 与下单响应的 `pay_expire_time` 都引用同一个模型常量，避免"用户看到的倒计时"与"系统真正取消的时刻"漂移
- 消费侧：**订单状态机前置校验天然幂等**，重复投递不会重复取消；解析失败 Ack 丢弃防无限重投递
- **系统取消不带 userId**：系统取消跳过归属校验（该订单不属于任何用户），归属校验只保留在用户主动取消路径 —— 早期消费者写死某个 `userId` 导致其余用户全部取消失败，已被回归用例钉死
- 解耦设计：MQ 包不依赖业务层，通过接口由调用方注入，接口倒置

### 6. 可靠消息双保险 —— outbox 意图账本 + 超时扫描兜底

MQ 只保证"进了 Broker 且已落盘就不丢"，**不保证"发送意图"本身不丢**：订单事务提交后如果进程崩溃，"该发一条延迟消息"这个意图只存在于内存里。解法是把它换成数据库里的一行。

- **outbox 本地消息表**：下单事务内写消息行（`message_id` 唯一索引做幂等键），投递器独立轮询投递 —— 消息不再依赖发送那一刻的网络与进程存活
- **投递器**：`1s` 轮询、单轮 `LIMIT 100`、失败按 `min(2^n 秒, 5min)` 指数退避；分布式锁排除多实例在同一轮内并发投递
- **语义诚实声明为 at-least-once**：`MarkSent` 在 Publish **之后**执行，因此"Publish 成功但 MarkSent 失败"必然重发。能承诺的是**下游仅一次生效**（消费侧按 `message_id` 去重 + 唯一索引兜底），不是"零重复投递"
- **超时扫描任务**：MQ 消费者对取消失败是 **Ack 丢弃、有意不重试**，所以一次 DB 瞬时超时就会让订单永久停在 `pending_pay`、库存永久锁定。扫描任务不依赖 MQ，按 `created_at + TTL` 周期收敛 —— **它是这条链路唯一的安全网，不是可有可无的备份**
- **结果分流**：`cancelled` / `skipped`（已被 MQ 处理或用户已支付，正常）/ `failed`（真故障，值得告警）三态分开计数，避免"正常跳过"淹没告警
- **幂等闸与亮点 4 同构**：取消时先用条件 UPDATE 抢占订单状态（`RowsAffected=0` 即已被并发取消，直接返回），再退券、释放库存；不用事务外"先查后做"

### 7. 跨服务一致性 —— Saga 编排 + 每步幂等补偿

拆分把"一个本地事务"变成了"跨三个服务的调用链"。这里不用分布式事务，而是 **Saga（正向编排 + 反向补偿）**：

```
下单: 锁库存(product) ─▶ 核销券(marketing) ─▶ 写订单(trade, 本地事务)
         │失败                  │失败
         ▼                      ▼
      无需补偿            补偿: 释放库存(product)
```

- **每一步都是幂等 RPC**：补偿可以安全重试，这是选"幂等接口 + 补偿"而不是"两阶段提交"的前提
- **补偿失败不让主流程假装成功**：订单落库前任何一步失败都回滚本地事务并把错误上抛；补偿本身失败会记入日志与订单日志，由超时扫描兜底收敛
- trade → product / marketing 是**单向依赖**，不存在循环；服务之间不反向调 BFF

### 8. 幂等下单 + 雪花算法分布式 ID

- 下单接口携带 `idempotent_key`，订单主表上建**唯一索引**，前端重复提交直接返回已有订单，杜绝脏单
- 自研**雪花算法**：标准 64 位划分（41 时间戳 / 10 workerId / 12 序列号），处理**同毫秒序列溢出自旋**与**时钟回拨回退**；订单号 = `"DS" + 雪花 ID 转 36 进制`
- 最大事务将「核销优惠券 → 写订单主表 → 写明细 → 逐项锁库存 → 删购物车 → 写订单日志」在 trade-service 内原子完成
- 拆分后 **WorkerId 成为部署配置**：单实例保持 1；扩多实例时每实例必须不同，否则雪花 ID 碰撞（表现为订单号唯一约束冲突，而非静默脏数据 —— 这点比静默好，但排查时要知道是这个原因）

### 9. Redis 缓存体系 —— 一致性策略 + 版本号失效 + 弱依赖降级

| 缓存键 | TTL | 说明 |
|---|---|---|
| `user:perm:{id}` | 30 min | 用户权限码，省去每次请求的关联查询 |
| `api:perm:v{version}:{method}:{path}` | 30 min | 接口权限点（单体时代使用；现改为生成期固化，见亮点 2） |
| `product:detail:{user/admin}:{id}` | 10 min | 商品详情（用户端/管理端视图分离） |
| `sku:stock:{id}` | 30 s | SKU 实时库存，短 TTL 兜底 |
| `category:{id}` | 1 h | 类目树 |

- **一致性策略**：Cache-Aside 读模式（miss 回源回写）+ **写库后删除** + 短 TTL 兜底
- **版本号批量失效**：接口权限缓存键带版本号，权限变更时 `INCR api:perm:version` 一次失效全部接口缓存，避免逐 key 删除。拆分为微服务后缓存**跨服务共享同一个 Redis** —— 失效仍是一次原子操作，这是不拆 Redis 的直接收益
- **弱依赖降级**：Redis 不可用时打 WARN 后业务直查 DB，缓存层全部判空处理，**不阻塞主流程**

### 10. Elasticsearch 商品搜索 —— 增量 + 全量双周期对账

- 业务写路径同步写 ES，失败仅 WARN 降级不阻塞主流程，**靠对账任务保证最终一致**
- **增量对账（5 分钟）**：基于 Redis 时间水位（RFC3339）拉取变更商品重新索引，**水位在写入成功后才推进**（否则漏数据）
- **全量对账（12 小时）**：全量比对 DB 与 ES 的 SPU ID 差集，重建缺失文档、清理孤儿文档
- **搜索可整体关闭**：`ES.Addresses` 留空即不启用（商品搜索降级为直查 DB，同时不启动对账任务）—— 对账任务随搜索一起关，避免"没有索引却在对账"

### 11. 支付网关抽象 —— 策略模式 + 回调多重校验

- 定义 `PayGateway` 接口（创建支付 / 解析回调 / 主动查询）+ 注册中心，业务层只依赖接口；当前 Mock 实现已跑通支付闭环，接入微信/支付宝零业务改动
- 支付回调安全链：**签名验证（真实网关）→ 幂等检查（已 success 直接返回）→ 金额双重校验（回调金额 = 支付单金额 = 订单应付）→ 状态机校验（仅待支付可支付）→ 单事务落库**（更新支付单 → 更新订单 → 写日志 → 逐项扣减库存）

### 12. RBAC 权限体系 —— 六维数据模型 + 三级权限控制

- **数据模型**：用户-角色、角色-权限、角色-菜单、菜单-权限、用户-部门、数据权限（scope）六类关联，全部落在 user-service 一个库里 —— 关联校验不需要跨服务
- **接口级鉴权**：见亮点 2（生成期固化权限码）
- **菜单驱动动态路由**：后端返回菜单树，前端动态生成路由与侧边栏；页面按权限码控制按钮显隐，前后端双重校验
- **两级缓存**：用户权限码与接口权限点分别缓存，权限变更走版本号批量失效（见亮点 9）

### 13. 文件上传 —— 对象存储取代共享卷

单体时代文件写本地盘，靠"两个实例挂同一个 volume"在单机 compose 下成立。上 K8s / 多机后那个前提就没了（同路径不再是同一块盘），故改对象存储（MinIO / S3 协议）。

- **上传直落 S3**：BFF 收 multipart 后转存对象存储并返回对象 URL；不落本地盘，BFF 保持无状态
- **两个地址必须分开**（本域最容易配错的一处）：一个供服务端上传（容器内走服务名），一个用于生成给前端的 URL —— 后者**必须是浏览器可达的地址**。配成同一个，前端图片会全部裂开
- **对象键优先用文件 MD5**：同一张图重复上传天然覆盖自己，不在桶里堆垃圾
- **扩展名白名单**：落点从"后端静态目录"变成"匿名可读的对象存储"后，任意 `html`/`svg` 都能被浏览器直接打开，是存储型 XSS 入口，故收紧（单体不校验）
- **一个反直觉的实现约束**：go-zero 的请求解析**不支持 multipart 文件**（其映射解析器里没有任何文件字段处理），所以上传 handler 是手写的，文件用 `r.FormFile` 取

### 14. 请求契约的坑（工程细节，值得单列）

拆分过程中撞到的、**编译期看不见**的契约问题，全部已在代码注释里留下成因：

- **请求方向不能用空接口类型**：go-zero 的映射解析器按「声明类型的 Kind == 实际值的 Kind」判定，而空接口的 Kind 是 `reflect.Interface`，与任何具体 Kind 都不相等 ⇒ 任何非数字值都必然报 `type mismatch` → 400。实测踩过（商品规格模板字段因此建品直接 400）。故请求侧一律给具体形状（结构体切片 / 字符串映射）
- **未声明的字段被静默丢弃**：建品请求里的商品/SKU 状态字段若不在 `.api` 声明，会被解析器丢掉 → 空串 → 撞数据库的 CHECK 约束 → 表现为 503（像依赖挂了，其实是契约缺字段）
- **局部更新的"传没传"要用指针表达**：非指针字段把"没传"与"传了零值"抹成同一个值 —— 于是"隐藏类目"（`false`）与"移回根层级"（`0`）都表达不出来。单体靠把请求体绑成 `map[string]interface{}` 天然保留字段是否存在，生成物只能用指针补回来
- **HTTP 契约 ≠ RPC 契约**：字段改名、JSON 文本 ↔ 对象、`optional` 对布尔不产生指针……转换全部按域集中

### 15. 一次真实的排障 —— 401 让前端把用户踢下线

BFF 的判权在取不到身份时**原本回 401**。语义上"未登录"没错，但状态码选错了：

```
某个管理端路由漏挂认证中间件
  → 判权回 401 "未登录"
  → 前端把 401 一律当作"令牌过期",去调刷新接口
  → 刷新也失败(令牌本没问题,是路由配置错了)
  → **前端清空用户令牌并跳登录页**
  → 之后所有请求都没有 Authorization 头 → 全部"未登录"
```

用户看到的是"突然所有接口都未登录"，而真实原因只是**一个路由的中间件配错了**。排查代价很高：症状与"令牌失效"完全一致，最后靠 **curl 逐条对比**才确认后端正常、问题在前端令牌已被清空。

**修法**：改为 **500 + "服务配置错误:该接口未挂认证中间件"**。原则是 ——
**401 留给"凭据真的有问题"（前端跳登录是期望行为），500 用于"服务端自己配错了"**。

### 16. 工程化与代码质量

- **版本化数据库迁移**：golang-migrate 管理各服务的迁移（up/down 可回滚），表结构、约束、索引全部落在 SQL；**种子与迁移分离**（版本账本表，事务化记账、可重入）
- **迁移与就绪校验**：每个服务带独立的迁移命令（可校验 schema 版本与表清单、可打印各表行数供割接对账）。**用显式表清单而非"数表个数"** —— 迁移跑成功但漏建一张表，数个数是查不出来的
- **PostgreSQL 特性利用**：部分唯一索引（软删除不占唯一约束）、CHECK 约束防脏数据、幂等键唯一索引
- **容器化**：五个服务均多阶段构建，**构建上下文是仓库根**（服务依赖共享的 proto 生成模块，靠 Go workspace 才编译得过）；镜像内置入口脚本，启动前自检二进制 / 配置 / 密钥，并**警告配置里残留的 `127.0.0.1`**（那是"挂成了本地配置"的典型症状）
- **CI 门禁**：按模块矩阵并行执行 `go build` + `go vet`（**不能用仓库根的 `go build ./...`** —— 根目录不是模块），外加一个 workspace 一致性 job
- **统一响应 / 错误码**：`{code, message, data}` 信封；200 成功、400 业务失败、401 凭据、403 无权限、429 限流、500 内部、503 下游不可用

### 17. 并发正确性验证 —— 真实 PG 并发测试 + race 检测 + CI 门禁

- **并发/幂等/一致性用例**：300 并发抢 100 张券**恰好发出 100 张**（发放计数 / 账本行数 / 响应数三方复核）、同用户限领、库存防超卖、按「订单 + SKU + 操作类型」流水幂等、多 SKU 同订单回归、超时取消幂等与并发抢占、闸门两阶段收敛
- 断言精确到数字并**回查数据库账本**，不只看返回值；发令枪（barrier）保证竞争窗口重叠；测试基建自带独立测试库引导与数据工厂
- `go test -race` 全绿，由 GitHub Actions 在 Linux 上强制执行
- 测试驱动出真实缺陷：幂等键未含 SKU 导致多 SKU 订单绕过库存校验 —— 已修复并有回归用例钉死

### 18. 热点路径 Redis 预扣闸门 —— Lua 原子预扣 + 对账收敛（压测 10×）

在领券/扣库存入口增加 Redis Lua 预扣闸门：**闸门挡量，DB 账本保真**。

- **O(1) 拒绝无效流量**：Lua 脚本原子完成「查余量-校验-扣减」，售罄/超限请求不触碰数据库；DB 悲观锁+条件 UPDATE 双防线原样保留为正确性锚点
- **失败补偿**：DB 事务失败时 Lua 补偿脚本归还闸门额度，标记位区分「扣了没还」与「没扣就还」
- **对账收敛**：定时任务以 DB 为准单向修正闸门计数（SETNX 回填 + 偏差强制对齐），闸门允许瞬时偏差、最终一致
- **一键熔断**：运行时开关，Redis 故障时自动降级直走 DB，不重启实例

**压测对比**（200 并发，售罄拒绝场景，3 轮中位）：

| 指标 | 纯 DB 双防线 | Redis 闸门态 | 提升 |
|---|---|---|---|
| QPS | 707 | **7135** | **10.1×** |
| p50 / p99 | 237ms / 927ms | **22.9ms / 93.9ms** | 10.3× / 9.9× |

> 放行路径刻意设计为持平（发放 2000 张：23.2s → 24.9s）——闸门的价值是挡量而非加速放行，放行仍受 DB 事务约束保证正确性。压测后 SQL 三方复核（发放计数=账本行数=成功响应数）0 超发 0 漏发。

### 19. 限流与定时任务分布式锁 —— 令牌桶 + Redsync

- **两级令牌桶限流**：BFF 层按 IP 对登录/注册/领券收紧（防爆破），429 携带 `Retry-After`；惰性建桶 + 周期清理空闲桶防内存泄漏
- **定时任务分布式锁**：ES 对账、库存/闸门对账、outbox 投递各任务跨实例互斥（抢不到即跳过、到期自释放防死锁），Redis 不可用降级进程内互斥不停摆
- 多实例语义诚实声明：进程内限流在多实例下为每实例各限一份，跨实例全局限流是 Redis Lua 的演进方向

### 20. Prometheus 可观测体系 —— 指标随所有权迁移

- **业务指标跟着数据所有权走**：领券埋点从单体迁到 marketing-service —— 迁的原因是那个位置只能观察到"RPC 这一跳"，而 `path` 标签（闸门判定 vs 降级走 DB）在调用方必然是假值：**闸门跑在被调进程里，调用方读不到闸门是否生效**
- **独立指标端口**：go-zero 的 gRPC 服务无法在同端口再服务 HTTP，故指标另开端口
- **指标基数的坑**：HTTP 层 route 标签强制使用**路由模板**而非实际 path，未匹配请求归入 `UNMATCHED`，杜绝基数爆炸



---

##  核心链路设计

**领券并发控制**（详见[亮点 4](#4-优惠券领取并发控制--悲观锁--乐观锁双防线)）：

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

**下单 → 跨服务 Saga → 超时取消**（详见[亮点 7](#7-跨服务一致性--saga-编排--每步幂等补偿)）：

```
BFF ──▶ trade-service
          │ 幂等检查(idempotent_key) ──▶ Saga 编排:
          │      ① 锁库存(product, 幂等 RPC)
          │      ② 核销券(marketing, 幂等 RPC)
          │      ③ 写订单 + 明细 + 日志 + outbox(本地事务)
          │            └─ 任一步失败 → 反向补偿已成功的步骤
          ▼
        COMMIT ──▶ outbox 投递器(1s 轮询) ──▶ 15min 延迟消息
                                                    │ TTL 到期死信
                                                    ▼
                            消费 ──▶ 状态机校验(仅 pending_pay) ──▶ 取消订单
                                                                     ├─ 释放库存(幂等 RPC)
                                                                     └─ 归还优惠券(幂等 RPC)
                            超时扫描任务 ──▶ 兜底收敛(MQ 不可用时的唯一安全网)
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
| 商品运营 | 类目树、SPU/SKU 管理、图片上传(对象存储)、库存台账与流水日志 |
| 订单管理 | 全状态流转、发货、退款处理(库存/优惠券自动归还) |
| 优惠券运营 | 模板创建(满减/直减、有效期双模式)、发放统计 |

---

## 技术栈

| 端 | 技术 |
|----|------|
| 网关 | **BFF**：111 条路由 · 统一鉴权 · 协议转换 · 限流 |
| 服务框架 | [go-zero](https://go-zero.dev) 1.10（gRPC + HTTP）· [gRPC](https://grpc.io) · [Protocol Buffers](https://protobuf.dev) · goctl 1.9.2 |
| 数据层 | PostgreSQL 14（**四个库 / 28 张业务表**）· Redis 6 · Elasticsearch 7.17 · RabbitMQ 3.13 · MinIO（S3） |
| 中间件 | JWT 鉴权 · Redis 缓存 · MQ 延迟消息 · ES 全文检索 · 对象存储 · 服务发现(etcd) |
| 前端 | [Vue 3](https://vuejs.org) `<script setup>` · [TypeScript](https://www.typescriptlang.org) · [Vite](https://vitejs.dev) · [Element Plus](https://element-plus.org) · [Pinia](https://pinia.vuejs.org) · Tailwind CSS |
| 基础设施 | Docker 多阶段构建 · docker compose 一键全栈 · GitHub Actions CI |

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
│                      BFF (go-zero rest · :9003)                 │
│   认证中间件(JWT 验签) → 判权 → 协议转换 → 限流                  │
│   111 条路由 · 78 条管理端路由挂判权 · 权限码生成期固化            │
└───┬──────────────┬──────────────┬──────────────┬───────────────┘
    │ gRPC         │ gRPC         │ gRPC         │ gRPC
    ▼              ▼              ▼              ▼
┌─────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐
│  user   │  │ product  │  │  trade   │  │  marketing   │
│  :9004  │  │  :9005   │  │  :9007   │  │    :9009     │
│─────────│  │──────────│  │──────────│  │──────────────│
│ 用户    │  │ 类目/SPU │  │ 购物车   │  │ 券模板/用户券 │
│ 角色    │  │ SKU/图片 │  │ 订单     │  │ Redis 预扣闸门│
│ 权限    │  │ 库存流水 │  │ 支付     │  │ 闸门对账任务  │
│ 菜单    │  │ ES 搜索  │  │ Saga 编排│  │ 领券指标      │
│ 部门    │  │ ES 对账  │  │ outbox   │  │              │
│ 数据权限│  │ 库存对账 │  │ 超时扫描 │  │              │
│ 地址    │  │          │  │          │  │              │
└────┬────┘  └────┬─────┘  └────┬─────┘  └──────┬───────┘
     │            │             │               │
     ▼            ▼             ▼               ▼
  user_db     product_db     trade_db     marketing_db
  (15 表)      (5 表)         (5 表)         (3 表)
     └────────────┴─────────────┴───────────────┘
                          │
        etcd(服务发现) · Redis(缓存/闸门) · RabbitMQ(延迟消息)
        MinIO(对象存储) · Elasticsearch(商品搜索·可选)
```

**请求链路**：

```
请求 → nginx(前端镜像) → BFF
     → middleware：请求元信息(请求 ID / 日志 / 路由模板注入) → 认证(JWT 验签)
     → 判权：用户权限码 × 路由上固化的权限码
     → logic：参数校验 / proto 转换 / 一次 RPC
     → 下游服务：业务规则 / 事务 / 状态机 / 并发控制 → 数据访问(行锁 / 条件更新)
```

---

## 项目结构

```
demo-shop/
├── api/                        # proto 定义与生成代码(各服务共享)
│   ├── proto/{user,product,trade,marketing}/v1/
│   └── gen/                    # protoc 生成物(Go)
├── pkg/                        # 预留的共享包目录
├── services/
│   ├── bff/                    # 唯一对外入口(HTTP :9003)
│   │   ├── bff.api             # 路由与类型定义(111 条路由 / 25 个前缀块)
│   │   ├── cmd/genroutes/      # 生成期固化权限码的代码生成器
│   │   ├── internal/middleware/    # 请求元信息 / 认证 / 限流 / 公开路由
│   │   ├── internal/guard/         # 判权 + 路由模板注入
│   │   ├── internal/infra/rpc/     # gRPC 薄封装 + 错误三态分类
│   │   ├── internal/storage/       # 对象存储(MinIO/S3)
│   │   ├── internal/logic/         # 按域分目录的协议转换与 RPC 调用
│   │   └── internal/handler/       # 生成的路由注册(判权包装由生成器注入)
│   ├── user/                   # user-service(gRPC :9004)
│   │   ├── internal/logic/     # 按域分目录：用户/角色/权限/菜单/部门/数据权限/地址
│   │   ├── internal/server/    # gRPC 服务实现(逻辑薄层，转调 logic)
│   │   ├── internal/repository/# 数据访问
│   │   ├── cmd/migrate/        # 迁移 + schema 就绪校验 + 行数对账
│   │   ├── migrations/         # 19 个版本化迁移(up/down)
│   │   └── Dockerfile
│   ├── product/                # product-service(gRPC :9005)
│   │   ├── internal/logic/     # 类目 / 商品 / 库存三个域
│   │   ├── internal/infra/es/  # ES 搜索
│   │   ├── internal/task/      # ES 增量/全量对账、库存对账
│   │   └── migrations/         # 3 个版本化迁移
│   ├── trade/                  # trade-service(gRPC :9007)
│   │   ├── internal/service/   # 订单 / 购物车 / 支付 / Saga 编排
│   │   ├── internal/infra/mq/  # RabbitMQ 客户端、延迟消息、outbox
│   │   ├── internal/task/      # outbox 投递、延迟消费、超时扫描
│   │   └── migrations/         # 5 个版本化迁移
│   └── marketing/              # marketing-service(gRPC :9009)
│       ├── internal/logic/     # 券域(领券双防线)
│       ├── internal/infra/gate/ # Redis 预扣闸门(Lua)
│       ├── internal/task/      # 闸门对账
│       └── migrations/         # 2 个版本化迁移
├── docker/
│   ├── entrypoint.sh           # 五个镜像共用的入口脚本(启动前自检)
│   ├── postgres-init/          # 首次初始化时建出四个库
│   ├── prometheus.yml          # 指标抓取配置
│   └── docker-init.yml         # 仅中间件(本地开发用)
├── demo_shop_front/            # Vue3 前端
│   ├── nginx.conf.template     # 生产镜像的 nginx 配置(反代 BFF)
│   └── src/
│       ├── views/platform/     # 管理端业务页(商品/订单/库存/支付/优惠券)
│       ├── views/system/       # 管理端 RBAC 页(用户/角色/权限/菜单/部门/数据权限)
│       ├── views/shop/         # 用户端页(商品/购物车/结算/订单/优惠券)
│       ├── views/layout + userLayout # 管理端 / 用户端布局(动态侧边栏)
│       ├── pinia/modules/      # 状态管理(动态路由 / 用户权限码)
│       ├── utils/              # axios 拦截器(令牌注入 / 401 刷新)
│       └── api/                # 接口封装
├── docker-compose.yml          # 全栈编排(中间件 + 5 服务 + 前端)
└── go.work                     # Go workspace(7 个模块)
```

---

## 快速开始

### 方式 A：一键全栈（推荐）

```bash
git clone <repo> && cd demo-shop
docker compose up -d --build      # 中间件 + BFF + 4 服务 + 前端
docker compose ps                 # bff 应 healthy,迁移容器应 Exited (0)
```

浏览器访问 **http://localhost:8080** （nginx 入口，前端与 API 同源，`/api/v1/` 反代到 BFF）。

| 服务 | 地址 | 说明 |
|---|---|---|
| 前端 + API 入口 | http://localhost:8080 | nginx：静态资源 + `/api/v1/` 反代到 BFF |
| BFF | http://localhost:9003 | 直连时用（前端经 nginx 同源访问，不跨域） |
| MinIO 控制台 | http://localhost:9001 | 对象存储（demoshop/demoshop123），bucket 由 BFF 启动时自动创建 |
| RabbitMQ 管理台 | http://localhost:15672 | 消息队列（demoShop/demoShop） |
| Prometheus / Grafana | http://localhost:9090 / :3000 | **需显式启用**：`docker compose --profile observability up -d` |

容器启动顺序由 healthcheck 门禁控制：

```
中间件 healthy → 迁移容器(一次性,建表 + 校验) → 四个服务 → BFF healthy → 前端
```

> **迁移容器是硬门禁**：服务等待它们"成功退出"。若库不存在，迁移会失败并让整条链卡住 ——
> 四个库由初始化脚本在**数据卷首次初始化时**建出。已有数据卷的环境该脚本不会执行（库早就在了）；
> 若那种情况下缺库，需手工建库或删卷重建。

> **为什么可观测性是可选 profile**：Elasticsearch + Prometheus + Grafana 合计约
> 2.5~3G 内存，而它们对"把服务跑起来"不是必需的 —— product-service 的搜索地址留空
> 即不启用搜索（降级为直查 DB），指标端点仍在各服务上暴露。

### 方式 B：本地开发（中间件容器 + 本地进程）

> 前置条件：Docker Desktop、Go 1.25、Node.js 18+ & npm

**第 1 步：起中间件**

```bash
docker compose up -d postgres redis rabbitmq etcd minio
```

**第 2 步：跑迁移**

```bash
cd services/user      && go run ./cmd/migrate -f etc/user.yaml      && cd ../..
cd services/product   && go run ./cmd/migrate -f etc/product.yaml   && cd ../..
cd services/trade     && go run ./cmd/migrate -f etc/trade.yaml     && cd ../..
cd services/marketing && go run ./cmd/migrate -f etc/marketing.yaml && cd ../..
```

**第 3 步：启动四个后端服务与 BFF**（各开一个终端）

```bash
cd services/user      && go run .    # :9004
cd services/product   && go run .    # :9005
cd services/trade     && go run .    # :9007
cd services/marketing && go run .    # :9009
cd services/bff       && go run .    # :9003(HTTP,对外入口)
```

> 各服务的本地配置在 `services/<名>/etc/<名>.yaml`（地址为 `127.0.0.1`）。
> 容器内用的是 `etc/<名>.docker.yaml`（地址为 compose 服务名）—— 两份配置只差地址字段。

**第 4 步：启动前端**

```bash
cd demo_shop_front
npm install
npm run dev
```

浏览器访问 http://localhost:5173 。

> **改了 `.api` 之后必须重跑两个命令**（顺序不能反）：
> ```bash
> cd services/bff
> goctl api go -api bff.api -dir . --style gozero --home ../../.goctl
> go run ./cmd/genroutes
> ```
> 前者会**覆盖**生成的路由文件，把生成器注入的判权包装全部清掉；
> 后者按 `.api` 的 111 条路由重新查库并注入。只跑前者 = 判权静默失效。

### 运行测试

```bash
# 需要一个可用的 PostgreSQL
docker compose up -d postgres redis

cd services/product   && go test ./... -race
cd services/trade     && go test ./... -race
cd services/marketing && go test ./...
```

### 停止服务

```bash
docker compose down                                   # 方式 A
docker compose --profile observability down           # 含可观测性时
docker compose down -v                                # 连数据卷一起删(会丢数据)
```

> 本地体验账号见 user-service 的迁移种子数据。

---

## 关于单体

项目最初是 Gin 单体（`demo-shop-back/`），已按数据所有权完成微服务拆分并**退役**：

- 代码归档在 **`archive/monolith` 分支** 与 **`monolith-final` tag**
- 退役前 BFF 已覆盖全部四个下游服务，前端早已指向 BFF；退役时**无任何代码引用单体**（仅两处注释）
- 删除暴露了一个隐藏的反向依赖：product-service 的一批依赖（ES 客户端、ORM、迁移工具等）原本由**单体的依赖清单**提供，借 Go workspace 的模块图间接解析
---

## License

[Apache License 2.0](LICENSE)
