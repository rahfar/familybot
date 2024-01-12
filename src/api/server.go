package api

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func StartServer() {
	http.HandleFunc("/ping", PingHandler)
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":8080", nil)
	panic(err)
}
