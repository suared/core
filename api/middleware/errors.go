package middleware

import (
	"net/http"

	"github.com/gorilla/mux"

	coreerrors "github.com/suared/core/errors"
	//Base infra setup for package
	_ "github.com/suared/core/infra"
)

//SetupErrorHandler - Will setup last resort error handler
//For now, will always setup and will add future flags when needed to support on/off
//For now I am not writing a speific http response here but that may be the right thing to do eventually.  Preference is these are all exception to rule so no need
//The assumption is this is the last middleware to be called to enable the most available context to be available for error logging separately
func SetupErrorHandler(router *mux.Router) {
	router.Use(errorMiddleware)
}

//Intended to be called after setting up auth and other middleware to enable user id capture
func errorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		//setup recover call
		defer coreerrors.LastResortHandlerWithContext(ctx, "ErrorMiddleware")

		//call next
		next.ServeHTTP(w, r)
	})
}
