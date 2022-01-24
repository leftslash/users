package users

import (
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/leftslash/xerror"
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

type Database struct{ *sql.DB }

func OpenDatabase(dbfile string) (db *Database) {
	sqldb, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
		return
	}
	db = &Database{sqldb}
	return
}

func (db *Database) GetAll() (users []User, e xerror.Error) {
	stmt := "SELECT id, email, name, country FROM users"
	rows, err := db.Query(stmt)
	if err != nil {
		e = xerror.Errorf(err, 0xe049, "retrieving users")
		e.Log()
		return
	}
	for rows.Next() {
		var u User
		err = rows.Scan(&u.Id, &u.Email, &u.Name, &u.Country)
		if err != nil {
			e = xerror.Errorf(err, 0xe085, "retrieving users")
			e.Log()
			return
		}
		users = append(users, u)
	}
	return
}

func (db *Database) Get(id string) (u User, e xerror.Error) {
	intId, err := strconv.Atoi(id)
	if err != nil {
		e = xerror.Errorf(err, 0x1698, "invalid user id %q", id)
		return
	}
	stmt := "SELECT id, email, name, country FROM users WHERE id = ?"
	err = db.QueryRow(stmt, intId).Scan(&u.Id, &u.Email, &u.Name, &u.Country)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			e = xerror.Errorf(fmt.Errorf("no user with id %q", id), 0x731c, "retrieving user")
			return
		}
		e = xerror.Errorf(err, 0x8758, "retrieving user")
		e.Log()
		return
	}
	return
}

func (db *Database) Delete(id string) (e xerror.Error) {
	intId, err := strconv.Atoi(id)
	if err != nil {
		e = xerror.Errorf(err, 0xce9a, "invalid user id %q", id)
		return
	}
	stmt := "DELETE FROM users WHERE id = ?"
	result, err := db.Exec(stmt, intId)
	if err != nil {
		e = xerror.Errorf(err, 0xce4c, "deleting user")
		e.Log()
		return
	}
	n, err := result.RowsAffected()
	if err != nil {
		e = xerror.Errorf(err, 0xfd15, "deleting user")
		e.Log()
		return
	}
	if n != 1 {
		e = xerror.Errorf(fmt.Errorf("no user with id %q", id), 0xbe69, "deleting user")
		return
	}
	return
}

func (db *Database) Add(u User) (e xerror.Error) {
	stmt := "INSERT INTO users (email, name, country, password) VALUES (?, ?, ?, ?)"
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), passwordHashCost)
	if err != nil {
		e = xerror.Errorf(err, 0x90f9, "adding user")
		e.Log()
		return
	}
	result, err := db.Exec(stmt, u.Email, u.Name, u.Country, passwordCryptPrefix+string(hash))
	if err != nil {
		e = xerror.Errorf(err, 0x4f9e, "adding user")
		e.Log()
		return
	}
	n, err := result.RowsAffected()
	if err != nil {
		e = xerror.Errorf(err, 0xe822, "adding user")
		e.Log()
		return
	}
	if n != 1 {
		e = xerror.Errorf(fmt.Errorf("no user added"), 0xd838, "adding user")
		e.Log()
		return
	}
	return
}

func (db *Database) Update(u User) (e xerror.Error) {
	var result sql.Result
	if u.Password != "" {
		stmt := "UPDATE users SET email = ?, name = ?, country = ?, password = ? WHERE id = ?"
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), passwordHashCost)
		if err != nil {
			e = xerror.Errorf(err, 0xa7ea, "updating user")
			e.Log()
			return
		}
		result, err = db.Exec(stmt, u.Email, u.Name, u.Country, passwordCryptPrefix+string(hash), u.Id)
		if err != nil {
			e = xerror.Errorf(err, 0x54a5, "updating user")
			e.Log()
			return
		}
	} else {
		stmt := "UPDATE users SET email = ?, name = ?, country = ? WHERE id = ?"
		var err error
		result, err = db.Exec(stmt, u.Email, u.Name, u.Country, u.Id)
		if err != nil {
			e = xerror.Errorf(err, 0x37cd, "updating user")
			e.Log()
			return
		}
	}
	n, err := result.RowsAffected()
	if err != nil {
		e = xerror.Errorf(err, 0x2aa0, "updating user")
		e.Log()
		return
	}
	if n != 1 {
		e = xerror.Errorf(fmt.Errorf("no user with id %q", u.Id), 0x4c3a, "deleting user")
		e.Log()
		return
	}
	return
}

func (db *Database) IsValid(email, password string) (ok bool) {
	stmt := "SELECT password FROM users WHERE email = ?"
	row := db.QueryRow(stmt, email)
	var hash string
	err := row.Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		xerror.Errorf(err, 0x780f, "validating user").Log()
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

func (db *Database) SetTempPassword(email, password string) (e xerror.Error) {
	stmt := "UPDATE users SET password = ? WHERE email = ?"
	result, err := db.Exec(stmt, passwordTempPrefix+password, email)
	if err != nil {
		e = xerror.Errorf(err, 0xbb61, "setting temporary password")
		e.Log()
		return
	}
	n, err := result.RowsAffected()
	if err != nil {
		e = xerror.Errorf(err, 0xc9f9, "setting temporary password")
		e.Log()
		return
	}
	if n != 1 {
		e = xerror.Errorf(errors.New("not found"), 0x399e, "setting temporary password")
		return
	}
	return
}

func (db *Database) GetUserByTempPassword(password string) (u User, e xerror.Error) {
	stmt := "SELECT id, email, name, country FROM users WHERE password = ?"
	err := db.QueryRow(stmt, passwordTempPrefix+password).Scan(&u.Id, &u.Email, &u.Name, &u.Country)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			e = xerror.Errorf(fmt.Errorf("no user"), 0x9e17, "retrieving user")
			return
		}
		e = xerror.Errorf(err, 0x40f6, "retrieving users")
		e.Log()
		return
	}
	return
}
