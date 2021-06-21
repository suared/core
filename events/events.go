package events

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/suared/core/types"
)

var debugLogger zerolog.Logger
var eventLogger zerolog.Logger

func init() {
	//For Developer Output, not intended to be removed, create new method/level if have temp only need but try to avoid to extent possible outside dev env
	debugLogger = zerolog.New(os.Stderr).With().Timestamp().Str("type", "MSG").Str("level", "DEBUG").Logger()
	//Events to integrate into event processes
	eventLogger = zerolog.New(os.Stderr).With().Timestamp().Str("type", "EVENT").Str("level", "EVENT").Logger()

}

//Debug write debug message to stdout
//Using Loc as user provided to enable flex on how to define - e.g. domain.method or unique identifier, etc as makes sense.  Will ensure it is always thought about to start
func Debug(loc string, msg string, keyVals ...types.KeyVal) {
	if keyVals == nil {
		debugLogger.Log().Str("loc", loc).Msg(msg)
	} else {
		//dictionary, string, int, float
		logger := debugLogger.Log().Str("loc", loc)
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
func Event(loc string, msg string, keyVals ...types.KeyVal) {
	if keyVals == nil {
		eventLogger.Log().Str("loc", loc).Msg(msg)
	} else {
		//dictionary, string, int, float
		logger := eventLogger.Log().Str("loc", loc)
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
