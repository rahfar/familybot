package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	PingCallCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "familybot_ping_call_total",
		Help: "The total number of ping calls",
	})
)
var (
	WeatherCallCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "familybot_weather_call_total",
		Help: "The total number of weather calls",
	})
)
var (
	GPTCallCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "familybot_gpt_call_total",
		Help: "The total number of gpt calls",
	})
)
var (
	AnecdoteCallCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "familybot_anecdote_call_total",
		Help: "The total number of anecdote calls",
	})
)
var (
	NewsCallCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "familybot_news_call_total",
		Help: "The total number of news calls",
	})
)
var (
	TranscriptCallCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "familybot_transcript_call_total",
		Help: "The total number of transcript calls",
	})
)
var (
	ImageCallCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "familybot_image_call_total",
		Help: "The total number of image calls",
	})
)
var (
	RevisionCallCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "familybot_revision_call_total",
		Help: "The total number of revision calls",
	})
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
