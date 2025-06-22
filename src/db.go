package main

import (
	"log"

	"database/sql"

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
	return DBClient {
		conn: db,
	}
}

func (db *DBClient) InitDB() error {
	dbInitStatement := `
		CREATE TABLE IF NOT EXISTS Users (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL,
			name TEXT, 
			pass TEXT, 
			gold INTEGER,
			UNIQUE(id),
			UNIQUE(email)
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
	statement,err := transaction.Prepare(updateStatement)
	if err != nil {
		return err
	}
	_, err = statement.Exec(difference,name)
	if err != nil  {
		return err
	}
	return nil
}

func (db *DBClient) CheckAddUser(name string) (int64, error) {
	insertStatement := `
		INSERT INTO Users (name, email, pass, gold) VALUES (?, ?, ?, ?);
	`
	result, err := db.conn.Exec(insertStatement,name,"test@gmail.com",DEFAULT_PASS,DefaultGold)
	if err != nil {
		return 0,err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0,err
	}
	return id, nil
}


