package test

import (
	"context"
	"fmt"
	"strings"

	"github.com/suared/core/v2/repository/dynamodb"
	"github.com/suared/core/v2/security"
)

// TestUserModel is the user-facing model for testing.
type TestUserModel struct {
	ID    string
	Title string
}

func (m TestUserModel) IsEmpty() bool {
	return m.ID == ""
}

func (m TestUserModel) String() string {
	return fmt.Sprintf("ID: %s, Title: %s", m.ID, m.Title)
}

// #############################################################################
// # Simple DAO - Represents the most minimal DAO a user would create.
// #############################################################################

// SimpleTestDAO is the minimal custom DAO for testing.
// It embeds the generic BaseDAO and defines the key fields for DynamoDB.
type SimpleTestDAO struct {
	dynamodb.BaseDAO[TestUserModel]
	HKey string `dynamodbav:"MyCleverHashKeyName"`
	SKey string `dynamodbav:"MyCleverSortKeyName"`
}

// FromModelSimple creates a minimal DAO from a user model.
func FromModelSimple(ctx context.Context, model TestUserModel) SimpleTestDAO {
	auth := security.GetAuth(ctx)
	return SimpleTestDAO{
		BaseDAO: dynamodb.BaseDAO[TestUserModel]{Model: model},
		HKey:    "TestTable_" + auth.GetUser(),
		SKey:    model.ID,
	}
}

// ToModelSimple converts the minimal DAO back to the user model.
func ToModelSimple(dao SimpleTestDAO) TestUserModel {
	return dao.Model
}

// #############################################################################
// # Advanced DAO - Represents a DAO with additional, non-model fields.
// #############################################################################

// AdvancedTestDAO is a more complex DAO for testing, with extra fields.
type AdvancedTestDAO struct {
	dynamodb.BaseDAO[TestUserModel]
	HKey           string   `dynamodbav:"MyCleverHashKeyName"`
	SKey           string   `dynamodbav:"MyCleverSortKeyName"`
	SearchKeywords []string `dynamodbav:"searchKeywords"`
	IsArchived     bool     `dynamodbav:"isArchived"`
}

// FromModelAdvanced creates an advanced DAO from a user model.
func FromModelAdvanced(ctx context.Context, model TestUserModel) AdvancedTestDAO {
	auth := security.GetAuth(ctx)
	return AdvancedTestDAO{
		BaseDAO:        dynamodb.BaseDAO[TestUserModel]{Model: model},
		HKey:           "TestTable_" + auth.GetUser(),
		SKey:           model.ID,
		SearchKeywords: strings.Split(model.Title, " "),
		IsArchived:     false,
	}
}

// ToModelAdvanced converts the advanced DAO back to the user model.
func ToModelAdvanced(dao AdvancedTestDAO) TestUserModel {
	return dao.Model
}
