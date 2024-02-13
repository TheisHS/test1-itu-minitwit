package main

import (
	"database/sql"
    "net/http"
    "github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
	"log"
	"github.com/gorilla/sessions"
	"os"
)

var (
	db *sql.DB
	err error
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
)


func main() {
	//os.Remove("./minitwit.db")

    r := mux.NewRouter()

    r.HandleFunc("/", indexHandler).Methods("GET")
	r.HandleFunc("/timeline", timelineHandler).Methods("GET")
	r.HandleFunc("/public_timeline", publicTimelineHandler).Methods("GET")
	r.HandleFunc("/{username}", userTimelineHandler).Methods("GET")
	r.HandleFunc("/{username}/follow", followUserHandler).Methods("GET")
	r.HandleFunc("/{username}/unfollow", unfollowUserHandler).Methods("GET")
	r.HandleFunc("/add_message", addMessageHandler).Methods("POST")
	r.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	r.HandleFunc("/register", registerHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", logoutHandler).Methods("GET")

	fmt.Println("Server is running on port 5000")
	r.Use(beforeRequest)
    http.ListenAndServe(":5000", r)
}

func connectDB() (*sql.DB, error) {
    db, err := sql.Open("sqlite3", "./minitwit.db")
    if err != nil {
        return nil, err
    }
    return db, nil
}

func beforeRequest(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Logic to be executed before passing the request to the main handler
        db, err = connectDB()
		error_handler(err)
		defer db.Close()
        // Pass the request to the next handler in the chain
        next.ServeHTTP(w, r)
    })
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to minitwit!")
	rows, err := db.Query("select username from user where user_id = 1")
	error_handler(err)
	defer rows.Close()
	for rows.Next() {
		fmt.Println("User found")
	}
}

func init_db() {

}

func error_handler(err error) {
	if err != nil {
       log.Fatal(err)
    }
}

func timelineHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func publicTimelineHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func userTimelineHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func followUserHandler(w http.ResponseWriter, r *http.Request) {
   http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func addMessageHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}