// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"next-terminal/handler/api"
	"next-terminal/models"
	ntcache "next-terminal/pkg/cache"
	"next-terminal/pkg/task"
	repository2 "next-terminal/repository"
)

func newweb() *cobra.Command {
	reset := &cobra.Command{
		Use:   "web",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			ntcache.MemCache = api.SetupCache()
			db := models.SetupDB()
			e := api.SetupRoutes(db)

			sessionRepo := repository2.NewSessionRepository(db)
			configRepo := repository2.NewConfigsRepository(db)
			ticker := task.NewTicker(sessionRepo, configRepo)
			ticker.SetupTicker()
			addr := viper.GetString("core.http")
			e.Run(addr)
		},
	}
	return reset
}

func init() {
	rootCmd.AddCommand(newweb())
}
