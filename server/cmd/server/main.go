// Modernchat-sso：Modernchat 家族的 SSO 认证服务入口。
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/modern-day-team/Modernchat-sso/server/internal/config"
	"github.com/modern-day-team/Modernchat-sso/server/internal/router"
	"github.com/modern-day-team/Modernchat-sso/server/internal/service"
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

	// 在线更新（站长侧）：WEB_UPDATE=false 整体关闭；可用性由 Updater.Available()
	// 检测（systemd 模板单元 + bin/ 下二进制）。更新源优先 RELEASE_API_URL/RELEASE_DL_URL，
	// 未配置时按 RELEASE_REPO_GITHUB 或默认官方仓库派生 GitHub Release 地址。
	var upd *service.Updater
	if cfg.WebUpdate {
		wd, wdErr := os.Getwd()
		root := ""
		if wdErr == nil {
			root = filepath.Dir(wd)
		}
		upd = service.NewUpdater(root, cfg.Port, service.ReleaseSource{
			API:   cfg.ReleaseAPIURL,
			DL:    cfg.ReleaseDLURL,
			Repo:  cfg.ReleaseRepoGitHub,
			Token: cfg.GitHubToken,
		})
		if upd.Available() {
			slog.Info("在线更新已启用", "repo", upd.Repo(), "root", upd.RootDir(), "port", upd.Port())
		}
	} else {
		upd = service.NewUpdater("", cfg.Port, service.ReleaseSource{})
		slog.Info("在线更新已关闭（WEB_UPDATE=false）")
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router.Setup(cfg, store.NewUserStore(db), upd),
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
