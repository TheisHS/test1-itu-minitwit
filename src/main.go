package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"

	"crypto/md5"
	"encoding/hex"
	"encoding/json"

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
	if password == "" { return "" }
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

type Post struct {
	Content string
	PubDate int
	Username string
}

type TimelinePageData struct {
	User *User
	ProfileUser string
	IsPublic bool
	Followed bool
	Posts []Post
	Usermessages []UserMessage
	Flashes []interface{}
}

type LoginPageData struct {
	User *User
	Error string
	Flashes []interface{}
}

type ServerError struct {
	Status int
	ErrorMsg string
}

var (
	timelineTmpl *template.Template
	loginTmpl *template.Template
	registerTmpl *template.Template
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

var registerRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "test1",
		Name: "register_requests",
		Help: "The amount of successful register requests on the web app",
	},
)

var loginRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "test1",
		Name: "login_requests",
		Help: "The amount of successful login requests on the web app",
	},
)

var unsuccessfulLoginRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "test1",
		Name: "login_requests_failed",
		Help: "The amount of unsuccessful login requests on the web app",
	},
)

var tweetRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "test1",
		Name: "tweets_requests",
		Help: "The amount of tweet requests on the web app",
	},
)

var serverEndpoint = "http://minitwit_api:5001"

func main() {
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


func getLoggedInUser(r *http.Request) (*User) {
	var loggedInUser *User
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int) 
	if ok {
		loggedInUser, _ = getUser(userID)
	}
	return loggedInUser
}


func getUser(userID int) (*User, error) {
	requestURL := fmt.Sprintf("%s/user/%d", serverEndpoint, userID)
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request to %s: %s\n", requestURL, err)
		return nil, err
	}
	var user User
	json.NewDecoder(res.Body).Decode(&user)
	return &user, nil
}


func messageToPost(res *http.Response) ([]Post) {
	var ins []map[string]interface{}
	body, _ := io.ReadAll(res.Body)
	json.Unmarshal(body, &ins)

	var posts []Post
	for _, in := range ins {
 		userIn := in["user"].(string)
		pubDateIn := int(in["pub_date"].(float64))
		contentIn := in["content"].(string)
		post := Post{
			Content: contentIn,
			Username: userIn,
			PubDate: pubDateIn,
		}
		posts = append(posts, post)
	}
	return posts
}


func timelineHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	loggedInUser := getLoggedInUser(r)
	if loggedInUser == nil {
		http.Redirect(w, r, "/public_timeline", http.StatusFound)
		return
	}

	requestURL := fmt.Sprintf("%s/msgsMy/%s?no=%d", serverEndpoint, loggedInUser.Username, perPage)
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request to %s: %s\n", requestURL, err)
		totalErrors.Inc()
		http.Error(w, fmt.Sprintf("error making http request to %s: %s\n", requestURL, err), http.StatusNotFound)
	}
	posts := messageToPost(res)

	data := TimelinePageData{
		User: loggedInUser,
		Posts: posts,
		Flashes: session.Flashes(),
	}

	session.Save(r, w)
	
	timelineTmpl.Execute(w, data)
}


func publicTimelineHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	loggedInUser := getLoggedInUser(r)

	requestURL := fmt.Sprintf("%s/msgs?no=%d", serverEndpoint, perPage)
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request to %s: %s\n", requestURL, err)
		totalErrors.Inc()
		http.Error(w, fmt.Sprintf("error making http request to %s: %s\n", requestURL, err), http.StatusNotFound)
	}
	posts := messageToPost(res)

	data := TimelinePageData{
		User: loggedInUser,
		Posts: posts,
		IsPublic: true,
		Flashes: session.Flashes(),
	}

	session.Save(r, w)
	
	timelineTmpl.Execute(w, data)
}


func userTimelineHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
  vars := mux.Vars(r)
	profileUser := vars["username"]

	followed := false

	loggedInUser := getLoggedInUser(r)
	if loggedInUser != nil {
		requestURLa := fmt.Sprintf("%s/fllws/%s/%s", serverEndpoint, loggedInUser.Username, profileUser)
		fllwsRes, err := http.Get(requestURLa)
		if err != nil {
			fmt.Printf("error making http request to %s: %s\n", requestURLa, err)
			totalErrors.Inc()
			http.Error(w, fmt.Sprintf("error making http request to %s: %s\n", requestURLa, err), http.StatusNotFound)	
		}
		fllwsBody, _ := io.ReadAll(fllwsRes.Body)
		followed, err = strconv.ParseBool(string(fllwsBody))
		if err != nil {
			fmt.Printf("could not convert %s to boolean, returning false instead: %s.\n", string(fllwsBody), err)
			followed = false
		}
	}

	requestURLb := fmt.Sprintf("%s/msgs/%s?no=%d", serverEndpoint, profileUser, perPage)
	res, err := http.Get(requestURLb)
	if err != nil {
		fmt.Printf("error making http request to %s: %s\n", requestURLb, err)
		totalErrors.Inc()
		http.Error(w, fmt.Sprintf("error making http request to %s: %s\n", requestURLb, err), http.StatusNotFound)
	}
	posts := messageToPost(res)
	
	// for rendering the HTML template
	data := TimelinePageData{
		User: loggedInUser,
		ProfileUser: profileUser,
		Followed: followed,
		Posts: posts,
		Flashes: session.Flashes(),
	}

	session.Save(r, w)
	
	timelineTmpl.Execute(w, data)
}

func handleFollow(w http.ResponseWriter, r *http.Request, isFollow bool) {
	//Adds the current user as follower of the given user.
	session, _ := store.Get(r, "session")
	whoUser := getLoggedInUser(r)
	if whoUser == nil {
		totalErrors.Inc()
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	whomUsername := vars["username"]

	var key string
	if (isFollow) { 
		key = "follow"
	} else {
		key = "unfollow"
	}
	body := []byte(fmt.Sprintf(`{"%s": "%s"}`, key, whomUsername))

	requestURL := fmt.Sprintf("%s/fllws/%s", serverEndpoint, whoUser.Username)
	_, err := http.Post(requestURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
		totalErrors.Inc()
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	var flashText string
	if isFollow {
		flashText = fmt.Sprintf(`You are now following "%s"`, whomUsername)
	} else {
	  flashText = fmt.Sprintf(`You are no longer following "%s"`, whomUsername)
	}

	session.AddFlash(flashText)
	session.Save(r, w)

	http.Redirect(w, r, fmt.Sprintf("/%s", whomUsername), http.StatusSeeOther)
}

func followUserHandler(w http.ResponseWriter, r *http.Request) {
	//Adds the current user as follower of the given user.
	handleFollow(w, r, true)
}


func unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	//Removes the current user as follower of the given user."
	handleFollow(w, r, false)
}


func addMessageHandler(w http.ResponseWriter, r *http.Request) {
  //Registers a new message for the user.
	session, _ := store.Get(r, "session")
	loggedInUser := getLoggedInUser(r)
	if loggedInUser == nil {
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
		fmt.Println("Tried posting with empty body")
		return
	}

	body := []byte(fmt.Sprintf(`{"Content": "%s"}`, text))
	requestURL := fmt.Sprintf("%s/msgs/%s", serverEndpoint, loggedInUser.Username)
	res, err := http.Post(requestURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
		totalErrors.Inc()
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if res.StatusCode == http.StatusNoContent {
		tweetRequests.Inc()
		session.AddFlash("Your message was recorded")
		session.Save(r, w)
	
		http.Redirect(w, r, "/timeline", http.StatusSeeOther)
	}
}


func registerHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	loggedInUser := getLoggedInUser(r)
	if loggedInUser != nil {
		http.Redirect(w, r, "/timeline", http.StatusSeeOther)
		return
	}

	var registerError string
	isPostRequest := r.Method == http.MethodPost

	defer func() {
		if isPostRequest && registerError == "" { return }
		data := LoginPageData{
			Error: registerError,
			Flashes: session.Flashes(),
		}
		
		session.Save(r, w)
		registerTmpl.Execute(w, data)
	}()
	
	if isPostRequest {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		password2 := r.FormValue("password2")

		if password != password2 {
			registerError = "The two passwords do not match"
			return
		} 

		body := []byte(fmt.Sprintf(`{
			"username": "%s",
			"email": "%s",
			"pwdHash": "%s"
		}`, username, email, GeneratePasswordHash(password)))

		requestURL := fmt.Sprintf("%s/register", serverEndpoint)
		res, err := http.Post(requestURL, "application/json" ,bytes.NewBuffer(body))
		if err != nil {
			fmt.Println(err)
			totalErrors.Inc()
			registerError = "Server error"
			return
		}

		var resErr ServerError
		json.NewDecoder(res.Body).Decode(&resErr)

		if resErr.ErrorMsg != "" {
			registerError = resErr.ErrorMsg
			return
		}

		session.AddFlash("You were successfully registered and can login now")
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}


func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	loggedInUser := getLoggedInUser(r)
	if loggedInUser != nil {
		http.Redirect(w, r, "/timeline", http.StatusSeeOther)
		return
	}
	
	var loginError string
	isPostRequest := r.Method == http.MethodPost

	defer func() {
		if isPostRequest && loginError == "" { return }
		data := LoginPageData{
			Error: loginError,
			Flashes: session.Flashes(),
		}
		
		session.Save(r, w)
		loginTmpl.Execute(w, data)
	}()
	
	if isPostRequest {
		username := r.FormValue("username")
		password := r.FormValue("password")

		body := []byte(fmt.Sprintf(`{
			"username": "%s"
		}`, username))

		requestURL := fmt.Sprintf("%s/login", serverEndpoint)
		res, err := http.Post(requestURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Println(err)
			loginError = "Server error"
			return
		}

		type Salt struct {
			ErrorMsg string
			Salt string
		}
		var saltResponse Salt
		json.NewDecoder(res.Body).Decode(&saltResponse)
		if saltResponse.ErrorMsg != "" {
			loginError = saltResponse.ErrorMsg
			return
		}

		body = []byte(fmt.Sprintf(`{
			"username": "%s",
			"pwdHash": "%s"
		}`, username, hashString(saltResponse.Salt, password)))

		res, err = http.Post(requestURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			totalErrors.Inc()
			fmt.Println(err)
			loginError = "Server error"
			return
		}

		type LoginResponse struct {
			ErrorMsg string
			UserID int
		}
		var loginResponse LoginResponse
		json.NewDecoder(res.Body).Decode(&loginResponse)
		if loginResponse.ErrorMsg != "" {
			loginError = loginResponse.ErrorMsg
			return
		}

		if loginResponse.UserID == 0 {
			fmt.Printf("%v", loginResponse)
			loginError = "Unexpected error"
			return
		}

		session.Values["user_id"] = loginResponse.UserID
		session.AddFlash("You were logged in")
		saveError := session.Save(r, w)
		if saveError != nil {
			totalErrors.Inc()
			fmt.Println(saveError.Error())
			loginError = "Unexpected error"
			return
		}

		loginError = ""
		http.Redirect(w, r, "/timeline", http.StatusSeeOther)
	}
}


func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	delete(session.Values, "user_id")
	session.AddFlash("You were logged out")
	session.Save(r,w)
	http.Redirect(w,r, "/public_timeline", http.StatusSeeOther)
}


func (p Post) Gravatar(size int) (string) {
	// Return the gravatar image for the user's username.
	hash := md5.Sum([]byte(p.Username))
	encoded := hex.EncodeToString(hash[:])
	return fmt.Sprintf("http://www.gravatar.com/avatar/%v?d=identicon&s=%v", encoded, size)
}


func (p Post) FormatDatetime() (string) {
	// Format a timestamp for display.
	t := time.Unix(int64(p.PubDate), 0)
	return t.Local().Format(time.ANSIC)
}