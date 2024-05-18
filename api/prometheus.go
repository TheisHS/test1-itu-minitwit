package main

import (
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"
)

var totalRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "api",
		Name:      "http_requests_total",
		Help:      "Number of get requests.",
	},
)

var databaseAccesses = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "api",
		Name:      "database_accesses_total",
		Help:      "Amount of database accesses or operations",
	},
)

var totalErrors = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "api",
		Name:      "errors_total",
		Help:      "Amount of errors",
	},
)

// add counters for requests and errors for the most used endpoints according to database cluster statistics.

// This is the counter for the register requests
// Register is the only endpoint that calls the getUserID function in the API
var totalGetUserIDRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "api",
		Name:      "get_user_id_requests",
		Help:      "The amount of get user id requests on the api - for API this is only the Register that calls it",
	},
)
var unsuccessfulGetUserIDRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "api",
		Name:      "get_user_id_requests_failed",
		Help:      "The amount of unsuccessful get user id requests on the api",
	},
)

var totalFollowRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "api",
		Name:      "follow_requests",
		Help:      "The amount of follow requests on the api",
	},
)
var unsuccessfulFollowRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "api",
		Name:      "follow_requests_failed",
		Help:      "The amount of unsuccessful follow requests on the api",
	},
)

var totalUnfollowRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "api",
		Name:      "unfollow_requests",
		Help:      "The amount of unfollow requests on the api",
	},
)

var unsuccessfulUnfollowRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "api",
		Name:      "unfollow_requests_failed",
		Help:      "The amount of unsuccessful unfollow requests on the api",
	},
)

var totalTweetMessageRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "api",
		Name:      "post_message_requests",
		Help:      "The amount of post message requests on the api",
	},
)

var unsuccessfulTweetMessageRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "api",
		Name:      "post_message_requests_failed",
		Help:      "The amount of failed post message requests on the api",
	},
)

var notFound = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "api",
		Name:      "not_found",
		Help:      "The amount of not found requests on the api - 404 error",
	},
)

var badRequest = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "api",
		Name:      "bad_request",
		Help:      "The amount of bad request requests on the api - 400 error",
	},
)

var internalServerError = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "api",
		Name:      "internal_server_error",
		Help:      "The amount of internal server error requests on the api - 500 error",
	},
)
