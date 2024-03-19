package main

import (
	"database/sql"
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
	_ "github.com/mattn/go-sqlite3"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
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

type Error struct{
	Status int
	ErrorMsg string
}

type M map[string]interface{}

var (
	db *sql.DB
	err error
	logger *zap.Logger
)

func init() {
	stdout := zapcore.AddSync(os.Stdout)

	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/go.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     7, // days
	})

	level := zap.NewAtomicLevelAt(zap.InfoLevel)

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
		zapcore.NewCore(fileEncoder, file, level),
	)

	logger = zap.New(core)
	defer logger.Sync()
}

func main() {
	_, err = os.Stat("./data/minitwit.db")
	if err != nil {
    initDB();
	}

	r := mux.NewRouter()
	r.HandleFunc("/latest", getLatestHandler).Methods("GET")
	r.HandleFunc("/register", registerHandler).Methods("POST")
	r.HandleFunc("/msgs", msgsHandler).Methods("GET")
	r.HandleFunc("/msgs/{username}", messagesPerUserHandler).Methods("GET", "POST")
	r.HandleFunc("/fllws/{username}", fllwsUserHandler).Methods("GET", "POST")

	fmt.Println("Server is running on port 5001")
	r.Use(beforeRequest)
  http.ListenAndServe(":5001", r)
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

func connectDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./data/minitwit.db")
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

func getUserID(username string) (int, error) {
    var userID int
    err = db.QueryRow("SELECT user_id FROM user WHERE username = ?", username).Scan(&userID)
    if err != nil {
        return 0, err
    }
    return userID, nil
}

func notReqFromSimulator(w http.ResponseWriter, r *http.Request) (bool) {
	fromSimulator := r.Header.Get("Authorization")
	if false && fromSimulator != "Basic c2ltdWxhdG9yOnN1cGVyX3NhZmUh" {
		errMsg := "You are not authorized to use this resource!"
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, fmt.Sprintf("{\"status\": 403, \"error_msg\": \"%v\"}", errMsg))
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
			http.Error(w, "Failed to open file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		_, err = fmt.Fprintf(file, "%d", parsedCommandID)
		if err != nil {
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
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}

	content, err := strconv.Atoi(string(file)) 
	if err != nil {
		io.WriteString(w, "{\"latest\":-1}")
	} else {
		io.WriteString(w, fmt.Sprintf("{\"latest\":%d}", content))
	}
}

func msgsHandler(w http.ResponseWriter, r *http.Request) {
	updateLatest(w, r)
	reqErr := notReqFromSimulator(w, r)
	if reqErr { return }

	noMsgs := r.URL.Query().Get("no")
	if r.Method == http.MethodGet {
		if noMsgs == "" {
			io.WriteString(w, "[]")
			return
		}
		rows, err := db.Query("SELECT message.*, user.* FROM message, user WHERE message.flagged = 0 AND message.author_id = user.user_id ORDER BY message.pub_date DESC LIMIT ?", noMsgs)
		if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
			}
		defer rows.Close()
	
		var filteredMessages []M
		for rows.Next() {
			var message Message
			var author User
			err = rows.Scan(&message.messageID, &message.authorID, &message.Text, &message.PubDate, &message.flagged, &author.UserID, &author.Username, &author.Email, &author.pwHash)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			filteredMessage := M{"content": message.Text, "pub_date": message.PubDate, "user": author.Username}
			filteredMessages = append(filteredMessages, filteredMessage)
		}	
		
		logger.Info("Messages retrieved", zap.Any("messages", filteredMessages))
		data, _ := json.Marshal(filteredMessages)
		io.WriteString(w, string(data))
	}
}

func messagesPerUserHandler(w http.ResponseWriter, r *http.Request) {
	updateLatest(w, r)
	reqErr := notReqFromSimulator(w, r)
	if reqErr { return }
	
	noMsgs := r.URL.Query().Get("no")
	vars := mux.Vars(r)
	username := vars["username"]
	userID, err := getUserID(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if r.Method == http.MethodGet {
		rows, err := db.Query("SELECT message.*, user.* FROM message, user WHERE message.flagged = 0 AND user.user_id = message.author_id AND user.user_id = ? ORDER BY message.pub_date DESC LIMIT ?", userID, noMsgs)
		if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
		}
		defer rows.Close()
	
		var filteredMessages []M
		for rows.Next() {
			var message Message
			var author User
			err = rows.Scan(&message.messageID, &message.authorID, &message.Text, &message.PubDate, &message.flagged, &author.UserID, &author.Username, &author.Email, &author.pwHash)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			filteredMessage := M{"content": message.Text, "pub_date": message.PubDate, "user": author.Username}
			filteredMessages = append(filteredMessages, filteredMessage)
		}	
		logger.Info("Retrieved messages for user", zap.String("username", username), zap.Any("messages", filteredMessages))
		data, _ := json.Marshal(filteredMessages)
		io.WriteString(w, string(data))
	} else if r.Method == http.MethodPost {
		type RegisterData struct {
			Content string
		}
		var data RegisterData
		json.NewDecoder(r.Body).Decode(&data)
		_, err := db.Exec("INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)", userID, data.Content, time.Now().Unix())
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
      return
		}
		logger.Info("Message added", zap.String("username", username), zap.String("content", data.Content))
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
			http.Error(w, "User not found", http.StatusNotFound)
			return
	}

	type FollowsData struct {
		Follow string
		Unfollow string
	}
	
	if r.Method == http.MethodPost {
		var data FollowsData
		json.NewDecoder(r.Body).Decode(&data)
		if data.Follow != "" {
			whomID, err := getUserID(data.Follow)
			if err != nil {
				// TODO: This has to be another error, likely 500 ???
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			_, err = db.Exec("INSERT INTO follower (who_id, whom_id) VALUES (?, ?)", whoID, whomID)
			if err != nil {
					http.Error(w, "Database error", http.StatusInternalServerError)
					return
			}
			logger.Info("User followed", zap.String("username", username), zap.String("followed", data.Follow))
			w.WriteHeader(http.StatusNoContent)
			io.WriteString(w, "")
			return
		}
		if data.Unfollow != "" {
			whomID, err := getUserID(data.Unfollow)
			if err != nil {
				// TODO: This has to be another error, likely 500 ???
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			_, err = db.Exec("DELETE FROM follower WHERE who_id=? and WHOM_ID=?", whoID, whomID)
			if err != nil {
					http.Error(w, "Database error", http.StatusInternalServerError)
					return
			}
			logger.Info("User unfollowed", zap.String("username", username), zap.String("unfollowed", data.Unfollow))
			w.WriteHeader(http.StatusNoContent)
			io.WriteString(w, "")
			return
		}
	}

	if r.Method == http.MethodGet {
		noFollowers, _ := strconv.Atoi(r.URL.Query().Get("no"))
		rows, err := db.Query("SELECT user.username FROM user INNER JOIN follower ON follower.whom_id=user.user_id WHERE follower.who_id=? LIMIT ?", whoID, noFollowers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var followers []string
		for rows.Next() {
			var username string
			err = rows.Scan(&username)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			followers = append(followers, username)
		}
		logger.Info("Retrieved followers for user", zap.String("username", username), zap.Any("followers", followers))
		followerJSON, _ := json.Marshal(followers)
		io.WriteString(w, fmt.Sprintf("{\"follows\": %v}", string(followerJSON)))
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	updateLatest(w, r)
	reqErr := notReqFromSimulator(w, r)
	if reqErr { return }

	var registerError string
	if r.Method == http.MethodPost {
		type RegisterData struct {
			Username string
			Email string
			Pwd string
		}
		var data RegisterData
		json.NewDecoder(r.Body).Decode(&data)
		userID, _ := getUserID(data.Username)
		if len(data.Username) == 0 {
			registerError = "You have to enter a username"
		} else if len(data.Email) == 0 || !strings.Contains(data.Email, "@") {
			registerError = "You have to enter a valid email address"
		} else if len(data.Pwd) == 0 {
			registerError = "You have to enter a password"
		} else if userID != 0 {
			registerError = "The username is already taken"
		} else {
			pwHash := GeneratePasswordHash(data.Pwd)
			_, err := db.Exec("insert into user (username, email, pw_hash) values (?, ?, ?)", data.Username, data.Email, pwHash)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			logger.Info("User registered", zap.String("username", data.Username))
			w.WriteHeader(http.StatusNoContent)
			io.WriteString(w, "")
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		errorData, _ := json.Marshal(Error {
			Status: 400,
			ErrorMsg: registerError,
		})
		io.WriteString(w, string(errorData))
	}
}