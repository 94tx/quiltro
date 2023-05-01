CREATE TABLE IF NOT EXISTS config (key TEXT, value TEXT);

CREATE TABLE IF NOT EXISTS users (
       userid INTEGER PRIMARY KEY,
       username TEXT NOT NULL,
       hash TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idxusers_username ON users(username);

CREATE TABLE IF NOT EXISTS sessions (
       sessionid INTEGER PRIMARY KEY,
       userid INTEGER NOT NULL,
       hash TEXT NOT NULL,
       expiry DATE NOT NULL,
       FOREIGN KEY(userid) REFERENCES users(userid)
);
CREATE INDEX IF NOT EXISTS idxsessions_userid ON sessions(userid);
CREATE INDEX IF NOT EXISTS idxsessions_hash ON sessions(hash);

CREATE TABLE IF NOT EXISTS posts (
       postid INTEGER PRIMARY KEY,
       author TEXT NOT NULL,
       permalink TEXT NOT NULL,
       published_at DATE
);
CREATE INDEX IF NOT EXISTS idxposts_permalink on posts(permalink);

CREATE TABLE IF NOT EXISTS revisions (
       revid INTEGER PRIMARY KEY,
       postid INTEGER NOT NULL,
       revised_at DATE,
       FOREIGN KEY(postid) REFERENCES posts(postid)
);
CREATE VIRTUAL TABLE IF NOT EXISTS revisions_fts USING fts5 (
       title,
       body,
       keywords
);

CREATE TABLE IF NOT EXISTS images (
       fileid INTEGER PRIMARY KEY,
       filename TEXT NOT NULL,
       uploaded_at DATE,
       x INTEGER,
       y INTEGER,
       data BLOB
);
CREATE UNIQUE INDEX IF NOT EXISTS idximages_filename ON images(filename);

CREATE TABLE IF NOT EXISTS files (
       fileid INTEGER PRIMARY KEY,
       filename TEXT NOT NULL,
       uploaded_at DATE,
       data BLOB
);
CREATE UNIQUE INDEX IF NOT EXISTS idxfiles_filename ON files(filename);
