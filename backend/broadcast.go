package main

import (
	"encoding/json"
	"fmt"
)

type BackendMessageCode int

const (
	BroadcastFriendship BackendMessageCode = iota
	BroadcastFriendRequest
	BroadcastChat
	BroadcastLoggedIn
)

type BackendMessage struct {
	Code    BackendMessageCode
	Payload interface{}
}

func AppListener(s *Server) {

	for message := range s.broadcast {
		switch message.Code {
		case BroadcastLoggedIn:
			if userId, ok := message.Payload.(apiKey); ok {

				// Get API key and get all friends in db
				friendIds, err := dbConn.GetFriendsById(string(userId))
				if err != nil {
					fmt.Println(err)
					break
				}

				res := UserMap[userId]

				for _, id := range *friendIds {

					fri := UserMap[apiKey(id)]

					if !fri.loggedIn {
						continue
					}
					go SendLoggedIn(id, res.username, s)

				}

			}
		case BroadcastFriendship:
			// handle friendship broadcast
			if userIds, ok := message.Payload.(*[]string); ok {

				// Get user content per id, if active in UserMap
				go SendFriendshipData((*userIds)[1], s)
				go SendFriendshipData((*userIds)[2], s)
			}
		case BroadcastFriendRequest:
			// handle friendship broadcast
			if friendRequestId, ok := message.Payload.(string); ok {

				var userIds *[]string
				var err error
				// Get user ids from friendship id
				userIds, err = dbConn.GetFriendRequestById(friendRequestId)
				if err != nil {
					fmt.Println(err)
					break
				}

				if len(*userIds) == 0 {
					// Get user ids from friendship id
					userIds, err = dbConn.GetFriendshipById(friendRequestId)
				}

				if err != nil {
					fmt.Println(err)
					break
				}

				// Get user content per id, if active in UserMap

				// First user id is always the requesting user
				// Second user id is always the receiving user
				go SendFriendshipData((*userIds)[1], s)
				go SendFriendshipData((*userIds)[2], s)
			}
		case BroadcastChat:
			// handle friendship broadcast
			if chatBroadcast, ok := message.Payload.(*ChatBroadcast); ok {

				if !ok {
					fmt.Println("Error getting chat broadcast")
					break
				}

				// First user id is always the requesting user
				// Second user id is always the receiving user
				go SendChatData((*chatBroadcast.Friendship)[1], chatBroadcast.Chat, s)
				go SendChatData((*chatBroadcast.Friendship)[2], chatBroadcast.Chat, s)
			}
		default:
			// Do nothing
		}
	}
}

func SendLoggedIn(friendId string, user string, s *Server) {

	// Generate client response
	clientResp := ClientResponse{
		Code:    NotifyLogin,
		Err:     nil,
		Message: "Friend logged in",
		Payload: nil,
	}

	clientResp.EncodePayload(user)

	jsonData, err := json.Marshal(&clientResp)

	if err != nil {
		return
	}

	_, errr := s.clients[apiKey(friendId)].conn.Write(jsonData)

	if errr != nil {
		fmt.Println(errr)
		return
	}
}

// On update to friendship status, then this sennds data to the parties involved
func SendFriendshipData(u string, s *Server) {

	res, _ := UserMap[apiKey(u)]
	if res.loggedIn {

		userContent, err := dbConn.GetAllFriendsContent(apiKey(u))

		if err != nil {
			return
		}

		// Generate client response
		clientResp := ClientResponse{
			Code:    UpdateFriendContent,
			Err:     nil,
			Message: "All friend content",
			Payload: nil,
		}

		clientResp.EncodePayload(userContent)

		jsonData, err := json.Marshal(&clientResp)

		if err != nil {
			return
		}

		_, errr := s.clients[apiKey(u)].conn.Write(jsonData)

		if errr != nil {
			fmt.Println(errr)
			return
		}

	}

}

// On update to friendship status, then this sennds data to the parties involved
func SendChatData(u string, chat *Message, s *Server) {

	res, _ := UserMap[apiKey(u)]

	if res.loggedIn {
		// Generate client response
		clientResp := ClientResponse{
			Code:    ReceiveMessage,
			Err:     nil,
			Message: "New Message",
			Payload: nil,
		}

		clientResp.EncodePayload(chat)

		jsonData, err := json.Marshal(&clientResp)

		if err != nil {
			return
		}

		_, errr := s.clients[apiKey(u)].conn.Write(jsonData)

		if errr != nil {
			fmt.Println(errr)
			return
		}

	}

}
