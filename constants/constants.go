// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package constants

import "fmt"

var (
	Commit  = "unknown"
	Date    = "unknown"
	Release = "unknown"
	Version = fmt.Sprintf("%s-%s-%s", Release, Date, Commit)
	CfgFile string
)

const (
	// Defaultcfgpath 默认配置文件
	Defaultcfgpath = "/conf/nt.yaml"
)