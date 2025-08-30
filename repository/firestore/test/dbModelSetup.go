package test

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/suared/core/v2/repository/firestore"
	"github.com/suared/core/v2/security"
)

// TestUserModel is the user-facing model for testing.
// It implements the firestore.ModelObject interface.
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

// MyCustomDAO is a custom DAO for the "advanced path" test.
// It embeds the generic BaseDAO and adds custom fields.
type MyCustomDAO struct {
	firestore.BaseDAO[TestUserModel]
	SearchKeywords []string `firestore:"searchKeywords"`
	IsArchived     bool     `firestore:"isArchived"`
}

// ToModel converts the custom DAO back to the user model.
func (d MyCustomDAO) ToModel() TestUserModel {
	return d.Model
}

// FromModel creates a custom DAO from a user model.
func FromModel(ctx context.Context, model TestUserModel) MyCustomDAO {
	auth := security.GetAuth(ctx)
	user := auth.GetUser()
	now := time.Now().UTC()

	return MyCustomDAO{
		BaseDAO: firestore.BaseDAO[TestUserModel]{
			Model:     model,
			CreatedBy: user,
			UpdatedBy: user,
			CreatedDt: now,
			UpdatedDt: now,
		},
		SearchKeywords: strings.Split(model.Title, " "),
		IsArchived:     false,
	}
}