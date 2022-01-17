package main

import (
	"fmt"
	"net/http"

	"github.com/leftslash/config"
	"github.com/leftslash/mux"
	"github.com/leftslash/users"
)

func main() {

	conf := config.NewConfig()
	conf.Load()

	env := conf.Get("env")
	dbfile := conf.Get(env, "db.file")
	host := conf.Get(env, "net.host")
	port := conf.GetInt(env, "net.port")

	r := mux.NewRouter()
	r.Use(mux.Logger(nil))
	h := users.NewHandler(dbfile)
	r.HandleFunc(http.MethodGet, "/users", h.GetAll)
	r.HandleFunc(http.MethodGet, "/users/{id}", h.Get)
	r.HandleFunc(http.MethodPost, "/users", h.Add)
	r.HandleFunc(http.MethodPut, "/users", h.Update)
	r.HandleFunc(http.MethodDelete, "/users/{id}", h.Delete)
	r.Run(fmt.Sprintf("%s:%d", host, port))
}
