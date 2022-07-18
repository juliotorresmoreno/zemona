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
	api := mux.NewRouter().StrictSlash(true)
	router := mux.NewRouter().StrictSlash(true)

	router.Use(middlewares.Cors)
	router.Use(middlewares.LoggerMiddleware)

	attach("/profile", api, controllers.NewProfileController())
	attach("/twitter", api, controllers.NewTwitterController())
	attach("/session", api, controllers.NewSessionController())
	attach("/", api, controllers.NewHomeController(), false)

	attach("/api/v1", router, api)

	return router
}
