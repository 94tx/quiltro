package main

import (
	"database/sql"
	"fmt"
	"time"
)

type err struct {
	scope string
	inner error
}
func (e *err) Error() string {
	return fmt.Sprintf("%s: %v", e.scope, e.inner)
}

func makeFailFn[T any](fn string) func(err error) (T, *err) {
	var t T
	return func(e error) (T, *err) {
		return t, &err{fn, e}
	}
}

// Create a new post in the database. This post will have no content and
// will need to be revised (see revisePost) for it to become valid.
func insertPost(db *sql.DB, post Post) (int64, *err) {
	fail := makeFailFn[int64]("insertPost")
	res, err := db.Exec(
		`INSERT INTO post(revision_count, author, permalink, published_at)
		 VALUES(?, ?, ?, ?);`,
		0, post.Author, post.Permalink, post.Published,
	)
	if err != nil {
		return fail(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fail(err)
	}

	return id, nil
}

func addRevision(db *sql.DB, postID int64, rev Revision) (int64, *err) {
	fail := makeFailFn[int64]("addRevision")
	tx, err := db.Begin()
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	var revisionCount int64
	row := tx.QueryRow(
		`SELECT revision_count FROM post WHERE id = ?;`,
		postID,
	)
	if err := row.Scan(&revisionCount); err != nil {
		return fail(err)
	}

	res, err := tx.Exec(
		`INSERT INTO revision(post_id, revision_index, title, body, created_at)
		 VALUES(?, ?, ?, ?, ?);`,
		postID,
		revisionCount,
		rev.Title,
		rev.Body,
		maketimestamp(time.Now()),
	)
	if err != nil {
		return fail(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fail(err)
	}

	_, err = tx.Exec(
		`UPDATE post SET revision_count = ? WHERE id = ?;`,
		revisionCount+1, postID,
	)
	if err != nil {
		return fail(err)
	}

	if err := tx.Commit(); err != nil {
		return fail(err)
	}
	return id, nil
}

func lookPost(db *sql.DB, permalink string) (Post, *err) {
	fail := makeFailFn[Post]("lookPost")
	var post Post
	
	row := db.QueryRow(
		`SELECT post.id, post.revision_count, post.author, post.permalink, post.published_at, revision.id, revision.revision_index, revision.title, revision.body, revision.created_at FROM post INNER JOIN revision ON revision.post_id = post.id WHERE post.permalink = ? AND revision.revision_index = post.revision_count - 1;`,
		permalink,
	)
	if err := row.Scan(
		&post.ID,
		&post.RevisionCount,
		&post.Author,
		&post.Permalink,
		&post.Published,
		&post.Latest.ID,
		&post.Latest.RevisionIndex,
		&post.Latest.Title,
		&post.Latest.Body,
		&post.Latest.Created,
	); err != nil {
		return fail(err)
	}
	
	return post, nil
}
