package bootstrap

// Run 启动应用
// 使用 Wire 进行依赖注入，所有依赖关系由 Wire 自动管理
func Run(service string, opts ...Option) error {
	app, err := InitializeApp(service, opts...)
	if err != nil {
		return err
	}
	return app.Run()
}
