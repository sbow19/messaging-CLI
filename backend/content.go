package main

/*
Sub set of clientData type in backend
*/
type Friend struct {
	Active   bool   `json:"active"`
	Message  string `json:"message"`
	Username string `json:"username"`
}

type FriendReqDetails struct {
	Username   string `json:"username"`
	RequestId  string `json:"request_id"`
	FromClient bool   `json:"from_client"`
}

// Username
type Messages map[string][]Message

type Message struct {
	Text     string `json:"text"`
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Date     string `json:"date"`
}

// All data
type UserContent struct {
	Friends        []Friend           `json:"friends"`
	FriendRequests []FriendReqDetails `json:"friend_requests"`
	Messages       Messages           `json:"messages"`
}
