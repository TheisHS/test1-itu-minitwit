package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/joho/godotenv"
)

var (
	err error
	env string
)

func main() {
	flag.StringVar(&env, "env", "dev", "Environment to run the server in")
	flag.Parse()
	if env == "test" {
		_, err = os.Stat(dbPath)
		if err != nil {
			initDB()
		}
	}
	if env == "dev" {
		if err := godotenv.Load(); err != nil {
			fmt.Println("No .env file found")
		}
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(totalRequests, databaseAccesses, totalErrors, totalGetUserIDRequests, unsuccessfulGetUserIDRequests, totalFollowRequests, unsuccessfulFollowRequests, totalUnfollowRequests, unsuccessfulUnfollowRequests, totalTweetMessageRequests, unsuccessfulTweetMessageRequests, notFound, badRequest, internalServerError)
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	r := mux.NewRouter()
	r.Path("/metrics").Handler(promHandler)
	r.HandleFunc("/latest", getLatestHandler).Methods("GET")
	r.HandleFunc("/register", registerHandler).Methods("POST")
	r.HandleFunc("/login", loginHandler).Methods("POST")
	r.HandleFunc("/msgs", msgsHandler).Methods("GET")
	r.HandleFunc("/msgs/{username}", messagesPerUserHandler).Methods("GET", "POST")
	r.HandleFunc("/msgsMy/{username}", msgsPersonalHandler).Methods("GET")
	r.HandleFunc("/fllws/{username}", fllwsUserHandler).Methods("GET", "POST")
	r.HandleFunc("/fllws/{whoUsername}/{whomUsername}", doesFllwUserHandler).Methods("GET")
	r.HandleFunc("/userID/{username}", getUserIDHandler).Methods("GET")
	r.HandleFunc("/getUser", getUserHandler).Methods("GET")

	fmt.Println("Server is running on port 5001")
	r.Use(beforeRequest)
	http.ListenAndServe(":5001", r)
}

func beforeRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Logic to be executed before passing the request to the main handler
		db, err = connectDB()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()
		// Pass the request to the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

func notReqFromSimulator(w http.ResponseWriter, r *http.Request) bool {
	fromSimulator := r.Header.Get("Authorization")
	if false && fromSimulator != "Basic c2ltdWxhdG9yOnN1cGVyX3NhZmUh" {
		errMsg := "You are not authorized to use this resource!!"
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, fmt.Sprintf(`{"status": 403, "error_msg": "%v"}`, errMsg))
		return true
	}
	return false
}
