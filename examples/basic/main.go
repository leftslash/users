package main

import (
	"net/http"

	"github.com/leftslash/config"
	"github.com/leftslash/mux"
	"github.com/leftslash/users"
)

func makeAuthFunc(a mux.Middleware) func(h http.HandlerFunc) http.Handler {
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
	auth := mux.Auth(mux.AuthOptions{Validator: users.IsValid, FailURL: "/login"})
	authFunc := makeAuthFunc(auth)

	router.Use(mux.Logger(nil))

	router.Handle(http.MethodGet, "/api/users", authFunc(users.GetAll))
	router.Handle(http.MethodGet, "/api/users/{id}", authFunc(users.Get))
	router.Handle(http.MethodPost, "/api/users", authFunc(users.Add))
	router.Handle(http.MethodPut, "/api/users", authFunc(users.Update))
	router.Handle(http.MethodDelete, "/api/users/{id}", authFunc(users.Delete))
	router.Handle(http.MethodGet, "/*", auth(http.FileServer(http.Dir("."))))

	router.HandleFunc(http.MethodGet, "/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "login.html")
	})

	router.Run(addr)
}
