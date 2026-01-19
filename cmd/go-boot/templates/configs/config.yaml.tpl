# go-boot 配置文件示例
# 将此文件复制为 config.yaml 并根据需要修改

server:
  grpc:
    addr: ":9000"      # gRPC 服务地址，留空则不启动 gRPC 服务
    timeout: "30s"     # gRPC 请求超时时间（可选，如 "30s", "1m"）
  http:
    addr: ":8000"      # HTTP 服务地址，留空则不启动 HTTP 服务
    timeout: "30s"     # HTTP 请求超时时间（可选，如 "30s", "1m"）

middleware:
  enableMetrics: false  # 是否启用监控指标（基于 OpenTelemetry）
  enableTracing: false  # 是否启用链路追踪（基于 OpenTelemetry）

app:
  version: "v1.0.0"     # 应用版本（可选，默认 v1.0.0）
  stopTimeout: "10s"    # 优雅关闭超时时间（可选，默认 10s）
  # 服务注册与发现配置（可选）
  # discovery:
  #   type: "etcd"        # 注册中心类型：etcd, consul（nacos 等后续支持）
  #   register: true      # 是否开启服务注册（默认 false）
  #   endpoints:          # 注册中心地址列表（配置后自动连接 Discovery，供客户端服务发现使用）
  #     - "127.0.0.1:2379"  # etcd 示例，consul 通常使用 "127.0.0.1:8500"
  #   timeout: "5s"       # 连接超时时间（可选）
  # 服务元数据（可选，用于服务注册时的标签）
  metadata:
    env: "dev"
    zone: "zone-a"

log:
  output: "logs/app.log"  # 日志输出位置：stdout, stderr, 或文件路径（默认 logs/app.log）
  level: "info"           # 日志级别：debug, info, warn, error（默认 info）

data:
  database:
    driver: mysql
    source: root:root@tcp(127.0.0.1:3306)/test?parseTime=True&loc=Local
  redis:
    addr: 127.0.0.1:6379
    read_timeout: "1s"
    write_timeout: "1s"