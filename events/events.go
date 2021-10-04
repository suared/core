package events

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/suared/core/security"
	"github.com/suared/core/types"
)

var debugLogger zerolog.Logger
var eventLogger zerolog.Logger
var environment string //Pull from PROCESS_ENV environment variable

func init() {
	environment = os.Getenv("PROCESS_ENV")
	//For Developer Output, not intended to be removed, create new method/level if have temp only need but try to avoid to extent possible outside dev env
	debugLogger = zerolog.New(os.Stderr).With().Timestamp().Str("type", "MSG").Str("level", "DEBUG").Str("env", environment).Logger()
	//Events to integrate into event processes
	eventLogger = zerolog.New(os.Stderr).With().Timestamp().Str("type", "EVENT").Str("env", environment).Logger()
}

//Debug write debug message to stdout
//Using Loc as user provided to enable flex on how to define - e.g. domain.method or unique identifier, etc as makes sense.  Will ensure it is always thought about to start
func Debug(ctx context.Context, loc string, msg string, keyVals ...types.KeyVal) {
	auth := security.GetAuth(ctx)
	userID := ""
	if auth != nil {
		userID = auth.GetUser()
	}
	if keyVals == nil {
		debugLogger.Log().Str("loc", loc).Str("userID", userID).Msg(msg)
	} else {
		//dictionary, string, int, float
		logger := debugLogger.Log().Str("loc", loc).Str("userID", userID)
		for _, keyVal := range keyVals {
			switch keyVal.Typ {
			case types.TYPESTRING:
				logger = logger.Str(keyVal.Key, keyVal.Val)
			case types.TYPEINT:
				logger = logger.Int(keyVal.Key, keyVal.ValueAsInt())
			case types.TYPEFLOAT:
				logger = logger.Float64(keyVal.Key, keyVal.ValueAsFloat())
			}

		}
		logger.Msg(msg)
	}

}

//Event write event message to stdout - note: events are handled outside the scope of the module by design
//See debug comments, same concept here, is just marked as event to enable later processing of event stream
//<domain>-<event>:  have event in the topic name so that listeners can focus on the key events interested in
//Ex:  ops-log, task-complete, task-create, task-skip, etc…

func Event(ctx context.Context, loc string, msg string, keyVals ...types.KeyVal) {
	auth := security.GetAuth(ctx)
	userID := ""
	if auth != nil {
		userID = auth.GetUser()
	}
	if keyVals == nil {
		eventLogger.Log().Str("loc", loc).Str("userID", userID).Msg(msg)
	} else {
		//dictionary, string, int, float
		logger := eventLogger.Log().Str("loc", loc).Str("userID", userID)
		for _, keyVal := range keyVals {
			switch keyVal.Typ {
			case types.TYPESTRING:
				logger = logger.Str(keyVal.Key, keyVal.Val)
			case types.TYPEINT:
				logger = logger.Int(keyVal.Key, keyVal.ValueAsInt())
			case types.TYPEFLOAT:
				logger = logger.Float64(keyVal.Key, keyVal.ValueAsFloat())
			}

		}
		logger.Msg(msg)
	}

}
