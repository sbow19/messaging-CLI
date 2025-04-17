package main

import (
	"context"
	"net/http"
)

// Parse incoming http requests and determine if they are valid
func parseIncomingReq(ctx context.Context, res chan<- ClientResponse, r *http.Request) {

	// path := r.URL.Path
	// method := r.Methods

	if !authenticationCycle(ctx, res, r) {
		return
	}

}
