package events

import (
	"testing"
	"time"

	"github.com/suared/core/types"
)

func TestEvents(t *testing.T) {
	Debug("testA.x", "Whatever")
	time.Sleep(1 * time.Second)
	Debug("testB.x", "Dude")
	time.Sleep(1 * time.Second)
	Debug("test1.x", "First Test!", types.Str("name", "david"), types.Int("age", 46))
	time.Sleep(1 * time.Second)
	Event("test2.x", "Second Test!", types.Str("name", "colleen"), types.Int("age", 48))
}
