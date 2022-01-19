package main

import (
	"net/http"

	"github.com/leftslash/config"
	"github.com/leftslash/mux"
	"github.com/leftslash/users"
)

func main() {

	conf := config.NewConfig()
	conf.Flag("env", "environment")
	conf.Load()

	env := conf.Get("env")
	dbfile := conf.Get(env, "db.file")
	addr := conf.Get(env, "net.host") + ":" + conf.Get(env, "net.port")
	timeout := conf.Get(env, "password.reset.timeout")

	router := mux.NewRouter()
	users := users.NewHandler(dbfile)
	auth := mux.NewAuth("/login/", "/login/", users.IsValid)

	router.Use(mux.Logger(nil))

	router.Handle(http.MethodGet, "/api/users", auth.HandleFunc(users.GetAll))
	router.Handle(http.MethodGet, "/api/users/{id}", auth.HandleFunc(users.Get))
	router.Handle(http.MethodPost, "/api/users", auth.HandleFunc(users.Add))
	router.Handle(http.MethodPut, "/api/users", auth.HandleFunc(users.Update))
	router.Handle(http.MethodDelete, "/api/users/{id}", auth.HandleFunc(users.Delete))

	router.Handle(http.MethodGet, "/lib/*", http.FileServer(http.Dir("static")))
	router.Handle(http.MethodGet, "/favicon.ico", http.FileServer(http.Dir("static")))
	router.Handle(http.MethodGet, "/login/", http.FileServer(http.Dir("static")))

	router.HandleFunc(http.MethodPost, "/forgot/", users.Forgot(timeout))
	router.HandleFunc(http.MethodGet, "/reset/", users.SetupReset)
	router.HandleFunc(http.MethodPost, "/reset/", users.PerformReset)

	router.HandleFunc(http.MethodGet, "/logout/", auth.Logout)

	router.Handle(http.MethodGet, "/*", auth.Handle(http.FileServer(http.Dir("static"))))
	router.Handle(http.MethodPost, "/*", auth.Handle(http.FileServer(http.Dir("static"))))

	router.Run(addr)
}
