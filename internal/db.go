package internal

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const DefaultGold = 20
const DEFAULT_PASS = "EASILYGUESSABLE"

var lastID = 0

type DBClient struct {
	conn *sql.DB
}

func CreateClient() DBClient {
	db, err := sql.Open("sqlite3", "./users.db")
	if err != nil {
		log.Fatal(err)
	}
	return DBClient{
		conn: db,
	}
}

func (db *DBClient) InitDB() error {
	dbInitStatement := `
		CREATE TABLE IF NOT EXISTS Users (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL, 
			pass TEXT NOT NULL, 
			gold INTEGER,
			UNIQUE(id),
			UNIQUE(name)
		);
	`
	_, err := db.conn.Exec(dbInitStatement)
	return err
}

func (db *DBClient) GetUserGold(name string) (int, error) {
	var gold int
	queryString := `
		SELECT gold FROM Users WHERE name = ?;
	`
	transaction, err := db.conn.Begin()
	defer db.conn.Close()
	if err != nil {
		return 0, err
	}
	statement, err := transaction.Prepare(queryString)
	if err != nil {
		return 0, err
	}
	err = statement.QueryRow(name).Scan(&gold)
	if err != nil {
		return 0, err
	}
	return gold, nil
}

func (db *DBClient) ChangeUserGold(name string, difference int) error {
	updateStatement := `
			UPDATE Users SET gold = gold + ? WHERE name = ?;
	`
	transaction, err := db.conn.Begin()
	defer db.conn.Close()
	if err != nil {
		return err
	}
	statement, err := transaction.Prepare(updateStatement)
	if err != nil {
		return err
	}
	_, err = statement.Exec(difference, name)
	if err != nil {
		return err
	}
	return nil
}

func (db *DBClient) CheckAddUser(name string, pass string) (int64, error) {
	selectStatement := `
		SELECT id FROM Users WHERE name == ? AND pass == ?;
	`
	foundUser := db.conn.QueryRow(selectStatement, name, pass)
	var foundId int64
	err := foundUser.Scan(&foundId)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User attempted to login but was not found, resorting to create the user")
		} else {
			return foundId, nil
		}
	} else {
		return foundId, nil
	}

	insertStatement := `
		INSERT INTO Users (name, pass, gold) VALUES (?, ?, ?);
	`
	insertedUser, err := db.conn.Exec(insertStatement, name, pass, DefaultGold)
	if err != nil {
		return 0, err
	}
	id, err := insertedUser.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
