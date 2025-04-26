package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

/*
		Load client data into UserMap

		type clientData struct {
		message      string
		welcomeSent  bool
		loginDetails LoginDetails
		loggedIn     bool
		apiKey       apiKey
		active       bool
		err          error
		mu           sync.Mutex
	}
*/

type DBConn struct {
	db *sql.DB
}

var dbConn = &DBConn{
	db: nil,
}

func (c *DBConn) CreateNewUser(d *clientData) error {

	var err error
	var stmt *sql.Stmt

	// Create transaction
	tx, err := c.db.Begin()

	if err != nil {
		goto retErr
	}

	// Prepare save statement
	stmt, err = tx.Prepare(
		`
	INSERT INTO users (id, welcomeSent, accountMade, username, password) VALUES (
		?,?,?,?,?
	);
	`,
	)

	if err != nil {
		goto retErr
	}
	defer stmt.Close()

	// Execute statement
	_, err = stmt.Exec(
		d.apiKey,
		0,
		0,
		d.loginDetails.Username,
		d.loginDetails.Password,
	)

	if err != nil {
		goto rollback
	}

	err = tx.Commit()

	if err != nil {
		goto rollback
	}

	return nil

	// Cleanup
rollback:
	{
		tx.Rollback()
	}
retErr:
	{
		return err
	}

}

func btobool(b uint8) bool {

	if b == 1 {
		return true
	} else if b == 0 {
		return false
	}

	return false

}

func booltob(b bool) uint8 {

	if b {
		return 1
	} else if !b {
		return 0
	}

	return 2

}

func (c *DBConn) UpdateClient(d *clientData) error {
	var err error
	var stmt *sql.Stmt
	var accMade uint8
	var welcomeSent uint8

	// Create transaction
	tx, err := c.db.Begin()

	if err != nil {
		goto retErr
	}

	// Prepare save statement
	stmt, err = tx.Prepare(
		`
	UPDATE users
	SET welcomeSent= ?, accountMade = ?, username = ?, password = ?
	WHERE id = ?
	;
	`,
	)

	if err != nil {
		goto retErr
	}
	defer stmt.Close()

	// Set binary flags
	accMade = booltob(d.accountMade)
	welcomeSent = booltob(d.welcomeSent)

	// Execute statement
	_, err = stmt.Exec(
		welcomeSent,
		accMade,
		d.loginDetails.Username,
		d.loginDetails.Password,
		d.apiKey,
	)

	if err != nil {
		goto rollback
	}

	err = tx.Commit()

	if err != nil {
		goto rollback
	}

	return nil

	// Cleanup
rollback:
	{
		tx.Rollback()
	}
retErr:
	{
		return err
	}
}

// Get all users and log to db.txt
func (c *DBConn) GetAll() error {

	var err error
	var rows *sql.Rows

	var outputString string

	// Query db
	rows, err = c.db.Query(
		`
		SELECT * FROM users
		;
		`,
	)
	if err != nil {
		goto retErr
	}
	defer rows.Close()

	// Data to write into output text file
	for rows.Next() {
		var apiKey apiKey
		var welcomeSent uint8
		var accountMade uint8
		var username string
		var password string

		err = rows.Scan(
			&apiKey,
			&welcomeSent,
			&accountMade,
			&username,
			&password,
		)

		if err != nil {
			goto retErr
		}

		outputString += fmt.Sprintf("id: %q, u: %q, p: %q, welcome:%d, acc:%d\n",
			apiKey,
			username,
			password,
			welcomeSent,
			accountMade,
		)

		// Convert
		welsent := btobool(welcomeSent)
		accMade := btobool(accountMade)

		// Add data to UserMap
		UserMap[apiKey] = &clientData{
			apiKey:      apiKey,
			username:    username,
			message:     "No new users",
			accountMade: accMade,
			welcomeSent: welsent,
			loginDetails: LoginDetails{
				Username: username,
				Password: password,
			},
			loggedIn: false,
			active:   false,
			err:      nil,
			mu:       sync.Mutex{},
			rwmu:     sync.RWMutex{},
		}
	}

	/*
		Friend requests
	*/

	outputString += "\n\n"
	rows, err = c.db.Query(
		`
		SELECT * FROM friend_requests
		;
		`,
	)
	if err != nil {
		goto retErr
	}
	defer rows.Close()

	// Data to write into output text file
	for rows.Next() {
		var id string
		var user1 string
		var user2 string

		err = rows.Scan(
			&id,
			&user1,
			&user2,
		)

		if err != nil {
			goto retErr
		}

		outputString += fmt.Sprintf("requestid: %q, reqId: %q, resId: %q\n",
			id,
			user1,
			user2,
		)
	}

	/*
		Friendships
	*/
	outputString += "\n\n"
	rows, err = c.db.Query(
		`
		SELECT * FROM friends
		;
		`,
	)
	if err != nil {
		goto retErr
	}
	defer rows.Close()

	// Data to write into output text file
	for rows.Next() {
		var id string
		var user1 string
		var user2 string

		err = rows.Scan(
			&id,
			&user1,
			&user2,
		)

		if err != nil {
			goto retErr
		}

		outputString += fmt.Sprintf("friendshipId: %q, reqId: %q, resId: %q\n",
			id,
			user1,
			user2,
		)
	}

	os.WriteFile("db.txt", []byte(outputString), 0644)
	// Reveal any errors encountered why executing query
	err = rows.Err()

	if err != nil {
		goto retErr
	}

	return nil

retErr:
	{
		return err
	}
}

type UsersSearch []string

// Get users by search string on username
func (c *DBConn) GetUsers(s string) (*UsersSearch, error) {
	var err error
	var rows *sql.Rows
	var stmt *sql.Stmt

	outputUsers := UsersSearch{}

	// Query db
	stmt, err = c.db.Prepare(
		`
		SELECT username FROM users
		WHERE username LIKE ?
		;
		`,
	)
	if err != nil {
		goto retErr
	}
	defer stmt.Close()

	rows, err = stmt.Query(s + "%")
	if err != nil {
		goto retErr
	}
	defer rows.Close()

	// Data to write into output text file
	for rows.Next() {

		var username string
		if err := rows.Scan(&username); err != nil {
			goto retErr
		}
		outputUsers = append(outputUsers, username)
	}
	return &outputUsers, nil

retErr:
	{
		return nil, err
	}
}

// Get id by username
func (c *DBConn) GetUserAPI(s string) (*UsersSearch, error) {
	var err error
	var rows *sql.Rows
	var stmt *sql.Stmt

	outputUsers := UsersSearch{}

	// Query db
	stmt, err = c.db.Prepare(
		`
		SELECT id FROM users
		WHERE username = ?
		;
		`,
	)
	if err != nil {
		goto retErr
	}
	defer stmt.Close()

	rows, err = stmt.Query(s)
	if err != nil {
		goto retErr
	}
	defer rows.Close()

	// Data to write into output text file
	for rows.Next() {

		var id string
		if err := rows.Scan(&id); err != nil {
			goto retErr
		}
		outputUsers = append(outputUsers, id)

	}
	return &outputUsers, nil

retErr:
	{
		return nil, err
	}
}

// Get a friend request by both reqid andr esId
func (c *DBConn) GetFriendRequestByIds(reqId string, resId string) (*UsersSearch, error) {
	var err error
	var rows *sql.Rows
	var stmt *sql.Stmt

	outputUsers := UsersSearch{}

	// Query db
	stmt, err = c.db.Prepare(
		`
		SELECT id FROM friend_requests
		WHERE reqId = ? 
		AND resId = ? 
		;
		`,
	)
	if err != nil {
		goto retErr
	}
	defer stmt.Close()

	rows, err = stmt.Query(reqId, resId)
	if err != nil {
		goto retErr
	}
	defer rows.Close()

	// Data to write into output text file
	for rows.Next() {

		var id string
		if err := rows.Scan(&id); err != nil {
			goto retErr
		}
		outputUsers = append(outputUsers, id)
	}
	return &outputUsers, nil

retErr:
	{
		return nil, err
	}
}

// By friend_request ID
func (c *DBConn) GetFriendRequestById(friendshipId string) (*[]string, error) {
	var err error
	var rows *sql.Rows
	var stmt *sql.Stmt
	var output []string

	// Query db
	stmt, err = c.db.Prepare(
		`
		SELECT * FROM friend_requests
		WHERE id = ?
		;
		`,
	)
	if err != nil {
		goto retErr
	}
	defer stmt.Close()

	rows, err = stmt.Query(friendshipId)
	if err != nil {
		goto retErr
	}
	defer rows.Close()

	// Data to write into output text file
	for rows.Next() {

		var id string
		var reqId string
		var resId string

		if err := rows.Scan(&id, &reqId, &resId); err != nil {
			goto retErr
		}
		output = append(output, id, reqId, resId)
	}
	return &output, nil

retErr:
	{
		return nil, err
	}
}

// By friendship ID
func (c *DBConn) GetFriendshipById(friendshipId string) (*[]string, error) {
	var err error
	var rows *sql.Rows
	var stmt *sql.Stmt
	var output []string

	// Query db
	stmt, err = c.db.Prepare(
		`
		SELECT * FROM friends
		WHERE id = ?
		;
		`,
	)
	if err != nil {
		goto retErr
	}
	defer stmt.Close()

	rows, err = stmt.Query(friendshipId)
	if err != nil {
		goto retErr
	}
	defer rows.Close()

	// Data to write into output text file
	for rows.Next() {

		var id string
		var user1 string
		var user2 string

		if err := rows.Scan(&id, &user1, &user2); err != nil {
			goto retErr
		}
		output = append(output, id, user1, user2)
	}
	return &output, nil

retErr:
	{
		return nil, err
	}
}

// Get  friendship id by  friend ids
func (c *DBConn) GetFriendshipByIds(id1 string, id2 string) (*[]string, error) {
	var err error
	var rows *sql.Rows
	var stmt *sql.Stmt
	var output []string

	// Query db
	stmt, err = c.db.Prepare(
		`
		SELECT * FROM friends
		WHERE (user1 = ? AND user2 = ?)
   			OR (user1 = ? AND user2 = ?)
		;
		`,
	)
	if err != nil {
		goto retErr
	}
	defer stmt.Close()

	rows, err = stmt.Query(id1, id2, id2, id1)
	if err != nil {
		goto retErr
	}
	defer rows.Close()

	// Data to write into output text file
	for rows.Next() {

		var friendshipId string
		var user1 string
		var user2 string

		if err := rows.Scan(&friendshipId, &user1, &user2); err != nil {
			goto retErr
		}
		output = append(output, friendshipId, user1, user2)
	}

	return &output, nil

retErr:
	{
		return nil, err
	}
}

func generateId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (c *DBConn) SetFriendRequest(name string, reqId string) (string, error) {

	var err error
	var stmt *sql.Stmt
	var userSearch *UsersSearch
	var id string
	var resId string

	userSearch, err = c.GetUserAPI(name)

	var tx *sql.Tx
	if err != nil {
		goto retErr
	}

	// resId is receiving request, req is the requesting user
	resId = (*userSearch)[0]

	// Check if reverse request has already been made
	userSearch, err = c.GetFriendRequestByIds(resId, reqId)

	if err != nil {
		goto retErr
	}

	if len(*userSearch) > 0 {
		return (*userSearch)[0], fmt.Errorf("friend request already exists")
	}

	//Friend request id
	id, err = generateId()
	if err != nil {
		goto retErr
	}

	// Create transaction
	tx, err = c.db.Begin()

	if err != nil {
		goto retErr
	}

	// Prepare save statement
	stmt, err = tx.Prepare(
		`
	INSERT INTO friend_requests (id, reqId, resId) VALUES (
		?,?,?
	);
	`,
	)

	if err != nil {
		goto retErr
	}
	defer stmt.Close()

	// Execute statement
	_, err = stmt.Exec(
		id,
		reqId,
		resId,
	)

	if err != nil {
		goto rollback
	}

	err = tx.Commit()

	if err != nil {
		goto rollback
	}

	return id, nil

	// Cleanup
rollback:
	{
		tx.Rollback()
	}
retErr:
	{
		return "", err
	}

}

func (c *DBConn) DeleteFriendRequest(requestId string) error {

	var err error
	var stmt *sql.Stmt

	var tx *sql.Tx
	if err != nil {
		goto retErr
	}

	// Create transaction
	tx, err = c.db.Begin()

	if err != nil {
		goto retErr
	}

	// Prepare delete statement
	stmt, err = tx.Prepare(
		`
	DELETE FROM friend_requests WHERE id = ?;
	`,
	)

	if err != nil {
		goto retErr
	}
	defer stmt.Close()

	// Execute statement
	_, err = stmt.Exec(
		requestId,
	)

	if err != nil {
		goto rollback
	}

	err = tx.Commit()

	if err != nil {
		goto rollback
	}

	return nil

	// Cleanup
rollback:
	{
		tx.Rollback()
	}
retErr:
	{
		return err
	}

}

func (c *DBConn) CreateFriend(f *FriendAcceptData, userId string) error {

	var err error
	var stmt *sql.Stmt
	var tx *sql.Tx
	var friendshipId string
	var res *[]string

	//Friend request id
	friendshipId, err = generateId()
	if err != nil {
		goto retErr
	}

	// Get reqid and resid
	res, err = c.GetFriendRequestById(f.RequestId)
	if err != nil {
		goto retErr
	}

	// Create transaction
	tx, err = c.db.Begin()

	if err != nil {
		goto retErr
	}

	// Prepare delete statement
	stmt, err = tx.Prepare(
		`
	INSERT INTO friends VALUES (?,?,?);
	`,
	)

	if err != nil {
		goto retErr
	}
	defer stmt.Close()

	// Execute statement
	_, err = stmt.Exec(
		friendshipId,
		(*res)[1], // Requester ID
		(*res)[2], // Receiver ID
	)

	if err != nil {
		goto rollback
	}

	// Delete friendship request
	// Prepare delete statement
	stmt, err = tx.Prepare(
		`
	DELETE FROM friend_requests WHERE id = ?;
	`,
	)

	if err != nil {
		goto retErr
	}

	// Execute statement
	_, err = stmt.Exec(
		f.RequestId,
	)

	if err != nil {
		goto rollback
	}

	err = tx.Commit()

	if err != nil {
		goto rollback
	}

	return nil

	// Cleanup
rollback:
	{
		tx.Rollback()
	}
retErr:
	{
		fmt.Println(err)
		return err
	}

}

// Save message
func (c *DBConn) SaveMessage(chat *Chat, userId apiKey) (*[]string, error) {

	var err error
	var stmt *sql.Stmt
	var tx *sql.Tx
	var res *UsersSearch
	var messageId string
	var friendship *[]string
	var id1 string

	//Message id
	messageId, err = generateId()
	if err != nil {
		goto retErr
	}

	// Get receiver id
	res, err = c.GetUserAPI(chat.Receiver)

	if err != nil || len(*res) == 0 {
		goto retErr
	}

	id1 = (*res)[0]
	// Get friendship id
	friendship, err = c.GetFriendshipByIds(id1, string(userId))

	// Create transaction
	tx, err = c.db.Begin()

	if err != nil {
		goto retErr
	}

	// Prepare delete statement
	stmt, err = tx.Prepare(
		`
	INSERT INTO messages (id, friendId, senderId, message) VALUES (?,?,?,?);
	`,
	)

	if err != nil {
		goto retErr
	}
	defer stmt.Close()

	// Execute statement
	_, err = stmt.Exec(
		messageId,
		(*friendship)[0],
		userId,
		chat.Text,
	)

	if err != nil {
		goto rollback
	}

	err = tx.Commit()

	if err != nil {
		goto rollback
	}

	return friendship, nil

	// Cleanup
rollback:
	{
		tx.Rollback()
	}
retErr:
	{
		fmt.Println(err)
		return nil, err
	}

}

func (c *DBConn) GetAllUserContent(k apiKey) (*UserContent, error) {

	var err error
	var rows *sql.Rows
	userContent := UserContent{}

	matchedFriendids := make(map[string]string)
	friends := []Friend{}

	friendRequests := []FriendReqDetails{}
	messages := Messages{}

	// Get friends
	rows, err = c.db.Query(
		`
		SELECT * FROM friends 
		WHERE user1 = ?
		OR  user2 = ?
		;
		`, k, k,
	)

	if err != nil {
		goto retErr
	}

	// Data to write into output text file
	for rows.Next() {

		var id string
		var user1 string
		var user2 string

		if err := rows.Scan(&id, &user1, &user2); err != nil {
			goto retErr
		}

		if string(k) != user1 {
			matchedFriendids[user1] = id
		}

		if string(k) != user2 {
			matchedFriendids[user2] = id
		}
	}

	// Get friend details
	for key, _ := range matchedFriendids {
		result, ok := UserMap[apiKey(key)]
		if !ok {
			continue
		}
		friends = append(friends, Friend{
			Username: result.username,
			Active:   result.active,
			Message:  result.message,
		})
	}

	// Get friend requests
	rows, err = c.db.Query(
		`
		SELECT * FROM friend_requests
		WHERE reqId = ?
		OR  resId = ?
		;
		`, k, k,
	)

	if err != nil {
		goto retErr
	}

	for rows.Next() {

		var id string
		var reqId string
		var resId string

		if err := rows.Scan(&id, &reqId, &resId); err != nil {
			continue
		}

		if string(k) != reqId {

			result, _ := UserMap[apiKey(reqId)]

			friendRequests = append(friendRequests, FriendReqDetails{
				Username:   result.username,
				RequestId:  id,
				FromClient: false,
			})
		}

		if string(k) != resId {
			result, _ := UserMap[apiKey(resId)]

			friendRequests = append(friendRequests, FriendReqDetails{
				Username:   result.username,
				RequestId:  id,
				FromClient: true,
			})
		}

	}

	// Get messages. Cycle through matched friends lists and fetch message  data from database
	for friendId, friendshipId := range matchedFriendids {

		rows, err = c.db.Query(
			`
			SELECT * FROM messages
			WHERE friendId = ?
			AND date > datetime('now', '-3 days')
			;
			`, friendshipId,
		)

		friendName := UserMap[apiKey(friendId)].username
		messages[friendName] = []Message{}

		if err != nil {
			continue
		}

		// Cycle through messages and add to messages
		for rows.Next() {
			var messageId string
			var friendshipId string
			var senderId string
			var message string
			var date string
			var sender string

			if err := rows.Scan(&messageId, &friendshipId, &senderId, &message, &date); err != nil {
				continue
			}

			if senderId == string(k) {
				result, _ := UserMap[k]
				sender = result.username
			} else {
				sender = friendName
			}

			// Parse it using the correct layout
			t, err := time.Parse(time.RFC3339, date)
			if err != nil {
				panic(err)
			}

			// Convert to your desired format
			layout := "2006-01-02 15:04"
			formatted := t.Format(layout)

			messages[friendName] = append(messages[friendName], Message{
				Text:   message,
				Date:   formatted,
				Sender: sender,
			})

		}

	}

	// Set user content
	userContent.Friends = friends
	userContent.FriendRequests = friendRequests
	userContent.Messages = messages

	return &userContent, nil

retErr:
	{
		rows.Close()
		return nil, err
	}

}

// Get all user content
func (c *DBConn) GetAllFriendsContent(k apiKey) (*UserContent, error) {

	var err error
	var rows *sql.Rows
	userContent := UserContent{}

	matchedFriendids := make(map[string]string)
	friends := []Friend{}

	friendRequests := []FriendReqDetails{}

	// Get friends
	rows, err = c.db.Query(
		`
		SELECT * FROM friends 
		WHERE user1 = ?
		OR  user2 = ?
		;
		`, k, k,
	)

	if err != nil {
		goto retErr
	}

	// Data to write into output text file
	for rows.Next() {

		var id string
		var user1 string
		var user2 string

		if err := rows.Scan(&id, &user1, &user2); err != nil {
			goto retErr
		}

		if string(k) != user1 {
			matchedFriendids[user1] = id
		}

		if string(k) != user2 {
			matchedFriendids[user2] = id
		}
	}

	// Get friend details
	for key, _ := range matchedFriendids {
		result, ok := UserMap[apiKey(key)]
		if !ok {
			continue
		}
		friends = append(friends, Friend{
			Username: result.username,
			Active:   result.active,
			Message:  result.message,
		})
	}

	// Get friend requests
	rows, err = c.db.Query(
		`
		SELECT * FROM friend_requests
		WHERE reqId = ?
		OR  resId = ?
		;
		`, k, k,
	)

	if err != nil {
		goto retErr
	}

	for rows.Next() {

		var id string
		var reqId string
		var resId string

		if err := rows.Scan(&id, &reqId, &resId); err != nil {
			continue
		}

		if string(k) != reqId {

			result, _ := UserMap[apiKey(reqId)]

			friendRequests = append(friendRequests, FriendReqDetails{
				Username:   result.username,
				RequestId:  id,
				FromClient: false,
			})
		}

		if string(k) != resId {
			result, _ := UserMap[apiKey(resId)]

			friendRequests = append(friendRequests, FriendReqDetails{
				Username:   result.username,
				RequestId:  id,
				FromClient: true,
			})
		}

	}
	// Set user content
	userContent.Friends = friends
	userContent.FriendRequests = friendRequests
	userContent.Messages = nil

	return &userContent, nil

retErr:
	{
		rows.Close()
		return nil, err
	}

}

func (c *DBConn) Close() error {
	err := c.db.Close()

	if err != nil {
		return err
	}
	return nil
}

// Create db and load all users into UserMap
func loadDB() error {
	db, err := sql.Open("sqlite3", "./cli.db?_foreign_keys=on")
	if err != nil {
		log.Fatal(err)
	}

	dbConn.db = db

	// Users table
	_, err = dbConn.db.Exec(createUsersDBStmt)
	if err != nil {
		return err
	}

	// Friend requests table
	_, err = dbConn.db.Exec(createFriendRequestsTable)
	if err != nil {
		return err
	}

	// Friends table
	_, err = dbConn.db.Exec(createFriendsTable)
	if err != nil {
		return err
	}

	// Messages table
	_, err = dbConn.db.Exec(messagesTable)
	if err != nil {
		return err
	}
	return nil

}

var createUsersDBStmt = `
	CREATE TABLE IF NOT EXISTS users (
	id TEXT NOT NULL PRIMARY KEY, 
	welcomeSent INTEGER NOT NULL DEFAULT 0, 
	accountMade INTEGER NOT NULL DEFAULT 0,
	username TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL
	);
	`
var createFriendRequestsTable = `
	CREATE TABLE IF NOT EXISTS friend_requests (
		id TEXT NOT NULL PRIMARY KEY,
		reqId TEXT NOT NULL, 
		resId TEXT NOT NULL, 
		FOREIGN KEY(reqId) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY(resId) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(reqId, resId)

	);
`
var createFriendsTable = `
CREATE TABLE IF NOT EXISTS friends (
	id TEXT NOT NULL PRIMARY KEY,
	user1 TEXT NOT NULL,
	user2 TEXT NOT NULL,
	FOREIGN KEY(user1) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY(user2) REFERENCES users(id) ON DELETE CASCADE,
	UNIQUE(user1, user2)
);
`

var messagesTable = `
	CREATE TABLE IF NOT EXISTS messages (
	id TEXT NOT NULL PRIMARY KEY,
	friendId TEXT NOT NULL,
	senderId TEXT NOT NULL,
	message TEXT, 
	date DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY(friendId)  REFERENCES friends(id)
);
`
