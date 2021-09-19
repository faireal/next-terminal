package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"next-terminal/repository"
)

func UserCount(users repository.UserRepository) {
	prometheus.MustRegister(
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "nt_user_count",
			Help: "Total number of users.",
		}, func() float64 {
			i, _ := users.Count()
			return float64(i)
		}),
	)
}
