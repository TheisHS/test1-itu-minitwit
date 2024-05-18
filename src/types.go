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

type Post struct {
	Content  string
	PubDate  int
	Username string
}

type TimelinePageData struct {
	User         *User
	ProfileUser  string
	IsPublic     bool
	Followed     bool
	Posts        []Post
	Usermessages []UserMessage
	Flashes      []interface{}
	Endpoint     string
}

type LoginPageData struct {
	User     *User
	Error    string
	Flashes  []interface{}
	Endpoint string
}

type ServerError struct {
	Status   int
	ErrorMsg string
	Endpoint string
}
