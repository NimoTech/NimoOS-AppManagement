# NimoOS App Store Architecture and Data Flow Summary

This document summarizes the data flow mechanism, core state management, and related files/directory structure of the App Store within NimoOS-AppManagement.

## 1. Core Concepts and Data Flow

The App Store's data flow is divided into three main stages: fetching from the remote source, parsing into local memory, and finally aggregating and returning it to the API.

### Stage 1: Remote sync ➡️ local filesystem
The App Store is essentially a compressed archive hosted remotely (containing each app's `docker-compose.yaml` file).
- **Trigger mechanisms**:
  - Triggered once on system startup.
  - A background `cron` job automatically runs a global update (`UpdateCatalog`) every 10 minutes.
  - The user manually triggers registration or sync via the API.
- **Incremental update**: an HTTP `HEAD` request fetches the remote file's `Content-Length` and compares it against the in-memory `lastAPPStoreSize`. If unchanged, the download is skipped to save bandwidth.
- **Atomic replacement mechanism**:
  - Download the new package into a `.tmp` directory.
  - After extraction, rename the current `workdir` to `.backup`.
  - Rename `.tmp` to the new `workdir`.
  - If any error occurs during the process, the system automatically restores `.backup` to `workdir`, preventing corruption of store data.

### Stage 2: Local filesystem ➡️ in-memory objects
- **File scanning**: the system calls `BuildCatalog`, walking every subdirectory under the local working directory (e.g. `Apps/`) looking for a `docker-compose.yaml` or `.yml` file.
- **Metadata extraction**: `NewComposeAppFromYAML` parses the file, focusing specifically on NimoOS's proprietary top-level extension field (currently the code constant is `x-casaos`, or the legacy `x-nimoos`). It extracts display metadata such as icon, title, category, and supported architectures.
- **In-memory residency**: successfully parsed app objects are wrapped as `ComposeApp` and stored in each `appStore` instance's `catalog` (a Map structure), kept resident in memory for fast reads.

### Stage 3: In-memory objects ➡️ API response
- **Multi-source aggregation**: `AppStoreManagement`, acting as the global manager, iterates over all registered `appStore` instances and merges their `catalog`s.
- **Dynamic filtering**: at the routing layer (e.g. `ComposeAppStoreInfoList`), apps are filtered based on the current device's CPU architecture (e.g. excluding apps that don't support `arm64`), the API request's category, etc.
- **State composition**: local installed-app state (installed or not, upgradable or not) is combined with the static metadata, and the result is serialized as a JSON response to the frontend.

---

## 2. Core State Management

NimoOS-AppManagement uses several caching and locking mechanisms to maintain app state:

- **Upgradable state cache (`isAppUpgradable`)**
  - **Component**: an in-memory cache built on `gcache`.
  - **Purpose**: checking whether an app has a new version usually requires querying the remote Docker Registry's image digest, which is slow. The system caches the check result (typically valid for 1 hour), ensuring millisecond-level response when the frontend fetches the app list.
- **Upgrading-state lock (`isAppUpgrading`)**
  - **Component**: `sync.Map`
  - **Purpose**: tracks the IDs of apps currently undergoing an upgrade task. When the user clicks upgrade, the system checks this lock first, preventing repeated clicks or concurrent multi-client requests from causing a concurrent-update conflict on the same app.
- **App lifecycle states**
  - **Not Installed**: exists only in `appStore.catalog`.
  - **Installed**: the app's config exists locally under `/var/lib/nimoos/apps/`, and the Docker container has been created.
  - **Upgradable**: the local image's tag or digest doesn't match the latest version defined in the store's `compose.yaml`.

---

## 3. Key Files and Physical Paths Involved

### 3.1 Physical storage paths
- **App Store display data (cache source)**
  - Base path: `/var/lib/nimoos/appstore/`
  - Default source's Compose file location: `/var/lib/nimoos/appstore/default/Apps/<AppName>/docker-compose.yaml`
- **Installed app data (runtime state)**
  - Base path: `/var/lib/nimoos/apps/`
  - Compose file location after user installation: `/var/lib/nimoos/apps/<AppName>/docker-compose.yaml`

### 3.2 Core source files
- **`service/appstore.go`**
  - Responsible for managing a **single App Store**.
  - Core methods: `UpdateCatalog()` (download and atomic replace), `BuildCatalog()` (scan local directory and load into memory), `StoreRoot()` (get the store's working directory).
- **`service/appstore_management.go`**
  - **Global App Store manager**, singleton pattern.
  - Responsible for aggregating data across multiple stores (`Catalog()`), and managing app upgrade state and cache locks (`IsUpdateAvailable()`, `isAppUpgradable`).
- **`service/compose_app.go`**
  - Core business entity `ComposeApp`, wrapping the Docker Compose spec's structure.
  - Core methods: `StoreInfo()` (extracts the `x-casaos` extension metadata), `PullAndApply()` (pulls the new image and attempts to start it, rolling back on failure).
- **`route/v2/appstore.go`**
  - Handles HTTP API requests related to the App Store.
  - Core method: `ComposeAppStoreInfoList()` (fetches the list and applies CPU-architecture, category, etc. filter rules).
- **`common/constants.go`**
  - Holds constant definitions.
  - Key constant: `ComposeExtensionNameXNimoOS = "x-casaos"` (determines which top-level extension field the system reads).

---

## 4. Common Pitfalls and FAQ

- **Extension field naming conflict (`x-nimoos` vs `x-casaos`)**
  The constant in the code may be changed to `"x-casaos"`, but some legacy apps or the default sample `docker-compose.yaml` still use `"x-nimoos:"`. This causes `StoreInfo()` to throw an `extension x-casaos not found` error, and the app fails to show up correctly in the store list.
  - **Suggested fix**: unify the field name across the App Store source's YAML files, or add backward-compatibility handling in `StoreInfo()`'s parsing logic (checking both field names).
- **Update rollback trigger**
  If, during development, you find that updates to `/var/lib/nimoos/appstore/default/` are failing, check whether a leftover `.tmp` or `.backup` folder exists — this is usually caused by a permission issue or an interrupted download triggering the atomic replacement's safety rollback.
