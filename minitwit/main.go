package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

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
	r.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	r.HandleFunc("/register", registerHandler).Methods("GET", "POST")
	r.HandleFunc("/add_message", addMessageHandler).Methods("POST")
	r.HandleFunc("/logout", logoutHandler).Methods("GET")
	r.HandleFunc("/public", publicTimelineHandler).Methods("GET")
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
