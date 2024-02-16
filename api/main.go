package main

import (
	"database/sql"
	"fmt"
	"io"
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

type Error struct{
	Status int
	Error_msg string
}

var (
	db *sql.DB
	err error
)

func main() {
	os.Remove("./minitwit.db")
	initDB()

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


func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./minitwit.db")
	if err != nil {
			return nil, err
	}
	
	schema, err := os.ReadFile("../schema.sql")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return nil, err
	}
	return db, nil
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
	if fromSimulator != "Basic c2ltdWxhdG9yOnN1cGVyX3NhZmUh" {
		errMsg := "You are not authorized to use this resource!"
		w.WriteHeader(http.StatusUnauthorized)
		error_data, _ := json.Marshal(Error {
			Status: 403,
			Error_msg: errMsg,
		})
		io.WriteString(w, string(error_data))
		return true
	}
	return false
}

func updateLatest(w http.ResponseWriter, r *http.Request) {
	parsedCommandID, err := strconv.Atoi(r.URL.Query().Get("latest"))
		if err != nil {
			http.Error(w, "Invalid latest parameter", http.StatusBadRequest)
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
	file, err := os.ReadFile("./latest_processed_sim_action_id.txt")
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}

	type Latest struct {
		Latest int
	}

	content, err := strconv.Atoi(string(file)) 
	if err != nil {
		io.WriteString(w, fmt.Sprintf("{\"latest\":-1}"))
	} else {
		io.WriteString(w, fmt.Sprintf("{\"latest\":%d}", content))
	}
}

func msgsHandler(w http.ResponseWriter, r *http.Request) {
	updateLatest(w, r)
	req_err := notReqFromSimulator(w, r)
	if req_err { return }

	no_msgs := r.URL.Query().Get("no")
	if r.Method == http.MethodGet {
		rows, err := db.Query("SELECT message.*, user.* FROM message, user WHERE message.flagged = 0 AND message.author_id = user.user_id ORDER BY message.pub_date DESC LIMIT ?", no_msgs)
		if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
			}
		defer rows.Close()
		type M map[string]interface{}
	
		var filteredMessages []M
		for rows.Next() {
			var message Message
			var author User
			err = rows.Scan(&message.message_id, &message.author_id, &message.Text, &message.Pub_date, &message.flagged, &author.User_id, &author.Username, &author.Email, &author.pw_hash)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			filteredMessage := M{"content": message.Text, "pub_date": message.Pub_date, "user": author.Username}
			filteredMessages = append(filteredMessages, filteredMessage)
		}	

		data, _ := json.Marshal(filteredMessages)
		io.WriteString(w, string(data))
	}
}

func messagesPerUserHandler(w http.ResponseWriter, r *http.Request) {
	updateLatest(w, r)
	req_err := notReqFromSimulator(w, r)
	if req_err { return }
	
	no_msgs := r.URL.Query().Get("no")
	username := r.URL.Query().Get("username")
	userID, err := getUserID(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if r.Method == http.MethodGet {

		rows, err := db.Query("SELECT message.*, user.* FROM message, user WHERE message.flagged = 0 AND user.user_id = message.author_id AND user.user_id = ? ORDER BY message.pub_date DESC LIMIT ?", userID, no_msgs)
		if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
		}
		defer rows.Close()
		type M map[string]interface{}
	
		var filteredMessages []M
		for rows.Next() {
			var message Message
			var author User
			err = rows.Scan(&message.message_id, &message.author_id, &message.Text, &message.Pub_date, &message.flagged, &author.User_id, &author.Username, &author.Email, &author.pw_hash)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			filteredMessage := M{"content": message.Text, "pub_date": message.Pub_date, "user": author.Username}
			filteredMessages = append(filteredMessages, filteredMessage)
		}	

		data, _ := json.Marshal(filteredMessages)
		io.WriteString(w, string(data))
	} else if r.Method == http.MethodPost {
		type RegisterData struct {
			Content string
		}
		var data RegisterData
		json.NewDecoder(r.Body).Decode(&data)
		fmt.Println(data)
		_, err := db.Exec("INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)", userID, data.Content, time.Now().Unix())
		if err != nil {
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
	req_err := notReqFromSimulator(w, r)
	if req_err { return }
	
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
		fmt.Println(data)
		if data.Follow != "" {
			whomID, err := getUserID(data.Follow)
			if err != nil {
				// TODO: This has to be another error, likely 500 ???
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			_, err = db.Exec("INSERT INTO follower (who_id, whom_id) VALUES (?, ?)", whoID, whomID)
			if err != nil {
				fmt.Println(err)
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
				// TODO: This has to be another error, likely 500 ???
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			_, err = db.Exec("DELETE FROM follower WHERE who_id=? and WHOM_ID=?", whoID, whomID)
			if err != nil {
				fmt.Println(err)
					http.Error(w, "Database error", http.StatusInternalServerError)
					return
			}
			w.WriteHeader(http.StatusNoContent)
			io.WriteString(w, "")
			return
		}
	}

	if r.Method == http.MethodGet {
		no_followers, _ := strconv.Atoi(r.URL.Query().Get("no"))
		rows, err := db.Query("SELECT user.username FROM user INNER JOIN follower ON follower.whom_id=user.user_id WHERE follower.who_id=? LIMIT ?", whoID, no_followers)
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
		follower_json, _ := json.Marshal(followers)
		io.WriteString(w, string(follower_json))
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	updateLatest(w, r)
	req_err := notReqFromSimulator(w, r)
	if req_err { return }

	var register_error string
	if r.Method == http.MethodPost {
		type RegisterData struct {
			Username string
			Email string
			Pwd string
		}
		var data RegisterData
		json.NewDecoder(r.Body).Decode(&data)
		user_id, _ := getUserID(data.Username)
		if len(data.Username) == 0 {
			register_error = "You have to enter a username"
		} else if len(data.Email) == 0 || !strings.Contains(data.Email, "@") {
			register_error = "You have to enter a valid email address"
		} else if len(data.Pwd) == 0 {
			register_error = "You have to enter a password"
		} else if user_id != 0 {
			register_error = "The username is already taken"
		} else {
			pw_hash := GeneratePasswordHash(data.Pwd)
			db.QueryRow("insert into user (username, email, pw_hash) values (?, ?, ?)", data.Username, data.Email, pw_hash)
			w.WriteHeader(http.StatusNoContent)
			io.WriteString(w, "")
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		error_data, _ := json.Marshal(Error {
			Status: 400,
			Error_msg: register_error,
		})
		io.WriteString(w, string(error_data))
	}
}