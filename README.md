# PetVerse

PetVerse 是一个 `monorepo` 形式的一站式宠物管理应用，覆盖宠物档案、健康管理、智能设备、服务预约、社区互动、训练计划、商城推荐和通知推送。后端使用 `Go + Gin + GORM`，前端使用 `Expo Router + React Native + TypeScript + React Query + Zustand`。

## 当前状态

- `Phase 1`：认证、用户、宠物基础流程
- `Phase 2`：健康记录、AI 健康分析、设备控制、设备 WebSocket
- `Phase 3`：服务市场、预约、社区帖子/评论/点赞
- `Phase 4`：训练计划、AI 训练生成、商城推荐、通知中心与全局 WebSocket

当前 AI 默认仍会走本地规则 fallback，但后端已经支持配置外部模型客户端。配置 `ai.provider / ai.base_url / ai.api_key / ai.model` 后，健康问答、社区问答和训练计划生成会优先走真实模型；未配置或请求失败时会自动退回本地规则实现。

认证链路方面，微信登录仍然是演示模式；Apple / Google 已支持客户端提交真实 `identity_token`，后端会基于官方 JWKS 做签名、issuer 和 audience 校验。未配置 `social_auth` 或移动端 client ID 时，应用会继续回退到演示模式。

通知链路现在也支持 Expo Push token 注册和服务端下发。用户登录后，移动端会在权限允许时自动注册 Expo push token；后端在创建训练或系统通知时，会同时走 WebSocket 和可选的 Expo Push dispatcher。

基础设施层也已经支持可选的 `TimescaleDB / NATS / MinIO` 接入：设备时序点位可切到独立时序库，设备/训练/通知可发布到 NATS，对象上传可从本地目录切到 S3 兼容存储。

## 仓库结构

```text
.
├── apps/mobile            # Expo 应用
├── server                 # Go API 服务
├── docker-compose.yml     # 本地 Postgres / TimescaleDB / Redis / NATS / MinIO
├── Makefile               # 根命令入口
└── .github/workflows      # Backend / Mobile CI
```

## 快速启动

### 1. 启动基础设施

```bash
docker compose up -d
```

会启动：

- `postgres:16`
- `timescale/timescaledb:latest-pg16`
- `redis:7`
- `nats:2.11`
- `minio/minio`

### 2. 启动后端

后端配置文件默认在 `server/config/config.yaml`，示例见 `server/config/config.example.yaml`。AI 相关配置如下：

```yaml
ai:
  provider: local # local / openai_compatible / anthropic
  base_url: ""
  api_key: ""
  model: ""
  temperature: 0.2
  timeout: 20s
```

Push 相关配置如下：

```yaml
push:
  provider: none # none / expo
  expo_url: https://exp.host/--/api/v2/push/send
  access_token: ""
  timeout: 10s
```

社交登录校验配置如下：

```yaml
social_auth:
  google_client_ids:
    - your-google-web-or-native-client-id
  apple_audiences:
    - com.example.petverse
  http_timeout: 10s
  cache_ttl: 1h
```

如果要启用独立时序库、NATS 和 MinIO，可以把下面几组配置打开：

```yaml
timeseries:
  enabled: true
  host: 127.0.0.1
  port: 5433
  user: postgres
  password: postgres
  name: petverse_ts
  sslmode: disable
  timezone: Asia/Shanghai

nats:
  enabled: true
  url: nats://127.0.0.1:4222

object_storage:
  provider: minio # local / minio
  local_dir: ./uploads
  endpoint: 127.0.0.1:9000
  access_key: minioadmin
  secret_key: minioadmin
  use_ssl: false
  bucket: petverse
  public_base_url: http://127.0.0.1:9000
```

说明：

- `timeseries.enabled=true` 时，后端会连接独立时序库并自动创建 `device_data_points` 表及 hypertable。
- `nats.enabled=true` 时，设备创建/设备数据点、训练计划、通知会额外发布事件。
- `object_storage.provider=minio` 时，上传接口、用户头像和宠物头像会写入 MinIO；默认仍走本地 `./uploads`。

```bash
make run
```

常用命令：

```bash
make build
make test
make smoke
make migrate-up
make migrate-down
```

### 3. 启动移动端

先准备环境变量：

```bash
cp apps/mobile/.env.example apps/mobile/.env
```

如果要启用 Expo Push token 注册，补上：

```bash
EXPO_PUBLIC_EXPO_PROJECT_ID=your-expo-project-id
```

如果要启用真实 Google 登录，再补上：

```bash
EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID=
EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID=
EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID=
```

然后安装依赖并启动：

```bash
make mobile-install
make mobile-start
```

或直接使用：

```bash
cd apps/mobile
npm install
npx expo start
```

## 校验命令

仓库级快速校验：

```bash
make check
```

它会执行：

- `server` 的 `go test ./...`
- `apps/mobile` 的 `npx tsc --noEmit`

如果后端已经启动，还可以直接跑一条完整烟测：

```bash
make smoke
```

这条命令会按 Phase 1-4 依次验证：

- 注册 / 登录 / 刷新 token / 当前用户
- 宠物、健康、设备、设备 WebSocket
- 服务、预约、社区
- 训练、商城、通知、全局 WebSocket

## 已实现的主要接口

后端当前已经覆盖以下模块：

- `auth`：注册、登录、刷新 token
- `users`：当前用户信息
- `pets`：增删改查
- `health`：记录 CRUD、摘要、预警、AI 问答
- `devices`：CRUD、状态、数据查询、指令、设备 WS
- `events`：可选 NATS 发布，覆盖设备、训练、通知
- `services`：列表、详情、可预约时段、评价
- `bookings`：创建、列表、详情、取消、评价
- `posts` / `comments`：社区内容和互动
- `training`：计划 CRUD、AI 生成
- `shop`：商品列表、详情、宠物个性化推荐
- `notifications`：列表、单条已读、全部已读、全局 WS
- `notifications/push-token`：注册 / 注销移动端 push token
- `upload`：本地目录或 MinIO/S3 兼容存储

## CI

已提供两个 GitHub Actions 工作流：

- `.github/workflows/backend.yml`：运行 `go test` 和 `go build`
- `.github/workflows/mobile.yml`：运行 `npm ci` 和 `tsc`

## 已知限制

- 当前环境中未完成 `docker compose up` 的真实联调验证，因为执行环境没有可用 Docker 运行时。
- `TimescaleDB / NATS / MinIO` 已完成代码接入和本地 compose 编排，但当前环境没有 Docker 运行时，所以还没做 live 基础设施联调。
- 当前服务端推送走 Expo Push Service；还没有直接接 APNs / FCM 原生 provider，也没有在当前环境做真实云端送达验证。
- Google / Apple 已接入真实 `identity_token` 验签链路，但当前环境没有真实 Google client ID、Apple audience 和真机构建环境，所以还没做 live 登录验证。
- 商城推荐仍是规则引擎版本；健康问答、社区问答和训练生成已支持外部模型配置，但当前环境没有实际接第三方密钥做 live 调用验证。
