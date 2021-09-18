// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package models

type Cloud struct {
	Model
	Provider   string `json:"provider"`
	ProviderCN string `json:"provider_cn"`
	Key        string `json:"key"`
	Secret     string `json:"secret"`
}

func init() {
	Migrate(Cloud{})
}
