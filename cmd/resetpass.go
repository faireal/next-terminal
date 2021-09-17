// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package cmd

import (
	"github.com/spf13/cobra"
	"next-terminal/handler/api"
)

func newResetPass() *cobra.Command {
	reset := &cobra.Command{
		Use:   "reset-pass",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			api.ResetPassword(username)
		},
	}
	reset.PersistentFlags().StringVar(&username, "user", "", "user")
	return reset
}

func init() {
	rootCmd.AddCommand(newResetPass())
}
