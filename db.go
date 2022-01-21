package users

import (
	"crypto/subtle"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const (
	passwordHashCost    = 10
	passwordCryptPrefix = "crypt:"
	passwordTempPrefix  = "temp:"
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
	stmt := "INSERT INTO users (email, name, country, password) VALUES (?, ?, ?, ?)"
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), passwordHashCost)
	if err != nil {
		e = Errorf(err, "adding user (0x243)")
		return
	}
	result, err := d.db.Exec(stmt, u.Email, u.Name, u.Country, passwordCryptPrefix+string(hash))
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
		stmt := "UPDATE users SET email = ?, name = ?, country = ?, password = ? WHERE id = ?"
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), passwordHashCost)
		if err != nil {
			e = Errorf(err, "adding user (0x243)")
			return
		}
		result, err = d.db.Exec(stmt, u.Email, u.Name, u.Country, passwordCryptPrefix+string(hash), u.Id)
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
	stmt := "SELECT password FROM users WHERE email = ?"
	row := d.db.QueryRow(stmt, email)
	var hash string
	err := row.Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		e = Errorf(err, "validating user (0x242)")
		return
	}
	if strings.HasPrefix(hash, passwordCryptPrefix) {
		if len(hash) <= len(passwordCryptPrefix) {
			return
		}
		hash = hash[len(passwordCryptPrefix):]
		ok = nil == bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		return
	}
	if strings.HasPrefix(hash, passwordTempPrefix) {
		if len(hash) <= len(passwordTempPrefix) {
			return
		}
		hash = hash[len(passwordTempPrefix):]
		ok = 1 == subtle.ConstantTimeCompare([]byte(hash), []byte(password))
		return
	}
	return
}

func (d Database) SetTempPassword(email, password string) (e *Error) {
	stmt := "UPDATE users SET password = ? WHERE email = ?"
	result, err := d.db.Exec(stmt, passwordTempPrefix+password, email)
	if err != nil {
		e = Errorf(err, "setting temporary password (0x243)")
		return
	}
	n, err := result.RowsAffected()
	if err != nil {
		e = Errorf(err, "setting temporary password (0x243)")
		return
	}
	if n != 1 {
		e = Errorf(errors.New("not found"), "setting temporary password (0x243)")
		return
	}
	return
}

func (d Database) GetUserByTempPassword(password string) (u User, e *Error) {
	stmt := "SELECT id, email, name, country FROM users WHERE password = ?"
	err := d.db.QueryRow(stmt, passwordTempPrefix+password).Scan(&u.Id, &u.Email, &u.Name, &u.Country)
	if err != nil {
		e = Errorf(err, "retrieving user (0x243)")
		return
	}
	return
}
