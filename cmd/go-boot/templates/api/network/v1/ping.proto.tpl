syntax = "proto3";

package api.network.v1;

import "google/api/annotations.proto";

option go_package = "%s/api/network/v1";

// 提供 ping 接口
service Ping {
  // Ping 健康检查接口
  rpc Ping (PingRequest) returns (PingReply) {
    option (google.api.http) = {
      get: "/v1/ping"
    };
  }
}

message PingRequest {
}

message PingReply {
  string message = 1;
}
