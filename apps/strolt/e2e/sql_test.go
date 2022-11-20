package e2e_test

import (
	"errors"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type User struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
}

var (
	user = User{
		ID:       1,
		Username: "username",
		Password: "password",
	}
)

type Conn struct {
	db *sqlx.DB
}

func sqlConnect(driverName string, dataSourceName string) (*Conn, error) {
	db, err := sqlx.Connect(driverName, dataSourceName)

	return &Conn{
		db,
	}, err
}

func (c *Conn) insertData(isMysql bool) {
	query := "INSERT INTO users (id, username, password) VALUES ($1, $2, $3)"
	if isMysql {
		query = "INSERT INTO users (id, username, password) VALUES (?, ?, ?)"
	}

	c.db.MustExec(query, user.ID, user.Username, user.Password)
}

func (c *Conn) createTable() {
	c.db.MustExec(`
	CREATE TABLE users (
			id integer,
			username text,
			password text
	);
	`)
}

func (c *Conn) dropTable() {
	c.db.MustExec(`DROP TABLE IF EXISTS users;`)
}

func (c *Conn) checkValidData() error {
	users := []User{}

	if err := c.db.Select(&users, "SELECT * FROM users"); err != nil {
		return err
	}

	if len(users) == 0 {
		return errors.New("not found records")
	}

	if len(users) != 1 {
		return errors.New("count records > 1")
	}

	if users[0].Username != user.Username || users[0].Password != user.Password || users[0].ID != user.ID {
		return errors.New("record not match with mock")
	}

	return nil
}
