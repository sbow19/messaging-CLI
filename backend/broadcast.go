package main

import (
	"encoding/json"
	"fmt"
)

type BackendMessageCode int

const (
	BroadcastFriendship BackendMessageCode = iota
)

type BackendMessage struct {
	Code    BackendMessageCode
	Payload interface{}
}

func AppListener(mess chan *BackendMessage, s *Server) {

	for message := range mess {
		switch message.Code {
		case BroadcastFriendship:
			// handle friendship broadcast
			if friendshipId, ok := message.Payload.(string); ok {

				// Get user ids from friendship id
				userIds, err := dbConn.GetFriendshipById(friendshipId)
				if err != nil {
					fmt.Println(err)
					return
				}

				// Get user content per id, if active in UserMap
				go SendFriendshipData((*userIds)[1], s)
				go SendFriendshipData((*userIds)[2], s)
			}
		default:
			// Do nothing
		}
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
