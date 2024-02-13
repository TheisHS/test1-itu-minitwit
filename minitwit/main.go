package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
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
	r.HandleFunc("/add_message", addMessageHandler).Methods("POST")
	r.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	r.HandleFunc("/register", registerHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", logoutHandler).Methods("GET")
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

func getUserID(username string) (int, error) {
    var userID int
    err = db.QueryRow("SELECT user_id FROM user WHERE username = ?", username).Scan(&userID)
    if err != nil {
        return 0, err
    }
    return userID, nil
}

func timelineHandler(w http.ResponseWriter, r *http.Request) {
    //http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func publicTimelineHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT message.*, user.* FROM message, user WHERE message.flagged = 0 AND message.author_id = user.user_id ORDER BY message.pub_date DESC LIMIT ?", 30)
    fmt.Println(rows)
    error_handler(err)
    defer rows.Close()
}

func userTimelineHandler(w http.ResponseWriter, r *http.Request) {
    //http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func followUserHandler(w http.ResponseWriter, r *http.Request) {
	//Adds the current user as follower of the given user.
	session, _ := store.Get(r, "session-name")
    userID, ok := session.Values["user_id"].(int)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    vars := mux.Vars(r)
    username := vars["username"]

    whomID, err := getUserID(username)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

	_, err = db.Exec("INSERT INTO follower WHERE (who_id, whom_id) VALUES (?, ?)", userID, whomID)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

	//TODO: flash('You are now following "%s"' % username) -> Implement flash in Go
	http.Redirect(w, r, fmt.Sprintf("/%s", username), http.StatusSeeOther)
}

func unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
    //Removes the current user as follower of the given user."
	session, _ := store.Get(r, "session-name")
    userID, ok := session.Values["user_id"].(int)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    vars := mux.Vars(r)
    username := vars["username"]

    whomID, err := getUserID(username)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

	_, err = db.Exec("DELETE FROM follower WHERE (who_id, whom_id) VALUES (?, ?)", userID, whomID)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

	//TODO: flash('You are no longer following "%s"' % username) -> Implement flash in Go
	http.Redirect(w, r, fmt.Sprintf("/%s", username), http.StatusSeeOther)
}

func addMessageHandler(w http.ResponseWriter, r *http.Request) {
    //Registers a new message for the user.
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
    if err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    text := r.Form.Get("text")
    if text == "" {
        http.Error(w, "Bad Request: Empty message", http.StatusBadRequest)
        return
    }

	_, err = db.Exec("INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)", userID, text, time.Now().Unix())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/timeline", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    //http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
    //http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	//http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}