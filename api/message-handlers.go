package main

import (
	"io"
	"net/http"
	"time"

	"encoding/json"

	"github.com/gorilla/mux"
)

func getMessages(query string, args ...any) ([]M, error) {
	promtailClient := getPromtailClient("getMessages")
	defer promtailClient.Shutdown()
	databaseAccesses.Inc()
	rows, err := db.Query(query, args...)
	if err != nil {
		promtailClient.Errorf("Error in db.Query(%s)", query)
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
	promtailClient := getPromtailClient("msgsHandler")
	defer promtailClient.Shutdown()
	updateLatest(w, r)
	reqErr := notReqFromSimulator(w, r)
	if reqErr {
		return
	}

	noMsgs := r.URL.Query().Get("no")
	if r.Method == http.MethodGet {
		promtailClient.Infof("Get request for %s messages from the public timeline", noMsgs)
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
			internalServerError.Inc()
			promtailClient.Errorf("Error while fetching %s messages from the public timeline", noMsgs)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, _ := json.Marshal(messages)
		io.WriteString(w, string(data))
	}
}

func msgsPersonalHandler(w http.ResponseWriter, r *http.Request) {
	promtailClient := getPromtailClient("msgsPersonalHandler")
	defer promtailClient.Shutdown()
	noMsgs := r.URL.Query().Get("no")
	vars := mux.Vars(r)
	username := vars["username"]
	userID, err := getUserID(username)
	if err != nil {
		totalErrors.Inc()
		notFound.Inc()
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
			internalServerError.Inc()
			promtailClient.Errorf("Error while fetching %s messages", noMsgs)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, _ := json.Marshal(messages)
		io.WriteString(w, string(data))
	}
}

func messagesPerUserHandler(w http.ResponseWriter, r *http.Request) {
	promtailClient := getPromtailClient("messagesPerUserHandler")
	defer promtailClient.Shutdown()
	updateLatest(w, r)
	reqErr := notReqFromSimulator(w, r)
	if reqErr {
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]
	userID, err := getUserID(username)
	if err != nil {
		totalErrors.Inc()
		notFound.Inc()
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if r.Method == http.MethodGet {
		noMsgs := r.URL.Query().Get("no")
		promtailClient.Infof("Get request for %s messages from %s", noMsgs, username)
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
			internalServerError.Inc()
			promtailClient.Errorf("Error while fetching %s messages from %s (userID: %d)", noMsgs, username, userID)
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
		totalTweetMessageRequests.Inc()
		_, err := db.Exec(`
        INSERT INTO message (author_id, text, pub_date, flagged) 
        VALUES ($1, $2, $3, 0)
      `, userID, data.Content, time.Now().Unix())
		if err != nil {
			totalErrors.Inc()
			unsuccessfulTweetMessageRequests.Inc()
			internalServerError.Inc()
			promtailClient.Errorf("Error while creating post by %s (userID: %d): %s", username, userID, data.Content)
			devLog("Error in db.Exec in messagesPerUserHandler()")
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		promtailClient.Infof("New post created by %s: %s", username, data.Content)
		w.WriteHeader(http.StatusNoContent)
		io.WriteString(w, "")
		return
	}
}
