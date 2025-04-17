package main

import "net/http"

// Handler functions for each http route
type routesMap map[string]func(r *http.Request)

var routes = routesMap{
	"/": func(r *http.Request) {

	},
	"/login": func(r *http.Request) {

	},
	"/error": func(r *http.Request) {

	},
}
