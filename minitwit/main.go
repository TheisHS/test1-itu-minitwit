package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"

	"crypto/md5"
	"encoding/hex"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

const (
	Method     = "pbkdf2:sha256"
	SaltLength = 8
	Iterations = 150000
)

const (
	satlChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	keyLength = 32
)

func GeneratePasswordHash(password string) string {
	salt := genSalt()
	hash := hashString(salt, password)
	return fmt.Sprintf("%s:%v$%s$%s", Method, Iterations, salt, hash)
}

func CheckPasswordHash(password string, hash string) bool {
	if strings.Count(hash, "$") < 2 {
		return false
	}
	ps := strings.Split(hash, "$")
	return ps[2] == hashString(ps[1], password)
}

func genSalt() string {
	var bytes = make([]byte, SaltLength)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = satlChars[v%byte(len(satlChars))]
	}
	return string(bytes)
}

func hashString(salt string, password string) string {
	hash := pbkdf2.Key([]byte(password), []byte(salt), Iterations, keyLength, sha256.New)
	return hex.EncodeToString(hash)
}

type UserMessage struct {
	User User
	Message Message
}

type User struct {
	User_id int
	Username string
	Email string
	pw_hash string
}

type Message struct {
	message_id int
	author_id int
	Text string
	Pub_date int
	flagged int
}

type Follower struct {
	who_id int
	whom_id int
}

type TimelinePageData struct {
	User *User
	Profile_user User
	Followed bool
	Usermessages []UserMessage
	Flashes []interface{}
}

type LoginPageData struct {
	User *User
	Error string
	Flashes []interface{}
}

var (
	timeline_tmpl *template.Template
	login_tmpl *template.Template
	register_tmpl *template.Template
	db *sql.DB
	err error
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))	
	PER_PAGE = 30
)

func main() {
	//os.Remove("./minitwit.db")

	store.Options = &sessions.Options{
		Domain:   "localhost",
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
        Secure:   true,
	}

	timeline_tmpl = template.Must(template.Must(template.ParseFiles("templates/layout.html")).ParseFiles("templates/timeline.html"))
	login_tmpl = template.Must(template.Must(template.ParseFiles("templates/layout.html")).ParseFiles("templates/login.html"))
	register_tmpl = template.Must(template.Must(template.ParseFiles("templates/layout.html")).ParseFiles("templates/register.html"))


	r := mux.NewRouter()
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
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()
        // Pass the request to the next handler in the chain
        next.ServeHTTP(w, r)
    }) 
}

func init_db() {
	
}

func getUserID(username string) (int, error) {
    var userID int
    err = db.QueryRow("SELECT user_id FROM user WHERE username = ?", username).Scan(&userID)
    if err != nil {
        return 0, err
    }
    return userID, nil
}

func getUser(user_id int) (*User) {
	var user User
	err = db.QueryRow("SELECT user_id, username, email, pw_hash FROM user WHERE user_id = ?", user_id).Scan(&user.User_id, &user.Username, &user.Email, &user.pw_hash)
	if err == sql.ErrNoRows {
		return nil
	} else {
		return &user
	}
}

func timelineHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	user_id, ok := session.Values["user_id"].(int) 
	fmt.Println(user_id)
	if !ok {
		http.Redirect(w, r, "/public_timeline", 302)
		return
	}
	user := getUser(user_id)
	var usermessages []UserMessage
	
	rows, err := db.Query("SELECT message.*, user.* FROM message, user WHERE message.flagged = 0 AND message.author_id = user.user_id AND (user.user_id = ? OR user.user_id in (SELECT whom_id FROM follower WHERE who_id = ?)) ORDER BY message.pub_date DESC LIMIT ?", user_id, user_id, PER_PAGE)
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	defer rows.Close()

	for rows.Next() {
		var message Message
		var author User
		err = rows.Scan(&message.message_id, &message.author_id, &message.Text, &message.Pub_date, &message.flagged, &author.User_id, &author.Username, &author.Email, &author.pw_hash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		um := UserMessage { User: author, Message: message }
		usermessages = append(usermessages, um)
	}

	data := TimelinePageData{
		User: user,
		Usermessages: usermessages,
	}
	
	timeline_tmpl.Execute(w, data)

	//rnd.HTML(w, http.StatusOK, "timeline", nil)
}

func publicTimelineHandler(w http.ResponseWriter, r *http.Request) {
	var user *User
	session, _ := store.Get(r, "session")
	user_id, ok := session.Values["user_id"].(int) 
	if ok {
		user = getUser(user_id)
	}
	var usermessages []UserMessage

  	rows, err := db.Query("SELECT message.*, user.* FROM message, user WHERE message.flagged = 0 AND message.author_id = user.user_id ORDER BY message.pub_date DESC LIMIT ?", PER_PAGE)
  	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var message Message
		var user User
		err = rows.Scan(&message.message_id, &message.author_id, &message.Text, &message.Pub_date, &message.flagged, &user.User_id, &user.Username, &user.Email, &user.pw_hash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		um := UserMessage { User: user, Message: message }
		usermessages = append(usermessages, um)
	}

	// for rendering the HTML template
	data := TimelinePageData{
		User: user,
		Usermessages: usermessages,
	}
	
	timeline_tmpl.Execute(w, data)

	//fmt.Println(usermessages)
}

func userTimelineHandler(w http.ResponseWriter, r *http.Request) {
  	vars := mux.Vars(r)
	username := vars["username"]
	var user User

	row := db.QueryRow("SELECT * FROM user WHERE username = ?", username)
	err := row.Scan(&user.User_id, &user.Username, &user.Email, &user.pw_hash)
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
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var messages []Message
	var users []User

	rows, err := db.Query("select message.*, user.* from message, user where user.user_id = message.author_id and user.user_id = ? order by message.pub_date desc limit ?", userID, PER_PAGE)
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	defer rows.Close()

	for rows.Next() {
		var message Message
		var user User
		err = rows.Scan(&message.message_id, &message.author_id, &message.Text, &message.Pub_date, &message.flagged, &user.User_id, &user.Username, &user.Email, &user.pw_hash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		messages = append(messages, message)
		users = append(users, user)
	}
	fmt.Println(messages)
	//rnd.HTML(w, http.StatusOK, "timeline", nil)
}

func followUserHandler(w http.ResponseWriter, r *http.Request) {
	//Adds the current user as follower of the given user.
	session, _ := store.Get(r, "session")
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
	session, _ := store.Get(r, "session")
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
	session, _ := store.Get(r, "session")
	if _, ok := session.Values["user_id"]; ok {
		http.Redirect(w, r, "/timeline", http.StatusSeeOther)
		return
	}
	var register_error string
	if r.Method == http.MethodPost {
		var user User
		username := r.FormValue("username")
		user_id, _ := getUserID(username)
		email := r.FormValue("email")
		password := r.FormValue("password")
		password2 := r.FormValue("password2")
		if len(username) == 0 {
			register_error = "You have to enter a username"
		} else if len(email) == 0 || !strings.Contains(email, "@") {
			register_error = "You have to enter a valid email address"
		} else if len(password) == 0 {
			register_error = "You have to enter a password"
		} else if password != password2 {
			register_error = "The two passwords do not match"
		} else if user_id != 0 {
			register_error = "The username is already taken"
		} else {
			pw_hash := GeneratePasswordHash(password)
			err = db.QueryRow("insert into user (username, email, pw_hash) values (?, ?, ?)", username, email, pw_hash).Scan(&user.User_id, &user.Username, &user.pw_hash)
			session.AddFlash("You were successfully registered and can login now")
			session.Save(r, w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}
	data := LoginPageData{
		Error: register_error,
	}
	
	register_tmpl.Execute(w, data)
	register_tmpl.Execute(os.Stdout, data)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	
	if _, ok := session.Values["user_id"]; ok {
		http.Redirect(w, r, "/timeline", http.StatusSeeOther)
		return
	}
	
	var login_error string
	
	if r.Method == http.MethodPost {
		var user User
		username := r.FormValue("username")
		password := r.FormValue("password")

		err = db.QueryRow("SELECT user_id, username, pw_hash FROM user WHERE username = ?", username).Scan(&user.User_id, &user.Username, &user.pw_hash)
		if err == sql.ErrNoRows {
			login_error = "Invalid username"
		} else if !CheckPasswordHash(password, user.pw_hash) { 
			login_error = "Invalid password"
		} else {
			session.Values["user_id"] = user.User_id
			save_error := session.Save(r, w)
			if save_error != nil {
				http.Error(w, save_error.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Println("User ID:", session.Values["user_id"])
			http.Redirect(w, r, "/timeline", http.StatusSeeOther)
            return
		}
	}
	
	data := LoginPageData{
		Error: login_error,
		Flashes: session.Flashes(),
	}
	
	session.Save(r, w)

	login_tmpl.Execute(w, data)
	login_tmpl.Execute(os.Stdout, data)
}


func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	delete(session.Values, "user_id")
	session.Save(r,w)
	http.Redirect(w,r, "/public_timeline", http.StatusSeeOther)
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func gravatar_url(email string, size int) (string) {
	// Return the gravatar image for the given email address.
	return fmt.Sprintf("http://www.gravatar.com/avatar/%v?d=identicon&s=%v", GetMD5Hash(email), size)
}

func (u User) Gravatar(size int) (string) {
	// Return the gravatar image for the user's email address.
	return fmt.Sprintf("http://www.gravatar.com/avatar/%v?d=identicon&s=%v", GetMD5Hash(u.Email), size)
}

func (m Message) Format_datetime() (string) {
	// Format a timestamp for display.
	t := time.Unix(int64(m.Pub_date), 0)
	return t.Local().Format(time.ANSIC)
}