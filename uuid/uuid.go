package uuid

import "github.com/segmentio/ksuid"

//NewUUID - Returns a new unique identifier
func NewUUID() string {
	return ksuid.New().String()
}
