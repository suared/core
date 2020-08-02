package test

import (
	"context"
	"errors"
	"log"
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
*/

//sample caller model
type TestModel struct {
	ID    string
	Title string
}

//sample caller model wrapper for non-model adds (e.g. ui bits) to show it works with embedded structs
type TestUserModel struct {
	dummy string
	TestModel
}

//sample repository model - what a caller would implement
type TestRepository struct {
	/*Required fields for dynamoDB implementation
	backend     string
	table       string
	region      string
	endpoint    string
	hashKeyName string
	sortKeyName string
	rcu         int
	wcu         int
	*/
	config  repository.Config
	session repository.Session
}

func (repo *TestRepository) Config() repository.Config {
	return repo.config
}

//TODO:  Create/Test active and audit options as well as zip conversion
//sample DAO, the DAO requires to have the HashKey and SortKey fields mathing the names - how you do that is up to the DAO implementor.  In this example, the data is being duplicated vs. naming the fields to show a potential indirection path to allow growth

//TestDAO - Caller would ipmlement the relevant DAO which would include the model to be saved along with any other key values (e.g. hash/sort for dynamo)
type TestDAO struct {
	UserID              string
	MyCleverHashKeyName string
	MyCleverSortKeyName string
	TestUserModel
}

//HashKey - This is the value that would be set as the dynamo hashkey
func (dao *TestDAO) HashKey() string {
	return dao.MyCleverHashKeyName
}

//SortKey - This is the value that would be set as the dynamo sortKey
func (dao *TestDAO) SortKey() string {
	return dao.MyCleverSortKeyName
}

//User - the user that made this call
func (dao *TestDAO) User() string {
	return dao.UserID
}

//New - creates a new instance of this specific type to support return values of the right type
func (dao *TestDAO) New() dynamodb.DAO {
	return new(TestDAO)
}

//Refresh - updates the Hashkey and SortKey.  Used by the library before calls
func (dao *TestDAO) Refresh() {
	dao.MyCleverHashKeyName = "TestTable_" + dao.UserID
	dao.MyCleverSortKeyName = dao.ID
}

//Populate - For any additional actions, e.g. calculated columns, zip/unzip, etc
func (dao *TestDAO) Populate() {
}

//NewTestDAO - Initializes this object with the user ID from context
func NewTestDAO(ctx context.Context) *TestDAO {
	dao := new(TestDAO)
	dao.UserID = security.GetAuth(ctx).GetUser()
	return dao
}

//DAO- Returns a DAO associated with this repository from a model object
func (r *TestRepository) DAO(ctx context.Context, userModel TestUserModel, active bool, audit bool) (dynamodb.DAO, error) {
	// Both "Un-Delete" and "Make Active" will be the exception to the rule so this will assume always from active pile and not deleted
	// Special methods will be created for the special cases vs. adding complexity to the base case
	// Populate the Data object First
	dao := NewTestDAO(ctx)
	dao.ID = userModel.ID
	dao.Title = userModel.Title
	return dao, nil
}

//Insert - Sample of a basic insert method with validation
func (r *TestRepository) Insert(ctx context.Context, userModel TestUserModel) error {
	// Populate the Data object First //  active?, audit?
	dao, err := r.DAO(ctx, userModel, false, false)

	if err != nil {
		log.Printf("Unable to Insert DAO, error Getting DAO: %v", err)
		return err
	}

	// Repository layer is responsible for validating auth rules
	err = dynamodb.ValidAction(ctx, "insert", dao)
	if err != nil {
		return err
	}

	return dynamodb.InsertOrUpdate(ctx, r, dao)
}

//Update - Sample of updating a DB entry
func (r *TestRepository) Update(ctx context.Context, userModel TestUserModel) error {
	dao, err := r.DAO(ctx, userModel, false, false)
	if err != nil {
		log.Printf("Unable to Update, error getting DAO, err: %v", err)
		return err
	}

	err = dynamodb.ValidAction(ctx, "update", dao)
	if err != nil {
		return err
	}

	return dynamodb.InsertOrUpdate(ctx, r, dao)

}

//Delete - Sample of deleting a DB entry
func (r *TestRepository) Delete(ctx context.Context, template TestUserModel) error {
	dao, err := r.DAO(ctx, template, false, false)
	if err != nil {
		log.Printf("Unable getting Dao in Delete, err: %v", err)
		return err
	}

	return dynamodb.Delete(ctx, r, dao)
}

//Select - Sample of a get all by hashkey
func (r *TestRepository) Select(ctx context.Context, template TestUserModel, strFilterVals ...string) ([]TestUserModel, error) {
	dao, err := r.DAO(ctx, template, false, false)
	if err != nil {
		log.Printf("Unable to Select, error getting DAO, err: %v", err)
		return nil, err
	}

	result, err := dynamodb.Select(ctx, r, dao, strFilterVals...)

	var outputList []TestUserModel
	//since the search is for user, validation only needs to occur on one item..
	var validated bool
	for i := range result {
		resultDAO := result[i]
		//Check once only...
		if !validated {
			//ignore if no result/ empty
			if resultDAO.HashKey() != "" {
				err = dynamodb.ValidAction(ctx, "selectAll", resultDAO)

				if err != nil {
					return []TestUserModel{}, err
				}
			}
		}
		//Convert DAO to Request here then add to list
		testDao, ok := resultDAO.(*TestDAO)
		if !ok {
			return []TestUserModel{}, errors.New("Unable to convert back to TestDAO, DB results unexpected")
		}
		resultItem := testDao.TestUserModel
		outputList = append(outputList, resultItem)
		//log.Printf("output list: %v", outputList)
	}

	return outputList, err

}

//SelectOne - Returns one model object, can be empty if no results
func (repo *TestRepository) SelectOne(ctx context.Context, template TestUserModel) (TestUserModel, error) {
	dao, err := repo.DAO(ctx, template, false, false)
	if err != nil {
		log.Printf("Unable to SelectOne, error getting DAO, err: %v", err)
		return TestUserModel{}, err
	}

	result, err := dynamodb.SelectOne(ctx, repo, dao)

	//Convert DAO to Request here then add to list
	testDao, ok := result.(*TestDAO)
	if !ok {
		return TestUserModel{}, errors.New("Unable to convert back to TestDAO, DB results unexpected")
	}
	resultItem := testDao.TestUserModel

	if resultItem.ID != "" {
		err = dynamodb.ValidAction(ctx, "selectOne", result)

		if err != nil {
			return TestUserModel{}, err
		}
	}

	return resultItem, err

}

//SetSession - enables the library to store/ reuse the session for efficiency vs. creating new on each call
func (r *TestRepository) SetSession(session repository.Session) {
	r.session = session
}

//Session - Returns the session associated with this repository
func (r *TestRepository) Session() repository.Session {
	return r.session
}

//NewRepository - Initializes a sample repository with config values set
func NewRepository() *TestRepository {
	testRepo := new(TestRepository)
	configMap := repository.NewBasicConfig("mytestdnyamodb")
	configMap.AddEntry("backend", "dynamoDB")
	configMap.AddEntry("table", "testTable")
	configMap.AddEntry("region", "us-east-1")
	configMap.AddEntry("endpoint", "http://localhost:9001")
	configMap.AddEntry("hashKeyName", "MyCleverHashKeyName")
	configMap.AddEntry("sortKeyName", "MyCleverSortKeyName")
	configMap.AddEntry("rcu", "1")
	configMap.AddEntry("wcu", "1")
	testRepo.config = configMap
	return testRepo
}

//Validation of the library data lifecycle is here - only the above items are needed to setup for a specific caller
func TestDBSetup(t *testing.T) {
	//Create Test Context
	ctx := context.TODO()
	ctx = security.SetupTestAuthFromContext(ctx, 1)

	//Test DB Creation Successful (will return existing db table if exists, during testing will not exist the first time)
	initializeRepo := NewRepository()
	initializedRepo, err := dynamodb.CreateTable(initializeRepo)
	if err != nil {
		t.Errorf("Repo initialization failed with: %v", err)
	}

	dynamoDBRepo := initializedRepo.(*TestRepository)

	if err != nil {
		t.Errorf("Repository creation failed with: %v", err)
	}

	testStruct := TestUserModel{}
	testStruct.ID = "A"
	testStruct.Title = "first test"

	//Test can insert/ retrieve
	err = dynamoDBRepo.Insert(ctx, testStruct)

	if err != nil {
		t.Errorf("Unable to insert value, received err: %v", err)
	}

	queryModel := TestUserModel{}
	queryModel.ID = "A"
	valObject, err := dynamoDBRepo.SelectOne(ctx, queryModel)

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

	err = dynamoDBRepo.Insert(ctx, testStruct2)

	if err != nil {
		t.Errorf("Unable to insert value 2, received err: %v", err)
	}

	queryModel2 := TestUserModel{}
	queryModel2.ID = "B"
	valObject, err = dynamoDBRepo.SelectOne(ctx, queryModel2)

	if err != nil {
		t.Errorf("SelectOne: Unable to query value, received err: %v", err)
	}

	if valObject.ID != "B" && valObject.Title != "2nd Item in table" {
		t.Errorf("SelectOne: Unexpected results, received: %v", valObject)
	}

	//select "all" matching the hashkey test
	valObjects, err := dynamoDBRepo.Select(ctx, queryModel2)

	if err != nil {
		t.Errorf("Select: Unable to query value, received err: %v", err)
	}

	if len(valObjects) != 2 {
		t.Errorf("Expected 2 return value in query, received: %v", len(valObjects))
	}

	//Use Select All to filter down to 1 value
	valObjects, err = dynamoDBRepo.Select(ctx, queryModel2, "Title = :title", "title:S:2nd Item in table")

	if err != nil {
		t.Errorf("Select: Unable to query value, received err: %v", err)
	}

	if len(valObjects) != 1 {
		t.Errorf("Expected 1 return value in query, received: %v", len(valObjects))
	}

	//Test update
	testStruct.Title = "Second Test"

	//Test update of title
	err = dynamoDBRepo.Update(ctx, testStruct)

	if err != nil {
		t.Errorf("Unable to update value, received err: %v", err)
	}

	valObject, err = dynamoDBRepo.SelectOne(ctx, queryModel)

	if err != nil {
		t.Errorf("Unable to query value, received err: %v", err)
	}

	if valObject != testStruct {
		t.Errorf("non-matching struct.., received: %v, expected: %v", valObject, testStruct)
	}

	//Test delete
	err = dynamoDBRepo.Delete(ctx, queryModel2)

	if err != nil {
		t.Errorf("Unable to query value, received err: %v", err)
	}

	valObjects, err = dynamoDBRepo.Select(ctx, queryModel2)

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
