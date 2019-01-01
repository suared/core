package api

import (
	"io"
	"net/http"
)

var healthChecker Health

func init() {
	//set default health check
	healthChecker = basicAPIHealth{}
}

//Health - Interface can be set by caller using --> setHealthChecker(health APIHealth)
//If not set, only core architecture health checks will be performed
type Health interface {
	Healthy() bool
}

//Default Health Checker if a replacement is not set by calling API
type basicAPIHealth struct {
}

func (checker basicAPIHealth) Healthy() bool {
	//add any core checks here, till then keep super simple
	return true
}

//HealthCheckHandler - Implements the Health check via http handler
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Mux example simple health  check extended to use interface from caller if available
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	// Add any general infra level checks here

	// Defer to set caller any api specific checks, e.g. db access and other infra supporting the api

	// Keeping the simple final result as the caller shouldn't care about the details
	if healthChecker.Healthy() {
		io.WriteString(w, `{"alive": true}`)
	} else {
		//In the future, may want to consider killing the process here or adding a recovery test
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"alive": false}`)
	}

}
