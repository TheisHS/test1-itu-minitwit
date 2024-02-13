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

type User struct {
	user_id int
	username string
	email string
	pw_hash string
}

type Message struct {
	message_id int
	author_id int
	text string
	pub_date int
	flagged int
}

type Follower struct {
	who_id int
	whom_id int
}

var (
	db *sql.DB
	err error
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	PER_PAGE = 30
)

func main() {
	//os.Remove("./minitwit.db")

    r := mux.NewRouter()

    r.HandleFunc("/", timelineHandler).Methods("GET")
	r.HandleFunc("/timeline", timelineHandler).Methods("GET")
	r.HandleFunc("/logout", logoutHandler).Methods("GET")
	r.HandleFunc("/public_timeline", publicTimelineHandler).Methods("GET")
	r.HandleFunc("/add_message", addMessageHandler).Methods("POST")
	r.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	r.HandleFunc("/register", registerHandler).Methods("GET", "POST")
	r.HandleFunc("/{username}", userTimelineHandler).Methods("GET")
	r.HandleFunc("/{username}/follow", followUserHandler).Methods("GET")
	r.HandleFunc("/{username}/unfollow", unfollowUserHandler).Methods("GET")

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

func init_db() {
	
}

func error_handler(err error) {
	if err != nil {
		log.Fatal(err)
    }
}

func timelineHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "We got a visitor FROM " + r.URL.Path)
	/* if store.user == nil {
		http.Redirect(w, r, "/public_timeline", 302)
	} */
	var messages []Message
	var users []User
	
	rows, err := db.Query("SELECT message.*, user.* FROM message, user WHERE message.flagged = 0 AND message.author_id = user.user_id AND (user.user_id = ? OR user.user_id in (SELECT whom_id FROM follower WHERE who_id = ?)) ORDER BY message.pub_date DESC LIMIT ?", 1, 1, PER_PAGE)
	error_handler(err)
	defer rows.Close()
	for rows.Next() {
		var message Message
		var user User
		err = rows.Scan(&message.message_id, &message.author_id, &message.text, &message.pub_date, &message.flagged, &user.user_id, &user.username, &user.email, &user.pw_hash)
		error_handler(err)
		messages = append(messages, message)
		users = append(users, user)
	}
	fmt.Println(messages)
	//rnd.HTML(w, http.StatusOK, "timeline", nil)
}

func publicTimelineHandler(w http.ResponseWriter, r *http.Request) {
	var messages []Message
	var users []User
    rows, err := db.Query("SELECT message.*, user.* FROM message, user WHERE message.flagged = 0 AND message.author_id = user.user_id ORDER BY message.pub_date DESC LIMIT ?", PER_PAGE)
	error_handler(err)
	defer rows.Close()

	for rows.Next() {
		var message Message
		var user User
		err = rows.Scan(&message.message_id, &message.author_id, &message.text, &message.pub_date, &message.flagged, &user.user_id, &user.username, &user.email, &user.pw_hash)
		error_handler(err)
		messages = append(messages, message)
		users = append(users, user)
	}
	fmt.Println(messages)

	//rnd.HTML(w, http.StatusOK, "timeline", nil)
}

func userTimelineHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
	username := vars["username"]
	var users []User

	row := db.QueryRow("SELECT * FROM user WHERE username = ?", username)
	followed := false

	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)

	if ok {
		row := db.QueryRow("SELECT 1 FROM follower WHERE who_id = ? AND whom_id = ?", userID, 1)
		err := row.Scan(&followed)
		error_handler(err)
	}

	var messages []Message
	var users []User

	rows, err := db.Query("select message.*, user.* from message, user where user.user_id = message.author_id and user.user_id = ? order by message.pub_date desc limit ?", userID, PER_PAGE)
	//rnd.HTML(w, http.StatusOK, "timeline", nil)

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