package main

import (
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"
)

var totalRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "test1",
		Name:      "http_requests_total",
		Help:      "Number of get requests.",
	},
)

var databaseAccesses = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "test1",
		Name:      "database_accesses_total",
		Help:      "Amount of database accesses or operations",
	},
)

var totalErrors = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "test1",
		Name:      "errors_total",
		Help:      "Amount of errors",
	},
)

var registerRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "test1",
		Name:      "register_requests",
		Help:      "The amount of successful register requests on the web app",
	},
)

var loginRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "test1",
		Name:      "login_requests",
		Help:      "The amount of successful login requests on the web app",
	},
)

var unsuccessfulLoginRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "test1",
		Name:      "login_requests_failed",
		Help:      "The amount of unsuccessful login requests on the web app",
	},
)

var tweetRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "test1",
		Name:      "tweets_requests",
		Help:      "The amount of tweet requests on the web app",
	},
)
