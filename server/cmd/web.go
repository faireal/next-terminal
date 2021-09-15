// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"next-terminal/pkg/global"
	"next-terminal/pkg/task"
	"next-terminal/server/api"
	"next-terminal/server/repository"
)

func newweb() *cobra.Command {
	reset := &cobra.Command{
		Use:   "web",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			global.Cache = api.SetupCache()
			db := api.SetupDB()
			e := api.SetupRoutes(db)

			sessionRepo := repository.NewSessionRepository(db)
			configRepo := repository.NewConfigsRepository(db)
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
