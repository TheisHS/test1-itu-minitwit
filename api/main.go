package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"

	"encoding/hex"
	"encoding/json"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/joho/godotenv"
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

type Error struct {
  Status int
  ErrorMsg string
}

type M map[string]interface{}

var (
  db *sql.DB
  err error
  env string
)

var totalRequests = prometheus.NewCounter(
  prometheus.CounterOpts{
    Namespace: "api",
    Name: "http_requests_total",
    Help: "Number of get requests.",
  },
)

var databaseAccesses = prometheus.NewCounter(
  prometheus.CounterOpts{
    Namespace: "api",
    Name: "database_accesses_total",
    Help: "Amount of database accesses or operations",
  },
)

var totalErrors = prometheus.NewCounter(
  prometheus.CounterOpts{
    Namespace: "api",
    Name: "errors_total",
    Help: "Amount of errors",
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

var dbPath = "./data/minitwit.db"


func main() {
  flag.StringVar(&env, "env", "dev", "Environment to run the server in")
  flag.Parse()
  if env == "test" {
    _, err = os.Stat(dbPath)
    if err != nil {
      initDB();
    }
  }
  if env == "dev" {
    if err := godotenv.Load(); err != nil {
      log.Print("No .env file found")
    }
  }

  reg := prometheus.NewRegistry()
  reg.MustRegister(totalRequests, databaseAccesses, totalErrors)
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
  r.HandleFunc("/user/{userID}", getUserHandler).Methods("GET")

  fmt.Println("Server is running on port 5001")
  r.Use(beforeRequest)
  http.ListenAndServe(":5001", r)
}


func devLog(str string) {
  if env == "dev" {
    fmt.Println(str)
  }
}


func initDB() {
  log.Println("Initialising the database...")

  os.Create(dbPath)
  db, err := sql.Open("sqlite3", dbPath)
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


func connectDB() (*sql.DB, error) {
  if env == "test" {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
      return nil, err
    }
    return db, nil
  }
  if env == "dev" {
    connStr, ok := os.LookupEnv("DATABASE_URL")
    if ok {
      db, err := sql.Open("postgres", connStr)
      if err != nil {
          return nil, err
      }
      return db, nil
    }  
    panic("DATABASE_URL not set!")
  }
  if env == "prod" {
    connStr, ok := os.LookupEnv("DATABASE_URL")
    if ok {
      db, err := sql.Open("postgres", connStr)
      if err != nil {
          return nil, err
      }
      return db, nil
    }  
    panic("DATABASE_URL not set!")
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


func getUserID(username string) (int, error) {
  var userID int
  databaseAccesses.Inc()
  err = db.QueryRow(`SELECT user_id FROM "user" WHERE username = $1`, username).Scan(&userID)
  if err != nil {
    totalErrors.Inc()
    return 0, err
  }
  return userID, nil
}


func getUserIDHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  username := vars["username"]
  userID, err := getUserID(username)
  if err != nil {
    http.Error(w, err.Error(), http.StatusNotFound)
    return
  }
  io.WriteString(w, strconv.Itoa(userID))
}


func getUser(userID int) (*User) {
  var user User
  databaseAccesses.Inc()
  err = db.QueryRow(`
          SELECT user_id, username, email 
          FROM "user" 
          WHERE user_id = $1
        `, userID).Scan(&user.UserID, &user.Username, &user.Email)
  if err == sql.ErrNoRows {
    totalErrors.Inc()
    return nil
  }
  return &user
}


func getUserHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  userID := vars["userID"]
  var user *User
  id, err := strconv.Atoi(userID)
  if err != nil {
    http.Error(w, err.Error(), http.StatusNotFound)
  }
  user = getUser(id)
  if user == nil {
    io.WriteString(w, "")
  } else {
    data, _ := json.Marshal(*user)
    io.WriteString(w, string(data))
  }
}


func notReqFromSimulator(w http.ResponseWriter, r *http.Request) (bool) {
  fromSimulator := r.Header.Get("Authorization")
  if false && fromSimulator != "Basic c2ltdWxhdG9yOnN1cGVyX3NhZmUh" {
    errMsg := "You are not authorized to use this resource!"
    w.WriteHeader(http.StatusUnauthorized)
    io.WriteString(w, fmt.Sprintf(`{"status": 403, "error_msg": "%v"}`, errMsg))
    return true
  }
  return false
}


func updateLatest(w http.ResponseWriter, r *http.Request) {
  parsedCommandID, err := strconv.Atoi(r.URL.Query().Get("latest"))
  if err != nil {
    return
  }

  if parsedCommandID != -1 {
    file, err := os.Create("./latest_processed_sim_action_id.txt")
    if err != nil {
      totalErrors.Inc()
      http.Error(w, "Failed to open file", http.StatusInternalServerError)
      return
    }
    defer file.Close()

    _, err = fmt.Fprintf(file, "%d", parsedCommandID)
    if err != nil {
      totalErrors.Inc()
      http.Error(w, "Failed to write to file", http.StatusInternalServerError)
      return
    }
  }    
}


func getLatestHandler(w http.ResponseWriter, r *http.Request) {
  _, err = os.Stat("./latest_processed_sim_action_id.txt")
  if err != nil {
    os.Create("./latest_processed_sim_action_id.txt")
  }

  file, err := os.ReadFile("./latest_processed_sim_action_id.txt")
  if err != nil {
    totalErrors.Inc()
    http.Error(w, "Failed to open file", http.StatusInternalServerError)
    return
  }

  content, err := strconv.Atoi(string(file)) 
  if err != nil {
    io.WriteString(w, `{"latest":-1}`)
  } else {
    io.WriteString(w, fmt.Sprintf(`{"latest":%d}`, content))
  }
}


func msgsHandler(w http.ResponseWriter, r *http.Request) {
  updateLatest(w, r)
  reqErr := notReqFromSimulator(w, r)
  if reqErr { 
    return 
  }

  noMsgs := r.URL.Query().Get("no")
  if r.Method == http.MethodGet {
    totalRequests.Inc()
    if noMsgs == "" {
      io.WriteString(w, "[]")
      return
    }
    databaseAccesses.Inc()
    rows, err := db.Query(`
                   SELECT message.text, message.pub_date, "user".username 
                   FROM message, "user" 
                   WHERE message.flagged = 0 AND message.author_id = "user".user_id 
                   ORDER BY message.pub_date DESC LIMIT $1
                 `, noMsgs)
    if err != nil {
        totalErrors.Inc()
        fmt.Println("error when db.Query in msgsHandler")
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
    defer rows.Close()
  
    var filteredMessages []M
    for rows.Next() {
      var message Message
      var author User
      err = rows.Scan(&message.Text, &message.PubDate, &author.Username)
      if err != nil {
        totalErrors.Inc()
        fmt.Println("error when rows.Scan() in msgsHandler")
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
      filteredMessage := M{"content": message.Text, "pub_date": message.PubDate, "user": author.Username}
      filteredMessages = append(filteredMessages, filteredMessage)
    }  

    data, _ := json.Marshal(filteredMessages)
    io.WriteString(w, string(data))
  }
}


func msgsPersonalHandler(w http.ResponseWriter, r *http.Request) {
  noMsgs := r.URL.Query().Get("no")
  vars := mux.Vars(r)
  username := vars["username"]
  userID, err := getUserID(username)

  if err != nil {
    totalErrors.Inc()
    http.Error(w, err.Error(), http.StatusNotFound)
    return
  }

  if r.Method == http.MethodGet {
    totalRequests.Inc()
    if noMsgs == "" {
      io.WriteString(w, "[]")
      return
    }
    databaseAccesses.Inc()
    rows, err := db.Query(`
                   SELECT message.text, message.pub_date, "user".username 
                   FROM message, "user" 
                   WHERE message.flagged = 0 
                     AND message.author_id = "user".user_id 
                     AND ("user".user_id = $1 OR "user".user_id in (
                       SELECT whom_id FROM follower WHERE who_id = $2
                      )) 
                   ORDER BY message.pub_date DESC LIMIT $3
                 `, userID, userID, noMsgs)
    if err != nil {
        totalErrors.Inc()
        fmt.Println("error when db.Query in msgsHandler")
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
    defer rows.Close()
  
    var filteredMessages []M
    for rows.Next() {
      var message Message
      var author User
      err = rows.Scan(&message.Text, &message.PubDate, &author.Username)
      if err != nil {
        totalErrors.Inc()
        fmt.Println("error when rows.Scan() in msgsHandler")
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
      filteredMessage := M{"content": message.Text, "pub_date": message.PubDate, "user": author.Username}
      filteredMessages = append(filteredMessages, filteredMessage)
    }  

    data, _ := json.Marshal(filteredMessages)
    io.WriteString(w, string(data))
  }
}


func messagesPerUserHandler(w http.ResponseWriter, r *http.Request) {
  updateLatest(w, r)
  reqErr := notReqFromSimulator(w, r)
  if reqErr { return }
  
  vars := mux.Vars(r)
  username := vars["username"]
  userID, err := getUserID(username)
  
  if err != nil {
    totalErrors.Inc()
    http.Error(w, err.Error(), http.StatusNotFound)
    return
  }
  
  if r.Method == http.MethodGet {
    noMsgs := r.URL.Query().Get("no")
    totalRequests.Inc()
    databaseAccesses.Inc()
    rows, err := db.Query(`
                   SELECT message.text, message.pub_date, "user".username 
                   FROM message, "user"
                   WHERE message.flagged = 0 AND "user".user_id = message.author_id AND "user".user_id = $1 
                   ORDER BY message.pub_date DESC LIMIT $2
                 `, userID, noMsgs)
    if err != nil {
      totalErrors.Inc()
      fmt.Println("Error in Query")
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    defer rows.Close()
  
    var filteredMessages []M
    for rows.Next() {
      var message Message
      var author User
      err = rows.Scan(&message.Text, &message.PubDate, &author.Username)
      if err != nil {
        fmt.Println("Error in scan")
        totalErrors.Inc()
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
      filteredMessage := M{"content": message.Text, "pub_date": message.PubDate, "user": author.Username}
      filteredMessages = append(filteredMessages, filteredMessage)
    }  

    data, _ := json.Marshal(filteredMessages)
    io.WriteString(w, string(data))
  } else if r.Method == http.MethodPost {
    totalRequests.Inc()
    type MessageData struct {
      Content string
    }
    var data MessageData
    json.NewDecoder(r.Body).Decode(&data)
    fmt.Printf("%v\n", data.Content)
    databaseAccesses.Inc()
    _, err := db.Exec(`
                INSERT INTO message (author_id, text, pub_date, flagged) 
                VALUES ($1, $2, $3, 0)
              `, userID, data.Content, time.Now().Unix())
    if err != nil {
      totalErrors.Inc()
      devLog("Error in db.Exec in messagesPerUserHandler()")
      http.Error(w, "Database error", http.StatusInternalServerError)
      return
    }
    w.WriteHeader(http.StatusNoContent)
    io.WriteString(w, "")
    return
  }
}


func fllwsUserHandler(w http.ResponseWriter, r *http.Request) {
  updateLatest(w, r)
  reqErr := notReqFromSimulator(w, r)
  if reqErr { return }
  
  vars := mux.Vars(r)
  username := vars["username"]
  whoID, err := getUserID(username)
  if err != nil {
    totalErrors.Inc()
    http.Error(w, "User not found", http.StatusNotFound)
    return
  }

  type FollowsData struct {
    Follow string
    Unfollow string
  }
  
  if r.Method == http.MethodPost {
    totalRequests.Inc()
    var data FollowsData
    json.NewDecoder(r.Body).Decode(&data)
    if data.Follow != "" {
      whomID, err := getUserID(data.Follow)
      if err != nil {
        totalErrors.Inc()
        http.Error(w, "User not found", http.StatusNotFound)
        return
      }
      databaseAccesses.Inc()
      _, err = db.Exec("INSERT INTO follower (who_id, whom_id) VALUES ($1, $2)", whoID, whomID)
      if err != nil {
        totalErrors.Inc()
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
      }
      w.WriteHeader(http.StatusNoContent)
      io.WriteString(w, "")
      return
    }
    if data.Unfollow != "" {
      whomID, err := getUserID(data.Unfollow)
      if err != nil {
        totalErrors.Inc()
        http.Error(w, "User not found", http.StatusNotFound)
        return
      }

      databaseAccesses.Inc()
      _, err = db.Exec("DELETE FROM follower WHERE who_id=$1 and WHOM_ID=$2", whoID, whomID)
      if err != nil {
        totalErrors.Inc()
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
      }
      w.WriteHeader(http.StatusNoContent)
      io.WriteString(w, "")
      return
    }
  }

  if r.Method == http.MethodGet {
    totalRequests.Inc()
    noFollowers, _ := strconv.Atoi(r.URL.Query().Get("no"))
    databaseAccesses.Inc()
    rows, err := db.Query(`
                  SELECT "user".username 
                  FROM "user" 
                  INNER JOIN follower ON follower.whom_id="user".user_id 
                  WHERE follower.who_id=$1 LIMIT $2
                `, whoID, noFollowers)
    if err != nil {
      totalErrors.Inc()
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    defer rows.Close()
    var followers []string
    for rows.Next() {
      var username string
      err = rows.Scan(&username)
      if err != nil {
        totalErrors.Inc()
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
      followers = append(followers, username)
    }
    followerJSON, _ := json.Marshal(followers)
    io.WriteString(w, fmt.Sprintf(`{"follows": %v}`, string(followerJSON)))
  }
}


func doesFllwUserHandler(w http.ResponseWriter, r *http.Request) {
  updateLatest(w, r)
  reqErr := notReqFromSimulator(w, r)
  if reqErr { return }
  
  vars := mux.Vars(r)
  whoUsername := vars["whoUsername"]
  whomUsername := vars["whomUsername"]
  whoID, err := getUserID(whoUsername)
  if err != nil {
    totalErrors.Inc()
    devLog(fmt.Sprintf("Could not find user %s", &whoUsername))
    http.Error(w, "who user not found", http.StatusNotFound)
    return
  }
  whomID, err := getUserID(whomUsername)
  if err != nil {
    totalErrors.Inc()
    devLog(fmt.Sprintf("Could not find user %s", &whomUsername))
    http.Error(w, "whom user not found", http.StatusNotFound)
    return
  }
  
  if r.Method == http.MethodGet {
    totalRequests.Inc()
    databaseAccesses.Inc()
    var x Follower
    err := db.QueryRow(`
             SELECT * FROM follower 
             WHERE follower.who_id=$1 AND follower.whom_id=$2
           `, whoID, whomID).Scan(&x)
    if err == sql.ErrNoRows {
      io.WriteString(w, "false")
      return
    }
    io.WriteString(w, "true")
  }
}


func registerHandler(w http.ResponseWriter, r *http.Request) {
  updateLatest(w, r)
  reqErr := notReqFromSimulator(w, r)
  if reqErr { return }

  var registerError string
  if r.Method == http.MethodPost {
    totalRequests.Inc()
    type RegisterData struct {
      Username string
      Email string
      Pwd string
      PwdHash string
    }
    var data RegisterData
    json.NewDecoder(r.Body).Decode(&data)
    userID, _ := getUserID(data.Username)
    if len(data.Username) == 0 {
      devLog("No username")
      registerError = "You have to enter a username"
    } else if len(data.Email) == 0 || !strings.Contains(data.Email, "@") {
      devLog(fmt.Sprintf("Not valid email, %s", data.Email))
      registerError = "You have to enter a valid email address"
    } else if len(data.Pwd) == 0 && len(data.PwdHash) == 0 {
      devLog("No password or password hash")
      registerError = "You have to enter a password"
    } else if userID != 0 {
      devLog(fmt.Sprintf("Username is already taken: %s, userid: %d", data.Username, userID))
      registerError = "The username is already taken"
    } else {
      devLog("Valid register")
      var pwHash string
      if data.PwdHash == "" {
        devLog("PwdHash not set, generating from Pwd")
        pwHash = GeneratePasswordHash(data.Pwd)
      } else {
        devLog("PwdHash set, using this instead of Pwd")
        pwHash = data.PwdHash
      }
      databaseAccesses.Inc()
      _, err := db.Exec(`
                  insert into "user" (username, email, pw_hash) 
                  values ($1, $2, $3)
                `, data.Username, data.Email, pwHash)
      if err != nil {
        totalErrors.Inc()
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
      w.WriteHeader(http.StatusNoContent)
      io.WriteString(w, "")
      return
    }
    devLog("Registration failed")
    w.WriteHeader(http.StatusBadRequest)
    errorData, _ := json.Marshal(Error {
      Status: 400,
      ErrorMsg: registerError,
    })
    io.WriteString(w, string(errorData))
  }
}


func loginHandler(w http.ResponseWriter, r *http.Request) {
  var loginError string
  if r.Method == http.MethodPost {
    totalRequests.Inc()
    type LoginData struct {
      Username string
      PwdHash string
    }
    var data LoginData
    json.NewDecoder(r.Body).Decode(&data)

    var user User
    databaseAccesses.Inc()
    err = db.QueryRow(`
            SELECT user_id, username, pw_hash 
            FROM "user" 
            WHERE username = $1
          `, data.Username).Scan(&user.UserID, &user.Username, &user.pwHash)

    if err == sql.ErrNoRows {
      devLog("Username does not exist")
      unsuccessfulLoginRequests.Inc()
      loginError = "Invalid username"
    } else { 
      ps := strings.Split(user.pwHash, "$")
      if data.PwdHash == "" {
        io.WriteString(w, fmt.Sprintf(`{"salt": "%s"}`, ps[1]))
        return
      } else {
        if ps[2] != data.PwdHash { 
          devLog(fmt.Sprintf(`Password hashes do not match: %s != %s`, ps[2], data.PwdHash))
          unsuccessfulLoginRequests.Inc()
          loginError = "Invalid password"
        } else {
          loginRequests.Inc()
          io.WriteString(w, fmt.Sprintf(`{"userID": %d}`, user.UserID))
          return
        }
      }
    }
    devLog("Login failed")
    w.WriteHeader(http.StatusBadRequest)
    errorData, _ := json.Marshal(Error {
      Status: 400,
      ErrorMsg: loginError,
    })
    io.WriteString(w, string(errorData))
  }
}