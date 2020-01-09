package errors

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

//Will use http status codes for simplicity though errors can come from server or api tier.  Each selected error is easily understandable in both contexts

const (
	//2xx
	StatusOK = http.StatusOK

	//4xx
	StatusBadRequest      = http.StatusBadRequest
	StatusUnauthorized    = http.StatusUnauthorized
	StatusPaymentRequired = http.StatusPaymentRequired
	StatusForbidden       = http.StatusForbidden
	StatusNotFound        = http.StatusNotFound

	//5xx
	StatusInternalServerError = http.StatusInternalServerError
	StatusNotImplemented      = http.StatusNotImplemented
)

/*
Error contains -
- errorType - from the const list, the type of error this represents.  The New constructors set reasonable default status that can be replaced if appropriate.
- developerMessage - message that may help the developer use the API correctly
  - userMessage - message for the end user display
  - errorCode - error to report to support or standardized user error code as applicable
  - moreInfo - link to documentation with more info

-  If eventing is enabled, event will be generated upon creation.  For now, have only setup for system errors to event.  For example, validation errors or client errors will not event.  Can change to config in the future

Future TODO:
   - add correlation id into errorCode or moreInfo when reporting
*/
type Error struct {
	ErrorType        int
	DeveloperMessage string
	UserMessage      string
	ErrorCode        string
	MoreInfo         string
}

//Error - implements the error interface
func (err Error) Error() string {
	return `ErrorType:` + string(err.ErrorType) + `
	DeveloperMessage:` + err.DeveloperMessage + `
	UserMessage:` + err.UserMessage + `
	ErrorCode:` + err.ErrorCode + `
	MoreInfo:` + err.MoreInfo
}

//NewValidationError - returns a default StatusOK error denoting a user validation error
func NewValidationError(err string) error {
	return Error{ErrorType: StatusOK,
		UserMessage: err}
}

//NewClientError - returns a default BadRequest denoting a client user error
func NewClientError(err string) error {
	return Error{ErrorType: StatusBadRequest,
		DeveloperMessage: err}
}

//NewSystemError - returns a default InternalServerError denoting an implementation issue
func NewSystemError(ctx context.Context, err string) error {
	theErr := Error{ErrorType: StatusInternalServerError,
		DeveloperMessage: err}
	exceptionEvent(ctx, theErr)
	return theErr
}

//NewError - returns a full core error struct from the default error type.  Type defafults to server error
//Because this takes an error, the assumption is the underlying type is preferred to minimize type conversions in calling code
func NewError(ctx context.Context, err error) Error {
	theErr := Error{ErrorType: StatusInternalServerError,
		DeveloperMessage: err.Error(),
		UserMessage:      "Oops!  Unexpected data and actions, please try a different change"}
	exceptionEvent(ctx, theErr)
	return theErr
}

//LastResortHandler - Setup at start of program to capture panics not solved elsewhere to enable logging prior to program close down
func LastResortHandler(loc string) {
	LastResortHandlerWithContext(nil, loc)
}

//LastResortHandlerWithContext - Setup at start of program to capture panics not solved elsewhere to enable logging prior to program close down
func LastResortHandlerWithContext(ctx context.Context, loc string) {
	if r := recover(); r != nil {
		log.Printf("LastResort: panic=%v", r)
		exceptionEvent(ctx, fmt.Errorf("LastResort:%v, panic:%v", loc, r))
		//If panicking, need to wait to be sure flush of events clears prior to program close
		if isEventing {
			time.Sleep(flushDuration() + time.Duration(time.Second*1))
		}
	}
}
