package main

type UserMessage struct {
	User    User
	Message Message
}

type User struct {
	UserID   int
	Username string
	Email    string
	pwHash   string
}

type Message struct {
	messageID int
	authorID  int
	Text      string
	PubDate   int
	flagged   int
}

type Follower struct {
	whoID  int
	whomID int
}

type Error struct {
	Status   int
	ErrorMsg string
}

type M map[string]interface{}