# 短链接服务

## 一、项目介绍

本短链接服务是一个基于 go-zero 框架开发的高性能短链接服务，用于将冗长的 URL 网址通过程序计算转换成简短易记的短链接字符串。

### 核心功能
- **短链接生成**：将长链接转换为短链接，支持自定义短域名
- **短链接解析**：通过短链接 302 重定向到原始长链接
- **幂等性保障**：相同长链接生成相同短链接，避免重复
- **黑名单机制**：防止敏感词汇作为短链接路径
- **多种发号器支持**：支持 MySQL 和 Redis 两种发号器实现

### 示例
原始长链接：
```
https://www.example.com/article/detail/20260524/news-info?id=897654321&type=original&source=official&page=1&lang=zh-CN
```

转换后短链接：
```
https://Felix.com/123456
```

### 技术栈
- **Web 框架**：go-zero
- **数据库**：MySQL
- **缓存**：Redis
- **参数校验**：validator

---

## 二、环境准备

### 软件环境要求

| 软件 | 版本要求 | 说明 |
|-----|---------|------|
| Go | 1.19+ | 编程语言 |
| MySQL | 5.7+ | 关系型数据库 |
| Redis | 6.0+ | 缓存和可选发号器 |

### 环境安装参考

#### 2.1 Go 安装
- 官网下载：https://golang.org/dl/
- 验证安装：`go version`

#### 2.2 MySQL 安装
- 官网下载：https://dev.mysql.com/downloads/mysql/
- 启动服务：`net start mysql`（Windows）或 `systemctl start mysql`（Linux）

#### 2.3 Redis 安装
- 官网下载：https://redis.io/download
- 启动服务：`redis-server`
- 验证连接：`redis-cli ping`

---

## 三、数据库初始化

### 3.1 创建数据库

```sql
CREATE DATABASE IF NOT EXISTS `shortlink` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `shortlink`;
```

### 3.2 创建发号器表

```sql
CREATE TABLE `sequence` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `stub` varchar(1) NOT NULL,
  `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_uniq_stub` (`stub`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT = '序号表';

-- 初始化发号器数据
INSERT INTO `sequence` (`stub`) VALUES ('a');
```

**注意**：统一使用 InnoDB 引擎，支持事务和更好的崩溃恢复。

### 3.3 创建长链接短链接映射表

```sql
CREATE TABLE `short_url_map` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `create_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `create_by` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建者',
  `is_del` tinyint UNSIGNED NOT NULL DEFAULT '0' COMMENT '是否删除：0正常1删除',
  `lurl` varchar(2048) DEFAULT NULL COMMENT '长链接',
  `md5` char(32) DEFAULT NULL COMMENT '长链接MD5',
  `surl` varchar(11) DEFAULT NULL COMMENT '短链接',
  PRIMARY KEY (`id`),
  INDEX `idx_is_del` (`is_del`),
  UNIQUE KEY `uk_md5` (`md5`),
  UNIQUE KEY `uk_surl` (`surl`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT = '长短链映射表';
```

### 3.4 数据库初始化脚本

项目根目录提供了完整的建表脚本：`short_url_map.sql`

---

## 四、项目构建

### 4.1 项目结构预览

在开始构建前，先进入项目目录：

```bash
cd shortlink/shorturl
```

### 4.2 API 定义（shorturl.api）

API 定义文件位于 `shorturl/shorturl.api`：

```api
syntax = "v1"

type ConvertRequest {
	LongURL string `json:"longUrl" validate:"required"`
}

type ConvertResponse {
	ShortURL string `json:"shortUrl"`
}

type ShowUrlRequest {
	ShortURL string `path:"shortUrl" validate:"required"`
}

type ShowUrlResponse {
	LongURL string `json:"longUrl"`
}

service shorturl-api {
	@handler ShorturlHandler
	post /convert (ConvertRequest) returns (ConvertResponse)

	@handler ShowUrlHandler
	get /:shortUrl (ShowUrlRequest) returns (ShowUrlResponse)
}
```

**重要说明**：`/:shortUrl` 接口使用 GET 方法，便于浏览器直接访问。

### 4.3 使用 goctl 命令生成代码

```bash
# 进入 shorturl 目录
cd shorturl

# 生成 API 服务代码
goctl api go -api shorturl.api -dir .
```

### 4.4 根据数据表生成 model 层

```bash
# 生成 short_url_map 表的 model
goctl model mysql datasource \
  -url="root:123456@tcp(127.0.0.1:3308)/shortlink" \
  -table="short_url_map" \
  -dir="./model"

# 生成 sequence 表的 model
goctl model mysql datasource \
  -url="root:123456@tcp(127.0.0.1:3308)/shortlink" \
  -table="sequence" \
  -dir="./model"
```

**注意**：请根据实际环境修改数据库连接信息。

### 4.5 配置文件修改

编辑 `etc/shorturl-api.yaml` 配置文件：

```yaml
Name: shorturl-api
Host: 0.0.0.0
Port: 8888

ShortUrlDB:
  DSN: root:123456@tcp(127.0.0.1:3308)/shortlink?parseTime=true&charset=utf8mb4

SequenceDB:
  DSN: root:123456@tcp(127.0.0.1:3308)/shortlink?parseTime=true&charset=utf8mb4

ShortURLBlacklist:
  - health
  - status
  - metrics
  - fuxk
  - convert

ShortDomain: "Felix.com"

CacheRedis:
  Host: 127.0.0.1:6379
  Pass: "123456"
  Type: node
  DB: 0
```

**关键配置项**：
- `parseTime=true`：必须设置，否则时间字段无法正确解析
- `charset=utf8mb4`：支持完整 Unicode，包括表情符号

### 4.6 下载项目依赖

```bash
go mod tidy
```

### 4.7 运行项目

```bash
# 进入 shorturl 目录
cd shorturl

# 启动服务
go run shorturl.go
```

**验证服务启动成功**：
看到以下日志表示服务启动成功：
```
Starting server at 0.0.0.0:8888...
```

---

## 五、项目相关原理及策略

本项目是一个典型的**读多写少**的系统。特点是数据写入后基本上不会改变，不需要考虑数据一致性问题，可以放心使用缓存来提高读性能。

### 5.1 短链生成机制

#### 5.1.1 Hash 方案
使用 Hash 函数对长链接进行 Hash，得到 Hash 值作为短链标识符。

| 特性 | 说明 |
|-----|------|
| 优势 | 实现简单，不需要维护序号状态 |
| 缺点 | 数据量大之后会出现哈希冲突 |

#### 5.1.2 发号器方案
每收到一个转链请求，使用发号器生成递增序号，将序号转为 62 进制，最后拼接到短域名，得到最终的短链接。

**示例**：
- 序号：1234567890
- 62 进制：1ly7vk
- 完整短链接：https://Felix.com/1ly7vk

| 特性 | 说明 |
|-----|------|
| 优势 | 生成的 ID 递增，理论容量足够满足现实需求 |
| 缺点 | 高并发下的发号器设计是难点 |

##### 基于 MySQL 主键实现发号器
利用 MySQL 的 REPLACE 语法实现高效发号：

**REPLACE 语法特点**：
- 工作方式与 INSERT 完全相同
- 如果表中的旧行与新行在 PRIMARY KEY 或 UNIQUE 索引具有相同值，则在插入新行之前删除旧行
- 可以在单条记录上进行自动更新，并获得新的自增 ID

**工作原理**：
1. 创建只有一条记录的 sequence 表
2. 使用 REPLACE INTO 更新同一行
3. 每次更新都会生成新的自增 ID

##### 基于 Redis 实现发号器
项目还提供了 Redis 发号器实现（`sequence/redis.go`）：
- 使用 Redis 的 INCR 命令实现原子自增
- 性能更高，适合高并发场景
- 需要配置 Redis 持久化策略

### 5.2 base62 编码

**字符集**：`0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`

**编码规则**：
- 0-9：数字本身
- 10-35：小写字母 a-z
- 36-61：大写字母 A-Z

**转换示例**：

| 十进制 | 62 进制 |
|-------|--------|
| 0 | 0 |
| 1 | 1 |
| 10 | a |
| 11 | b |
| 61 | Z |
| 62 | 10 |
| 63 | 11 |

**容量计算**：
- 1 位：62 种可能
- 2 位：62² = 3,844 种可能
- 3 位：62³ = 238,328 种可能
- 6 位：62⁶ ≈ 568 亿种可能

### 5.3 缓存策略

**方法一**：精简缓存（surl → lurl）
- 缓存内容：短链接 → 长链接映射
- 优势：节省缓存存储空间
- 缺点：需要自己实现缓存逻辑

**方法二**：Redis 完整缓存（surl → 数据行）
- 缓存内容：短链接 → 完整数据行
- 优势：开发量小，实现简单
- 适用场景：数据量适中，内存足够

**内存淘汰策略**：
使用 LRU（Least Recently Used）策略，移除最近最少使用的数据。

**缓存问题解决方案**：

| 问题 | 解决方案 |
|-----|---------|
| 缓存击穿 | 1. 增大过期时间 2. 加锁 3. 使用 single flight 合并请求 |
| 缓存穿透 | 使用布隆过滤器及其变种的布谷鸟过滤器 |
| 缓存雪崩 | 随机化过期时间，避免大量缓存同时过期 |

### 5.4 参数校验

引入第三方库 validator，对请求参数进行校验，确保数据合法性。

**项目地址**：[validator package - github.com/go-playground/validator/v10](https://pkg.go.dev/github.com/go-playground/validator/v10)

**下载安装**：
```bash
go get github.com/go-playground/validator/v10
```

**使用示例**：
在 API 定义中使用 validate 标签：
```go
type ConvertRequest {
	LongURL string `json:"longUrl" validate:"required"`
}
```

---

## 六、API 使用说明

### 6.1 转链接口

**功能**：将长链接转换为短链接

**请求信息**：

| 项目 | 值 |
|-----|-----|
| URL | `http://localhost:8888/convert` |
| Method | POST |
| Content-Type | application/json |

**请求参数**：

| 参数名 | 类型 | 必填 | 说明 |
|-------|------|-----|------|
| longUrl | string | 是 | 要转换的长链接地址 |

**请求示例**：

```json
{
    "longUrl": "https://www.example.com/article/detail/20260524/news-info?id=897654321"
}
```

**响应参数**：

| 参数名 | 类型 | 说明 |
|-------|------|------|
| shortUrl | string | 生成的短链接 |

**响应示例**：

```json
{
    "shortUrl": "https://Felix.com/1"
}
```

**Postman 请求示例**：

1. 选择 POST 方法
2. URL 输入 `http://localhost:8888/convert`
3. Body 选择 raw → JSON
4. 输入请求参数并发送

---

### 6.2 解析接口

**功能**：通过短链接查询并重定向到原始长链接

**请求信息**：

| 项目 | 值 |
|-----|-----|
| URL | `http://localhost:8888/{shortUrl}` |
| Method | GET |
| 参数位置 | URL 路径参数 |

**请求示例**：

```
GET http://localhost:8888/1
```

**响应说明**：

| 场景 | 响应 |
|-----|------|
| 短链接存在 | 302 重定向到原始长链接 |
| 短链接不存在 | 返回错误信息 "404-short链接不存在" |

**浏览器访问流程**：

1. 用户在浏览器输入：`http://localhost:8888/1`
2. 服务查询数据库获取对应的长链接
3. 服务返回 302 重定向响应
4. 浏览器自动跳转到原始长链接页面

---

### 6.3 完整使用示例

#### 步骤一：创建短链接

**请求**：

```bash
curl -X POST http://localhost:8888/convert \
  -H "Content-Type: application/json" \
  -d '{"longUrl": "https://www.example.com"}'
```

**响应**：

```json
{
    "shortUrl": "https://Felix.com/2"
}
```

#### 步骤二：访问短链接

在浏览器中输入：`http://localhost:8888/2`

浏览器将自动重定向到：`https://www.example.com`

---

## 七、项目结构

```
shorturl/
├── shorturl.go              # 程序入口
├── shorturl.api             # API 定义文件
├── go.mod                   # Go 模块依赖
├── go.sum                   # 依赖校验文件
├── etc/
│   └── shorturl-api.yaml    # 服务配置文件
├── internal/
│   ├── config/
│   │   └── config.go        # 配置结构体定义
│   ├── handler/
│   │   ├── convertlogic.go  # 转链业务逻辑
│   │   ├── converthandler.go # 转链请求处理
│   │   ├── showurllogic.go  # 查链业务逻辑
│   │   └── showurlhandler.go # 查链请求处理
│   ├── logic/
│   │   ├── shorturllogic.go # 转链核心逻辑
│   │   └── showurllogic.go  # 查链核心逻辑
│   ├── svc/
│   │   └── servicecontext.go # 服务上下文
│   └── types/
│       └── types.go         # 请求/响应结构体
├── model/
│   ├── shorturlmapmodel.go  # 短链接映射表模型
│   └── sequencemodel.go     # 发号器表模型
├── pkg/
│   └── base62/
│       └── base62.go        # 62进制编码工具
├── sequence/
│   ├── mysql.go             # MySQL 发号器实现
│   └── redis.go             # Redis 发号器实现
└── short_url_map.sql        # 数据库建表 SQL
```

---

## 八、核心代码说明

### 8.1 短链接生成流程（shorturllogic.go）

```
1. 接收长链接请求
       ↓
2. 参数校验（validator）
       ↓
3. MD5 计算长链接指纹
       ↓
4. 查询缓存/数据库是否已存在
       ↓
5. 使用发号器获取序号
       ↓
6. 序号转为 62 进制
       ↓
7. 拼接短域名生成短链接
       ↓
8. 存储到数据库
       ↓
9. 返回短链接
```

### 8.2 短链接解析流程（showurllogic.go）

```
1. 接收短链接请求
       ↓
2. 参数校验
       ↓
3. 查询数据库获取长链接
       ↓
4. 返回 302 重定向
       ↓
5. 浏览器跳转到长链接页面
```

---

## 九、配置说明

### 9.1 配置文件（shorturl-api.yaml）

```yaml
Name: shorturl-api                    # 服务名称
Host: 0.0.0.0                         # 监听地址
Port: 8888                            # 监听端口

ShortUrlDB:                           # 短链接数据库配置
  DSN: root:123456@tcp(127.0.0.1:3308)/shortlink?parseTime=true&charset=utf8mb4

SequenceDB:                           # 发号器数据库配置
  DSN: root:123456@tcp(127.0.0.1:3308)/shortlink?parseTime=true&charset=utf8mb4

ShortURLBlacklist:                    # 短链接黑名单
  - health
  - status
  - metrics
  - fuxk
  - convert

ShortDomain: "Felix.com"              # 短链接域名

CacheRedis:                          # Redis 缓存配置
  Host: 127.0.0.1:6379               # Redis 地址
  Pass: "123456"                     # Redis 密码
  Type: node                         # Redis 类型：node/cluster
  DB: 0                              # 数据库编号
```

### 9.2 关键配置项说明

| 配置项 | 说明 | 注意事项 |
|-------|------|---------|
| parseTime | 解析时间字段 | 必须设为 true，否则时间字段无法正确解析 |
| charset | 字符集 | 建议使用 utf8mb4 支持完整 Unicode |
| Type | Redis 类型 | node 表示单节点，cluster 表示集群模式 |

---

## 十、常见问题

### Q1: 访问短链接返回 405 Method Not Allowed？

**原因**：API 定义中使用了 POST 方法，而浏览器访问默认使用 GET 方法。

**解决**：修改 `shorturl.api` 文件，将 `post /:shortUrl` 改为 `get /:shortUrl`，然后重新生成代码。

### Q2: 返回 "sql: Scan error" 错误？

**原因**：数据库 DSN 缺少 `parseTime=true` 参数。

**解决**：在配置文件的 DSN 后添加 `?parseTime=true&charset=utf8mb4`

### Q3: 重定向失败，浏览器显示空白？

**原因**：数据库中的 longurl 缺少协议头（http:// 或 https://）。

**解决**：在 `showurlhandler.go` 中添加 URL 协议补全逻辑：

```go
if !strings.HasPrefix(resp.LongURL, "http://") && !strings.HasPrefix(resp.LongURL, "https://") {
    resp.LongURL = "https://" + resp.LongURL
}
```

### Q4: 转链接口返回 "无效链接"？

**原因**：系统会验证长链接是否可访问，如果无法访问则拒绝转换。

**解决**：
- 方案一：使用可访问的 URL 测试
- 方案二：临时注释掉 URL 验证逻辑（仅用于开发环境）

### Q5: Redis 连接失败？

**原因**：
- Redis 服务未启动
- 配置的地址、端口或密码不正确

**解决**：检查 `CacheRedis` 配置项，确保与实际 Redis 服务一致。

---

## 十一、运行与维护

### 11.1 启动服务

```bash
cd shorturl
go run shorturl.go
```

### 11.2 查看日志

服务启动后会输出日志，包含请求处理、错误信息等。

### 11.3 健康检查

访问根路径或任意不存在的短链接，检查服务是否正常运行。

---

*文档更新时间：2026-05-24*
