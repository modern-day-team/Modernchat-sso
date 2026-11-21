// Modernchat-sso：Modernchat 家族的 SSO 认证服务入口。
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modern-day-team/Modernchat-sso/server/internal/config"
	"github.com/modern-day-team/Modernchat-sso/server/internal/router"
	"github.com/modern-day-team/Modernchat-sso/server/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("配置加载失败", "err", err)
		os.Exit(1)
	}
	db, err := store.Open(cfg.DBDSN, true)
	if err != nil {
		slog.Error("数据库连接失败（本地开发先在仓库根目录 docker compose up -d）", "err", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router.Setup(cfg, store.NewUserStore(db)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("SSO 服务启动", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP 服务异常退出", "err", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	slog.Info("收到退出信号，正在优雅关机……")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("关机超时", "err", err)
	}
	slog.Info("已退出")
}
