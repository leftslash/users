package users

import (
	"database/sql"
	"errors"
	"log"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 10
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Country  string `json:"country"`
	Password string `json:"password"`
}

type Database struct {
	db *sql.DB
}

func OpenDatabase(dbfile string) (d *Database) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
		return
	}
	d = &Database{db: db}
	return
}

func (d Database) GetAll() (users []User, e *Error) {
	var u User
	stmt := "SELECT id, email, name, country FROM users"
	rows, err := d.db.Query(stmt)
	if err != nil {
		e = Errorf(err, "retrieving users (0x242)")
		return
	}
	for rows.Next() {
		err = rows.Scan(&u.Id, &u.Email, &u.Name, &u.Country)
		if err != nil {
			e = Errorf(err, "retrieving users (0x243)")
			return
		}
		users = append(users, u)
	}
	return
}

func (d Database) Get(id string) (u User, e *Error) {
	intId, err := strconv.Atoi(id)
	if err != nil {
		e = Errorf(err, "invalid user id (0x242)")
		return
	}
	stmt := "SELECT id, email, name, country FROM users WHERE id = ?"
	err = d.db.QueryRow(stmt, intId).Scan(&u.Id, &u.Email, &u.Name, &u.Country)
	if err != nil {
		e = Errorf(err, "retrieving user (0x243)")
		return
	}
	return
}

func (d Database) Delete(id string) (e *Error) {
	intId, err := strconv.Atoi(id)
	if err != nil {
		e = Errorf(err, "invalid user id (0x242)")
		return
	}
	stmt := "DELETE FROM users WHERE id = ?"
	result, err := d.db.Exec(stmt, intId)
	if err != nil {
		e = Errorf(err, "deleting user (0x243)")
		return
	}
	n, err := result.RowsAffected()
	if err != nil {
		e = Errorf(err, "deleting user (0x243)")
		return
	}
	if n != 1 {
		e = Errorf(errors.New("not found"), "deleting user (0x243)")
		return
	}
	return
}

func (d Database) Add(u User) (e *Error) {
	stmt := "INSERT INTO users (email, name, country, passhash) VALUES (?, ?, ?, ?)"
	passhash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcryptCost)
	if err != nil {
		e = Errorf(err, "adding user (0x243)")
		return
	}
	result, err := d.db.Exec(stmt, u.Email, u.Name, u.Country, passhash)
	if err != nil {
		e = Errorf(err, "adding user (0x243)")
		return
	}
	n, err := result.RowsAffected()
	if err != nil {
		e = Errorf(err, "adding user (0x243)")
		return
	}
	if n != 1 {
		e = Errorf(errors.New("no user added"), "adding user (0x243)")
		return
	}
	return
}

func (d Database) Update(u User) (e *Error) {
	var result sql.Result
	var err error
	if u.Password != "" {
		stmt := "UPDATE users SET email = ?, name = ?, country = ?, passhash = ? WHERE id = ?"
		passhash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcryptCost)
		if err != nil {
			e = Errorf(err, "adding user (0x243)")
			return
		}
		result, err = d.db.Exec(stmt, u.Email, u.Name, u.Country, passhash, u.Id)
		if err != nil {
			e = Errorf(err, "updating user (0x243)")
			return
		}
	} else {
		stmt := "UPDATE users SET email = ?, name = ?, country = ? WHERE id = ?"
		result, err = d.db.Exec(stmt, u.Email, u.Name, u.Country, u.Id)
		if err != nil {
			e = Errorf(err, "updating user (0x243)")
			return
		}
	}
	n, err := result.RowsAffected()
	if err != nil {
		e = Errorf(err, "updating user (0x243)")
		return
	}
	if n != 1 {
		e = Errorf(errors.New("no user updated"), "updating user (0x243)")
		return
	}
	return
}

func (d Database) IsValid(email, password string) (ok bool, e *Error) {
	stmt := "SELECT passhash FROM users WHERE email = ?"
	row := d.db.QueryRow(stmt, email)
	var passhash string
	err := row.Scan(&passhash)
	if err != nil {
		e = Errorf(err, "validating user (0x242)")
		return
	}
	ok = nil == bcrypt.CompareHashAndPassword([]byte(passhash), []byte(password))
	return
}

func (d Database) SetTempPassword(email, password string) (e *Error) {
	// stmt := "UPDATE users SET passhash = ? WHERE email = ?"
	return
}

func (d Database) GetUserByTempPassword(password string) (u User, e *Error) {
	// stmt := "SELECT id, email, name, country FROM users WHERE passhash = ?"
	return
}
