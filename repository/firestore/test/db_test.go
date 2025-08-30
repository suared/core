package test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/suared/core/v2/repository"
	"github.com/suared/core/v2/repository/firestore"
	"github.com/suared/core/v2/security"

	_ "github.com/suared/core/v2/infra"
)

func setupConfig() repository.Config {
	configMap := repository.NewBasicConfig("TestRepo")
	configMap.AddEntry("backend", "firestore")
	configMap.AddEntry("projectID", os.Getenv("PROCESS_FIRESTORE_PROJECT_ID"))
	configMap.AddEntry("endpoint", os.Getenv("FIRESTORE_EMULATOR_HOST"))
	configMap.AddEntry("env", os.Getenv("PROCESS_ENV"))
	return configMap
}

func TestSimplePath(t *testing.T) {
	ctx := security.SetupTestAuthFromContext(context.Background(), 1)
	config := setupConfig()

	repo, err := firestore.NewRepository(config, firestore.ModelMapping[TestUserModel]{
		Collection: "test-collection-simple",
		DocIDFunc:  func(m TestUserModel) string { return m.ID },
	})
	if err != nil {
		t.Fatalf("Repo initialization failed: %v", err)
	}

	// Run tests
	runTestSuite(t, ctx, repo, "Simple Path")
}

func TestAdvancedPath(t *testing.T) {
	ctx := security.SetupTestAuthFromContext(context.Background(), 1)
	config := setupConfig()

	repo, err := firestore.NewRepositoryWithDAO(config, firestore.DaoMapping[TestUserModel, MyCustomDAO]{
		Collection: "test-collection-advanced",
		DocIDFunc: func(dao MyCustomDAO) string {
			return dao.Model.ID
		},
		FromModelFunc: FromModel,
		ToModelFunc: func(dao MyCustomDAO) TestUserModel {
			return dao.ToModel()
		},
	})
	if err != nil {
		t.Fatalf("Repo initialization failed: %v", err)
	}

	// Run tests
	runTestSuite(t, ctx, repo, "Advanced Path")

	// Run specific checks for the advanced fields
	t.Run("AdvancedPath_ExtraFields", func(t *testing.T) {
		model := TestUserModel{ID: "C", Title: "advanced item check"}
		if err := repo.Insert(ctx, model); err != nil {
			t.Fatalf("Insert failed: %v", err)
		}

		// To check the extra fields, we need to query the raw DAO
		retrievedDAO, err := firestore.SelectOne[MyCustomDAO](ctx, repo, repo.Collection, model.ID)
		if err != nil {
			t.Fatalf("SelectOne for raw DAO failed: %v", err)
		}

		expectedKeywords := []string{"advanced", "item", "check"}
		if !reflect.DeepEqual(retrievedDAO.SearchKeywords, expectedKeywords) {
			t.Errorf("SearchKeywords mismatch: got %v, want %v", retrievedDAO.SearchKeywords, expectedKeywords)
		}

		if retrievedDAO.IsArchived != false {
			t.Errorf("IsArchived mismatch: got %v, want false", retrievedDAO.IsArchived)
		}

		// Clean up
		if err := repo.Delete(ctx, model); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	})
}

// runTestSuite contains the actual test logic, runnable for any repo implementation.
func runTestSuite(t *testing.T, ctx context.Context, repo *firestore.Repository[TestUserModel], suiteName string) {
	t.Run(suiteName, func(t *testing.T) {
		// 1. Insert
		testModel := TestUserModel{ID: "A", Title: "first test"}
		if err := repo.Insert(ctx, testModel); err != nil {
			t.Fatalf("Insert failed: %v", err)
		}

		// 2. SelectOne
		retrieved, err := repo.SelectOne(ctx, TestUserModel{ID: "A"})
		if err != nil {
			t.Fatalf("SelectOne failed: %v", err)
		}
		if retrieved != testModel {
			t.Errorf("SelectOne retrieved wrong model. Got %v, want %v", retrieved, testModel)
		}

		// 3. Insert another, then Select
		testModel2 := TestUserModel{ID: "B", Title: "second test"}
		if err := repo.Insert(ctx, testModel2); err != nil {
			t.Fatalf("Insert (2) failed: %v", err)
		}

		// 4. Select with filter
		all, err := repo.Select(ctx, TestUserModel{}, "model.Title", "==", "second test")
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}
		if len(all) != 1 {
			t.Fatalf("Select got %d results, want 1", len(all))
		}
		if all[0] != testModel2 {
			t.Errorf("Select retrieved wrong model. Got %v, want %v", all[0], testModel2)
		}

		// 5. Update
		updatedModel := TestUserModel{ID: "A", Title: "first test updated"}
		if err := repo.Update(ctx, updatedModel); err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		retrieved, err = repo.SelectOne(ctx, TestUserModel{ID: "A"})
		if err != nil {
			t.Fatalf("SelectOne after update failed: %v", err)
		}
		if retrieved != updatedModel {
			t.Errorf("SelectOne after update retrieved wrong model. Got %v, want %v", retrieved, updatedModel)
		}

		// 6. Delete
		if err := repo.Delete(ctx, TestUserModel{ID: "A"}); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		retrieved, err = repo.SelectOne(ctx, TestUserModel{ID: "A"})
		if err == nil && !retrieved.IsEmpty() {
			t.Errorf("SelectOne after delete should have failed or returned empty, but got: %v", retrieved)
		}

		// 7. Clean up second item
		if err := repo.Delete(ctx, TestUserModel{ID: "B"}); err != nil {
			t.Fatalf("Delete (2) failed: %v", err)
		}
	})
}