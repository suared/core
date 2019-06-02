package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
)

// FUTURE: Setup with this -->  https://github.com/rs/cors

//SetupCORS - Eables CORS Middleware.  Presently usurping all Options requests (TODO:  Future Optional)
func SetupCORS(router *mux.Router) {
	//For all other requests add in middleware
	router.Use(corsMiddleware)
	//For Options requests handle here - required, otherwise will not go through middleware if no handler
	router.Methods("OPTIONS").HandlerFunc(preflightHandler)
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	//If Options pre-flight request, handle...
	//If an Options request this is visited twice, once as the middleware for itself however the header doesn't dup and catching it for the odd scenarios seems not worth the effort right now
	//log.Printf("in pre-flight, method: %v, acess req: %v", r.Method, r.Header.Get("Access-Control-Request-Method"))

	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		headers := w.Header()
		headers.Add("Vary", "Origin")
		//headers.Add("Vary", "Access-Control-Request-Method")
		//headers.Add("Vary", "Access-Control-Request-Headers")
		//headers.Set("Vary", "Location")
		headers.Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		headers.Set("Access-Control-Allow-Methods", "OPTIONS,POST,GET,PUT,DELETE,PATCH,HEAD")
		headers.Set("Access-Control-Allow-Credentials", "true")
		headers.Set("Access-Control-Allow-Headers", "Authorization,Origin,Referer,User-Agent")
		//headers.Set("Access-Control-Allow-Headers", "Location,Authorization,Content-Length,Content-Type,Origin,Referer,User-Agent,Origin,Access-Control-Request-Headers,Access-Control-Request-Method")
		//headers.Set("Access-Control-Expose-Headers", "Location,Cache-Control,Content-Length,Content-Type,Date,ETag,Expires,Server,Access-Control-Allow-Origin,Access-Control-Allow-Methods,Access-Control-Allow-Credentials,Access-Control-Allow-Headers,Access-Control-Expose-Headers,Access-Control-Max-Age")
		headers.Set("Access-Control-Expose-Headers", "Location")
		headers.Set("Access-Control-Max-Age", "3600")
		//TODO:  Research these, move to separate middleware if appropriate
		//headers.Set("X-Content-Type-Options", "nosniff")
		//headers.Set("X-Frame-Options", "SAMEORIGIN")
		//headers.Set("X-XSS-Protection", "1; mode=block")

		//log.Printf("in pre-flight, returning")
		return
	} else {
		//Even for Get requests Chrome seems to block so adding the origin only here to work around
		headers := w.Header()
		headers.Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))

	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Add in CORS headers to all request
		preflightHandler(w, r)
		next.ServeHTTP(w, r)
	})
}
