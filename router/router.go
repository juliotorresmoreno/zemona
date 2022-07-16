package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/juliotorresmoreno/zemona/controllers"
	"github.com/juliotorresmoreno/zemona/middlewares"
)

func attach(prefix string, router *mux.Router, handler http.Handler, strip ...bool) *mux.Route {
	if len(strip) == 1 && !strip[0] {
		return router.PathPrefix(prefix).Handler(handler)
	}

	return router.PathPrefix(prefix).
		Handler(http.StripPrefix(prefix, handler))
}

func NewRouter() http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	router.Use(middlewares.LoggerMiddleware)

	attach("/twitter", router, controllers.NewTwitterController())
	attach("/", router, controllers.NewHomeController(), false)

	return router
}
