package users

import (
	"net/http"

	"github.com/leftslash/mux"
)

type Handler struct {
	db *Database
}

func NewHandler(dbfile string) (h *Handler) {
	return &Handler{db: OpenDatabase(dbfile)}
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.db.GetAll()
	if err != nil {
		err.Handler(w)
		return
	}
	mux.WriteJSON(w, users)
}

func (h *Handler) IsValid(email, password string) (ok bool) {
	var err *Error
	ok, err = h.db.IsValid(email, password)
	if err != nil {
		err.Log()
	}
	return
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
		e := Errorf(err, "adding user (0x243)")
		e.Handler(w)
		return
	}
	e := h.db.Add(u)
	if e != nil {
		e.Handler(w)
		return
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	u := User{}
	err := mux.ReadJSON(w, r, &u)
	if err != nil {
		e := Errorf(err, "updating user (0x243)")
		e.Handler(w)
		return
	}
	e := h.db.Update(u)
	if e != nil {
		e.Handler(w)
		return
	}
}
