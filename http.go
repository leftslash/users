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
)

type Handler struct {
	db      *Database
	timeout float64
}

func NewHandler(dbfile string) (h *Handler) {
	return &Handler{db: OpenDatabase(dbfile)}
}

func (h *Handler) IsValid(email, password string) (ok bool) {
	return h.db.IsValid(email, password)
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.db.GetAll()
	if err != nil {
		err.Handler(w)
		return
	}
	mux.WriteJSON(w, users)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := h.db.Get(mux.URLParam(r, "id"))
	if err != nil {
		err.Handler(w)
		return
	}
	mux.WriteJSON(w, user)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	err := h.db.Delete(mux.URLParam(r, "id"))
	if err != nil {
		err.Handler(w)
		return
	}
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	u := User{}
	err := mux.ReadJSON(w, r, &u)
	if err != nil {
		err.Handler(w)
		return
	}
	err = h.db.Add(u)
	if err != nil {
		err.Handler(w)
		return
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	u := User{}
	err := mux.ReadJSON(w, r, &u)
	if err != nil {
		err.Handler(w)
		return
	}
	err = h.db.Update(u)
	if err != nil {
		err.Handler(w)
		return
	}
}

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

func makeExpiringPassword() (password string) {
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, time.Now().Unix())
	timeBytes := buf.Bytes()
	password = base64.RawURLEncoding.EncodeToString(randomBytes)
	password += base64.RawURLEncoding.EncodeToString(timeBytes)
	return
}

func isPasswordExpired(password string, timeout float64) (ok bool) {
	timeBytes, err := base64.RawURLEncoding.DecodeString(password[22:])
	if err != nil {
		return
	}
	var unixTime int64
	err = binary.Read(bytes.NewBuffer(timeBytes), binary.BigEndian, &unixTime)
	if err != nil {
		return
	}
	t := time.Unix(unixTime, 0)
	ok = time.Since(t).Minutes() <= timeout
	return
}
