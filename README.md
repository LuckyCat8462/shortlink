# Shortlink

面向营销短信 / App 推送场景的短链接服务。支持长链转短链、302 跳转、过期管理与点击统计，并提供同源演示页便于联调与面试演示。

**技术栈：** Go · [go-zero](https://go-zero.dev/) · MySQL · Redis（缓存 / Bloom / HyperLogLog / 限流）

---

## 功能特性

- **定长短码**：发号 → Salt 模乘双射混淆 → 定长 6 位 Base62，降低可预测枚举
- **幂等建链**：`scene_id + URL` 哈希去重；`scene_id` 为空表示全局，表结构预留活动维度
- **安全跳转**：过期 / 不存在 / 禁用统一返回纯 404；跳转按 IP 限流
- **高可用读路径**：Bloom 防穿透 + Redis 缓存 + SingleFlight 防热点击穿
- **点击统计**：异步记录 PV，HyperLogLog 近似 UV，提供 stats 查询接口
- **管理鉴权**：转链 / 查询 / 统计使用 `X-API-Key`；跳转无需鉴权
- **演示页面**：`GET /` 同源静态页，可完成转链、打开短链、刷新统计

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

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 5.7+ / 8.x
- Redis 6+

### 1. 克隆与依赖

```bash
git clone https://github.com/<your-name>/shortlink.git
cd shortlink
go mod download
```

### 2. 初始化数据库

```sql
CREATE DATABASE IF NOT EXISTS `shortlink` DEFAULT CHARACTER SET utf8mb4;
USE `shortlink`;
```

执行仓库根目录 SQL：

- [`short_url_map.sql`](short_url_map.sql)
- [`sequence.sql`](sequence.sql)

### 3. 配置

从示例配置复制一份本地配置并编辑（`etc/shorturl-api.yaml` 已被 gitignore，不会提交）：

```bash
cp etc/shorturl-api.example.yaml etc/shorturl-api.yaml
```

至少检查以下项：

| 配置项 | 说明 |
|--------|------|
| `ShortUrlDB.DSN` / `SequenceDB.DSN` | MySQL 连接串 |
| `CacheRedis` | Redis 地址 |
| `ShortDomain` | 对外短链前缀，须与真实访问地址一致（含 `http/https`） |
| `ApiKey` | 管理 API 密钥 |
| `Base62Salt` | 短码混淆盐，勿使用默认示例值上线 |

### 4. 启动服务

```bash
go run .
# 或显式指定：go run . -f etc/shorturl-api.yaml
```

- 演示页：http://127.0.0.1:8888/
- 健康检查：http://127.0.0.1:8888/health

---

## API

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| `POST` | `/api/v1/links` | `X-API-Key` | 创建短链 |
| `GET` | `/api/v1/links/{shortCode}` | `X-API-Key` | 查询短链元数据 |
| `GET` | `/api/v1/links/{shortCode}/stats` | `X-API-Key` | 查询 PV / UV |
| `GET` | `/{shortCode}` | 无 | 302 跳转或统一 404 |
| `GET` | `/` | 无 | 演示页 |
| `GET` | `/health` | 无 | 健康检查 |

### 创建短链示例

```bash
curl -s -X POST http://127.0.0.1:8888/api/v1/links \
  -H "Content-Type: application/json" \
  -H "X-API-Key: <your-api-key>" \
  -d '{"longUrl":"https://www.example.com/hello","expireDays":90}'
```

响应字段说明：

- `shortUrl` / `shortCode`：完整短链与短码
- `created`：`true` 表示新建；`false` 表示幂等命中
- `expireAt`：过期时间；`expireDays=0` 时为永不过期（`null`）

### 其他常用命令

```bash
# 跳转（查看 302）
curl -i http://127.0.0.1:8888/{shortCode}

# 统计
curl -s http://127.0.0.1:8888/api/v1/links/{shortCode}/stats \
  -H "X-API-Key: <your-api-key>"

# 查询
curl -s http://127.0.0.1:8888/api/v1/links/{shortCode} \
  -H "X-API-Key: <your-api-key>"
```

---

## 项目结构

```text
shortlink/
├── main.go                 # 入口
├── shorturl.api            # go-zero API 定义
├── etc/                    # 配置文件
├── web/                    # 演示页（HTML/CSS/JS）
├── internal/
│   ├── config/             # 配置结构
│   ├── handler/            # HTTP Handler / 路由 / 静态资源
│   ├── logic/              # 业务逻辑
│   ├── middleware/         # API Key、限流
│   ├── svc/                # 依赖注入
│   ├── cachex/             # Redis 缓存封装
│   └── types/              # 请求响应类型
├── model/                  # MySQL 访问
├── sequence/               # 发号器（MySQL / Redis）
├── pkg/
│   ├── base62/             # 定长 Base62
│   ├── codec/              # Salt 双射混淆
│   ├── rediskey/           # Redis Key 常量
│   └── urltool/            # URL 规范化与 hash
├── short_url_map.sql
└── sequence.sql
```

---

## 测试

```bash
go test ./pkg/... ./sequence/...
```

---

## 安全说明

- **不要**将含真实数据库密码、`ApiKey`、`Base62Salt` 的配置推送到公开仓库；仓库仅保留 `etc/shorturl-api.example.yaml`
- 管理接口依赖 `X-API-Key`；演示页中的 Key 仅用于本地/受控演示，生产应由服务端保管
- 跳转接口对公网开放，已结合短码混淆、统一 404、Bloom 与 IP 限流降低盲扫成本
- UV 为 HyperLogLog **近似值**，非精确去重

---

## 路线图

- [ ] Docker Compose（api + MySQL + Redis）一键启动
- [ ] 配置敏感项支持环境变量覆盖
- [ ] 腾讯云轻量 + Nginx / HTTPS 部署示例
- [ ] `sceneId` 按活动维度拆码
- [ ] Salt 轮换（带 version）

---

