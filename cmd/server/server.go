package main

import (
	"net/http"

	"github.com/leftslash/config"
	"github.com/leftslash/mux"
	"github.com/leftslash/users"
)

func makeAuth(a mux.Middleware) func(h http.HandlerFunc) http.Handler {
	return func(h http.HandlerFunc) http.Handler {
		return a(h)
	}
}

func main() {

	conf := config.NewConfig()
	conf.Flag("env", "environment")
	conf.Load()

	env := conf.Get("env")
	dbfile := conf.Get(env, "db.file")
	addr := conf.Get(env, "net.host") + ":" + conf.Get(env, "net.port")

	router := mux.NewRouter()
	users := users.NewHandler(dbfile)
	auth := makeAuth(mux.Auth(mux.AuthOptions{Validator: users.IsValid, FailURL: "/login"}))

	router.Use(mux.Logger(nil))

	router.Handle(http.MethodGet, "/api/users", auth(users.GetAll))
	router.Handle(http.MethodGet, "/api/users/{id}", auth(users.Get))
	router.Handle(http.MethodPost, "/api/users", auth(users.Add))
	router.Handle(http.MethodPut, "/api/users", auth(users.Update))
	router.Handle(http.MethodDelete, "/api/users/{id}", auth(users.Delete))
	router.Handle(http.MethodGet, "/*", http.FileServer(http.Dir("static")))

	router.Run(addr)
}
