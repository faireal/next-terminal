// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package cmd

import (
	"github.com/spf13/cobra"
	"next-terminal/server/api"
)

var (
	username string
)

func newResetTotp() *cobra.Command {
	reset := &cobra.Command{
		Use:   "reset-totp",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			api.ResetTotp(username)
		},
	}
	reset.PersistentFlags().StringVar(&username, "user", "", "username")
	return reset
}

func init() {
	rootCmd.AddCommand(newResetTotp())
}
