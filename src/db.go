package main

import (
	"log"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

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
	dbInit := `
		create table Users (id integer not null primary key, name text, pass text, gold int);
	`
	_, err := db.conn.Exec(dbInit)
	if err != nil {
		return err
	}
	return nil
}

func (db *DBClient) GetUserGold(name string) int {
	var gold int

	queryString := `
		select gold from Users where name = ?
	`
	transaction, err := db.conn.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer db.conn.Close()
	statement, err := transaction.Prepare(queryString)
	if err != nil {
		log.Fatal(err)
	}
	err = statement.QueryRow(name).Scan(&gold)
	if err != nil {
		log.Fatal(err)
	}
	return gold
}
