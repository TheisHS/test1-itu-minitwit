package main

import (
	"bytes"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)


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
