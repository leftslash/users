/*
package users provides user functionality for applications.

This file includes HTTP handlers to perform CRUD (create, read,
update, delete) functionality and is backed by a database
using SQLite.  It includes password reset functionality
and uses secure logic for storing passwords

*/
package users

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"net/http"
	"strconv"
	"time"

	"github.com/leftslash/mux"
	"github.com/leftslash/xerror"
)

// Handler is the primary structure for HTTP requests.
// It includes the Database so that users may be stored
// in a persistent manner.  Timeout is the length of time
// a temporary password (created during reset) is permitted
// to remain valid.
type Handler struct {
	db      *Database
	timeout float64
}

// NewHandler creates a new Handler with the database
// opened and ready for access operations.
func NewHandler(dbfile string) (h *Handler) {
	return &Handler{db: OpenDatabase(dbfile)}
}

// IsValid uses the database to determine whether password
// is valid for a specific email.
func (h *Handler) IsValid(email, password string) (ok bool) {
	return h.db.IsValid(email, password)
}

// GetAll retrieves all users and their relevant details from
// the database.  It returns these in JSON format.  If an error
// occurs, it is handled via HTTP logic.  This method assumes
// no inputs and returns *all* users.
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.db.GetAll()
	if err != nil {
		err.Handler(w)
		return
	}
	mux.WriteJSON(w, users)
}

// Get assumes a URL input parameter "id" and will search
// the database and retrieve only the user which has this id.
// It returns data in JSON format and handles errors via HTTP.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := h.db.Get(mux.URLParam(r, "id"))
	if err != nil {
		err.Handler(w)
		return
	}
	mux.WriteJSON(w, user)
}

// Delete assumes a URL input parameter "id" and will search
// the database and delete only the user with this specific id.
// It returns *no* data, only an HTTP status code signifying the result.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	err := h.db.Delete(mux.URLParam(r, "id"))
	if err != nil {
		err.Handler(w)
		return
	}
}

// Add assumes a POST or PUT HTTP method with data in the request
// body formatted as JSON.  It uses this data to create the new
// user in the database which also creates a new id for this user
// which is not returned.  Only the HTTP status code signifies
// the result to the caller.
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	u := User{}
	err := mux.ReadJSON(w, r, &u)
	if err != nil {
		e := xerror.Errorf(err, 0x4d96, "adding user")
		e.Handler(w)
		return
	}
	e := h.db.Add(u)
	if e != nil {
		e.Handler(w)
		return
	}
}

// Update assumes a POST or PUT HTTP method with data in the request
// body formatted as JSON.  It uses this data to update the
// user in the database.  Only the HTTP status code signifies
// the result to the caller.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	u := User{}
	err := mux.ReadJSON(w, r, &u)
	if err != nil {
		e := xerror.Errorf(err, 0x3b15, "updating user")
		e.Handler(w)
		return
	}
	e := h.db.Update(u)
	if e != nil {
		e.Handler(w)
		return
	}
}

// Forgot creates an HTTP handler that will be used for all
// requests for a temporary password if the user has forgotten
// their password.  The handler expects a username/email
// stored in the body of the request in HTML form format.
// This handler will get a temporary password, store it in the
// database and email the user a link to reset their password.
func (h *Handler) Forgot(timeout string) http.HandlerFunc {
	var err error
	h.timeout, err = strconv.ParseFloat(timeout, 64)
	if err != nil {
		h.timeout = 5.0
	}
	return func(w http.ResponseWriter, r *http.Request) {
		// retrieve username/email from post form
		// generate temporary expiring password
		// store password in database without hash
		//   h.db.SetTempPassword(username, password)
		// email url with temporary password
	}
}

func (h *Handler) SetupReset(w http.ResponseWriter, r *http.Request) {
	// retrieve temporary expiring password from url query
	// validate password has not expired
	// retrieve username and validate password
	//   user, err := h.db.GetUserByTempPassword(password)
	// set username in reset template
	// set temporary password in reset template
}

func (h *Handler) PerformReset(w http.ResponseWriter, r *http.Request) {
	// retrieve username, temporary expiring password and new password from post form
	// validate temporary password has not expired
	// retrieve username and validate password
	//   user, err := h.db.GetUserByTempPassword(password)
	// confirm form user == db user
	// copy new password to user
	// update user
	//   h.db.Update(user)
	// redirect to Login
}

// makeExpiringPassword generates a temporary password that contains
// cryptographically secure random bytes as well as the current time.
// This temporary password must be URL friendly since it will be
// sent to the user via email as part of a link they will use
// to reset their password to one of their chosing.
// The password is encoded using base64 and concatenated.  The encoded
// time is used later to determine whether this temporary password
// has expired (i.e. it is older than the current time less a timeout)
func makeExpiringPassword() (password string) {

	// create a slice of 16 cryptographically secure random bytes
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)

	// create a byte buffer, add current time, get as byte slice
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, time.Now().Unix())
	timeBytes := buf.Bytes()

	// encode and concatenate random and time bytes
	password = base64.RawURLEncoding.EncodeToString(randomBytes)
	password += base64.RawURLEncoding.EncodeToString(timeBytes)
	return
}

// isPasswordExpired is the analog to the makeExpiringPassword
// function.  This function extracts the time portion of the
// provided password string and determines whether the extracted
// time has exceeded the provided timeout given the current time.
func isPasswordExpired(password string, timeout float64) (ok bool) {

	// base64 decode the portion of the password containing the time.
	// the first 22 encoded bytes contain the 16 random bytes created
	// via makeExpiringPassword and should be disregarded
	timeBytes, err := base64.RawURLEncoding.DecodeString(password[22:])
	if err != nil {
		return
	}

	// convert the time bytes into a standard unix time number
	var unixTime int64
	err = binary.Read(bytes.NewBuffer(timeBytes), binary.BigEndian, &unixTime)
	if err != nil {
		return
	}

	// compare current unix time with decoded unix time.
	// determine if the timeout period has been exceeded.
	t := time.Unix(unixTime, 0)
	ok = time.Since(t).Minutes() <= timeout
	return
}
