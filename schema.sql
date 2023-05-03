CREATE TABLE config(key TEXT, value TEXT);
CREATE TABLE user (
       id INTEGER PRIMARY KEY,
       username TEXT NOT NULL,
       hash TEXT NOT NULL
);
CREATE TABLE session (
       id INTEGER PRIMARY KEY,
       user_id INTEGER NOT NULL,
       hash TEXT NOT NULL,
       expiry DATE NOT NULL,
       FOREIGN KEY(user_id) REFERENCES user(id)
);
CREATE TABLE post (
       id INTEGER PRIMARY KEY,
       revision_count INTEGER NOT NULL,
       author TEXT NOT NULL,
       permalink TEXT NOT NULL,
       published_at DATE
);
CREATE TABLE revision (
       id INTEGER PRIMARY KEY,
       post_id INTEGER NOT NULL,
       revision_index INTEGER NOT NULL,
       title TEXT NOT NULL,
       body TEXT NOT NULL,
       created_at DATE NOT NULL,

       FOREIGN KEY(post_id) REFERENCES post(id)
);
CREATE VIRTUAL TABLE revision_fts USING fts5(
       title,
       body,
       content='revisions',
       content_rowid='revid',
       tokenize='porter unicode61 remove_diacritics 2'
);
CREATE TABLE file (
       id INTEGER PRIMARY KEY,
       filename TEXT NOT NULL,
       uploaded_at DATE NOT NULL,
       data BLOB
);
CREATE TABLE image (
       id INTEGER PRIMARY KEY,
       filename TEXT NOT NULL,
       uploaded_at DATE NOT NULL,
       data BLOB,
       x INTEGER NOT NULL,
       y INTEGER NOT NULL
);

CREATE TRIGGER revision_ai AFTER INSERT ON revision BEGIN
       INSERT INTO revision_fts(rowid, title, body)
       	      VALUES(new.id, new.title, new.body);
END;
CREATE TRIGGER revision_ad AFTER DELETE ON revision BEGIN
       INSERT INTO revisionfts(fts_idx, rowid, title, body)
       	      VALUES('delete', old.id, old.title, old.body);
END;
CREATE TRIGGER revision_au AFTER UPDATE ON revision BEGIN
       INSERT INTO revisionfts(fts_idx, rowid, title, body)
       	      VALUES('delete', old.id, old.title, old.body);
       INSERT INTO revisionfts(rowid, title, body)
       	      VALUES (new.id, new.title, new.body);
END;

CREATE UNIQUE INDEX idxuser_username ON user(username);
CREATE UNIQUE INDEX idxpost_permalink ON post(permalink);
CREATE UNIQUE INDEX idxfile_filename ON file(filename);
CREATE UNIQUE INDEX idximage_filename ON image(filename);
CREATE INDEX idxsession_user_id ON session(user_id);
CREATE INDEX idxsession_hash ON session(hash);
