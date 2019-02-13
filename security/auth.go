package security

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

//SetupAuthFromHTTP - Enables Auth for later retrieval in the request flow, value added to returned Context
func SetupAuthFromHTTP(r *http.Request) context.Context {
	log.Printf("setting up auth from http...")
	ctx := r.Context()
	return context.WithValue(ctx, authKey, &BasicAuth{user: "suared"})
}

//SetupTestAuthFromContext - Enables Auth for later retrieval in the request flow, , value added to returned Context
func SetupTestAuthFromContext(ctx context.Context, userNumber uint) context.Context {
	return context.WithValue(ctx, authKey, &BasicAuth{user: "testuser" + fmt.Sprint(userNumber)})
}
