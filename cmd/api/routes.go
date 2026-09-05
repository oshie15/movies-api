package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// Initialize a new httprouter router instance
	router := httprouter.New()

	// Convert the notFoundResponse() helper to a http.Handler using the
	// http.HandlerFunc() adapter, and then set it as the custom error handler
	// for 404 Not Found response.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// Likewise, convert the methodAllowedResponse() to a http.Handler and set
	// it as thr custom error handler for 405 Method Not Allowed reponse.
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)

	// Add the route for the PUT /v1/movies/:id endpoint.

	// Require a PATCH request, rather than PUT.
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)

	// Add the route for the DELETE /v1/movies/:id endpoint.
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)

	// Wrap the router with the panic recovery middleware.
	return app.recoverPanic(router)
}
