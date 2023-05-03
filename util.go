package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "embed"

	_ "humungus.tedunangst.com/r/go-sqlite3"
)

//go:embed schema.sql
var schema string 
var dbhandle *sql.DB
var getconfigStmt *sql.Stmt
var setconfigStmt *sql.Stmt

func makeUserError(err *err, fmtstr string, args... any) error {
	if err.inner == sql.ErrNoRows {
		args = append(args, "not found")
		return fmt.Errorf(fmtstr, args...)
	} else {
		return err
	}
}

func initializeDB(filename string) {
	// TODO: configuration stuff
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatalf("initializeDB: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("initializeDB: %v", err)
	}
}

func openDB(filename string) *sql.DB {
	if dbhandle != nil {
		return dbhandle
	}
	
	if _, err := os.Stat(filename); err != nil {
		log.Fatalf("openDB: %v", err)
	}
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatalf("openDB: %v", err)
	}

	getconfigStmt, err = db.Prepare("SELECT value FROM config WHERE key = ?")
	if err != nil {
		log.Fatalf("openDB: %v", err)
	}
	setconfigStmt, err = db.Prepare("UPDATE config SET value = ? WHERE key = ?")
	if err != nil {
		log.Fatalf("openDB: %v", err)
	}
	
	dbhandle = db
	return db
}

func getconfig(key string, value interface{}) error {
	row := getconfigStmt.QueryRow(key)
	err := row.Scan(value)
	if err == sql.ErrNoRows {
		err = nil
	}
	return err
}

func maketimestamp(tm time.Time) sql.NullTime {
	if tm.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: tm, Valid: true}
}
