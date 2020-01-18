package middleware

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/suared/core/errors"
	"github.com/suared/core/security"

	//Base infra setup for package
	_ "github.com/suared/core/infra"
)

var isTest, isCognito bool

func init() {
	authStyle := os.Getenv("AUTH_STYLE")
	if authStyle == "test" {
		isTest = true
	} else if authStyle == "cognito" {
		isCognito = true
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
		var err error
		ctx := r.Context()

		if isTest {
			ctx = security.SetupTestAuthFromContext(ctx, 1)
		} else if isCognito {
			//This is the scheme I am going to use for Cognito:
			//Authorization: COGNITO id_token="<idJWT>", access_token="<accessJWT>"
			//To start, only id_token will be used, this will just give me flex for future if needed
			//ctx = security.SetupTestAuthFromContext(ctx, 1)
			//token := r.Header.Get("X-AUTH-USER")
			ctx, err = security.SetupAuthFromHTTP(r)
			if err != nil {
				coreerror := errors.NewClientError("Invalid Authorization Header Structure: " + err.Error()).(errors.Error)
				//write error message
				w.WriteHeader(coreerror.ErrorType)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(coreerror)
				return
			}
		}
		// update request with new context
		r = r.WithContext(ctx)
		// call the next handler inthe chain - clean up of context done by libraries so unneeded to add cleanup here...
		next.ServeHTTP(w, r)
	})
}
