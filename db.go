package users

import (
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/leftslash/mux"

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

func (d Database) GetAll() (users []User, e mux.Error) {
	stmt := "SELECT id, email, name, country FROM users"
	rows, err := d.db.Query(stmt)
	if err != nil {
		e = mux.Errorf(err, 0, "retrieving users")
		e.Log()
		return
	}
	for rows.Next() {
		var u User
		err = rows.Scan(&u.Id, &u.Email, &u.Name, &u.Country)
		if err != nil {
			e = mux.Errorf(err, 0, "retrieving users")
			e.Log()
			return
		}
		users = append(users, u)
	}
	return
}

func (d Database) Get(id string) (u User, e mux.Error) {
	intId, err := strconv.Atoi(id)
	if err != nil {
		e = mux.Errorf(err, 0, "invalid user id %q", id)
		return
	}
	stmt := "SELECT id, email, name, country FROM users WHERE id = ?"
	err = d.db.QueryRow(stmt, intId).Scan(&u.Id, &u.Email, &u.Name, &u.Country)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			e = mux.Errorf(fmt.Errorf("no user with id %q", id), 0, "retrieving user")
			return
		}
		e = mux.Errorf(err, 0, "retrieving user")
		e.Log()
		return
	}
	return
}

func (d Database) Delete(id string) (e mux.Error) {
	intId, err := strconv.Atoi(id)
	if err != nil {
		e = mux.Errorf(err, 0, "invalid user id %q", id)
		return
	}
	stmt := "DELETE FROM users WHERE id = ?"
	result, err := d.db.Exec(stmt, intId)
	if err != nil {
		e = mux.Errorf(err, 0, "deleting user")
		e.Log()
		return
	}
	n, err := result.RowsAffected()
	if err != nil {
		e = mux.Errorf(err, 0, "deleting user")
		e.Log()
		return
	}
	if n != 1 {
		e = mux.Errorf(fmt.Errorf("no user with id %q", id), 0, "deleting user")
		return
	}
	return
}

func (d Database) Add(u User) (e mux.Error) {
	stmt := "INSERT INTO users (email, name, country, password) VALUES (?, ?, ?, ?)"
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), passwordHashCost)
	if err != nil {
		e = mux.Errorf(err, 0, "adding user")
		e.Log()
		return
	}
	result, err := d.db.Exec(stmt, u.Email, u.Name, u.Country, passwordCryptPrefix+string(hash))
	if err != nil {
		e = mux.Errorf(err, 0, "adding user")
		e.Log()
		return
	}
	n, err := result.RowsAffected()
	if err != nil {
		e = mux.Errorf(err, 0, "adding user")
		e.Log()
		return
	}
	if n != 1 {
		e = mux.Errorf(fmt.Errorf("no user added"), 0, "adding user")
		e.Log()
		return
	}
	return
}

func (d Database) Update(u User) (e mux.Error) {
	var result sql.Result
	if u.Password != "" {
		stmt := "UPDATE users SET email = ?, name = ?, country = ?, password = ? WHERE id = ?"
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), passwordHashCost)
		if err != nil {
			e = mux.Errorf(err, 0, "updating user")
			e.Log()
			return
		}
		result, err = d.db.Exec(stmt, u.Email, u.Name, u.Country, passwordCryptPrefix+string(hash), u.Id)
		if err != nil {
			e = mux.Errorf(err, 0, "updating user")
			e.Log()
			return
		}
	} else {
		stmt := "UPDATE users SET email = ?, name = ?, country = ? WHERE id = ?"
		var err error
		result, err = d.db.Exec(stmt, u.Email, u.Name, u.Country, u.Id)
		if err != nil {
			e = mux.Errorf(err, 0, "updating user")
			e.Log()
			return
		}
	}
	n, err := result.RowsAffected()
	if err != nil {
		e = mux.Errorf(err, 0, "updating user")
		e.Log()
		return
	}
	if n != 1 {
		e = mux.Errorf(fmt.Errorf("no user with id %q", u.Id), 0, "deleting user")
		e.Log()
		return
	}
	return
}

func (d Database) IsValid(email, password string) (ok bool) {
	stmt := "SELECT password FROM users WHERE email = ?"
	row := d.db.QueryRow(stmt, email)
	var hash string
	err := row.Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		mux.Errorf(err, 0, "validating user").Log()
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

func (d Database) SetTempPassword(email, password string) (e mux.Error) {
	stmt := "UPDATE users SET password = ? WHERE email = ?"
	result, err := d.db.Exec(stmt, passwordTempPrefix+password, email)
	if err != nil {
		e = mux.Errorf(err, 0, "setting temporary password")
		e.Log()
		return
	}
	n, err := result.RowsAffected()
	if err != nil {
		e = mux.Errorf(err, 0, "setting temporary password")
		e.Log()
		return
	}
	if n != 1 {
		e = mux.Errorf(errors.New("not found"), 0, "setting temporary password")
		return
	}
	return
}

func (d Database) GetUserByTempPassword(password string) (u User, e mux.Error) {
	stmt := "SELECT id, email, name, country FROM users WHERE password = ?"
	err := d.db.QueryRow(stmt, passwordTempPrefix+password).Scan(&u.Id, &u.Email, &u.Name, &u.Country)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			e = mux.Errorf(fmt.Errorf("no user"), 0, "retrieving user")
			return
		}
		e = mux.Errorf(err, 0, "retrieving users")
		e.Log()
		return
	}
	return
}
