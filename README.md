# Shortlink

面向营销短信 / App 推送场景的短链接服务：长链转短链、302 跳转、过期管理与点击统计，并提供同源演示页便于联调与面试演示。

**技术栈：** Go · [go-zero](https://go-zero.dev/) · MySQL · Redis（缓存 / Bloom / HyperLogLog / 限流）· Kubernetes（kind 演示）

---

## 亮点

- **定长短码**：发号器 → Salt 模乘双射混淆 → 定长 6 位 Base62，降低可预测枚举
- **幂等建链**：`scene_id + 规范化 URL` 哈希去重；DB 唯一约束兜底并发冲突
- **读路径防护**：Bloom 防穿透 + Redis 缓存 + SingleFlight 防热点击穿
- **点击统计**：异步 PV（精确）+ HyperLogLog 近似 UV
- **接口边界**：管理接口 `X-API-Key`；跳转公网开放，过期/不存在/禁用统一纯 404

---

## 架构概览

```text
调用方 --X-API-Key--> POST /api/v1/links ──► MySQL（映射表 / 发号）
                                      └──► Redis（幂等缓存 / Bloom）

用户点击 ------------> GET /{code}
                         │
                         ├─ Bloom → Cache → DB（SingleFlight）
                         ├─ 302 → 长链
                         └─ 异步：PV / UV / lastClickAt
```

| Redis Key | 用途 |
|-----------|------|
| `surl:c:{code}` | 跳转缓存 |
| `surl:h:{hash}` | 幂等缓存 |
| `surl:bf` | Bloom Filter |
| `surl:pv:{code}` | PV |
| `surl:uv:{code}` | HyperLogLog UV |
| `surl:la:{code}` | 最近点击时间 |

---

## 设计取舍

| 决策 | 为什么 | 代价 / 边界 |
|------|--------|-------------|
| 发号 + 混淆 + Base62，而不是 UUID / MD5 截断 | 定长可控、可逆映射、容量与碰撞可估算；Salt 降低顺序可猜 | Salt 变更会影响新码空间；轮换需带 version（未做） |
| 读路径加重防护，写路径相对简单 | 短链跳转是热路径；创建量远低于点击量 | 写侧未做分布式事务；依赖唯一键 + 查重 |
| UV 用 HyperLogLog | 省内存、适合海量点击去重示意 | **近似值**，非精确 UV |
| 管理鉴权用 API Key，跳转不鉴权 | 符合短信/推送「点开即跳」场景 | Key 泄露可滥发短链；生产应配合网关/配额 |
| Compose 单机一键跑通 | 面试官 / 自己可快速复现 | 非生产高可用；无多副本、无自动 HTTPS |

---

## 快速开始（Docker，推荐）

需 Docker Compose。请在**仓库根目录**执行：

```bash
cp .env.example .env          # Windows: Copy-Item .env.example .env
docker compose up -d --build
docker compose ps             # api / mysql / redis 应为 Up（healthy）
```

**验收（约 1 分钟）：**

1. `curl http://127.0.0.1:8890/health` → `ok`（端口见 `.env` 的 `API_HOST_PORT`，默认 `8890`）
2. 打开 http://127.0.0.1:8890/ ，API Key 填 `.env` 中 `SHORTLINK_API_KEY`（默认 `change-me`）
3. 创建短链 → 打开短链应 302 → 刷新统计见 PV/UV

停止：`docker compose down`（加 `-v` 清空数据卷）。

敏感配置可由环境变量覆盖 yaml：`SHORTLINK_DSN` / `SHORTLINK_REDIS` / `SHORTLINK_API_KEY` / `SHORTLINK_BASE62_SALT` / `SHORTLINK_DOMAIN`。

云主机可复用同一 Compose：安全组只开 22/80/443，Nginx 反代 `127.0.0.1:${API_HOST_PORT}`（默认 `8890`，见 [`deploy/nginx.conf.example`](deploy/nginx.conf.example)），并设置 `SHORTLINK_DOMAIN` 为真实访问地址。

> 不用 Docker、本机 `go run` 的步骤见文末 [本地开发](#本地开发可选)。

**Kubernetes（kind）：** 独立集群一键部署 API + MySQL + Redis，见 [`deploy/k8s/README.md`](deploy/k8s/README.md)。

---

## API

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| `POST` | `/api/v1/links` | `X-API-Key` | 创建短链 |
| `GET` | `/api/v1/links/{shortCode}` | `X-API-Key` | 查询元数据 |
| `GET` | `/api/v1/links/{shortCode}/stats` | `X-API-Key` | PV / UV（Redis 热路径） |
| `POST` | `/api/v1/assistant/chat` | `X-API-Key` | 分析助手（SSE） |
| `GET` | `/{shortCode}` | 无 | 302 或统一 404 |
| `GET` | `/` | 无 | 演示页 |
| `GET` | `/health` | 无 | 健康检查 |

```bash
curl -s -X POST http://127.0.0.1:8890/api/v1/links \
  -H "Content-Type: application/json" \
  -H "X-API-Key: change-me" \
  -d '{"longUrl":"https://www.example.com/hello","expireDays":90}'
```

- `created=true` 新建；`false` 幂等命中  
- `expireDays=0` 表示永不过期（`expireAt` 为 `null`）

---

## 关键代码入口

| 主题 | 路径 |
|------|------|
| 路由 / 演示页 | [`internal/handler/routes.go`](internal/handler/routes.go) |
| 创建短链 / 幂等 | [`internal/logic/createlinklogic.go`](internal/logic/createlinklogic.go) |
| 跳转 / Bloom / SingleFlight | [`internal/logic/redirectlogic.go`](internal/logic/redirectlogic.go) |
| Salt 双射混淆 | [`pkg/codec/mixer.go`](pkg/codec/mixer.go) |
| 发号器 | [`sequence/`](sequence/) |
| 依赖组装 | [`internal/svc/servicecontext.go`](internal/svc/servicecontext.go) |

---

## 项目结构

```text
shortlink/
├── main.go                 # 入口（SHORTLINK_* 环境变量覆盖）
├── Dockerfile / docker-compose.yml / .env.example
├── shorturl.api            # go-zero API 定义
├── web/                    # 演示页
├── deploy/                 # MySQL init、Nginx 示意、k8s（kind）清单
├── internal/               # handler / logic / middleware / svc / cachex / stats / channel / assistant
├── model/ · sequence/ · pkg/
├── short_url_map.sql · sequence.sql · click_events.sql
```

---

## 测试

```bash
go test ./pkg/... ./sequence/...
```

---

## 安全说明

- 勿提交真实 `.env` / `etc/shorturl-api.yaml`；仓库仅保留 example
- 演示页 API Key 仅用于本地 / 受控演示
- 跳转侧结合短码混淆、统一 404、Bloom 与 IP 限流降低盲扫成本
- UV 为 HyperLogLog **近似值**

---

## 分析助手（Eino Tool Calling + SSE）

自然语言问运营指标；模型只能通过工具读 MySQL `click_events`，禁止编造。演示页下方「分析助手」支持多轮与 SSE 流式输出。

| 环境变量 | 含义 |
|----------|------|
| `SHORTLINK_CLICK_LOG` | `true` 时跳转异步写入 `click_events` |
| `SHORTLINK_SEED_DEMO` | `true` 时启动 seed 演示短链与近 8 日点击 |
| `SHORTLINK_LLM_API_KEY` | DeepSeek（或兼容）API Key（项目专用，勿提交） |
| `SHORTLINK_LLM_BASE_URL` | 默认 `https://api.deepseek.com` |
| `SHORTLINK_LLM_MODEL` | 默认 `deepseek-flash` |

**口径：** 助手以 MySQL 为准；Redis PV/HLL 仍为跳转热路径。  
**时区：** 「昨天 / 近 7 日」按 `Asia/Shanghai`。  
**渠道：** `sms`→短信，`app_push`→App 推送，`wechat`→微信；演示短码 `demos1` / `demop1` / `demow1`。

**接口：** `POST /api/v1/assistant/chat`（`X-API-Key`，SSE）

```bash
curl -N -X POST http://127.0.0.1:8890/api/v1/assistant/chat \
  -H "X-API-Key: change-me" \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{"sessionId":"","message":"昨天哪个推广渠道点击更高？"}'
```

SSE 事件：`session` → `status` → `delta`* → `error?` → `done`。

**可问：** 短码总 PV；昨天渠道排行；近 7 日趋势（全站/短码）；短码 UV + 最近点击。

本地验收 CLI：`go run ./cmd/toolcheck`、`go run ./cmd/agentcheck`（需 LLM Key + MySQL `127.0.0.1:3307`）。

---

## 边界与后续

**当前定位：** 可演示的单机短链 + 自然语言分析助手（Compose 一键启动），并提供 kind 上的 Kubernetes 清单，适合作为 Go 后端面试项目。

**已知边界：** Compose / kind 均为演示向；集群内 MySQL/Redis 用 emptyDir；API Key 级鉴权；Salt 不支持在线轮换；UV 分析为库表精确去重（演示级 visitor hash）。

**可扩展方向：** 云托管 MySQL/Redis + Ingress/HTTPS；API 多副本（关 Seed）；Salt version；按活动拆码与运营后台。

---

## 本地开发（可选）

```bash
# 1. 依赖：本机 MySQL / Redis；执行 short_url_map.sql、sequence.sql
# 2. 配置
cp etc/shorturl-api.example.yaml etc/shorturl-api.yaml   # 编辑 DSN / Redis / ApiKey / ShortDomain
# 3. 启动
go run .
```

演示页与健康检查同上：http://127.0.0.1:8890/ 、`/health`（`go run` 默认仍是进程内 `8888`，见 `etc/shorturl-api.yaml`）。
