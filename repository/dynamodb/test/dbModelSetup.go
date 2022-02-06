package test

import (
	"context"
	"fmt"

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
	//  * Repository - Implementation in library, generated from DAO data
	//  * Internal DAO from Model - Implementation options is library, option selection is caller
	//  * Database - library
*/

//sample caller model
type TestModel struct {
	ID    string
	Title string
}

func (model *TestModel) String() string {
	return fmt.Sprintf("TestModel: model.ID=%v, model.Title=%v", model.ID, model.Title)
}

//sample caller model wrapper for non-model adds (e.g. ui bits) to show it works with embedded structs
type TestUserModel struct {
	dummy string
	TestModel
}

func (model TestUserModel) IsEmpty() bool {
	return model.ID == ""
}

func (model *TestUserModel) String() string {
	return fmt.Sprintf("TestuserModel: model.ID=%v, model.TestModel=%v", model.ID, model.TestModel)
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

func (dao *TestDAO) SetUser(ctx context.Context) {
	dao.UserID = security.GetAuth(ctx).GetUser()
}

func (dao *TestDAO) Setup(model interface{}) {
	thismodel := model.(TestUserModel)
	dao.ID = thismodel.ID
	dao.Title = thismodel.Title
	dao.TestUserModel = thismodel
}

func (dao *TestDAO) GetModel() interface{} {
	return dao.TestUserModel
}

func (dao *TestDAO) String() string {
	return fmt.Sprintf("TestDAO: dao.ID=%v, dao.Title=%v, dao.TestModel=%v", dao.ID, dao.Title, dao.TestUserModel)
}

//Refresh - updates the Hashkey and SortKey.  Used by the library before calls
func (dao *TestDAO) Refresh() {
	dao.MyCleverHashKeyName = "TestTable_" + dao.UserID
	dao.MyCleverSortKeyName = dao.ID
}

//Populate - For any additional actions, e.g. calculated columns, zip/unzip, etc
func (dao *TestDAO) Populate() {

}
