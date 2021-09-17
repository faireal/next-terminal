// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package cron

import "github.com/robfig/cron/v3"

var Cron *cron.Cron

func init() {
	Cron = cron.New(cron.WithSeconds())
	Cron.Start()
}
