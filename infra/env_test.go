package infra

import (
	"fmt"
	"os"
	"testing"
)

func TestEnvironment(t *testing.T) {
	//This is tested via the users of this service, dummy setup here for now
	fmt.Printf("In Tester... up, model dir set to: %v", os.Getenv("PROCESS_MODEL_TESTFILE_DIR"))
}
