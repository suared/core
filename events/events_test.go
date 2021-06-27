package events

import (
	"testing"
	"time"

	"github.com/suared/core/types"
)

func TestEvents(t *testing.T) {
	Debug(nil, "testA.x", "Whatever")
	time.Sleep(1 * time.Second)
	Debug(nil, "testB.x", "Dude")
	time.Sleep(1 * time.Second)
	Debug(nil, "test1.x", "First Test!", types.Str("name", "david"), types.Int("age", 46))
	time.Sleep(1 * time.Second)
	Event(nil, "test2.x", "Second Test!", types.Str("name", "colleen"), types.Int("age", 48))
	Event(nil, "test3.1", "Third Test!", types.Str("name", "colleen"), types.Int("age", 48))
	Event(nil, "test3.2", "Third Test!", types.Str("name", "colleen"), types.Int("age", 48))
	Event(nil, "test3.3", "Third Test!", types.Str("name", "colleen"), types.Int("age", 48))
	Event(nil, "test3.4", "Third Test!", types.Str("name", "colleen"), types.Int("age", 48))
	//doesn't drop, good - uncomment to test in future if create buffer
	//panic("not drop test3.4")
}
