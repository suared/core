package security

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

var authKey key

//key is the key value for the auth struct
type key struct{}

func (a key) Comparable(T interface{}) bool {
	_, ok := T.(key)
	return ok
}

//initialize the key for use
func init() {
	authKey = key{}
}

//Auth - Interface for setting up and retrieving authentication data
type Auth interface {
	GetUser() string
	IsAdmin() bool
}

//BasicAuth - authentication data holder
type BasicAuth struct {
	user    string
	isAdmin bool
}

func (t *BasicAuth) String() string {
	return "User:" + t.user
}

//GetUser - returns end user
func (t *BasicAuth) GetUser() string {
	return t.user
}

//IsAdmin - returns if user is an admin
//TODO: this always will be false till implemented
func (t *BasicAuth) IsAdmin() bool {
	return t.isAdmin
}

//GetAuth - returns Auth from the provided context
func GetAuth(ctx context.Context) Auth {
	authKey := ctx.Value(authKey)
	if authKey == nil {
		return nil
	}
	return authKey.(Auth)
}

//IsAnonymous - if security context is not set or user is empty or the string anonymousit  will return true/ not logged in
func IsAnonymous(ctx context.Context) bool {
	auth := GetAuth(ctx)
	if auth == nil {
		return true
	}
	username := auth.GetUser()
	if username == "" || username == "anonymous" {
		return true
	}
	return false
}

type authHeaderStruct struct {
	scheme   string
	valueMap map[string]string
	isScheme bool
	isKey    bool
	isVal    bool
	lastKey  string
}

func (header *authHeaderStruct) Set(val string) {
	//This is a stateful switch so returning after each block is required so only one case happens per turn
	if header.isScheme {
		header.scheme = val
		header.isKey = true
		header.isScheme = false
		return
	}

	if header.isKey {
		header.lastKey = val
		header.isKey = false
		header.isVal = true
		return
	}

	if header.isVal {
		header.valueMap[header.lastKey] = val
		header.isKey = true
		header.isVal = false
		return
	}
}

//At present this is only good for one time use, assumes previous call is a newly initialized struct
func (header *authHeaderStruct) setAuthString(val string) {
	//Create slices, ignore back to back "space" characters when processing
	authHeaderTokenizer := strings.NewReplacer("\"", "",
		"=", " ",
		",", " ",
	)
	authHeaderTokens := authHeaderTokenizer.Replace(val)
	authHeaderSlice := strings.Split(authHeaderTokens, " ")
	prev := " "
	for i := range authHeaderSlice {
		switch authHeaderSlice[i] {
		case "":
			//ignore spaces

		default:
			prev = authHeaderSlice[i]
			header.Set(prev)
		}
	}

}
func newAuthHeaderStruct(val string) *authHeaderStruct {
	obj := authHeaderStruct{}
	obj.isScheme = true
	obj.valueMap = make(map[string]string)
	retValue := &obj
	retValue.setAuthString(val)
	return retValue
}

//SetupAuthFromHTTP - Enables Auth for later retrieval in the request flow, value added to returned Context
func SetupAuthFromHTTP(r *http.Request) (context.Context, error) {
	//log.Printf("setting up auth from http...")
	ctx := r.Context()
	//r := strings.NewReplacer("<", "&lt;", ">", "&gt;")
	//func Map(mapping func(rune) rune, s string) string
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return context.WithValue(ctx, authKey, &BasicAuth{user: "anonymous"}), nil
	}

	authHeaderStruct := newAuthHeaderStruct(authHeader)

	//If Scheme is Cognito --> Authorization: COGNITO id_token="<idJWT>", access_token="<accessJWT>"
	//only supported Scheme so far is Cognito so I am just falling through here for now till others exist
	//Only support key is "id_token" so assumed for now
	idToken := authHeaderStruct.valueMap["id_token"]
	//validate jwt is still valid (signature check + expiration check); if not valid throw error
	basicAuth, err := validateJWT(idToken)
	if err != nil {
		return ctx, fmt.Errorf("unable to validate JWT: %v", err)
	}
	//

	return context.WithValue(ctx, authKey, &basicAuth), nil
}

func validateJWT(tokenString string) (BasicAuth, error) {

	//This only checks for AWS Cognito tokens right now, can expand it within later...
	token, err := validateToken(tokenString)

	if err != nil {
		return BasicAuth{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return BasicAuth{}, fmt.Errorf("Claims type unexpected: %v", claims)
	}

	validatedEmail, ok := claims["email_verified"].(bool)
	if !ok {
		return BasicAuth{}, fmt.Errorf("ID Token expected with email verification, received: %v", validatedEmail)
	}

	if validatedEmail {
		return BasicAuth{user: claims["email"].(string)}, nil
	}

	//non email validated users will be treated the same as anonymous users
	return BasicAuth{user: "anoymous"}, nil
}

//SetupTestAuthFromContext - Enables Auth for later retrieval in the request flow, , value added to returned Context
func SetupTestAuthFromContext(ctx context.Context, userNumber uint) context.Context {
	return context.WithValue(ctx, authKey, &BasicAuth{user: "testuser" + fmt.Sprint(userNumber)})
}
