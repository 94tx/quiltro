package main

import (
	"database/sql"
	"time"
)

type User struct {
	ID int64
	Username, Hash string
}

type Session struct {
	ID, UserID int64
	Hash string
	Expiry time.Time
}

type Post struct {
	ID            int64
	RevisionCount int64
	Author        string
	Permalink     string
	Published     sql.NullTime

	Latest Revision
}

type Revision struct {
	ID            int64
	RevisionIndex int64
	Created       time.Time
	Title, Body   string
}

type File struct {
	ID int64
	Filename string
	Uploaded time.Time
	Data []byte
}

type Image struct {
	File
	X, Y int64
}
