package middleware

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/suared/core/security"

	//Base infra setup for package
	_ "github.com/suared/core/infra"
)

var isTest, isCognito bool

func init() {
	authStyle := os.Getenv("AUTH_STYLE")
	if authStyle == "test" {
		isTest = true
	}
}

//SetupAuth - Will enable authentication middleware if appropriate environment variables are set.
//For now, will always setup and will add future flags when needed to support on/off + diff authentication styles
//The assumption for all users as that this will setup the auth only, the determination of correct level of auth will be determined by the api functions themselves
func SetupAuth(router *mux.Router) {

	router.Use(authMiddleware)
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//TODO:  Not yet reading from real token value, add testers in integration test of 2 different users to ensure this is fuly tested
		// get current context from request
		ctx := r.Context()

		if isTest {
			ctx = security.SetupTestAuthFromContext(ctx, 1)
		} else if isCognito {
			//token := r.Header.Get("X-Session-Token")
			//translate values into user object
			// ctx = security.SetupAuthFromContext(ctx...)
		}
		// update request with new context
		r = r.WithContext(ctx)
		// call the next handler inthe chain - clean up of context done by libraries so unneeded to add cleanup here...
		next.ServeHTTP(w, r)
	})
}
