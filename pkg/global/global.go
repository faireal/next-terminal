package global

// var Cache *cache.Cache

// var Config *config.Config

// var Cron *cron.Cron

type Security struct {
	Rule string
	IP   string
}

var Securities []*Security

//func init() {
//	Cron = cron.New(cron.WithSeconds())
//	Cron.Start()
//}
