package middleware

import (
	"github.com/gorilla/mux"
)

// SetUpMiddleware contains the middleware that applies to every request.
// Reference middlewares here --> https://github.com/urfave/negroni#logger and here --> https://github.com/go-chi/chi
// The Google example here should be used as an additional base element for the architecture:
// https://blog.golang.org/context
func SetUpMiddleware(router *mux.Router) {
	//Add each middleware needed as core setup here as they are ready
	//Future:  Placeholder middlewares in this package file names to be implemented
	//Future:  Add access.log equiv?   Any resource clearing needed here that would not be covered at project level?
	//leave to the middlewares to use environment variables to turn on/off/configure themsevles as needed vs. managing centrally

	//TODO:  Make Cors part of config options vs. always
	SetupCORS(router)
	SetupAuth(router)

}
