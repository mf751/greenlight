package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc(http.MethodGet+" /v1/healthcheck", app.healthCheckHandler)
	router.HandleFunc(http.MethodPost+" /v1/movies", app.createMovieHandler)
	router.HandleFunc(http.MethodGet+" /v1/movies/{id}", app.showMovieHandler)

	return router
}
