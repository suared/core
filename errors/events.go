package errors

import (
	"os"
	"strconv"
	"time"

	sentry "github.com/getsentry/sentry-go"
)

var isEventing bool

func flushDuration() time.Duration {
	//Default flushSeconds to 5
	flushSeconds := time.Second * 5
	flushInterval := os.Getenv("ERROR_EVENTING_FLUSH_SECS")
	flushTest, err := strconv.Atoi(flushInterval)
	if err != nil && flushTest != 0 {
		flushSeconds = (time.Second * time.Duration(flushTest))
	}
	return flushSeconds
}

//SetupEventing - Configures your eventing if enabled via env variable: ERROR_EVENTING = true
func SetupEventing() {
	var sendStackTrace bool

	isEventing = true

	uri := os.Getenv("ERROR_EVENTING_URI")
	stackTraces := os.Getenv("ERROR_EVENTING_TRACES")

	if uri == "" {
		panic("Error eventing set to true without a URI to notify")
	}
	if stackTraces == "true" {
		sendStackTrace = true
	}

	//Different options can be setup here in the future
	sentry.Init(sentry.ClientOptions{
		Dsn:              uri,
		Environment:      os.Getenv("PROCESS_ENV"),
		AttachStacktrace: sendStackTrace,
	})
	sentry.Flush(flushDuration())
}

//exceptionEvent - Adds exception event if eventing is configured
func exceptionEvent(err error) {
	if isEventing {
		sentry.CaptureException(err)
	}
}
