package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "humungus.tedunangst.com/r/go-sqlite3"
)

//go:embed schema.sql
var Schema string

type DB struct {
	*sql.DB
}

type Post struct {
	ID        int64
	Author    string
	Permalink string
	Published time.Time
}

type Revision struct {
	ID          int64
	Title, Body string
	Keywords    []string
	Timestamp   time.Time
}

type User struct {
	ID         int64
	Name, Hash string
}

type Session struct {
	ID, UserID int64
	Hash       string
	Expiry     time.Time
}

type File struct {
	ID       int64
	Name     string
	Uploaded time.Time
}

type Image struct {
	File
	X, Y int64
}

func Initialize(file string) error {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return fmt.Errorf("db.Initialize: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(Schema); err != nil {
		return fmt.Errorf("db.Initialize: %v", err)
	}
	return nil
}

func Open(file string) (*DB, error) {
	var (
		err error
		db  = new(DB)
	)

	db.DB, err = sql.Open("sqlite3", file)
	if err != nil {
		return nil, fmt.Errorf("db.Open: %v", err)
	}

	return db, nil
}

// Create a new post. Returns the newly created post's ID.
// The created post has no revisions, so it needs to be revised before it is
// apt for display. See `RevisePost`
func (db *DB) NewPost(data Post) (int64, error) {
	fail := func(err error) (int64, error) {
		return 0, fmt.Errorf("db.NewPost: %v", err)
	}

	res, err := db.Exec(`INSERT INTO posts(permalink, author, published_at) VALUES(?, ?, ?)`,
		data.Permalink,
		data.Author,
		data.Published,
	)
	if err != nil {
		return fail(err)
	}

	postID, err := res.LastInsertId()
	if err != nil {
		return fail(err)
	}

	return postID, nil
}

// Add a new revision to a post. Returns the newly created revision's ID.
func (db *DB) RevisePost(id int64, data Revision) (int64, error) {
	fail := func(err error) (int64, error) {
		return 0, fmt.Errorf("db.RevisePost: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO revisions(postid, revised_at) VALUES (?, ?)`, id, time.Now())
	if err != nil {
		return fail(err)
	}

	revID, err := res.LastInsertId()
	if err != nil {
		return fail(err)
	}

	for idx := range data.Keywords {
		data.Keywords[idx] = normalizeKeyword(data.Keywords[idx])
	}

	if _, err := tx.Exec(`INSERT INTO revisions_fts(rowid, title, body, keywords) VALUES (?, ?, ?, ?)`,
		revID,
		data.Title,
		data.Body,
		strings.Join(data.Keywords, ";"),
	); err != nil {
		return fail(err)
	}

	if err := tx.Commit(); err != nil {
		return fail(err)
	}

	return revID, nil
}

// Get nth revision for post, with n = 0 being the latest revision.
func (db *DB) GetNthRevision(id int64, n int64) (Revision, error) {
	fail := func(err error) (Revision, error) {
		return Revision{}, fmt.Errorf("db.GetNthRevision: %v", err)
	}

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	var rev Revision

	var res *sql.Row
	if n == 0 {
		res = tx.QueryRow(`SELECT revid, revised_at FROM revisions WHERE postid = ? ORDER BY revid DESC LIMIT 1`, id, n)	
	} else {
		res = tx.QueryRow(`SELECT revid, revised_at FROM revisions WHERE postid = ? ORDER BY revid ASC LIMIT 1 OFFSET ?`, id, n - 1)
	}
		var kws string
		if err := res.Scan(&rev.ID, &rev.Timestamp); err != nil {
			return fail(err)
		}

		res = tx.QueryRow(`SELECT title, body, keywords FROM revisions_fts WHERE rowid = ?`, rev.ID)
		if err := res.Scan(
			&rev.Title,
			&rev.Body,
			&kws,
		); err != nil {
			return fail(err)
		}

		rev.Keywords = strings.Split(kws, ";")
	
	if err := tx.Commit(); err != nil {
		return fail(err)
	}

	return rev, nil
}

/* Get a post's data and (if applicable) latest revision from the post's permalink.
 * The "complete" parameter will dictate if the method should query the latest revision
 * or not. */
func (db *DB) GetPostByPermalink(permalink string, complete bool) (Post, Revision, error) {
	var post Post
	var revision Revision

	fail := func(err error) (Post, Revision, error) {
		return Post{}, Revision{}, fmt.Errorf("db.GetPostByPermalink: %v", err)
	}

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	res := tx.QueryRow(`SELECT postid, author, permalink, published_at FROM posts WHERE permalink = ?`, permalink)

	if err := res.Scan(
		&post.ID,
		&post.Author,
		&post.Permalink,
		&post.Published,
	); err != nil {
		return fail(err)
	}

	if complete {
		revision, err = db.GetNthRevision(post.ID, 0)
		if err != nil {
			return fail(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fail(err)
	}

	return post, revision, nil
}

// Return the IDs of all revisions that belong to a post
func (db *DB) GetRevisionIDsForPost(id int64) ([]int64, error) {
	fail := func(err error) ([]int64, error) {
		return nil, fmt.Errorf("db.GetRevisionIDsForPost: %v", err)
	}

	revIDs := make([]int64, 0)
	res, err := db.Query(`SELECT revid FROM revisions WHERE postid = ? ORDER BY revid DESC`, id)
	if err != nil {
		return fail(err)
	}

	for res.Next() {
		var id int64
		if err := res.Scan(&id); err != nil {
			return fail(err)
		}

		revIDs = append(revIDs, id)
	}

	if err := res.Err(); err != nil {
		return fail(err)
	}

	return revIDs, nil
}

func (db *DB) GetRevisionsForPost(id int64) ([]Revision, error) {
	fail := func(err error) ([]Revision, error) {
		return nil, fmt.Errorf("db.GetRevisionsForPost: %v", err)
	}

	revIDs, err := db.GetRevisionIDsForPost(id)
	if err != nil {
		return fail(err)
	}

	revs := make([]Revision, len(revIDs))
	for idx, revID := range revIDs {
		rev, err := db.GetRevisionByID(revID)
		if err != nil {
			return fail(err)
		}

		revs[idx] = rev
	}

	return revs, nil
}

// Get a revision from its ID
func (db *DB) GetRevisionByID(id int64) (Revision, error) {
	var rev Revision
	fail := func(err error) (Revision, error) {
		return Revision{}, fmt.Errorf("db.GetRevisionByID: %v", err)
	}

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	res := tx.QueryRow(`SELECT revid, revised_at FROM revisions WHERE revid = ?`, id)
	if err := res.Scan(
		&rev.ID,
		&rev.Timestamp,
	); err != nil {
		return fail(err)
	}

	kws := ""
	res = tx.QueryRow(`SELECT title, keywords, body FROM revisions_fts WHERE rowid = ?`, id)
	if err := res.Scan(
		&rev.Title,
		&kws,
		&rev.Body,
	); err != nil {
		return fail(err)
	}

	rev.Keywords = strings.Split(kws, ";")

	if err := tx.Commit(); err != nil {
		return fail(err)
	}

	return rev, nil
}

var kwNormalizeRegexp = regexp.MustCompile(`(\pM|\pZ|\pS|\pP|\pC)+`)

func normalizeKeyword(kw string) string {
	return kwNormalizeRegexp.ReplaceAllString(
		strings.ToLower(strings.TrimSpace(kw)), "",
	)
}
