# NimoOS-AppManagement 详解

NimoOS-AppManagement 是负责容器化应用完整生命周期管理的微服务，整合 Docker Compose 和 NimoOS AppStore。当前版本 `v1.9.0-alpha1`（`common/constants.go`）。

---

## 核心职责

- Docker Compose 应用的安装、卸载、启动、停止、更新
- AppStore 目录同步（每 10 分钟自动刷新，`main.go` cron）
- 多 AppStore 注册与管理（含内置默认商店）
- Docker 镜像管理（拉取、版本检查）
- Docker 数据根目录迁移（换存储盘时停 Docker + rsync 搬迁）
- 应用健康检查（HTTP 端口可达性）
- 端口冲突检测
- Compose YAML 验证（支持 dry-run），兼容 `x-casaos` 旧扩展字段
- 向 MessageBus 发布应用生命周期事件
- 从 App 网格隐藏系统级 compose 应用/容器（`nimoos.system=true` label）

---

## 目录结构

```
NimoOS-AppManagement/
├── main.go                  # 启动入口（含 AppStore 10 分钟刷新 cron）
├── api/                     # OpenAPI 规范（V1/V2）
├── codegen/                 # oapi-codegen 生成的 server/types
├── common/                  # 常量、MessageBus 事件类型定义（message.go）
├── model/                   # 数据模型（App、Port、Env、Volume 映射）
├── service/
│   ├── service.go           # 服务注册表（单例）
│   ├── compose_app.go       # ComposeApp 类型：YAML 解析、x-nimoos 扩展、更新
│   ├── compose_service.go   # Compose 应用安装编排（工作目录在 AppsPath 下）
│   ├── container.go         # Docker 容器操作、容器统计
│   ├── image.go             # 镜像管理
│   ├── appstore.go          # AppStore 抽象接口（含内置 "default" 商店）
│   └── appstore_management.go  # AppStore 注册与目录刷新
├── pkg/
│   ├── docker/              # Docker 客户端封装（容器、镜像、认证）
│   └── config/              # INI 配置管理（默认路径见 pkg/config/init.go）
├── route/
│   ├── v1.go / v1/          # V1 传统路由（含 Docker daemon 配置/迁移）
│   ├── v2.go / v2/          # V2 OpenAPI 路由（主要实现；internal_web.go 出 App 网格）
│   └── （两侧均挂迁移锁中间件，见下文）
├── APP_STORE_ARCHITECTURE.md  # AppStore 架构说明（x-nimoos/x-casaos 兼容背景）
└── build/                   # 构建产物、systemd 服务文件、配置样例
```

---

## 应用安装流程

```
POST /v2/app_management/compose
  1. 解析 Compose YAML（compose-spec/compose-go）
  2. 校验端口冲突
  3. 创建工作目录：{AppsPath}/{appName}/（默认 /var/lib/nimoos/apps/）
  4. 写入 docker-compose.yml
  5. 异步执行 docker compose up
  6. 向 MessageBus 发布进度事件
  7. 失败时自动清理
```

兼容性：`NewComposeAppFromYAML` 加载时会把旧的 `x-casaos` 扩展字段（顶层与各 service）自动归一化为 `x-nimoos`（`service/compose_app.go`），CasaOS 生态的 compose 文件可直接安装。

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

**Web 内部接口**
```
GET /v2/app_management/web/appgrid       App 网格数据（internal use ONLY，UI 首页）
```

**全局配置**
```
GET    /v2/app_management/global        查看全局配置
PUT    /v2/app_management/global/{key}  设置全局配置
DELETE /v2/app_management/global/{key}  删除全局配置
```

### V1 API（传统，部分仍在用）

```
GET /v1/container/info    读取 Docker daemon 配置（含 docker_root_dir）
PUT /v1/container/info    修改 docker_root_dir → 触发数据根迁移（见下）
POST /v1/container/prune  清理 Docker 资源
```

---

## 系统组件隐藏与显示名（nimoos.* labels）

系统内部组件（如内置 Photos 背后的 ML 后端）以 compose 应用/容器形式运行，但不应出现在用户的 App 面板：

- **`nimoos.system: "true"`**：任一 service 带此 label 的 compose 应用被 `isSystemComposeApp()` 从 `/web/appgrid` 过滤（`route/v2/internal_web.go`）；带此 label 的裸容器同样在容器列表中被跳过（`service/container.go`）。
- **`nimoos.display_name`**：容器统计（`getContainerStats()`，`service/container.go`）的标题优先使用该 label，仪表盘显示友好名（如 "Photos AI"）而非裸容器名。

---

## Docker 数据根迁移

`PUT /v1/container/info` 传入 `docker_root_dir` 时（`route/v1/docker.go`）：

1. 读取 `/etc/docker/daemon.json` 得到当前根（缺省 `/var/lib/docker`），目标为 `{docker_root_dir}/docker`；路径未变则直接返回。
2. 停止 docker 服务 → `rsync -a --ignore-existing` 搬迁数据 → 写回 daemon.json 与 `/var/lib/nimoos/docker_root` → 重启 docker。
3. 全程通过 MessageBus 发布 `DockerMigrationBegin/Progress/End/Error` 事件（`common/message.go`），UI 据此展示进度。

**迁移锁**：迁移期间存在 `/var/run/nimoos/migration.lock`；V1/V2 路由均挂中间件（`route/v1.go`、`route/v2.go`），对所有 POST/PUT/DELETE 请求返回 503，避免 Docker 停止状态下安装/卸载/更新静默失败。

---

## MessageBus 事件

服务生命周期每个阶段都会发布 Begin/Progress/End/Error 事件：

- `AppInstall`, `AppUninstall`, `AppUpdate`, `AppStart`, `AppStop`, `AppRestart`
- `AppStoreRegister`, `AppStoreUnregister`
- `ImagePull`
- `DockerMigrationBegin/Progress/End/Error`（数据根迁移）

事件属性包含：`app:name`、`app:title`、`app:icon`、`app:progress`、`docker:image:name` 等。

---

## 数据存储

- 应用工作目录：`/var/lib/nimoos/apps/{appName}/`（`AppsPath`）
- AppStore 目录缓存：`/var/lib/nimoos/appstore/`（`AppStorePath`）；其中 `default/` 为内置默认商店（`service/appstore.go` 中 url 为 `"default"` 的特例，由 AppStore 包解压）
- Docker 根目录记录：`/var/lib/nimoos/docker_root`
- 日志：`/var/log/nimoos/app-management.log`
- 全局环境变量：`/etc/nimoos/env`

---

## 配置

默认值来自 `NimoOS-Common/utils/constants`（`DefaultDataPath = /var/lib/nimoos` 等），样例见 `build/sysroot/etc/nimoos/app-management.conf.sample`：

```ini
[common]
RuntimePath = /var/run/nimoos

[app]
AppStorePath = /var/lib/nimoos/appstore
AppsPath = /var/lib/nimoos/apps
LogPath = /var/log/nimoos

[server]
; 可配置多个第三方商店源（AllowShadows），样例当前指向 CasaOS 上游仓库
appstore = https://nimoos-public.s3.us-east-2.amazonaws.com/nimoos/appstore/store/main.zip
appstore = https://github.com/bigbeartechworld/big-bear-casaos/archive/refs/heads/master.zip
```

---

## 技术栈

- **框架**：Echo v4
- **Docker 集成**：docker/docker（官方 SDK）、docker/compose v2
- **Compose 解析**：compose-spec/compose-go（`x-nimoos` 扩展，兼容 `x-casaos`）
- **缓存**：bluele/gcache（升级检查结果缓存）
- **调度**：robfig/cron v3（AppStore 10 分钟刷新）
- **代码生成**：deepmap/oapi-codegen
