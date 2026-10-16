# Kubernetes（kind）演示部署

面试 / 简历向的**本地完整栈**：API + MySQL + Redis，跑在独立 kind 集群 `shortlink` 中。  
与本机其他项目（如 ihome）隔离；**不是**生产高可用方案。

## 前置

- Docker
- [kind](https://kind.sigs.k8s.io/) ≥ 0.20
- kubectl

## 一键步骤

在仓库根目录执行：

```bash
# 1) 独立集群（与其它 kind 集群分离）
kind create cluster --name shortlink

# 2) 构建并导入镜像（kind 不会拉本地未推送的 tag）
docker build -t shortlink:dev .
kind load docker-image shortlink:dev --name shortlink
# 若节点拉取 Docker Hub 较慢，可一并导入依赖镜像：
# docker pull redis:7-alpine && docker pull mysql:8.4
# kind load docker-image redis:7-alpine mysql:8.4 --name shortlink

# 3) 部署（确保 context 为 kind-shortlink）
kubectl --context kind-shortlink apply -k deploy/k8s

# 4) 等待就绪
kubectl -n shortlink rollout status deploy/mysql --timeout=180s
kubectl -n shortlink rollout status deploy/redis --timeout=60s
kubectl -n shortlink rollout status deploy/api --timeout=120s

# 5) 访问（另开终端可一直挂着；用 18888 避开本机常见的 :8888 占用）
kubectl -n shortlink port-forward --address 127.0.0.1 svc/api 18888:8888
```

验收：

```bash
curl -s http://127.0.0.1:18888/health
# → ok

curl -s -X POST http://127.0.0.1:18888/api/v1/links \
  -H "Content-Type: application/json" \
  -H "X-API-Key: change-me" \
  -d '{"longUrl":"https://www.example.com/hello","expireDays":90}'
```

演示页：http://127.0.0.1:18888/ （API Key 默认 `change-me`）

## 架构（演示）

```text
kubectl port-forward :18888 ──► Service/api ──► Deployment/api (replicas: 1)
                              │
                              ├── Service/mysql (headless) ──► MySQL 8.4
                              └── Service/redis (headless) ──► Redis 7
```

| 资源 | 说明 |
|------|------|
| Secret `shortlink` | DSN / API Key / Salt / 可选 LLM Key |
| ConfigMap `shortlink` | Domain、Redis 地址、Seed、LLM 非密钥项 |
| ConfigMap `mysql-init` | 建表 SQL（与 `deploy/mysql/init` 一致） |
| emptyDir | MySQL/Redis 数据卷；**删 Pod 即丢数据** |

## 设计取舍（面试可讲）

| 决策 | 为什么 | 边界 |
|------|--------|------|
| API 无状态 Deployment | 跳转热路径可水平扩展；配置走 env | 多副本须共享同一 Redis/MySQL |
| MySQL/Redis 也进集群 | kind 上完整可演示，不依赖本机 Compose | 演示级；生产应托管库或 Operator + PVC |
| 默认 `replicas: 1` + `SEED_DEMO=true` | 避免多副本重复灌演示数据 | 扩容前关 Seed |
| `port-forward` 暴露 | 最少网络配置，今天即可验收 | 无 Ingress / TLS |
| 独立 `kind` 集群名 `shortlink` | 与其它项目集群叙事分离 | 本机可并存多个 kind 集群 |

### 多副本（可选）

```bash
kubectl -n shortlink patch configmap shortlink --type merge -p '{"data":{"seed-demo":"false"}}'
kubectl -n shortlink rollout restart deploy/api
kubectl -n shortlink scale deploy/api --replicas=2
```

Bloom / PV / UV / 助手 session 均在 Redis，多副本 API 可共享。

### LLM 助手

Secret 中 `llm-api-key` 默认为空；填入后 `kubectl apply -f deploy/k8s/secret.yaml` 并重启 API。勿把真实 Key 提交进 Git。

## 清理

```bash
kubectl delete -k deploy/k8s
kind delete cluster --name shortlink
```

## 与 Docker Compose 的关系

| | Compose | 本目录 |
|--|---------|--------|
| 用途 | 日常开发默认路径 | 编排 / 简历演示 |
| 数据 | named volume 可保留 | emptyDir，演示向 |
| 访问 | 宿主机端口映射 | port-forward |

日常开发仍推荐根目录 `docker compose up`；需要讲 Kubernetes 时用本清单。
