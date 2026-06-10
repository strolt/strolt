package e2e_test

import (
	"bytes"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

// sqlRoundTrip backs up the given task, drops the schema and restores it
// from the latest snapshot. Data validity is asserted by the suite hooks.
func sqlRoundTrip(s *suite.Suite, c *Conn, serviceName, taskName, destinationName string) {
	s.T().Helper()

	s.Require().NoError(strolt("backup", "--service", serviceName, "--task", taskName, "--y"))

	c.dropSchema()

	latestSnapshotID, err := stroltGetLatestSnapshotID(serviceName, taskName, destinationName)
	s.Require().NoError(err)

	s.NoError(strolt("restore", "--service", serviceName, "--task", taskName, "--destination", destinationName, "--snapshot", latestSnapshotID, "--y"))
}

type User struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
}

type Post struct {
	ID      int    `db:"id"`
	UserID  int    `db:"user_id"`
	Title   string `db:"title"`
	Payload []byte `db:"payload"`
}

type UserPost struct {
	Username string `db:"username"`
	Title    string `db:"title"`
}

// Fixture data deliberately exercises what naive dump/restore breaks first:
// non-ASCII text, emoji, quotes and raw binary content.
var (
	user = User{
		ID:       1,
		Username: "üñîçødé-пользователь-🚀",
		Password: "p@ssw0rd",
	}

	post = Post{
		ID:      1,
		UserID:  1,
		Title:   `пост №1 — emoji 🎉, quotes 'single' "double"`,
		Payload: binaryPayload(),
	}
)

func binaryPayload() []byte {
	payload := make([]byte, 256)
	for i := range payload {
		payload[i] = byte(i)
	}

	return payload
}

type Conn struct {
	db      *sqlx.DB
	isMysql bool
}

func sqlConnect(driverName string, dataSourceName string) (*Conn, error) {
	db, err := sqlx.Connect(driverName, dataSourceName)

	return &Conn{
		db:      db,
		isMysql: driverName == "mysql",
	}, err
}

// createSchema creates two tables linked by a foreign key, an index and a
// view, so restores are verified against schema objects, not just rows.
func (c *Conn) createSchema() {
	if c.isMysql {
		c.db.MustExec(`
		CREATE TABLE users (
				id INT AUTO_INCREMENT PRIMARY KEY,
				username VARCHAR(190) NOT NULL UNIQUE,
				password TEXT NOT NULL
		);
		`)
		c.db.MustExec(`
		CREATE TABLE posts (
				id INT AUTO_INCREMENT PRIMARY KEY,
				user_id INT NOT NULL,
				title TEXT NOT NULL,
				payload BLOB NOT NULL,
				CONSTRAINT fk_posts_user FOREIGN KEY (user_id) REFERENCES users (id)
		);
		`)
	} else {
		c.db.MustExec(`
		CREATE TABLE users (
				id SERIAL PRIMARY KEY,
				username TEXT NOT NULL UNIQUE,
				password TEXT NOT NULL
		);
		`)
		c.db.MustExec(`
		CREATE TABLE posts (
				id SERIAL PRIMARY KEY,
				user_id INTEGER NOT NULL REFERENCES users (id),
				title TEXT NOT NULL,
				payload BYTEA NOT NULL
		);
		`)
		c.db.MustExec(`CREATE INDEX idx_posts_user_id ON posts (user_id);`)
	}

	c.db.MustExec(`CREATE VIEW user_posts AS SELECT u.username AS username, p.title AS title FROM users u JOIN posts p ON p.user_id = u.id;`)
}

// insertData inserts rows WITHOUT explicit ids so the auto-increment state is
// part of what backup/restore must preserve.
func (c *Conn) insertData() {
	if c.isMysql {
		c.db.MustExec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, user.Password)
		c.db.MustExec("INSERT INTO posts (user_id, title, payload) VALUES (?, ?, ?)", post.UserID, post.Title, post.Payload)
	} else {
		c.db.MustExec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, user.Password)
		c.db.MustExec("INSERT INTO posts (user_id, title, payload) VALUES ($1, $2, $3)", post.UserID, post.Title, post.Payload)
	}
}

func (c *Conn) dropSchema() {
	c.db.MustExec(`DROP VIEW IF EXISTS user_posts;`)
	c.db.MustExec(`DROP TABLE IF EXISTS posts;`)
	c.db.MustExec(`DROP TABLE IF EXISTS users;`)
}

func (c *Conn) checkValidData() error {
	if err := c.checkUsers(); err != nil {
		return err
	}

	if err := c.checkPosts(); err != nil {
		return err
	}

	if err := c.checkView(); err != nil {
		return err
	}

	return c.checkAutoIncrementUsable()
}

func (c *Conn) checkUsers() error {
	users := []User{}

	if err := c.db.Select(&users, "SELECT id, username, password FROM users"); err != nil {
		return err
	}

	if len(users) != 1 {
		return fmt.Errorf("expected 1 user, got %d", len(users))
	}

	if users[0] != user {
		return fmt.Errorf("user mismatch: want %+v, got %+v", user, users[0])
	}

	return nil
}

func (c *Conn) checkPosts() error {
	posts := []Post{}

	if err := c.db.Select(&posts, "SELECT id, user_id, title, payload FROM posts"); err != nil {
		return err
	}

	if len(posts) != 1 {
		return fmt.Errorf("expected 1 post, got %d", len(posts))
	}

	if posts[0].ID != post.ID || posts[0].UserID != post.UserID || posts[0].Title != post.Title {
		return fmt.Errorf("post mismatch: want %+v, got %+v", post, posts[0])
	}

	if !bytes.Equal(posts[0].Payload, post.Payload) {
		return errors.New("post binary payload corrupted by backup/restore")
	}

	return nil
}

func (c *Conn) checkView() error {
	rows := []UserPost{}

	if err := c.db.Select(&rows, "SELECT username, title FROM user_posts"); err != nil {
		return fmt.Errorf("view user_posts not usable: %w", err)
	}

	if len(rows) != 1 || rows[0].Username != user.Username || rows[0].Title != post.Title {
		return fmt.Errorf("view user_posts returned unexpected rows: %+v", rows)
	}

	return nil
}

// checkAutoIncrementUsable verifies the sequence / auto-increment counter
// survived the restore: inserting without an explicit id must not collide
// with existing rows. The probe row is rolled back; the counter advance is
// non-transactional in both engines, which is fine for the fixture.
func (c *Conn) checkAutoIncrementUsable() error {
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}

	query := "INSERT INTO users (username, password) VALUES ($1, $2)"
	if c.isMysql {
		query = "INSERT INTO users (username, password) VALUES (?, ?)"
	}

	if _, err := tx.Exec(query, "auto-increment-probe", "x"); err != nil {
		_ = tx.Rollback()

		return fmt.Errorf("auto-increment state not restored: %w", err)
	}

	return tx.Rollback()
}
