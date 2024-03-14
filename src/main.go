package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
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
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	UserID int
	Username string
	Email string
	pwHash string
}

type Message struct {
	messageID int
	authorID int
	Text string
	PubDate int
	flagged int
}

type Follower struct {
	whoID int
	whomID int
}

type TimelinePageData struct {
	User *User
	ProfileUser *User
	IsPublic bool
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
	timelineTmpl *template.Template
	loginTmpl *template.Template
	registerTmpl *template.Template
	db *sql.DB
	err error
	store = sessions.NewCookieStore([]byte("bb9cfb7ab2a6e36d683b0b209f96bb33"))
	perPage = 30
	env string
)

var totalRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "test1",
		Name: "http_requests_total",
		Help: "Number of get requests.",
	},
)

var databaseAccesses = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "test1",
		Name: "database_accesses_total",
		Help: "Amount of database accesses or operations",
	},
)

var totalErrors = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "test1",
		Name: "errors_total",
		Help: "Amount of errors",
	},
)

func main() {
	//os.Remove("./data/minitwit.db")

	store.Options = &sessions.Options{
		// Domain:   "localhost",
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
    //Secure:   true,
	}

	timelineTmpl = template.Must(template.Must(template.ParseFiles("templates/layout.html")).ParseFiles("templates/timeline.html"))
	loginTmpl = template.Must(template.Must(template.ParseFiles("templates/layout.html")).ParseFiles("templates/login.html"))
	registerTmpl = template.Must(template.Must(template.ParseFiles("templates/layout.html")).ParseFiles("templates/register.html"))

	flag.StringVar(&env, "env", "dev", "Environment to run the server in")
	flag.Parse()
	if env == "test" {
		_, err = os.Stat("./data/minitwit.db")
		if err != nil {
			initDB();
		}
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(totalRequests, databaseAccesses, totalErrors)
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
	r.Use(beforeRequest)
  http.ListenAndServe(":5000", r)
}

func connectDB() (*sql.DB, error) {
	if env == "test" {
		db, err := sql.Open("sqlite3", "./data/minitwit.db")
		if err != nil {
			return nil, err
		}
		return db, nil
	}
	if env == "dev" {
		var connStr = "postgres://postgres:mkw68nka@172.28.144.1/minitwit?sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
				return nil, err
		}
		return db, nil
	}
	if env == "prod" {
		var connStr = "postgres://postgres:mkw68nka@172.28.144.1/minitwit?sslmode=disable"
		db, err := sql.Open("postgres", connStr)
		if err != nil {
				return nil, err
		}
		return db, nil
	}
	panic("Unknown environment")
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

func initDB() {
	log.Println("Initialising the database...")

	os.Create("./data/minitwit.db")
	db, err := sql.Open("sqlite3", "./data/minitwit.db")
	if err != nil {
		log.Println(err)
	}
	
	schema, err := os.ReadFile("./schema.sql")
	if err != nil {
		log.Println(err) 
	}
	
	_, err = db.Exec(string(schema))
	if err != nil {
		log.Println(err) 
	}
	db.Close()
}

func getUserID(username string) (int, error) {
    var userID int
		databaseAccesses.Inc()
    err = db.QueryRow("SELECT user_id FROM \"user\" WHERE username = $1", username).Scan(&userID)
    if err != nil {
			totalErrors.Inc()
			return 0, err
    }
    return userID, nil
}

func getUser(userID int) (*User) {
	var user User
	databaseAccesses.Inc()
	err = db.QueryRow("SELECT user_id, username, email, pw_hash FROM \"user\" WHERE user_id = $1", userID).Scan(&user.UserID, &user.Username, &user.Email, &user.pwHash)
	if err == sql.ErrNoRows {
		totalErrors.Inc()
		return nil
	}
	
	return &user
}

func timelineHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/public_timeline", http.StatusFound)
		return
	}
	user := getUser(userID)
	var usermessages []UserMessage
	
	databaseAccesses.Inc()
	rows, err := db.Query("SELECT message.*, \"user\".* FROM message, \"user\" WHERE message.flagged = 0 AND message.author_id = \"user\".user_id AND (\"user\".user_id = $1 OR \"user\".user_id in (SELECT whom_id FROM follower WHERE who_id = $2)) ORDER BY message.pub_date DESC LIMIT $3", userID, userID, perPage)
	if err != nil {
		totalErrors.Inc()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var message Message
		var author User
		err = rows.Scan(&message.messageID, &message.authorID, &message.Text, &message.PubDate, &message.flagged, &author.UserID, &author.Username, &author.Email, &author.pwHash)
		if err != nil {
			totalErrors.Inc()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		um := UserMessage { User: author, Message: message }
		usermessages = append(usermessages, um)
	}

	data := TimelinePageData{
		User: user,
		Usermessages: usermessages,
		Flashes: session.Flashes(),
	}

	session.Save(r, w)
	
	timelineTmpl.Execute(w, data)

	//rnd.HTML(w, http.StatusOK, "timeline", nil)
}

func publicTimelineHandler(w http.ResponseWriter, r *http.Request) {
	var user *User
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int) 
	if ok {
		user = getUser(userID)
	}
	var usermessages []UserMessage

	databaseAccesses.Inc()
	rows, err := db.Query("SELECT message.*, \"user\".* FROM message, \"user\" WHERE message.flagged = 0 AND message.author_id = \"user\".user_id ORDER BY message.pub_date DESC LIMIT $1", perPage)
	if err != nil {
		totalErrors.Inc()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var message Message
		var user User
		err = rows.Scan(&message.messageID, &message.authorID, &message.Text, &message.PubDate, &message.flagged, &user.UserID, &user.Username, &user.Email, &user.pwHash)
		if err != nil {
			totalErrors.Inc()
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
		IsPublic: true,
		Flashes: session.Flashes(),
	}

	session.Save(r, w)
	
	timelineTmpl.Execute(w, data)
}

func userTimelineHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
	username := vars["username"]
	var user User

	databaseAccesses.Inc()
	row := db.QueryRow("SELECT * FROM \"user\" WHERE username = $1", username)
	err := row.Scan(&user.UserID, &user.Username, &user.Email, &user.pwHash)
	if err != nil {
		totalErrors.Inc()
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	followed := false

	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)

	if ok {
		databaseAccesses.Inc()
		row := db.QueryRow("SELECT 1 FROM follower WHERE who_id = $1 AND whom_id = $2", userID, user.UserID)
		err := row.Scan(&followed)
		if err == nil {
			followed = true
		} 
	}

	var usermessages []UserMessage
	var loggedInUser *User = getUser(userID)

	databaseAccesses.Inc()
	rows, err := db.Query("select message.*, \"user\".* from message, \"user\" where \"user\".user_id = message.author_id and \"user\".user_id = $1 order by message.pub_date desc limit $2", user.UserID, 30)
	if err != nil {
		totalErrors.Inc()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var message Message
		var user User
		err = rows.Scan(&message.messageID, &message.authorID, &message.Text, &message.PubDate, &message.flagged, &user.UserID, &user.Username, &user.Email, &user.pwHash)
		if err != nil {
			totalErrors.Inc()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		um := UserMessage { User: user, Message: message }
		usermessages = append(usermessages, um)
	}

	// for rendering the HTML template
	data := TimelinePageData{
		User: loggedInUser,
		ProfileUser: &user,
		Followed: followed,
		Usermessages: usermessages,
		Flashes: session.Flashes(),
	}

	session.Save(r, w)
	
	timelineTmpl.Execute(w, data)
}

func followUserHandler(w http.ResponseWriter, r *http.Request) {
	//Adds the current user as follower of the given user.
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		totalErrors.Inc()
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]

	whomID, err := getUserID(username)
	if err != nil {
		totalErrors.Inc()
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	databaseAccesses.Inc()
	_, err = db.Exec("INSERT INTO follower (who_id, whom_id) VALUES ($1, $2)", userID, whomID)
	if err != nil {
		fmt.Println(err)
		totalErrors.Inc()
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	session.AddFlash("You are now following \"" + username + "\"")
	session.Save(r, w)

	http.Redirect(w, r, fmt.Sprintf("/%s", username), http.StatusSeeOther)
}

func unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
    //Removes the current user as follower of the given user."
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		totalErrors.Inc()
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]

	whomID, err := getUserID(username)
	if err != nil {
		totalErrors.Inc()
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	databaseAccesses.Inc()
	_, err = db.Exec("DELETE FROM follower WHERE who_id=$1 and whom_id=$2", userID, whomID)
	if err != nil {
		totalErrors.Inc()
		fmt.Println(err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	session.AddFlash("You are no longer following \"" + username + "\"")
	session.Save(r, w)

	//TODO: flash('You are no longer following "%s"' % username) -> Implement flash in Go
	http.Redirect(w, r, fmt.Sprintf("/%s", username), http.StatusSeeOther)
}

func addMessageHandler(w http.ResponseWriter, r *http.Request) {
    //Registers a new message for the user.
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		totalErrors.Inc()
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		totalErrors.Inc()
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	text := r.Form.Get("text")
	if text == "" {
			http.Error(w, "Bad Request: Empty message", http.StatusBadRequest)
			return
	}

	databaseAccesses.Inc()
	_, err = db.Exec("INSERT INTO message (author_id, text, pub_date, flagged) VALUES ($1, $2, $3, 0)", userID, text, time.Now().Unix())
	totalRequests.Inc()
	if err != nil {
		totalErrors.Inc()
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	session.AddFlash("Your message was recorded")
	session.Save(r, w)

	http.Redirect(w, r, "/timeline", http.StatusSeeOther)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	if _, ok := session.Values["user_id"]; ok {
		http.Redirect(w, r, "/timeline", http.StatusSeeOther)
		return
	}
	var registerError string
	if r.Method == http.MethodPost {
		var user User
		username := r.FormValue("username")
		userID, _ := getUserID(username)
		email := r.FormValue("email")
		password := r.FormValue("password")
		password2 := r.FormValue("password2")
		if len(username) == 0 {
			registerError = "You have to enter a username"
		} else if len(email) == 0 || !strings.Contains(email, "@") {
			registerError = "You have to enter a valid email address"
		} else if len(password) == 0 {
			registerError = "You have to enter a password"
		} else if password != password2 {
			registerError = "The two passwords do not match"
		} else if userID != 0 {
			registerError = "The username is already taken"
		} else {
			pwHash := GeneratePasswordHash(password)
			databaseAccesses.Inc()
			err = db.QueryRow("insert into \"user\" (username, email, pw_hash) values ($1, $2, $3)", username, email, pwHash).Scan(&user.UserID, &user.Username, &user.pwHash)
			session.AddFlash("You were successfully registered and can login now")
			session.Save(r, w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}
	data := LoginPageData{
		Error: registerError,
		Flashes: session.Flashes(),
	}

	session.Save(r, w)
	
	registerTmpl.Execute(w, data)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		totalErrors.Inc()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := session.Values["user_id"]; ok {
		http.Redirect(w, r, "/timeline", http.StatusSeeOther)
		return
	}
	
	var loginError string
	
	if r.Method == http.MethodPost {
		var user User
		username := r.FormValue("username")
		password := r.FormValue("password")

		databaseAccesses.Inc()
		err = db.QueryRow("SELECT user_id, username, pw_hash FROM \"user\" WHERE username = $1", username).Scan(&user.UserID, &user.Username, &user.pwHash)
		if err == sql.ErrNoRows {
			loginError = "Invalid username"
		} else if !CheckPasswordHash(password, user.pwHash) { 
			loginError = "Invalid password"
		} else {
			session.Values["user_id"] = user.UserID
			session.AddFlash("You were logged in")
			saveError := session.Save(r, w)
			if saveError != nil {
				totalErrors.Inc()
				http.Error(w, saveError.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/timeline", http.StatusSeeOther)
			return
		}
	}
	
	data := LoginPageData{
		Error: loginError,
		Flashes: session.Flashes(),
	}
	
	session.Save(r, w)

	loginTmpl.Execute(w, data)
}


func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	delete(session.Values, "user_id")
	session.AddFlash("You were logged out")
	session.Save(r,w)
	http.Redirect(w,r, "/public_timeline", http.StatusSeeOther)
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (u User) Gravatar(size int) (string) {
	// Return the gravatar image for the user's email address.
	return fmt.Sprintf("http://www.gravatar.com/avatar/%v?d=identicon&s=%v", GetMD5Hash(u.Email), size)
}

func (m Message) FormatDatetime() (string) {
	// Format a timestamp for display.
	t := time.Unix(int64(m.PubDate), 0)
	return t.Local().Format(time.ANSIC)
}