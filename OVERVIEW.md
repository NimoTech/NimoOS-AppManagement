# CasaOS-AppManagement 详解

CasaOS-AppManagement 是负责容器化应用完整生命周期管理的微服务，整合 Docker Compose 和 CasaOS AppStore。

---

## 核心职责

- Docker Compose 应用的安装、卸载、启动、停止、更新
- AppStore 目录同步（每 10 分钟自动刷新）
- 多 AppStore 注册与管理
- Docker 镜像管理（拉取、版本检查）
- 应用健康检查（HTTP 端口可达性）
- 端口冲突检测
- Compose YAML 验证（支持 dry-run）
- 向 MessageBus 发布应用生命周期事件

---

## 目录结构

```
CasaOS-AppManagement/
├── main.go                  # 启动入口
├── api/                     # OpenAPI 规范（V1/V2）
├── model/                   # 数据模型（App、Port、Env、Volume 映射）
├── service/
│   ├── service.go           # 服务注册表（单例）
│   ├── compose_service.go   # Compose 应用安装/卸载核心逻辑
│   ├── container.go         # Docker 容器操作
│   ├── image.go             # 镜像管理
│   ├── appstore.go          # AppStore 抽象接口
│   └── appstore_management.go  # AppStore 注册与目录刷新
├── pkg/
│   ├── docker/              # Docker 客户端封装（容器、镜像、认证）
│   └── config/              # INI 配置管理
├── route/
│   ├── v1/                  # V1 传统路由
│   └── v2/                  # V2 OpenAPI 路由（主要实现）
└── build/                   # 构建产物、systemd 服务文件
```

---

## 应用安装流程

```
POST /v2/app_management/compose
  1. 解析 Compose YAML（compose-spec/compose-go）
  2. 校验端口冲突
  3. 创建工作目录：/data/Apps/{appName}/
  4. 写入 docker-compose.yml
  5. 异步执行 docker compose up
  6. 向 MessageBus 发布进度事件
  7. 失败时自动清理
```

---

## API 一览

### V2 API（主要，OpenAPI 3.0）

**AppStore 管理**
```
GET  /v2/app_management/appstore          列出已注册的 AppStore
POST /v2/app_management/appstore          注册新 AppStore（异步）
DELETE /v2/app_management/appstore/{id}   注销 AppStore
GET  /v2/app_management/apps             列出所有可用应用
GET  /v2/app_management/apps/upgradable  可升级应用列表
```

**Compose 应用管理**
```
GET    /v2/app_management/compose            列出已安装应用
POST   /v2/app_management/compose            安装应用（支持 dry-run）
GET    /v2/app_management/compose/{id}       查看应用详情
PUT    /v2/app_management/compose/{id}       应用配置变更
PATCH  /v2/app_management/compose/{id}       升级到最新版本
DELETE /v2/app_management/compose/{id}       卸载应用
PUT    /v2/app_management/compose/{id}/status 启动/停止/重启
GET    /v2/app_management/compose/{id}/logs   查看日志
```

**全局配置**
```
GET    /v2/app_management/global        查看全局配置
PUT    /v2/app_management/global/{key}  设置全局配置
DELETE /v2/app_management/global/{key}  删除全局配置
```

---

## MessageBus 事件

服务生命周期每个阶段都会发布 Begin/Progress/End/Error 事件：

- `AppInstall`, `AppUninstall`, `AppUpdate`, `AppStart`, `AppStop`, `AppRestart`
- `AppStoreRegister`, `AppStoreUnregister`
- `ImagePull`

事件属性包含：`app:name`、`app:title`、`app:icon`、`app:progress`、`docker:image:name` 等。

---

## 数据存储

- 应用工作目录：`/data/Apps/{appName}/`
- AppStore 目录缓存：`/data/appstore/`
- 日志：`/var/log/casaos/app-management.log`
- 全局环境变量：`/etc/casaos/env`

---

## 配置

```ini
[common]
RuntimePath = /var/run/casaos

[app]
AppStorePath = /data/appstore
AppsPath = /data/Apps
LogPath = /var/log/casaos

[server]
appstore = https://store.casaos.io
```

---

## 技术栈

- **框架**：Echo v4
- **Docker 集成**：docker/docker（官方 SDK）、docker/compose v2
- **Compose 解析**：compose-spec/compose-go
- **缓存**：bluele/gcache（升级检查结果缓存）
- **调度**：robfig/cron v3（AppStore 10 分钟刷新）
- **代码生成**：deepmap/oapi-codegen
