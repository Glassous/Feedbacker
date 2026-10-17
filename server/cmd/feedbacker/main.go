// Command feedbacker 启动意见反馈服务，或执行管理员密码重置等运维命令。
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"feedbacker/internal/config"
	"feedbacker/internal/database"
	"feedbacker/internal/repository"
	"feedbacker/internal/router"
	"feedbacker/internal/service"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "admin":
			runAdminCommand(os.Args[2:])
			return
		case "version", "--version", "-v":
			fmt.Printf("feedbacker %s\n", version)
			return
		case "help", "--help", "-h":
			printUsage()
			return
		}
	}

	if err := run(); err != nil {
		slog.Error("服务启动失败", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := newLogger(cfg)

	startupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, db, err := database.Connect(cfg, logger)
	if err != nil {
		return err
	}
	defer func() {
		disconnectCtx, cancelDisconnect := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelDisconnect()
		if err := client.Disconnect(disconnectCtx); err != nil {
			logger.Error("断开 MongoDB 连接失败", "error", err)
		}
	}()

	if err := database.EnsureIndexes(startupCtx, db); err != nil {
		return err
	}

	authService := service.NewAuthService(
		cfg,
		repository.NewAdminRepository(db),
		repository.NewSettingsRepository(db),
		logger,
	)
	if err := authService.Bootstrap(startupCtx, os.Stdout); err != nil {
		return err
	}

	feedbackService := service.NewFeedbackService(repository.NewFeedbackRepository(db))
	engine := router.New(cfg, logger, db, authService, feedbackService, version)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Info("Feedbacker 已启动",
		"addr", server.Addr,
		"database", cfg.MongoDatabase,
		"site", cfg.SiteName,
	)

	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("HTTP 服务异常退出: %w", err)
	case sig := <-stop:
		logger.Info("收到退出信号，开始关闭服务", "signal", sig.String())
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("关闭 HTTP 服务失败: %w", err)
	}

	logger.Info("服务已安全退出")
	return nil
}

func runAdminCommand(args []string) {
	if len(args) == 0 {
		printAdminUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "reset-password":
		flags := flag.NewFlagSet("reset-password", flag.ExitOnError)
		password := flags.String("password", "", "指定新密码；留空则随机生成一次性密码")
		username := flags.String("username", "", "管理员用户名；默认使用 ADMIN_USERNAME")
		_ = flags.Parse(args[1:])

		if err := resetPassword(*username, *password); err != nil {
			fmt.Fprintln(os.Stderr, "重置密码失败:", err)
			os.Exit(1)
		}
	default:
		printAdminUsage()
		os.Exit(1)
	}
}

func resetPassword(username, password string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := newLogger(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, db, err := database.Connect(cfg, logger)
	if err != nil {
		return err
	}
	defer func() {
		disconnectCtx, cancelDisconnect := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelDisconnect()
		_ = client.Disconnect(disconnectCtx)
	}()

	if err := database.EnsureIndexes(ctx, db); err != nil {
		return err
	}

	authService := service.NewAuthService(
		cfg,
		repository.NewAdminRepository(db),
		repository.NewSettingsRepository(db),
		logger,
	)

	result, err := authService.ResetPassword(ctx, username, password)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("密码重置成功")
	fmt.Printf("  用户名: %s\n", result.Username)
	fmt.Printf("  新密码: %s\n", result.Password)
	if result.Generated {
		fmt.Println("  该密码为一次性密码，登录后必须立即修改。")
	}
	if result.Created {
		fmt.Println("  已创建新的管理员账号。")
	}
	fmt.Println()
	return nil
}

func printUsage() {
	fmt.Print(`Feedbacker 意见反馈微服务

用法:
  feedbacker                              启动 HTTP 服务（默认）
  feedbacker admin reset-password         重置管理员密码（随机一次性密码）
  feedbacker admin reset-password -password <新密码>
  feedbacker version                      查看版本

环境变量见 docs/deploy.md 或 .env.example。
`)
}

func printAdminUsage() {
	fmt.Print(`管理员运维命令:

  feedbacker admin reset-password [--password <新密码>] [--username <用户名>]

不带 --password 时会生成随机一次性密码，并强制下次登录后修改。
`)
}

func newLogger(cfg *config.Config) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})
	return slog.New(handler).With("service", "feedbacker", "version", version)
}
