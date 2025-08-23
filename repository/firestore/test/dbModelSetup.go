
package test

import (
	"context"
	"fmt"

	"github.com/suared/core/v2/repository/firestore"
	"github.com/suared/core/v2/security"

	_ "github.com/suared/core/v2/infra"
)

// sample caller model
type TestModel struct {
	ID    string
	Title string
}

func (model *TestModel) String() string {
	return fmt.Sprintf("TestModel: model.ID=%v, model.Title=%v", model.ID, model.Title)
}

// sample caller model wrapper
type TestUserModel struct {
	TestModel
}

func (model TestUserModel) IsEmpty() bool {
	return model.ID == ""
}

func (model *TestUserModel) String() string {
	return fmt.Sprintf("TestUserModel: model.ID=%v, model.TestModel=%v", model.ID, model.TestModel)
}

// TestDAO - Caller would implement the relevant DAO
type TestDAO struct {
	UserID     string
	CollectionName string
	DocIDValue string
	TestUserModel
}

func (dao *TestDAO) Collection() string {
	return dao.CollectionName
}

func (dao *TestDAO) DocID() string {
	return dao.DocIDValue
}

func (dao *TestDAO) User() string {
	return dao.UserID
}

func (dao *TestDAO) New() firestore.DAO {
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

func (dao *TestDAO) Refresh() {
	dao.CollectionName = "test-collection"
	dao.DocIDValue = dao.ID
}

func (dao *TestDAO) Populate() {}
