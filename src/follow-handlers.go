package main

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)


func followUserHandler(w http.ResponseWriter, r *http.Request) {
	//Adds the current user as follower of the given user.
	handleFollow(w, r, true)
}


func unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	//Removes the current user as follower of the given user."
	handleFollow(w, r, false)
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