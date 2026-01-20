# github.com/addls/go-boot

公司级统一微服务基座 - 基于 Kratos 的 Go 微服务公共底座

## 概述

`go-boot` 是一个公司级的微服务公共底座，提供统一的标准、统一的升级机制和统一的治理能力。所有业务服务只需要关心业务代码，公共能力可以集中演进、自动升级。

## 核心特性

- ✅ **极简启动**：一行代码启动服务
- ✅ **配置驱动**：支持配置文件自动加载
- ✅ **统一日志**：基于 zap 的统一日志实现
- ✅ **统一中间件**：Recovery、Logging、Tracing、Metrics
- ✅ **统一响应格式**：标准化的 HTTP 响应结构
- ✅ **可扩展**：支持自定义中间件

## 快速开始

### 1. 安装 CLI 工具

```bash
go install github.com/addls/go-boot/cmd/go-boot@latest
```

### 2. 初始化项目

```bash
mkdir my-service && cd my-service
go-boot init
```

`go-boot init` 会自动完成：
- ✅ 初始化 `go.mod`
- ✅ 安装 `go-boot` 依赖
- ✅ 安装 protoc 插件
- ✅ 创建标准目录结构（`internal/service`, `internal/data`）
- ✅ 生成 `main.go`（一行代码启动）
- ✅ 创建 `protos/network/v1/` 目录和示例 `ping.proto`
- ✅ 复制第三方 proto 文件到 `third_party/`

### 3. 生成 API 代码

```bash
go-boot api
```

`go-boot api` 会：
- ✅ 执行 `make api` 生成 protobuf 代码（输出到 `api/` 目录）
- ✅ 自动为每个 proto 文件生成对应的 service 文件（`internal/service/*.go`）

**目录说明：**
- `protos/` - proto 源文件统一管理
- `api/` - 生成的 protobuf 代码
- `internal/service/` - 自动生成的 service 实现

### 启动服务

`go-boot init` 会自动生成 `main.go`，一行代码即可启动：

```go
package main

import "github.com/addls/go-boot/bootstrap"

func main() {
    bootstrap.Run("service-user")  // 自动查找 config.yaml 或使用默认配置
}
```

**配置文件示例 (config.yaml)**

```yaml
server:
  grpc:
    addr: ":9000"
  http:
    addr: ":8000"

middleware:
  enableMetrics: true
  enableTracing: true
```

### 进阶使用

**添加自定义中间件**

```go
package main

import (
    "github.com/addls/go-boot/bootstrap"
    "github.com/go-kratos/kratos/v2/middleware/validate"
)

func main() {
    bootstrap.Run("service-user",
        bootstrap.WithMiddleware(
            validate.Validator(),
        ),
    )
}
```

**直接传入配置（覆写配置文件）**

```go
package main

import (
    "github.com/addls/go-boot/bootstrap"
    "github.com/addls/go-boot/config"
)

func main() {
    // 配置会先加载文件，然后用这里传入的配置覆写文件中的对应字段
    cfg := &config.Config{
        Server: config.Server{
            GRPC: config.ServerConfig{Addr: ":9000"},
            HTTP: config.ServerConfig{Addr: ":8000"},
        },
        Middleware: config.Middleware{
            EnableMetrics: true,
            EnableTracing: true,
        },
    }
    bootstrap.Run("service-user", bootstrap.WithConfig(cfg))
}
```

**注册 HTTP 路由**

底座支持两种路由注册方式，业务代码可根据场景选择：

**方式1：使用 Protobuf 生成路由（Kratos 推荐，适合复杂 API）**

1. **定义 proto 文件**（在 `protos/network/v1/` 目录下）：

```protobuf
// protos/network/v1/ping.proto
syntax = "proto3";

package protos;

import "google/api/annotations.proto";

option go_package = "your-module/api/network/v1";

service PingService {
  rpc Ping (PingRequest) returns (PingReply) {
    option (google.api.http) = {
      get: "/v1/ping"
    };
  }
}

message PingRequest {}
message PingReply {
  string message = 1;
}
```

2. **生成代码**：

```bash
go-boot api
```

这会自动生成：
- `api/network/v1/ping.pb.go` - protobuf 消息定义
- `api/network/v1/ping_grpc.pb.go` - gRPC 服务定义
- `api/network/v1/ping_http.pb.go` - HTTP 路由定义
- `internal/service/ping.go` - service 实现文件（包含 `NewPingService()` 构造函数）

3. **实现 service 方法**（编辑 `internal/service/ping.go`）：

```go
package service

import (
    "context"
    "your-module/api/network/v1"
)

type PingService struct {
    v1.UnimplementedPingServiceServer
}

func NewPingService() *PingService {
    return &PingService{}
}

func (s *PingService) Ping(ctx context.Context, req *v1.PingRequest) (*v1.PingReply, error) {
    return &v1.PingReply{Message: "pong"}, nil
}
```

4. **注册路由**：

```go
package main

import (
    "github.com/addls/go-boot/bootstrap"
    "github.com/go-kratos/kratos/v2/transport/http"
    
    "your-module/api/network/v1"
    "your-module/internal/service"
)

func main() {
    // 创建服务实例
    pingService := service.NewPingService()
    
    // 生成 HTTP 处理器（Kratos 自动生成）
    pingHandler := v1.NewPingServiceHandler(pingService)
    
    bootstrap.Run("service-user",
        bootstrap.WithHTTPOptions(func(srv *http.Server) {
            // 注册 Protobuf 生成的 HTTP 处理器
            // 路径前缀需要与 proto 中定义的 HTTP 路径一致
            srv.HandlePrefix("/v1", pingHandler)
        }),
    )
}
```

**说明：**
- `go-boot api` 会自动为每个 proto 文件生成对应的 service 文件
- `v1.NewPingServiceHandler(pingService)` 是 Kratos 自动生成的函数，返回 `http.Handler`
- `srv.HandlePrefix("/v1", pingHandler)` 注册所有以 `/v1` 开头的路由
- 底座会自动应用统一响应格式和中间件
- 可以混合使用：部分接口用 Protobuf，部分用手动路由

**方式2：手动路由注册（适合简单接口，推荐方式，语义更清晰）**

```go
package main

import (
    "github.com/addls/go-boot/bootstrap"
    "github.com/go-kratos/kratos/v2/transport/http"
)

// Kratos HandlerFunc 示例（推荐）
func getUserHandler(ctx http.Context) error {
    // 从路径参数获取 ID
    id := ctx.Vars().Get("id")
    
    // 业务逻辑
    user := map[string]interface{}{
        "id":   id,
        "name": "Alice",
    }
    
    // 直接返回数据，底座会自动包装为统一响应格式
    return ctx.JSON(200, user)
}

func createUserHandler(ctx http.Context) error {
    var user map[string]interface{}
    // 绑定请求体
    if err := ctx.Bind(&user); err != nil {
        return err
    }
    // 业务逻辑...
    return ctx.JSON(200, user)
}

func main() {
    bootstrap.Run("service-user",
        bootstrap.WithHTTPOptions(func(srv *http.Server) {
            // 注册路由
            r := srv.Route("/api/v1")
            r.GET("/users/{id}", getUserHandler)
            r.POST("/users", createUserHandler)
            
            // 路由组
            users := r.Group("/users")
            users.GET("", listUsersHandler)
            users.DELETE("/{id}", deleteUserHandler)
        }),
    )
}
```

**使用标准 net/http Handler：**

```go
package main

import (
    "net/http"
    
    "github.com/addls/go-boot/bootstrap"
    "github.com/go-kratos/kratos/v2/transport/http"
)

func main() {
    bootstrap.Run("service-user",
        bootstrap.WithHTTPOptions(func(srv *http.Server) {
            // 使用 Handle 注册标准处理器
            srv.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(200)
                w.Write([]byte("ok"))
            }))
            
            // 使用 HandlePrefix 注册前缀路由
            srv.HandlePrefix("/static/", http.FileServer(http.Dir("./static")))
        }),
    )
}
```

**配置服务注册与发现**

只需在配置文件中配置注册中心，底座会自动创建并注册服务：

```go
package main

import "github.com/addls/go-boot/bootstrap"

func main() {
    bootstrap.Run("service-user")  // 底座自动根据配置文件创建并注册服务
}
```

配置文件 (config.yaml):
```yaml
app:
  discovery:
    type: "etcd"          # 注册中心类型：etcd 或 consul（nacos 等后续支持）
    register: true        # 是否开启服务注册；只做服务发现可以为 false
    endpoints:
      - "127.0.0.1:2379"  # etcd 支持多个地址，consul 通常只需要一个
    timeout: "5s"         # 连接/注册超时（可选）
  metadata:
    env: "prod"         # 服务标签，用于服务发现过滤
    zone: "zone-a"
```

**Consul 配置示例：**
```yaml
app:
  discovery:
    type: "consul"
    register: true
    endpoints:
      - "127.0.0.1:8500"  # consul 默认端口 8500
    timeout: "5s"
  metadata:
    env: "prod"
    zone: "zone-a"
```

**添加生命周期钩子**

```go
package main

import (
    "context"
    
    "github.com/addls/go-boot/bootstrap"
    "github.com/go-kratos/kratos/v2"
)

func main() {
    bootstrap.Run("service-user",
        bootstrap.WithAppOptions(
            kratos.BeforeStart(func(ctx context.Context) error {
                // 启动前初始化（如数据库迁移）
                return nil
            }),
            kratos.AfterStart(func(ctx context.Context) error {
                // 启动后处理（如健康检查）
                return nil
            }),
        ),
    )
}
```

## 配置说明

### 配置文件查找顺序

1. 先加载文件配置（按以下优先级）：
   - `WithConfigFile()` 指定的路径
   - `./config.yaml`
   - `./configs/config.yaml`
   - 默认配置
2. 如果使用了 `WithConfig()`，用传入的配置覆写文件配置中的对应字段

### 配置项

#### 底座统一管理的配置

**服务器配置：**
| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `server.grpc.addr` | gRPC 服务地址 | `:9000` |
| `server.grpc.timeout` | gRPC 请求超时（如 "30s", "1m"） | 使用 Kratos 默认值 |
| `server.http.addr` | HTTP 服务地址 | `:8000` |
| `server.http.timeout` | HTTP 请求超时（如 "30s", "1m"） | 使用 Kratos 默认值 |

**中间件配置：**
| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `middleware.enableMetrics` | 启用监控指标 | `false` |
| `middleware.enableTracing` | 启用链路追踪 | `false` |

**应用配置：**
| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `app.version` | 应用版本 | `v1.0.0` |
| `app.stopTimeout` | 优雅关闭超时（如 "10s", "30s"） | `10s` |
| `app.discovery.type` | 注册中心类型（etcd、consul，nacos 等后续支持） | 无（不启用） |
| `app.discovery.register` | 是否开启服务注册（只做服务发现时可为 false） | `false` |
| `app.discovery.endpoints` | 注册中心地址列表 | 无 |
| `app.discovery.timeout` | 注册/发现超时时间 | 无 |
| `app.metadata` | 服务元数据（用于服务注册时的标签，如 env、zone 等） | 无 |

> **Metadata 说明**：
> - **`app.metadata`**：服务注册时的静态标签（如 `env: prod`、`zone: zone-a`），用于服务发现和路由，通过 `kratos.Metadata()` 设置
> - **请求 Metadata 中间件**：已作为默认中间件自动启用，用于服务间传递动态元数据（如 `trace-id`、`request-id` 等），通过 `metadata.Server()` 实现

#### 业务代码控制的配置

**Kratos App 选项**（通过 `WithAppOptions()` 传入）：
- **生命周期钩子**：`kratos.BeforeStart()` / `kratos.AfterStart()` / `kratos.BeforeStop()` / `kratos.AfterStop()`
- **上下文**：`kratos.Context()` - 自定义上下文


## 中间件

### 底座统一管理（自动应用）

| 中间件 | 说明 | 默认状态 |
|--------|------|----------|
| **Recovery** | panic 恢复 | 必须启用 |
| **Metadata** | 请求元数据传递（服务间通信基础能力） | 必须启用 |
| **Logging** | 请求日志 | 必须启用 |
| **Tracing** | 链路追踪（OpenTelemetry） | 可选，配置启用 |
| **Metrics** | 监控指标（OpenTelemetry） | 可选，配置启用 |

### 业务代码扩展（按需添加）

| 中间件 | 说明 | 使用方式 |
|--------|------|----------|
| **Validate** | 参数校验 | `WithMiddleware(validate.Validator())` |
| **Auth** | 认证授权 | `WithMiddleware(auth.JWT(...))` |
| **RateLimit** | 限流 | `WithMiddleware(ratelimit.Server(...))` |

## 统一响应格式

底座自动统一所有 HTTP 接口的响应格式，无需业务代码手动处理。

### 响应结构

```json
{
  "code": 200,           // 状态码：200 表示成功，其他表示失败
  "message": "success", // 消息
  "data": {}            // 数据（成功时返回）
}
```

错误响应：
```json
{
  "code": 500,              // 错误码
  "message": "错误信息",     // 错误消息
  "error": "错误信息"       // 错误详情
}
```

### 使用方式

底座会自动包装所有 HTTP 响应为统一格式，业务代码只需返回数据或错误：

```go
// Kratos HandlerFunc（推荐）
func getUserHandler(ctx http.Context) error {
    id := ctx.Vars().Get("id")
    user := &User{ID: 1, Name: "Alice"}
    return ctx.JSON(200, user)  // 自动包装为统一响应格式
}

// 错误处理
func createUserHandler(ctx http.Context) error {
    var user User
    if err := ctx.Bind(&user); err != nil {
        return err  // 自动包装为统一错误格式
    }
    // 或使用 Kratos 错误
    return errors.BadRequest("INVALID_ID", "用户ID不能为空")
}
```

### 响应格式说明

- **成功响应**：`code=200`，`data` 字段包含业务数据
- **错误响应**：`code≠200`，`error` 字段包含错误信息
- **HTTP 状态码**：统一返回 `200 OK`，错误信息在响应体的 `code` 字段中

## 目录结构

### go-boot 底座项目结构

```
github.com/addls/go-boot/
├── go.mod
├── bootstrap/              # 统一启动器
│   ├── app.go              # 对外暴露 Run 接口
│   ├── options.go          # 启动参数 Option 定义
│   ├── providers.go        # Wire Provider 集合与具体 Provider
│   ├── wire.go             # Wire 声明文件（开发环境使用）
│   └── wire_gen.go         # Wire 生成的依赖注入代码（自动生成）
├── config/                 # 统一配置
│   └── config.go
├── common/                 # 通用组件（常量等）
│   └── constants.go
├── log/                    # 统一日志
│   ├── logger.go
│   └── adapter.go
├── response/               # 统一响应格式
│   ├── response.go
│   └── encoder.go
├── middleware/             # 统一中间件
│   ├── recovery.go         # panic 恢复
│   ├── metadata.go         # 元数据传递
│   ├── logging.go          # 统一日志
│   ├── tracing.go          # 链路追踪
│   └── metrics.go          # 指标采集
├── registry/               # 服务注册与发现
│   ├── registry.go
│   ├── etcd/               # etcd 实现
│   └── consul/             # consul 实现
├── cmd/go-boot/            # CLI 工具
│   └── main.go
└── README.md
```

### 业务项目结构（使用 go-boot init 生成）

```
your-service/
├── go.mod
├── main.go                 # 自动生成，一行代码启动
├── config.yaml             # 配置文件（可选）
├── Makefile                # 临时生成，用于 make api
├── protos/                 # proto 源文件统一管理
│   └── network/
│       └── v1/
│           └── ping.proto  # 示例文件
├── api/                    # 生成的 protobuf 代码（go-boot api 生成）
│   └── network/
│       └── v1/
│           ├── ping.pb.go
│           ├── ping_grpc.pb.go
│           ├── ping_http.pb.go
│           └── ping_errors.pb.go
├── internal/
│   ├── service/            # service 实现（go-boot api 自动生成）
│   │   └── ping.go         # 自动生成，包含 NewPingService()
│   └── data/               # 数据访问层
└── third_party/            # 第三方 proto 文件（go-boot init 自动复制）
    ├── google/
    │   └── api/
    └── errors/
```

## API 参考

### bootstrap.Run

```go
func Run(service string, opts ...Option) error
```

启动应用。`service` 为服务名称。

### 选项函数

| 函数 | 说明 |
|------|------|
| `WithConfigFile(path)` | 指定配置文件路径 |
| `WithConfig(cfg)` | 直接传入配置（覆写文件配置） |
| `WithMiddleware(...)` | 添加自定义中间件 |
| `WithGRPCOptions(...)` | 额外的 gRPC 服务器选项 |
| `WithHTTPOptions(...)` | 额外的 HTTP 服务器选项（可用于注册路由） |
| `WithAppOptions(...)` | 额外的 Kratos App 选项（生命周期钩子等） |

## License

MIT
