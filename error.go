package users

import (
	"fmt"
	"log"
	"net/http"
)

type Error struct {
	Internal error
	External error
}

func Errorf(i error, format string, a ...interface{}) *Error {
	format = "error: " + format
	return &Error{
		Internal: i,
		External: fmt.Errorf(format, a...),
	}
}

func (e *Error) Log() {
	log.Printf("error: %s", e.Internal.Error())
}

func (e *Error) Handler(w http.ResponseWriter) {
	log.Printf("%s: %s", e.External.Error(), e.Internal.Error())
	http.Error(w, e.External.Error(), http.StatusInternalServerError)
}
