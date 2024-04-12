package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	timelineTmpl *template.Template
	loginTmpl *template.Template
	registerTmpl *template.Template
	store = sessions.NewCookieStore([]byte("bb9cfb7ab2a6e36d683b0b209f96bb33"))
	perPage = 30
	env string
	serverEndpoint string
)

func main() {
	store.Options = &sessions.Options{
		// Domain:   "localhost",
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
    //Secure:   true,
	}

	flag.StringVar(&env, "env", "dev", "Environment to run the server in")
	flag.Parse()
	if env == "dev" {
		if err := godotenv.Load(); err != nil {
			fmt.Println("No .env file found")
		}
	}
	if env == "test" {
		serverEndpoint = "http://minitwit_api:5001"
	} else {
		ip, _ := os.LookupEnv("API_IP")
		serverEndpoint = "http://" + ip + ":4001"
	}

	timelineTmpl = template.Must(template.Must(template.ParseFiles("templates/layout.html")).ParseFiles("templates/timeline.html"))
	loginTmpl = template.Must(template.Must(template.ParseFiles("templates/layout.html")).ParseFiles("templates/login.html"))
	registerTmpl = template.Must(template.Must(template.ParseFiles("templates/layout.html")).ParseFiles("templates/register.html"))

	flag.StringVar(&env, "env", "dev", "Environment to run the server in")
	flag.Parse()

	reg := prometheus.NewRegistry()
	reg.MustRegister(totalRequests, databaseAccesses, totalErrors, registerRequests, tweetRequests, loginRequests, unsuccessfulLoginRequests)
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	r := mux.NewRouter()
	r.Path("/metrics").Handler(promHandler)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.HandleFunc("/", timelineHandler).Methods("GET")
	r.HandleFunc("/timeline", timelineHandler).Methods("GET")
	r.HandleFunc("/public_timeline", publicTimelineHandler).Methods("GET")
	r.HandleFunc("/add_message", addMessageHandler).Methods("POST")
	r.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	r.HandleFunc("/register", registerHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", logoutHandler).Methods("GET")
	r.HandleFunc("/{username}", userTimelineHandler).Methods("GET")
	r.HandleFunc("/{username}/follow", followUserHandler).Methods("GET")
	r.HandleFunc("/{username}/unfollow", unfollowUserHandler).Methods("GET")


	fmt.Println("Server is running on port 5000")
  http.ListenAndServe(":5000", r)
}


