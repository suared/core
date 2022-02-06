package test

import (
	"context"
	"os"
	"testing"

	"github.com/suared/core/repository"
	"github.com/suared/core/repository/dynamodb"

	"github.com/suared/core/security"

	_ "github.com/suared/core/infra"
)

/*

In addition to validating the library this provides and easy reference on the bits that need to be implemented by a caller and can be copy/pasted as a start with a specific model based on the use case

Debugging:
CLI Reference: https://docs.aws.amazon.com/cli/latest/reference/dynamodb/index.html#cli-aws-dynamodb
Command Example:  aws dynamodb scan --table-name "testTable" --endpoint-url http://localhost:9001
Other:  not done intentionally (yet) is to force insert/ update to use condition-expression attribute_exists or not exists as appropriate

	//Flow:
	//Typical:  Callers (Model) --> Repository -->  Internal DAO from Model -->  Database
	//Responsibilities Arch:
	//	* Callers(Model) - Callers In  (( above ))
	//  * Repository - Definition is Caller, Implementation is library (( above for caller ))
	//  * Internal DAO from Model - Implementation options is library, option selection is caller
	//  * Database - library
	// Create convenienience functions for the DAO to handle: zipme bool, active bool, audit bool
*/

//sample caller model
//TODO: TestMode, TestUserModel, DAO are in the v1 file to be merged in the future

//Validation of the library data lifecycle is here - only the above items are needed to setup for a specific caller

func TestDBSetupv2(t *testing.T) {
	//Create Test Context
	ctx := context.TODO()
	ctx = security.SetupTestAuthFromContext(ctx, 1)

	//setup db props
	configMap := repository.NewBasicConfig("TestDao")
	configMap.AddEntry("backend", os.Getenv("PROCESS_REPOSITORY"))
	configMap.AddEntry("table", os.Getenv("PROCESS_AWS_DYNAMOTABLE_CATEGORY"))
	configMap.AddEntry("region", os.Getenv("PROCESS_AWS_REGION"))
	configMap.AddEntry("endpoint", os.Getenv("PROCESS_AWS_DYNAMOENDPOINT"))
	configMap.AddEntry("rcu", os.Getenv("PROCESS_AWS_DYNAMOTABLE_RCU"))
	configMap.AddEntry("wcu", os.Getenv("PROCESS_AWS_DYNAMOTABLE_WCU"))

	configMap.AddEntry("hashKeyName", "MyCleverHashKeyName")
	configMap.AddEntry("sortKeyName", "MyCleverSortKeyName")
	configMap.AddEntry("env", os.Getenv("PROCESS_ENV"))

	//Test DB Creation Successful (will return existing db table if exists, during testing will not exist the first time)
	initializedRepo, err := dynamodb.NewRepository(configMap, &TestDAO{}, TestUserModel{})
	if err != nil {
		t.Errorf("Repo initialization failed with: %v", err)
	}

	testStruct := TestUserModel{}
	testStruct.ID = "A"
	testStruct.Title = "first test"

	//Test can insert/ retrieve
	err = initializedRepo.Insert(ctx, testStruct)

	if err != nil {
		t.Errorf("Unable to insert value, received err: %v", err)
	}

	queryModel := TestUserModel{}
	queryModel.ID = "A"
	valObject, err := initializedRepo.SelectOne(ctx, queryModel)

	if err != nil {
		t.Errorf("Unable to query value, received err: %v", err)
	}

	if valObject != testStruct {
		t.Errorf("non-matching struct.., received: %v, expected: %v", valObject, testStruct)
	}

	//Query for a specific value/ selectOne test
	testStruct2 := TestUserModel{}
	testStruct2.ID = "B"
	testStruct2.Title = "2nd Item in table"

	err = initializedRepo.Insert(ctx, testStruct2)

	if err != nil {
		t.Errorf("Unable to insert value 2, received err: %v", err)
	}

	queryModel2 := TestUserModel{}
	queryModel2.ID = "B"
	valObject, err = initializedRepo.SelectOne(ctx, queryModel2)

	if err != nil {
		t.Errorf("SelectOne: Unable to query value, received err: %v", err)
	}

	if valObject.ID != "B" && valObject.Title != "2nd Item in table" {
		t.Errorf("SelectOne: Unexpected results, received: %v", valObject)
	}

	//select "all" matching the hashkey test
	valObjects, err := initializedRepo.Select(ctx, queryModel2)

	if err != nil {
		t.Errorf("Select: Unable to query value, received err: %v", err)
	}

	if len(valObjects) != 2 {
		t.Errorf("Expected 2 return value in query, received: %v", len(valObjects))
	}

	//Use Select All to filter down to 1 value
	valObjects, err = initializedRepo.Select(ctx, queryModel2, "Title = :title", "title:S:2nd Item in table")

	if err != nil {
		t.Errorf("Select: Unable to query value, received err: %v", err)
	}

	if len(valObjects) != 1 {
		t.Errorf("Expected 1 return value in query, received: %v", len(valObjects))
	}

	//Test update
	testStruct.Title = "Second Test"

	//Test update of title
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

	//Test delete
	err = initializedRepo.Delete(ctx, queryModel2)

	if err != nil {
		t.Errorf("Unable to query value, received err: %v", err)
	}

	valObjects, err = initializedRepo.Select(ctx, queryModel2)

	if err != nil {
		t.Errorf("Select: Unable to query value, received err: %v", err)
	}

	if len(valObjects) != 1 {
		t.Errorf("Expected 1 return value in query, received: %v", len(valObjects))
	}

	if valObjects[0] != testStruct {
		t.Errorf("Expected struct to be: %v, received: %v", testStruct, valObjects[0])
	}

}
