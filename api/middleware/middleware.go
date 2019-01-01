package middleware

import (
	"github.com/gorilla/mux"
)

// SetUpMiddleware contains the middleware that applies to every request.
// Reference middlewares here --> https://github.com/urfave/negroni#logger
func SetUpMiddleware(router *mux.Router) {
	// Add each middleware needed as core setup here as they are ready

	//Future:  Placeholder middlewares in this package file names to be implemented
	//Future:  Add access.log equiv?   Any resource clearing needed here that would not be covered at project level?

	// Do nothing for now
	return router
}
