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
	var user User

	row := db.QueryRow("SELECT * FROM user WHERE username = ?", username)
	err := row.Scan(&user.user_id, &user.username, &user.email, &user.pw_hash)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
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

func registerHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not yet implemented", http.StatusNotImplemented)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	if _, ok := session.Values["user"]; ok {
		http.Redirect(w, r, "/timeline", http.StatusSeeOther)
		return
	}
	var error string
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		var user struct {
			user_id int
			username string
			pw_hash string
		}
		err = db.QueryRow("SELECT user_id, username, pw_hash FROM user WHERE username = ?", username).Scan(&user.user_id, &user.username, &user.pw_hash)
		if err == sql.ErrNoRows {
			error = "Invalid username"
		} else if user.pw_hash != password {
			error = "Invalid password"
		} else if user.username == username && user.pw_hash == password{
			session.Values["user_id"] = user.user_id
			session.Save(r, w)
			http.Redirect(w, r, "/timeline", http.StatusSeeOther)
			fmt.Println(session.Values["user_id"])
			return
		}
		fmt.Println(error)
	}
}


func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Values["user"] = nil
	session.Save(r,w)
	fmt.Println(session.Values["user"])
	http.Redirect(w,r, "/public_timeline", http.StatusSeeOther)
}