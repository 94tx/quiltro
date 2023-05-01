package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"quiltro/database"
	"strings"
)

var db *database.DB

func main() {
	os.RemoveAll("quiltro.db")

	var err error
	if err := database.Initialize("quiltro.db"); err != nil {
		log.Fatal(err)
	}

	db, err = database.Open("quiltro.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	postID, err := db.NewPost(database.Post{
		Permalink: "test-post",
		Author: "lobo",
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.RevisePost(postID, database.Revision{
		Title: "A test post",
		Body: `# A test post
This is a damn test post.
All this post is for is to test.`,
		Keywords: []string{"test meta"},
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.RevisePost(postID, database.Revision{
		Title: "A test post",
		Body: `# A test post
This is a damn test post.
All this post is for is to test.

With a new revision, just for tests.`,
		Keywords: []string{"test", "meta"},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("listneing")
	log.Fatalln(http.ListenAndServe("127.0.0.1:8081", http.HandlerFunc(PostHandler)))
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	post, _, err := db.GetPostByPermalink(path, false)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	revs, err := db.GetRevisionsForPost(post.ID)

	postTmpl.Execute(w, struct{
		Post database.Post
		Revisions []database.Revision
	}{post, revs})
}

var postTmpl = template.Must(template.New("").Parse(postShow))
const postShow = `<!doctype html>
<h1>Post information:</h1>
<dl id="post-metadata">
<dt>ID:</dt>
<dd>{{.Post.ID}}</dd>
<dt>Author:</dt>
<dd>{{.Post.Author}}</dd>
<dt>Permalink:</dt>
<dd>{{.Post.Permalink}}</dd>
<dt>Status:</dt>
<dd>{{if .Post.Published.IsZero}}draft{{else}}published{{end}}</dd>
{{if not .Post.Published.IsZero}}
<dt>Publish date:</dt>
<dd>{{.Post.Published.String}}</dd>
{{end}}
</dl>
<h1>Revisions</h1>
<table>
  <tr>
    <th>ID</th>
    <th>Title</th>
    <th>Body</th>
    <th>Keywords</th>
    <th>Timestamp</th>
  </tr>
{{range .Revisions}}
  <tr>
    <td>{{.ID}}</td>
    <td>{{.Title}}</td>
    <td><pre>{{.Body}}</pre></td>
    <td>{{.Keywords}}</td>
    <td>{{.Timestamp.String}}</td>
  </tr>
{{end}}
</table>`

/*

import (
	"encoding/gob"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"quiltro/auth"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)


var sessionKey []byte
var store *sessions.CookieStore

func init() {
	if _, err := os.Stat("key"); err != nil {
		key := securecookie.GenerateRandomKey(32)
		file, err := os.Create("key")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		sessionKey = key
		file.Write(key)
		log.Println("wrote new session key")
	} else {
		keyFile, err := os.Open("key")
		if err != nil {
			panic(err)
		}
		if sessionKey, err = io.ReadAll(keyFile); err != nil {
			panic(err)
		}
	}

	store = sessions.NewCookieStore(sessionKey)

	gob.Register(auth.User{})
}

func main() {
	router := mux.NewRouter()
	CSRF := csrf.Protect(sessionKey, csrf.Secure(false))

	log.Println("listening")
	log.Fatalln(http.ListenAndServe("127.0.0.1:8081", CSRF(router)))
}

var loginFormTemplate = template.Must(
	template.New("adm/login").Parse(loginForm),
)
const loginForm = `<!DOCTYPE html>
<meta name=viewport content="width=device-width, initial-scale=1">
<meta charset=utf-8>
<style type=text/css>body { max-width: 800px; margin: auto; }</style>
<title>quiltro &ndash; adm/login</title>

<h1>adm/login</h1>
{{range .flashes}}
<blockquote>{{.}}</blockquote>
{{end}}

<form method=POST action="/adm/login">
{{.csrfField}}
<label for="username">username:</label>
<input type="text" name="username" required>
<br>
<label for="password">password:</label>
<input type="password" name="password" required>
<br>
<input type="submit" value="login">
</form>`
*/
