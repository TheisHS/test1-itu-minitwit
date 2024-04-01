package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"encoding/json"

	"github.com/gorilla/mux"
)


func getMessages(query string, args ...any) ([]M, error) {
  databaseAccesses.Inc()
  rows, err := db.Query(query, args...)
  if err != nil {
    devLog(fmt.Sprintf("Error in db.Query(%s)", query))
    return nil, err
  }

  defer rows.Close()
  
  var filteredMessages []M

  for rows.Next() {
    var message Message
    var author User
    err = rows.Scan(&message.Text, &message.PubDate, &author.Username)
    if err != nil {
      devLog("Error in rows.Scan in getMessageList()")
      return nil, err
    }
    filteredMessage := M{"content": message.Text, "pub_date": message.PubDate, "user": author.Username}
    filteredMessages = append(filteredMessages, filteredMessage)
  }

  return filteredMessages, nil
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
    messages, err := getMessages(`
        SELECT message.text, message.pub_date, "user".username 
        FROM message, "user" 
        WHERE message.flagged = 0 AND message.author_id = "user".user_id 
        ORDER BY message.pub_date DESC LIMIT $1
      `, noMsgs)

    if err != nil {
      totalErrors.Inc()
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    data, _ := json.Marshal(messages)
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

    messages, err := getMessages(`
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
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    data, _ := json.Marshal(messages)
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

    messages, err := getMessages(`
        SELECT message.text, message.pub_date, "user".username 
        FROM message, "user"
        WHERE message.flagged = 0 
          AND "user".user_id = message.author_id 
          AND "user".user_id = $1 
        ORDER BY message.pub_date DESC LIMIT $2
      `, userID, noMsgs)

    if err != nil {
      totalErrors.Inc()
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    data, _ := json.Marshal(messages)
    io.WriteString(w, string(data))
  } else if r.Method == http.MethodPost {
    totalRequests.Inc()
    type MessageData struct {
      Content string
    }
    var data MessageData
    json.NewDecoder(r.Body).Decode(&data)
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

