package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"time"

	"encoding/hex"
	"encoding/json"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func getLoggedInUser(r *http.Request) (*User) {
	var loggedInUser *User
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(int) 
	if ok {
		loggedInUser, _ = getUser(userID)
	}
	return loggedInUser
}


func getUser(userID int) (*User, error) {
	requestURL := fmt.Sprintf("%s/user/%d", serverEndpoint, userID)
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request to %s: %s\n", requestURL, err)
		return nil, err
	}
	var user User
	json.NewDecoder(res.Body).Decode(&user)
	return &user, nil
}


func messageToPost(res *http.Response) ([]Post) {
	var ins []map[string]interface{}
	body, _ := io.ReadAll(res.Body)
	json.Unmarshal(body, &ins)

	var posts []Post
	for _, in := range ins {
 		userIn := in["user"].(string)
		pubDateIn := int(in["pub_date"].(float64))
		contentIn := in["content"].(string)
		post := Post{
			Content: contentIn,
			Username: userIn,
			PubDate: pubDateIn,
		}
		posts = append(posts, post)
	}
	return posts
}


func (p Post) Gravatar(size int) (string) {
	// Return the gravatar image for the user's username.
	hash := md5.Sum([]byte(p.Username))
	encoded := hex.EncodeToString(hash[:])
	return fmt.Sprintf("http://www.gravatar.com/avatar/%v?d=identicon&s=%v", encoded, size)
}


func (p Post) FormatDatetime() (string) {
	// Format a timestamp for display.
	t := time.Unix(int64(p.PubDate), 0)
	return t.Local().Format(time.ANSIC)
}