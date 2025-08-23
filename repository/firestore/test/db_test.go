
package test

import (
	"context"
	"os"
	"testing"

	"github.com/suared/core/v2/repository"
	"github.com/suared/core/v2/repository/firestore"
	"github.com/suared/core/v2/security"

	_ "github.com/suared/core/v2/infra"
)

func TestDBSetup(t *testing.T) {
	// Create Test Context
	ctx := context.TODO()
	ctx = security.SetupTestAuthFromContext(ctx, 1)

	// Setup db props
	configMap := repository.NewBasicConfig("TestDao")
	configMap.AddEntry("backend", "firestore")
	configMap.AddEntry("projectID", os.Getenv("PROCESS_FIRESTORE_PROJECT_ID"))
	configMap.AddEntry("endpoint", os.Getenv("FIRESTORE_EMULATOR_HOST"))
	configMap.AddEntry("env", os.Getenv("PROCESS_ENV"))

	// Test DB Creation Successful
	initializedRepo, err := firestore.NewRepository(configMap, &TestDAO{}, TestUserModel{})
	if err != nil {
		t.Fatalf("Repo initialization failed with: %v", err)
	}

	testStruct := TestUserModel{}
	testStruct.ID = "A"
	testStruct.Title = "first test"

	// Test can insert/retrieve
	err = initializedRepo.Insert(ctx, testStruct)
	if err != nil {
		t.Errorf("Unable to insert value, received err: %v", err)
	}

	queryModel := TestUserModel{TestModel: TestModel{ID: "A"}}
	valObject, err := initializedRepo.SelectOne(ctx, queryModel)
	if err != nil {
		t.Errorf("Unable to query value, received err: %v", err)
	}

	if valObject != testStruct {
		t.Errorf("non-matching struct.., received: %v, expected: %v", valObject, testStruct)
	}

	// Query for a specific value/selectOne test
	testStruct2 := TestUserModel{TestModel: TestModel{ID: "B", Title: "2nd Item in table"}}
	err = initializedRepo.Insert(ctx, testStruct2)
	if err != nil {
		t.Errorf("Unable to insert value 2, received err: %v", err)
	}

	queryModel2 := TestUserModel{TestModel: TestModel{ID: "B"}}
	valObject, err = initializedRepo.SelectOne(ctx, queryModel2)
	if err != nil {
		t.Errorf("SelectOne: Unable to query value, received err: %v", err)
	}

	if valObject.ID != "B" || valObject.Title != "2nd Item in table" {
		t.Errorf("SelectOne: Unexpected results, received: %v", valObject)
	}

	// Select "all" matching the hashkey test
	valObjects, err := initializedRepo.Select(ctx, queryModel2)
	if err != nil {
		t.Errorf("Select: Unable to query value, received err: %v", err)
	}

	if len(valObjects) != 2 {
		t.Errorf("Expected 2 return value in query, received: %v", len(valObjects))
	}

	// Use Select All to filter down to 1 value
	valObjects, err = initializedRepo.Select(ctx, queryModel2, "Title", "==", "2nd Item in table")
	if err != nil {
		t.Errorf("Select: Unable to query value with filter, received err: %v", err)
	}

	if len(valObjects) != 1 {
		t.Errorf("Expected 1 return value in query, received: %v", len(valObjects))
	}

	// Test update
	testStruct.Title = "Second Test"
	err = initializedRepo.Update(ctx, testStruct)
	if err != nil {
		t.Errorf("Unable to update value, received err: %v", err)
	}

	valObject, err = initializedRepo.SelectOne(ctx, queryModel)
	if err != nil {
		t.Errorf("Unable to query value, received err: %v", err)
	}

	if valObject != testStruct {
		t.Errorf("non-matching struct.., received: %v, expected: %v", valObject, testStruct)
	}

	// Test delete
	err = initializedRepo.Delete(ctx, queryModel2)
	if err != nil {
		t.Errorf("Unable to delete value, received err: %v", err)
	}

	valObjects, err = initializedRepo.Select(ctx, queryModel2)
	if err != nil {
		t.Errorf("Select: Unable to query value after delete, received err: %v", err)
	}

	if len(valObjects) != 1 {
		t.Errorf("Expected 1 return value in query, received: %v", len(valObjects))
	}

	if valObjects[0] != testStruct {
		t.Errorf("Expected struct to be: %v, received: %v", testStruct, valObjects[0])
	}
}
