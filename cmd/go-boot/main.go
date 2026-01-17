package main

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// go install github.com/addls/go-boot/cmd/go-boot@latest

//go:embed templates/third_party
var thirdPartyTemplates embed.FS

var (
	Version      = "v0.1.0"
	ProtoVersion = "v0.1.0"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		runInit()
	case "api":
		runAPI()
	case "version":
		printVersion()
	default:
		printHelp()
		os.Exit(1)
	}
}

// --------------------
// commands
// --------------------

func runInit() {
	fmt.Println("[go-boot] init project")

	// 初始化 go.mod（如果不存在）
	initGoMod()

	// 安装 go-boot 依赖
	installGoBoot()

	// 安装 protoc 插件
	installProtocPlugins()

	// 创建标准目录结构
	createStandardDirs()

	// 生成 main.go
	generateMainGo()

	// 下载第三方 proto 文件
	downloadThirdPartyProtos()

	// 创建 protos 目录和示例文件
	createProtosDir()

	fmt.Println("[go-boot] init done")
}

func runAPI() {
	fmt.Println("[go-boot] generate api")

	// 确保 api 目录存在
	ensureDir("api")

	// 临时创建 Makefile
	makefilePath := "Makefile"
	createTempMakefile := false

	// 检查 Makefile 是否存在
	if _, err := os.Stat(makefilePath); os.IsNotExist(err) {
		createTempMakefile = true
		_ = os.WriteFile(makefilePath, []byte(defaultMakefile), 0644)
		fmt.Println("[go-boot] created temporary Makefile")
	}

	// 确保执行完后删除临时 Makefile
	defer func() {
		if createTempMakefile {
			if err := os.Remove(makefilePath); err == nil {
				fmt.Println("[go-boot] removed temporary Makefile")
			}
		}
	}()

	// 执行 make api
	cmd := exec.Command("make", "api")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("[go-boot] make api failed")
		os.Exit(1)
	}

	// 生成 service 文件
	generateServiceFiles()
}

func printVersion() {
	fmt.Printf("go-boot version: %s\n", Version)
	fmt.Printf("proto rules:     %s\n", ProtoVersion)
}

// --------------------
// helpers
// --------------------

func printHelp() {
	fmt.Println(`go-boot - minimal enterprise bootstrap cli

Usage:
  go-boot init       initialize project
  go-boot api        generate proto api code
  go-boot version    show version
`)
}

func ensureDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.MkdirAll(path, 0755)
	}
}

func createStandardDirs() {
	fmt.Println("[go-boot] creating standard directory structure...")

	dirs := []string{
		"internal/service",
		"internal/data",
	}

	for _, dir := range dirs {
		ensureDir(dir)
		fmt.Printf("[go-boot] created %s\n", dir)
	}
}

func ensureFile(path, content string) {
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("[go-boot] %s already exists, skip\n", path)
		return
	}

	_ = os.WriteFile(path, []byte(content), 0644)
	fmt.Printf("[go-boot] create %s\n", path)
}

func initGoMod() {
	// 检查 go.mod 是否存在
	if _, err := os.Stat("go.mod"); err == nil {
		return
	}

	fmt.Println("[go-boot] initializing go.mod...")

	// 获取模块名（使用当前目录名）
	moduleName := getModuleName()
	if moduleName == "your-service" {
		// 如果获取失败，使用当前目录名
		wd, err := os.Getwd()
		if err == nil {
			moduleName = filepath.Base(wd)
		} else {
			moduleName = "your-service"
		}
	}

	// 执行 go mod init
	cmd := exec.Command("go", "mod", "init", moduleName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("[go-boot] warning: failed to init go.mod: %v\n", err)
		return
	}
	fmt.Println("[go-boot] go.mod initialized")
}

func installGoBoot() {
	fmt.Println("[go-boot] installing go-boot...")

	// 执行 go get
	cmd := exec.Command("go", "get", "github.com/addls/go-boot")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("[go-boot] warning: failed to install go-boot: %v\n", err)
		fmt.Println("[go-boot] you can manually run: go get github.com/addls/go-boot")
		return
	}
	fmt.Println("[go-boot] go-boot installed successfully")
}

func installProtocPlugins() {
	fmt.Println("[go-boot] installing protoc plugins...")

	// 需要安装的插件列表
	plugins := []struct {
		name string
		path string
	}{
		{"protoc-gen-go", "google.golang.org/protobuf/cmd/protoc-gen-go"},
		{"protoc-gen-go-grpc", "google.golang.org/grpc/cmd/protoc-gen-go-grpc"},
		{"protoc-gen-go-http", "github.com/go-kratos/kratos/cmd/protoc-gen-go-http"},
		{"protoc-gen-go-errors", "github.com/go-kratos/kratos/cmd/protoc-gen-go-errors"},
	}

	for _, plugin := range plugins {
		fmt.Printf("[go-boot] installing %s...\n", plugin.name)
		cmd := exec.Command("go", "install", plugin.path+"@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("[go-boot] warning: failed to install %s: %v\n", plugin.name, err)
			fmt.Printf("[go-boot] you can manually run: go install %s@latest\n", plugin.path)
			continue
		}
		fmt.Printf("[go-boot] %s installed successfully\n", plugin.name)
	}

	fmt.Println("[go-boot] protoc plugins installation completed")
}

func generateMainGo() {
	// 获取服务名
	serviceName := getServiceName()

	// 生成 main.go 内容
	mainGoContent := fmt.Sprintf(defaultMainGo, serviceName)
	ensureFile("main.go", mainGoContent)
}

func getModuleName() string {
	// 尝试读取 go.mod 文件
	goModPath := "go.mod"
	if _, err := os.Stat(goModPath); err == nil {
		file, err := os.Open(goModPath)
		if err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "module ") {
					moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module "))
					if moduleName != "" {
						return moduleName
					}
				}
			}
		}
	}

	// 如果没有 go.mod，使用当前目录名
	wd, err := os.Getwd()
	if err == nil {
		return filepath.Base(wd)
	}

	return "your-service"
}

func getServiceName() string {
	// 尝试从 go.mod 获取模块名，去掉路径部分
	moduleName := getModuleName()
	if moduleName != "" {
		// 如果模块名包含路径分隔符，取最后一部分
		if strings.Contains(moduleName, "/") {
			parts := strings.Split(moduleName, "/")
			if len(parts) > 0 {
				lastPart := parts[len(parts)-1]
				if lastPart != "" {
					// 如果最后一部分是 "service-xxx" 格式，直接返回
					// 否则添加 "service-" 前缀
					if strings.HasPrefix(lastPart, "service-") {
						return lastPart
					}
					return "service-" + lastPart
				}
			}
		}
		// 如果模块名本身就是一个简单的名字，添加 service- 前缀
		if strings.HasPrefix(moduleName, "service-") {
			return moduleName
		}
		return "service-" + moduleName
	}

	// 使用当前目录名
	wd, err := os.Getwd()
	if err == nil {
		dirName := filepath.Base(wd)
		if strings.HasPrefix(dirName, "service-") {
			return dirName
		}
		return "service-" + dirName
	}

	return "service-user"
}

func downloadThirdPartyProtos() {
	thirdPartyDir := "third_party"

	// 检查 third_party 目录是否已存在且有内容
	if info, err := os.Stat(thirdPartyDir); err == nil && info.IsDir() {
		// 检查是否已经有 google/api 目录
		if _, err := os.Stat(filepath.Join(thirdPartyDir, "google", "api")); err == nil {
			fmt.Printf("[go-boot] third_party already exists, skip\n")
			return
		}
	}

	fmt.Println("[go-boot] copying third party proto files...")

	// 创建 third_party 目录
	ensureDir(thirdPartyDir)

	// 从嵌入的文件系统复制文件
	if err := copyEmbeddedFiles(thirdPartyTemplates, "templates/third_party", thirdPartyDir); err != nil {
		fmt.Printf("[go-boot] warning: failed to copy third party proto files: %v\n", err)
		return
	}

	fmt.Println("[go-boot] third party proto files copied successfully")
}

func copyEmbeddedFiles(embedFS embed.FS, srcDir, dstDir string) error {
	return fs.WalkDir(embedFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 计算目标路径（去掉 templates/third_party 前缀）
		relPath, err := filepath.Rel("templates/third_party", path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dstDir, relPath)

		if d.IsDir() {
			// 创建目录
			return os.MkdirAll(targetPath, 0755)
		}

		// 如果文件已存在，跳过
		if _, err := os.Stat(targetPath); err == nil {
			return nil
		}

		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("create directory failed: %w", err)
		}

		// 读取嵌入的文件
		data, err := embedFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded file failed: %w", err)
		}

		// 写入目标文件
		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return fmt.Errorf("write file failed: %w", err)
		}

		fmt.Printf("[go-boot] copied %s\n", relPath)
		return nil
	})
}

// createProtosDir 创建 protos 目录和示例 ping.proto 文件
func createProtosDir() {
	protosDir := "protos"
	ensureDir(protosDir)

	// 创建 network/v1 子目录
	networkV1Dir := filepath.Join(protosDir, "network", "v1")
	ensureDir(networkV1Dir)

	pingProtoPath := filepath.Join(networkV1Dir, "ping.proto")

	// 获取模块名，生成正确的 go_package
	moduleName := getModuleName()
	goPackage := moduleName + "/api/network/v1"
	if goPackage == "your-service/api/network/v1" {
		// 如果模块名是默认值，使用相对路径格式
		goPackage = "./api/network/v1"
	}

	pingProtoContent := fmt.Sprintf(defaultPingProto, goPackage)
	ensureFile(pingProtoPath, pingProtoContent)
}

// generateServiceFiles 扫描 protos 目录下的 proto 文件，为每个文件生成对应的 service 文件
func generateServiceFiles() {
	fmt.Println("[go-boot] generating service files...")

	protosDir := "protos"
	if _, err := os.Stat(protosDir); os.IsNotExist(err) {
		fmt.Printf("[go-boot] protos directory not found, skip service generation\n")
		return
	}

	// 扫描所有 proto 文件
	var protoFiles []string
	err := filepath.Walk(protosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".proto") {
			protoFiles = append(protoFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("[go-boot] warning: failed to scan proto files: %v\n", err)
		return
	}

	if len(protoFiles) == 0 {
		fmt.Printf("[go-boot] no proto files found, skip service generation\n")
		return
	}

	// 获取模块名
	moduleName := getModuleName()

	// 为每个 proto 文件生成对应的 service 文件
	for _, protoFile := range protoFiles {
		// 获取文件名（不含扩展名），例如 ping.proto -> ping
		baseName := strings.TrimSuffix(filepath.Base(protoFile), ".proto")

		// 生成 service 文件路径，例如 internal/service/ping.go
		serviceFileName := baseName + ".go"
		serviceFilePath := filepath.Join("internal/service", serviceFileName)

		// 如果文件已存在，跳过
		if _, err := os.Stat(serviceFilePath); err == nil {
			fmt.Printf("[go-boot] %s already exists, skip\n", serviceFilePath)
			continue
		}

		// 解析 proto 文件路径，确定 package 路径
		// proto 文件在 protos/ 目录，但生成的代码在 api/ 目录
		// 例如: protos/network/v1/network.proto -> api/network/v1/network.pb.go
		relPath, err := filepath.Rel("protos", protoFile)
		if err != nil {
			relPath = filepath.Base(protoFile)
		}
		protoDir := filepath.Dir(relPath)
		var packagePath string
		if protoDir == "." {
			// protos/ping.proto -> api/ping.pb.go，package 路径是 {moduleName}/api
			packagePath = moduleName + "/api"
		} else {
			// protos/network/v1/network.proto -> api/network/v1/network.pb.go，package 路径是 {moduleName}/api/network/v1
			packagePath = moduleName + "/api/" + protoDir
		}

		// 生成 service 名称（首字母大写），例如 ping -> Ping
		var serviceName string
		if len(baseName) > 0 {
			serviceName = strings.ToUpper(baseName[:1]) + baseName[1:]
		} else {
			serviceName = "Service"
		}

		// 生成 service 文件内容（package 名称从 packagePath 推导）
		content := generateServiceContent(serviceName, packagePath)

		// 确保 internal/service 目录存在
		ensureDir("internal/service")

		// 写入文件
		if err := os.WriteFile(serviceFilePath, []byte(content), 0644); err != nil {
			fmt.Printf("[go-boot] warning: failed to generate %s: %v\n", serviceFilePath, err)
			continue
		}

		fmt.Printf("[go-boot] generated service file: %s\n", serviceFilePath)
	}
}

// generateServiceContent 生成 service 文件内容
func generateServiceContent(serviceName, packagePath string) string {
	// serviceName 是首字母大写的文件名，例如 "Ping"
	// 结构体名称应该是 serviceName + "Service"，例如 "PingService"
	structName := serviceName + "Service"
	constructorName := "New" + structName

	// 从 packagePath 推导 package 名称（路径的最后一部分）
	// 例如: your-module/api/network/v1 -> v1
	packageName := filepath.Base(packagePath)
	if packageName == "" || packageName == "." {
		packageName = "api"
	}

	unimplementedType := packageName + ".Unimplemented" + structName + "Server"

	return fmt.Sprintf(`package service

import (
	"context"
	"%s"
)

type %s struct {
	%s
	// TODO: 添加依赖项
	// data *data.DataRepo
}

// %s 创建新的 %s 实例
func %s() *%s {
	return &%s{
		// TODO: 初始化依赖项
	}
}

// TODO: 实现 proto 中定义的 rpc 方法
// 示例：
// func (s *%s) Ping(ctx context.Context, req *%s.PingRequest) (*%s.PingReply, error) {
// 	// 实现业务逻辑
// 	return &%s.PingReply{}, nil
// }
`,
		packagePath,
		structName,
		unimplementedType,
		constructorName,
		structName,
		constructorName,
		structName,
		structName,
		structName,
		packageName,
		packageName,
		packageName,
	)
}

// --------------------
// templates
// --------------------

var defaultMakefile = `.PHONY: api

api:
	protoc \
	  --proto_path=protos \
	  --proto_path=third_party \
	  --go_out=paths=source_relative:api \
	  --go-grpc_out=paths=source_relative:api \
	  --go-http_out=paths=source_relative:api \
	  --go-errors_out=paths=source_relative:api \
	  $(shell find protos -name "*.proto")
`

var defaultPingProto = `syntax = "proto3";

package protos;

import "google/api/annotations.proto";

option go_package = "%s";

// PingService 提供 ping 接口
service PingService {
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
`

var defaultMainGo = `package main

import "github.com/addls/go-boot/bootstrap"

func main() {
	bootstrap.Run("%s")  // 自动查找 config.yaml 或使用默认配置
}
`
