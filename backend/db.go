package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

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
	INSERT INTO messages (id, welcomeSent, accountMade, username, password) VALUES (
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
	UPDATE messages
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

func (c *DBConn) GetAllUsers() error {

	var err error
	var rows *sql.Rows

	var outputString string

	// Query db
	rows, err = c.db.Query(
		`
		SELECT * FROM messages
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
			message:     "No new messages",
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

func (c *DBConn) GetUsers(s string) (*UsersSearch, error) {
	fmt.Println(s)
	var err error
	var rows *sql.Rows
	var stmt *sql.Stmt

	outputUsers := UsersSearch{}

	// Query db
	stmt, err = c.db.Prepare(
		`
		SELECT username FROM messages
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

func (c *DBConn) Close() error {
	err := c.db.Close()

	if err != nil {
		return err
	}
	return nil
}

// Create db and load all users into UserMap
func loadDB() error {
	db, err := sql.Open("sqlite3", "./message.db")
	if err != nil {
		log.Fatal(err)
	}

	dbConn.db = db

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS messages (
	id TEXT NOT NULL PRIMARY KEY, 
	welcomeSent INTEGER NOT NULL DEFAULT 0, 
	accountMade INTEGER NOT NULL DEFAULT 0,
	username TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL
	);
	`
	_, err = dbConn.db.Exec(sqlStmt)
	if err != nil {
		return err
	}
	return nil

}
