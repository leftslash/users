package main

import (
	"net/http"

	"github.com/leftslash/config"
	"github.com/leftslash/mux"
	"github.com/leftslash/users"
)

func main() {

	// setup configuration and flags
	// allow command line flags to override file settings and edfaults
	conf := config.NewConfig()
	conf.Flag("env", "environment")
	conf.Flag("config", "config file")
	conf.Load()

	// retrieve config variables
	// note that the env variable drives selection of all other variables
	env := conf.Get("env")
	dbfile := conf.Get(env, "db.file")
	addr := conf.Get(env, "net.host") + ":" + conf.Get(env, "net.port")
	timeout := conf.Get(env, "password.reset.timeout")

	// setup router, handler and auth handler
	router := mux.NewRouter()
	users := users.NewHandler(dbfile)
	auth := mux.NewAuth("/login", "/login", users.IsValid)

	// establish logging for all handlers
	router.Use(mux.Logger(nil))

	// register User API REST handlers, all require authentication
	router.Handle(http.MethodGet, "/api/users", auth.HandleFunc(users.GetAll))
	router.Handle(http.MethodGet, "/api/users/{id}", auth.HandleFunc(users.Get))
	router.Handle(http.MethodPost, "/api/users", auth.HandleFunc(users.Add))
	router.Handle(http.MethodPut, "/api/users", auth.HandleFunc(users.Update))
	router.Handle(http.MethodDelete, "/api/users/{id}", auth.HandleFunc(users.Delete))

	// TODO: create static dir filesys *only* once not a zillion times

	// register public urls to fixed assets and login logic
	router.Handle(http.MethodGet, "/lib/*", http.FileServer(http.Dir("static")))
	router.Handle(http.MethodGet, "/favicon.ico", http.FileServer(http.Dir("static")))
	router.Handle(http.MethodGet, "/login", http.FileServer(http.Dir("static")))

	// register routes to handle forgotten passwords and reset logic
	// note that users.Forgot() requires a timeout, how long a temporary password is legit
	router.HandleFunc(http.MethodPost, "/forgot", users.Forgot(timeout))
	router.HandleFunc(http.MethodGet, "/reset", users.SetupReset)
	router.HandleFunc(http.MethodPost, "/reset", users.PerformReset)

	// register logout handler
	router.HandleFunc(http.MethodGet, "/logout", auth.Logout)

	// setup private / authenticated access to everything else on the server
	router.Handle(http.MethodGet, "/*", auth.Handle(http.FileServer(http.Dir("static"))))
	router.Handle(http.MethodPost, "/*", auth.Handle(http.FileServer(http.Dir("static"))))

	// start the web server
	router.Run(addr)
}
