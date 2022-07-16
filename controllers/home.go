package controllers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type homeController struct{}

func NewHomeController() http.Handler {
	controller := &homeController{}
	router := mux.NewRouter()

	router.HandleFunc("/", controller.getHome).Methods("GET")

	return router
}

func (c *homeController) getHome(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	fmt.Fprint(w, "OK")
}
