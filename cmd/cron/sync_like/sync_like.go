package sync_like

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/luyb177/silent-sign-backend/cmd/cron/internal"
	"github.com/luyb177/silent-sign-backend/internal/logic/cron"
)

var (
	daemon       bool
	syncInterval time.Duration

	// Cmd sync_like 子命令
	Cmd = &cobra.Command{
		Use:   "sync_like",
		Short: "sync like data from Redis to MySQL",
		RunE:  run,
	}
)

func init() {
	Cmd.Flags().BoolVar(&daemon, "daemon", false, "run as daemon, sync periodically")
	Cmd.Flags().DurationVar(&syncInterval, "interval", 60*time.Second, "sync interval in daemon mode")
}

func run(_ *cobra.Command, _ []string) error {
	defer internal.SvcCtx.RedisClient.Client.Close()

	if daemon {
		logx.Infof("starting like sync daemon, interval=%v", syncInterval)
		for {
			doSync()
			time.Sleep(syncInterval)
		}
	}

	doSync()
	return nil
}

func doSync() {
	logx.Info("starting like sync...")
	if err := cron.SyncLikes(internal.SvcCtx); err != nil {
		logx.Errorf("like sync failed: %v", err)
	} else {
		logx.Info("like sync completed")
	}
}
