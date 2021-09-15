// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package cmd

import (
	"github.com/spf13/cobra"
	"next-terminal/server/api"
	"next-terminal/server/utils"
)

var encryptionKey string

func newResetEncryptionKey() *cobra.Command {
	reset := &cobra.Command{
		Use:   "reset-key",
		Short: "重制加密的key",
		Run: func(cmd *cobra.Command, args []string) {
			api.ChangeEncryptionKey(utils.GetEncryptionKey(), encryptionKey)
		},
	}
	reset.PersistentFlags().StringVar(&encryptionKey, "newkey", "", "新key")
	return reset
}

func init() {
	rootCmd.AddCommand(newResetEncryptionKey())
}
