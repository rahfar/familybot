package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	MourningJobCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "familybot_mourning_job_total",
		Help: "The total number of mourning job",
	})
)
var (
	RecvMsgCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "familybot_recieved_msg_total",
		Help: "The total number of recieved messages",
	})
)
var (
	SentMsgCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "familybot_sent_msg_total",
		Help: "The total number of sent messages",
	})
)

var (
	CommandCallsCaounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "familybot_command_calls_total",
		Help: "The total number of command calls",
	}, []string{"command"})
)
