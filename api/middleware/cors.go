package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
)

// FUTURE: Setup with this -->  https://github.com/rs/cors

//SetupCORS - Eables CORS Middleware.  Presently usurping all Options requests (TODO:  Future Optional)
func SetupCORS(router *mux.Router) {
	//router.Use(corsMiddleware)
	router.Methods("OPTIONS").HandlerFunc(preflightHandler)
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	//If Options pre-flight request, handle...
	//log.Printf("in pre-flight, method: %v, acess req: %v", r.Method, r.Header.Get("Access-Control-Request-Method"))

	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		headers := w.Header()
		headers.Add("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")
		headers.Set("Access-Control-Allow-Origin", "*")
		headers.Set("Access-Control-Allow-Methods", "OPTIONS,POST,GET,PUT,DELETE")
		headers.Set("Access-Control-Allow-Credentials", "true")
		headers.Set("Access-Control-Allow-Headers", "Content-Type")
		//log.Printf("in pre-flight, returning")
		return
	}
}

//NOT IN USE RIGHT NOW, ONLY USE RIGHT NOW IS FOR PRE_FLIGHT, FUTURE FIX>...
/*func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//If Options pre-flight request, handle...
		log.Printf("in pre-flight, method: %v, acess req: %v", r.Method, r.Header.Get("Access-Control-Request-Method"))

		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			headers := w.Header()
			headers.Add("Vary", "Origin")
			headers.Add("Vary", "Access-Control-Request-Method")
			headers.Add("Vary", "Access-Control-Request-Headers")
			headers.Set("Access-Control-Allow-Origin", "*")
			headers.Set("Access-Control-Allow-Methods", "OPTIONS,POST,GET,PUT,DELETE")
			headers.Set("Access-Control-Allow-Credentials", "true")
			headers.Set("Access-Control-Allow-Headers", "Content-Type")
			log.Printf("in pre-flight, returning")
			return
		}
		//Go to the next handler in the chain only if not pre-flight
		log.Printf("in pre-flight, skipping")
		next.ServeHTTP(w, r)
	})
}
*/
