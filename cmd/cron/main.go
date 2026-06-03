package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/zeromicro/go-zero/core/conf"

	"github.com/luyb177/silent-sign-backend/cmd/cron/internal"
	"github.com/luyb177/silent-sign-backend/cmd/cron/sync_like"
	"github.com/luyb177/silent-sign-backend/internal/config"
	"github.com/luyb177/silent-sign-backend/internal/svc"
)

const codeFailure = 1

var (
	confPath string

	rootCmd = &cobra.Command{
		Use:   "cron",
		Short: "exec cron job",
		Long:  "exec cron job",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(codeFailure)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&confPath, "config", "etc/silent_sign.yaml", "config file")

	// 注册子命令
	rootCmd.AddCommand(sync_like.Cmd)
}

func initConfig() {
	var c config.Config
	conf.MustLoad(confPath, &c)
	internal.SvcCtx = svc.NewServiceContext(c)
}

func main() {
	Execute()
}
