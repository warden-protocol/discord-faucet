package discord

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

//nolint:gochecknoglobals // This is a Prometheus metric
var (
	reqCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "discord_faucet",
		Name:      "requests_total",
		Help:      "The total number of Faucet requests",
	})
	reqFailed = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "discord_faucet",
		Name:      "requests_failed_total",
		Help:      "The total number of failed Faucet requests",
	})
	reqDenied = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "discord_faucet",
		Name:      "requests_denied_total",
		Help:      "The total number of denied Faucet requests",
	})
	reqBad = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "discord_faucet",
		Name:      "requests_bad_total",
		Help:      "The total number of bad Faucet requests",
	})
	usersCooldown = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "discord_faucet",
		Name:      "users_cooldown",
		Help:      "The total number of denied Faucet requests",
	})
)
