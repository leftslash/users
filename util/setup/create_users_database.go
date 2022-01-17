package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

const dbfile = "users.db"

func main() {

	db, err := sql.Open("sqlite", dbfile)
	if err != nil {
		fmt.Printf("cannot create db: %s: %s\n", dbfile, err)
		os.Exit(1)
	}

	stmt := `CREATE TABLE users (
	id INTEGER PRIMARY KEY ON CONFLICT ABORT AUTOINCREMENT, 
	email TEXT UNIQUE ON CONFLICT ABORT, 
	name TEXT, 
	country TEXT, 
	passhash TEXT
);`
	_, err = db.Exec(stmt)
	if err != nil {
		fmt.Printf("cannot create db: %s: %s\n", dbfile, err)
		os.Exit(1)
	}

}
