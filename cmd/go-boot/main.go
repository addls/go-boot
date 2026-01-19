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

//go:embed templates
var templatesFS embed.FS

var (
	Version      = "v0.1.3"
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

	// 安装 kratos CLI
	if !checkKratosCLI() {
		fmt.Println("[go-boot] kratos CLI not found, installing...")
		installKratosCLI()
	}

	// 安装 protoc 插件
	installProtocPlugins()

	// 创建标准目录结构
	createStandardDirs()

	// 复制 templates 目录下的所有文件到当前目录（包括 main.go）
	copyTemplates()

	// 安装 go-boot 依赖（在生成 main.go 之后，因为它会导入 go-boot）
	installGoBoot()

	// 整理依赖关系
	runGoModTidy()

	fmt.Println("[go-boot] init done")
}

func runAPI() {
	fmt.Println("[go-boot] generate api")

	// 确保 api 目录存在
	ensureDir("api")

	// 检查 kratos CLI 是否安装（应该在 init 时已安装）
	if !checkKratosCLI() {
		fmt.Println("[go-boot] error: kratos CLI not found")
		fmt.Println("[go-boot] please run: go-boot init first, or manually install: go install github.com/go-kratos/kratos/cmd/kratos/v2@latest")
		os.Exit(1)
	}

	// 扫描 api 目录下的所有 proto 文件
	apiDir := "api"
	if _, err := os.Stat(apiDir); os.IsNotExist(err) {
		fmt.Printf("[go-boot] api directory not found\n")
		os.Exit(1)
	}

	var protoFiles []string
	err := filepath.Walk(apiDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".proto") {
			protoFiles = append(protoFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("[go-boot] failed to scan proto files: %v\n", err)
		os.Exit(1)
	}

	if len(protoFiles) == 0 {
		fmt.Printf("[go-boot] no proto files found in api directory\n")
		os.Exit(1)
	}

	// 使用 kratos proto client 生成 API 代码
	for _, protoFile := range protoFiles {
		fmt.Printf("[go-boot] generating api code for %s...\n", protoFile)
		cmd := exec.Command("kratos", "proto", "client", protoFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("[go-boot] failed to generate api code for %s: %v\n", protoFile, err)
			os.Exit(1)
		}
	}

	fmt.Println("[go-boot] proto api code generation completed")

	// 使用 kratos proto server 生成 service 代码到 internal/service
	ensureDir("internal/service")
	for _, protoFile := range protoFiles {
		fmt.Printf("[go-boot] generating service code for %s...\n", protoFile)
		cmd := exec.Command("kratos", "proto", "server", protoFile, "-t", "internal/service")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("[go-boot] failed to generate service code for %s: %v\n", protoFile, err)
			os.Exit(1)
		}
	}

	fmt.Println("[go-boot] service code generation completed")
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

func runGoModTidy() {
	fmt.Println("[go-boot] running go mod tidy...")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("[go-boot] warning: failed to run go mod tidy: %v\n", err)
		fmt.Println("[go-boot] you can manually run: go mod tidy")
		return
	}
	fmt.Println("[go-boot] go mod tidy completed")
}

func installProtocPlugins() {
	fmt.Println("[go-boot] installing protoc plugins...")

	// 需要安装的插件列表
	// 官方文档明确要求: protoc-gen-go
	// 使用 kratos proto 命令时还需要: protoc-gen-go-grpc, protoc-gen-go-http, protoc-gen-go-errors
	plugins := []struct {
		name string
		path string
	}{
		{"protoc-gen-go", "google.golang.org/protobuf/cmd/protoc-gen-go"},
		{"protoc-gen-go-grpc", "google.golang.org/grpc/cmd/protoc-gen-go-grpc"},
		{"protoc-gen-go-http", "github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2"},
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

func checkKratosCLI() bool {
	// 使用 exec.LookPath 检查命令是否存在
	_, err := exec.LookPath("kratos")
	return err == nil
}

func installKratosCLI() {
	fmt.Println("[go-boot] installing kratos CLI...")
	cmd := exec.Command("go", "install", "github.com/go-kratos/kratos/cmd/kratos/v2@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("[go-boot] warning: failed to install kratos CLI: %v\n", err)
		fmt.Println("[go-boot] you can manually run: go install github.com/go-kratos/kratos/cmd/kratos/v2@latest")
		return
	}
	fmt.Println("[go-boot] kratos CLI installed successfully")
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

func copyTemplates() {
	fmt.Println("[go-boot] copying template files...")

	moduleName := getModuleName()

	// 从嵌入的文件系统复制所有文件
	if err := copyEmbeddedFiles(templatesFS, "templates", ".", moduleName); err != nil {
		fmt.Printf("[go-boot] warning: failed to copy template files: %v\n", err)
		return
	}

	fmt.Println("[go-boot] template files copied successfully")
}

func copyEmbeddedFiles(embedFS embed.FS, srcDir, dstDir string, moduleName string) error {
	return fs.WalkDir(embedFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 计算目标路径（去掉 templates 前缀）
		relPath, err := filepath.Rel("templates", path)
		if err != nil {
			return err
		}

		// 对于所有 .tpl 文件，去掉 .tpl 后缀
		targetRelPath := relPath
		if strings.HasSuffix(relPath, ".tpl") {
			targetRelPath = strings.TrimSuffix(relPath, ".tpl")
		}

		targetPath := filepath.Join(dstDir, targetRelPath)

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

		// 如果是模板文件（.go.tpl 或 .proto），替换模板变量
		content := string(data)
		if strings.HasSuffix(relPath, ".tpl") {
			// 统计 %s 的数量，用于 fmt.Sprintf
			count := strings.Count(content, "%s")
			if count > 0 {
				args := make([]interface{}, count)
				// 所有 %s 都替换为模块名
				for i := range args {
					args[i] = moduleName
				}
				content = fmt.Sprintf(content, args...)
			}
		}

		// 写入目标文件
		if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("write file failed: %w", err)
		}

		fmt.Printf("[go-boot] copied %s\n", targetRelPath)
		return nil
	})
}

// --------------------
// templates
// --------------------
