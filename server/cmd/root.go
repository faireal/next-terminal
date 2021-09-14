// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package cmd

import (
	"fmt"
	"github.com/ergoapi/util/color"
	"github.com/ergoapi/util/zos"
	"github.com/ergoapi/zlog"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"next-terminal/constants"
)

var Version = "v0.5.0"

var (
	rootCmd = &cobra.Command{
		Use:        "nt",
		Short:      "",
		Long:       ``,
		PreRun: func(cmd *cobra.Command, args []string) {
			fmt.Printf(`
 _______                   __    ___________                  .__              .__   
 \      \   ____ ___  ____/  |_  \__    ___/__________  _____ |__| ____ _____  |  |  
 /   |   \_/ __ \\  \/  /\   __\   |    |_/ __ \_  __ \/     \|  |/    \\__  \ |  |  
/    |    \  ___/ >    <  |  |     |    |\  ___/|  | \/  Y Y  \  |   |  \/ __ \|  |__
\____|__  /\___  >__/\_ \ |__|     |____| \___  >__|  |__|_|  /__|___|  (____  /____/
        \/     \/      \/                     \/            \/        \/     \/      ` + constants.Version + "\n\n")
		},
	}
)

func init()  {
	cfg := zlog.Config{
		Simple:      false,
		HookFunc:    nil,
		WriteLog:    false,
		WriteJSON:   false,
		WriteConfig: zlog.WriteConfig{},
		ServiceName: "nt",
	}
	zlog.InitZlog(&cfg)
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&constants.CfgFile, "config", "", "config file (default is /conf/config.yml)")
}

func initConfig() {
	if constants.CfgFile == "" {
		constants.CfgFile = constants.Defaultcfgpath
		if zos.IsMacOS() {
			constants.CfgFile = "./config.yml"
		}
	}
	viper.SetConfigFile(constants.CfgFile)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		zlog.Debug("Using config file: %v", color.SGreen(viper.ConfigFileUsed()))
	}
	// reload
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		zlog.Debug("config changed: %v", color.SGreen(in.Name))
	})
}

func Execute() error {
	return rootCmd.Execute()
}