package middleware

// TODO / FUTURE:  Placeholder - move this to use auth0 below: 
// Use with this instead --> https://github.com/auth0/go-jwt-middleware

//SAMPLE code from MUX on auth middleware setup
// Define our struct
type authenticationMiddleware struct {
	tokenUsers map[string]string
}

// Initialize it somewhere
func (amw *authenticationMiddleware) Populate() {
	amw.tokenUsers["00000000"] = "user0"
	amw.tokenUsers["aaaaaaaa"] = "userA"
	amw.tokenUsers["05f717e5"] = "randomUser"
	amw.tokenUsers["deadbeef"] = "user0"
}

// Middleware function, which will be called for each request
func (amw *authenticationMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("X-Session-Token")

        if user, found := amw.tokenUsers[token]; found {
        	// We found the token in our map
        	log.Printf("Authenticated user %s\n", user)
        	// Pass down the request to the next middleware (or final handler)
        	next.ServeHTTP(w, r)
        } else {
        	// Write an error and stop the handler chain
        	http.Error(w, "Forbidden", http.StatusForbidden)
        }
    })
}
r := mux.NewRouter()
r.HandleFunc("/", handler)

amw := authenticationMiddleware{}
amw.Populate()

r.Use(amw.Middleware)