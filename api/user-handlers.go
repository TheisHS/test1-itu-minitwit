package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"encoding/json"

	"github.com/gorilla/mux"
)

func getUserID(username string) (int, error) {
  var userID int
  databaseAccesses.Inc()
  err = db.QueryRow(`
      SELECT user_id 
      FROM "user" 
      WHERE username = $1
    `, username).Scan(&userID)
  if err != nil {
    devLog(fmt.Sprintf("Could not find user %s", username))
    return 0, err
  }
  return userID, nil
}


func getUserIDHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  username := vars["username"]
  userID, err := getUserID(username)
  if err != nil {
    totalErrors.Inc()
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
          INSERT INTO "user" (username, email, pw_hash) 
          VALUES ($1, $2, $3)
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
      } 
      if ps[2] == data.PwdHash { 
        loginRequests.Inc()
        io.WriteString(w, fmt.Sprintf(`{"userID": %d}`, user.UserID))
        return
      }
      devLog(fmt.Sprintf(`Password hashes do not match: %s != %s`, ps[2], data.PwdHash))
      unsuccessfulLoginRequests.Inc()
      loginError = "Invalid password"
      
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