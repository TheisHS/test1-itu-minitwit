package main

import (
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

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	UserID int     				`gorm:"primaryKey"`
  CreatedAt time.Time
  UpdatedAt time.Time
	Username string
	Email string
	pwHash string
}
type Follower struct {
	whoID int
	whomID int
}
type Message struct {
	messageID int  				`gorm:"primaryKey"`
  CreatedAt time.Time
  UpdatedAt time.Time
	authorID int
	Text string
	PubDate int64
	flagged int
}

type Error struct{
	Status int
	ErrorMsg string
}

type M map[string]interface{}

var (
	db *gorm.DB
	err error
)

func main() {
	_, err = os.Stat("./data/minitwit.db")
	if err != nil {
    initDB();
	} else {
		db, err = gorm.Open(sqlite.Open("./data/minitwit.db"), &gorm.Config{})
		db.Table("user").AutoMigrate(&User{})
		db.Table("follower").AutoMigrate(&Follower{})
		db.Table("message").AutoMigrate(&Message{})
	}

	r := mux.NewRouter()
	r.HandleFunc("/latest", getLatestHandler).Methods("GET")
	r.HandleFunc("/register", registerHandler).Methods("POST")
	r.HandleFunc("/msgs", msgsHandler).Methods("GET")
	r.HandleFunc("/msgs/{username}", messagesPerUserHandler).Methods("GET", "POST")
	r.HandleFunc("/fllws/{username}", fllwsUserHandler).Methods("GET", "POST")

	fmt.Println("Server is running on port 5001")
	// r.Use(beforeRequest)
  http.ListenAndServe(":5001", r)
}


func initDB() {
	log.Println("Initialising the database...")

	os.Create("./data/minitwit.db")
	db, err = gorm.Open(sqlite.Open("./data/minitwit.db"), &gorm.Config{})
	if err != nil {
		log.Println(err)
	}
	
	db.AutoMigrate(&User{}, &Follower{}, &Message{})
	// db.AutoMigrate(&Follower{})
	// db.AutoMigrate(&Message{})
}

// func connectDB() (*sql.DB, error) {
// 	db, err = gorm.Open(sqlite.Open("./data/minitwit.db"), &gorm.Config{})
// 	if err != nil {
// 			return nil, err
// 	}
// 	return db, nil
// }

// func beforeRequest(next http.Handler) http.Handler {
//   return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Logic to be executed before passing the request to the main handler
// 		db, err = connectDB()
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		defer db.Close()
// 		// Pass the request to the next handler in the chain
// 		next.ServeHTTP(w, r)
//   }) 
// }

func getUserID(username string) (int, error) {
	var user User
	result := db.Table("user").Where(&User{Username: username}).First(&user)
	if result.Error != nil {
			return 0, result.Error
	}
  return user.UserID, nil
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

	if r.Method == http.MethodGet {
		noMsgs := r.URL.Query().Get("no")
		if noMsgs == "" {
			io.WriteString(w, "[]")
			return
		}
		
		limit, err := strconv.Atoi(noMsgs)
		if err != nil {
			io.WriteString(w, "[]")
			return
		}

		rows, err := db.Limit(limit).Table("user").Select("message.text, message.pub_date, user.username").Joins("join message on message.author_id = user.user_id").Order("message.pub_date desc").Rows()
		
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
	
		var filteredMessages []M
		for rows.Next() {
			type MessageDTO struct {
				text string
				pubdate string
				username string
			}
			var dto MessageDTO


			err = rows.Scan(&dto.text, &dto.pubdate, &dto.username)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			filteredMessage := M{"content": dto.text, "pub_date": dto.pubdate, "user": dto.username}
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
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if r.Method == http.MethodGet {
		noMsgs := r.URL.Query().Get("no")
		if noMsgs == "" {
			io.WriteString(w, "[]")
			return
		}

		limit, err := strconv.Atoi(noMsgs)
		if err != nil {
			io.WriteString(w, "[]")
			return
		}

		rows, err := db.Limit(limit).Table("user").Select("message.text, message.pub_date, user.username").Joins("join message on message.author_id = user.user_id").Where("user.user_id = ? AND message.flagged = 0", userID).Order("message.pub_date desc").Rows()
		
		if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
		}
		defer rows.Close()
	
		var filteredMessages []M
		for rows.Next() {
			type MessageDTO struct {
				text string
				pubdate string
				username string
			}
			var dto MessageDTO


			err = rows.Scan(&dto.text, &dto.pubdate, &dto.username)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			filteredMessage := M{"content": dto.text, "pub_date": dto.pubdate, "user": dto.username}
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

		newMessage := Message { authorID: userID, Text: data.Content, PubDate: time.Now().Unix() }
		result := db.Create(&newMessage)
		if result.Error != nil {
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
			newFollower := Follower {whoID: whoID, whomID: whomID}
			result := db.Create(&newFollower)
			if result.Error != nil {
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
			result := db.Where("who_id = ? AND whom_id = ?", whoID, whomID).Delete(&Follower{})
			if result.Error != nil {
					http.Error(w, "Database error", http.StatusInternalServerError)
					return
			}
			w.WriteHeader(http.StatusNoContent)
			io.WriteString(w, "")
			return
		}
	}

	if r.Method == http.MethodGet {
		noFollowers := r.URL.Query().Get("no")
		if noFollowers == "" {
			io.WriteString(w, "[]")
			return
		}
		limit, err := strconv.Atoi(noFollowers)
		if err != nil {
			io.WriteString(w, "[]")
			return
		}
		rows, err := db.Limit(limit).Table("user").Select("user.username").Joins("inner join follower on follower.whom_id = user.user_id").Where("follower.who_id = ?", whoID).Rows()
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

			newUser := User { Username: data.Username, Email: data.Email, pwHash: pwHash }
			result := db.Create(&newUser)
			if result.Error != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
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