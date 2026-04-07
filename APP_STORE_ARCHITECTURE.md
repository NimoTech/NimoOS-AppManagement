# NimoOS App Store 架构与数据流总结

本文档总结了 NimoOS-AppManagement 中应用商店（App Store）的数据流转机制、核心状态管理以及相关的文件与目录结构。

## 1. 核心概念与数据流转 (Data Flow)

应用商店的数据流分为三个主要阶段：从远端获取，解析到本地内存，最终聚合返回给 API。

### 阶段一：远端同步 ➡️ 本地文件系统
应用商店本质上是托管在远端的压缩包（包含各个应用的 `docker-compose.yaml` 文件）。
- **触发机制**：
  - 系统启动时触发一次。
  - 后台 `cron` 任务每 10 分钟自动执行全局更新（`UpdateCatalog`）。
  - 用户通过 API 手动触发注册或同步。
- **增量更新**：通过 HTTP `HEAD` 请求获取远端文件的 `Content-Length`，与内存记录的 `lastAPPStoreSize` 比对。若无变化，则跳过下载以节省带宽。
- **原子替换机制**：
  - 将新包下载到 `.tmp` 目录。
  - 解压后，将当前的 `workdir` 重命名为 `.backup`。
  - 将 `.tmp` 重命名为新的 `workdir`。
  - 若过程中发生任何错误，系统会自动将 `.backup` 恢复为 `workdir`，防止商店数据损坏。

### 阶段二：本地文件系统 ➡️ 内存对象
- **文件扫描**：系统调用 `BuildCatalog`，遍历本地工作目录（如 `Apps/`）下的每一个子目录，寻找 `docker-compose.yaml` 或 `.yml` 文件。
- **元数据提取**：调用 `NewComposeAppFromYAML` 解析文件，特别关注 NimoOS 专有的顶级扩展字段（当前代码常量为 `x-casaos` 或旧版的 `x-nimoos`）。从中提取图标（icon）、标题（title）、分类（category）、支持的架构等展示用的元数据。
- **内存驻留**：解析成功的应用对象被封装为 `ComposeApp`，存储在各个 `appStore` 实例的 `catalog` (Map结构) 中，常驻内存以便快速读取。

### 阶段三：内存对象 ➡️ API 响应
- **多源聚合**：`AppStoreManagement` 作为全局管理器，遍历所有已注册的 `appStore`，将它们的 `catalog` 合并。
- **动态过滤**：在路由层（如 `ComposeAppStoreInfoList`），根据当前设备的 CPU 架构（如剔除不支持 `arm64` 的应用）、API 请求的分类（Category）等进行过滤。
- **状态拼接**：结合本机已安装应用的状态（是否安装、是否可升级），将静态元数据与动态状态组合，最终序列化为 JSON 响应给前端。

---

## 2. 核心状态管理 (State Management)

NimoOS-AppManagement 采用了多种缓存和锁机制来维护应用的状态：

- **可升级状态缓存 (`isAppUpgradable`)**
  - **组件**：基于 `gcache` 实现的内存缓存。
  - **作用**：检查应用是否有新版本通常需要查询远端 Docker Registry 的 Image Digest，这非常耗时。系统将检测结果缓存（通常有效期 1 小时），确保前端获取应用列表时实现毫秒级响应。
- **升级中状态锁 (`isAppUpgrading`)**
  - **组件**：`sync.Map`
  - **作用**：记录当前正在执行升级任务的应用 ID。当用户点击升级时，系统先检查此锁，防止用户连续点击或多端并发导致同一个应用发生并发更新冲突。
- **应用生命周期状态**
  - **未安装 (Not Installed)**：仅存在于 `appStore.catalog` 中。
  - **已安装 (Installed)**：本地 `/var/lib/nimoos/apps/` 下存在该应用的配置，且 Docker 容器已创建。
  - **可更新 (Upgradable)**：本地镜像的 Tag 或 Digest 与商店中 `compose.yaml` 定义的最新版本不一致。

---

## 3. 涉及的关键文件与物理路径

### 3.1 物理存储路径
- **App Store 展示数据（缓存源）**
  - 基础路径：`/var/lib/nimoos/appstore/`
  - 默认源的 Compose 文件存放位置：`/var/lib/nimoos/appstore/default/Apps/<AppName>/docker-compose.yaml`
- **已安装应用数据（运行态）**
  - 基础路径：`/var/lib/nimoos/apps/`
  - 用户安装后的 Compose 文件存放位置：`/var/lib/nimoos/apps/<AppName>/docker-compose.yaml`

### 3.2 核心源代码文件
- **`service/appstore.go`**
  - 负责**单个应用商店**的管理。
  - 核心方法：`UpdateCatalog()`（下载并原子替换）、`BuildCatalog()`（扫描本地目录并加载到内存）、`StoreRoot()`（获取商店工作目录）。
- **`service/appstore_management.go`**
  - **全局应用商店管理器**，单例模式。
  - 负责聚合多个商店的数据（`Catalog()`），以及管理应用的升级状态和缓存锁（`IsUpdateAvailable()`, `isAppUpgradable`）。
- **`service/compose_app.go`**
  - 核心业务实体 `ComposeApp`，封装了 Docker Compose 规范的结构。
  - 核心方法：`StoreInfo()`（提取 `x-casaos` 扩展元数据），`PullAndApply()`（拉取新镜像并尝试启动，失败则回滚）。
- **`route/v2/appstore.go`**
  - 负责处理应用商店相关的 HTTP API 请求。
  - 核心方法：`ComposeAppStoreInfoList()`（获取列表并应用 CPU 架构、分类等过滤规则）。
- **`common/constants.go`**
  - 存放常量定义。
  - 关键常量：`ComposeExtensionNameXNimoOS = "x-casaos"`（决定了系统去读取哪个顶级扩展字段）。

---

## 4. 常见问题与避坑指南

- **扩展字段命名冲突 (`x-nimoos` vs `x-casaos`)**
  代码中的常量可能被修改为 `"x-casaos"`，但某些旧的应用或默认提供的示例 `docker-compose.yaml` 仍使用 `"x-nimoos:"`。这会导致 `StoreInfo()` 方法抛出 `extension x-casaos not found` 的错误，应用无法在商店列表中正常显示。
  - **修复建议**：统一应用商店源里的 YAML 文件字段名，或者在 `StoreInfo()` 的解析逻辑中加入后向兼容处理（同时检查两个字段名）。
- **更新回滚触发**
  如果在开发过程中发现 `/var/lib/nimoos/appstore/default/` 目录更新失败，可以检查是否存在 `.tmp` 或 `.backup` 文件夹遗留，通常是因为权限问题或下载中断触发了原子替换的安全回滚。