package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"encoding/json"

	"github.com/gorilla/mux"
)


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
      _, err = db.Exec(`
          INSERT INTO follower (who_id, whom_id) 
          VALUES ($1, $2)
        `, whoID, whomID)
      if err != nil {
        totalErrors.Inc()
        // http.Error(w, "Database error", http.StatusInternalServerError)
        // return
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
      _, err = db.Exec(`
          DELETE FROM follower 
          WHERE who_id=$1 and WHOM_ID=$2
        `, whoID, whomID)
      if err != nil {
        totalErrors.Inc()
        // http.Error(w, "Database error", http.StatusInternalServerError)
        // return
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
    http.Error(w, fmt.Sprintf("User %s not found", whoUsername), http.StatusNotFound)
    return
  }
  whomID, err := getUserID(whomUsername)
  if err != nil {
    totalErrors.Inc()
    http.Error(w, fmt.Sprintf("User %s not found", whomUsername), http.StatusNotFound)
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
