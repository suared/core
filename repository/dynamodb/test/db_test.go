package test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/suared/core/v2/repository"
	"github.com/suared/core/v2/repository/dynamodb"
	"github.com/suared/core/v2/security"

	_ "github.com/suared/core/v2/infra"
)

func setupConfig(t *testing.T) repository.Config {
	configMap := repository.NewBasicConfig("TestDao")
	os.Setenv("PROCESS_REPOSITORY", "dynamoDB")
	os.Setenv("PROCESS_AWS_DYNAMOTABLE_CATEGORY", "test-table-dyn")
	os.Setenv("PROCESS_AWS_REGION", "us-east-1")
	os.Setenv("PROCESS_AWS_DYNAMOENDPOINT", "http://localhost:9001")
	os.Setenv("PROCESS_AWS_DYNAMOTABLE_RCU", "5")
	os.Setenv("PROCESS_AWS_DYNAMOTABLE_WCU", "5")
	os.Setenv("PROCESS_ENV", "dev")
	os.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")

	configMap.AddEntry("backend", os.Getenv("PROCESS_REPOSITORY"))
	configMap.AddEntry("table", os.Getenv("PROCESS_AWS_DYNAMOTABLE_CATEGORY"))
	configMap.AddEntry("region", os.Getenv("PROCESS_AWS_REGION"))
	configMap.AddEntry("endpoint", os.Getenv("PROCESS_AWS_DYNAMOENDPOINT"))
	configMap.AddEntry("rcu", os.Getenv("PROCESS_AWS_DYNAMOTABLE_RCU"))
	configMap.AddEntry("wcu", os.Getenv("PROCESS_AWS_DYNAMOTABLE_WCU"))
	configMap.AddEntry("hashKeyName", "MyCleverHashKeyName")
	configMap.AddEntry("sortKeyName", "MyCleverSortKeyName")
	configMap.AddEntry("env", os.Getenv("PROCESS_ENV"))
	return configMap
}

func TestSimplePath(t *testing.T) {
	ctx := security.SetupTestAuthFromContext(context.Background(), 1)
	config := setupConfig(t)

	mapping := dynamodb.DaoMapping[TestUserModel, SimpleTestDAO]{
		FromModelFunc: FromModelSimple,
		ToModelFunc:   ToModelSimple,
		HashKeyFunc:   func(dao SimpleTestDAO) string { return dao.HKey },
		SortKeyFunc:   func(dao SimpleTestDAO) string { return dao.SKey },
	}

	repo, err := dynamodb.NewRepository(config, mapping)
	if err != nil {
		t.Fatalf("Repo initialization failed: %v", err)
	}

	runTestSuite(t, ctx, repo, "SimplePath")
}

func TestAdvancedPath(t *testing.T) {
	ctx := security.SetupTestAuthFromContext(context.Background(), 2) // Use a different user for a clean hash key
	config := setupConfig(t)

	mapping := dynamodb.DaoMapping[TestUserModel, AdvancedTestDAO]{
		FromModelFunc: FromModelAdvanced,
		ToModelFunc:   ToModelAdvanced,
		HashKeyFunc:   func(dao AdvancedTestDAO) string { return dao.HKey },
		SortKeyFunc:   func(dao AdvancedTestDAO) string { return dao.SKey },
	}

	repo, err := dynamodb.NewRepository(config, mapping)
	if err != nil {
		t.Fatalf("Repo initialization failed: %v", err)
	}

	// Run the standard test suite
	runTestSuite(t, ctx, repo, "AdvancedPath")

	// Run specific checks for the advanced fields
	t.Run("AdvancedDAO_ExtraFields", func(t *testing.T) {
		model := TestUserModel{ID: "C", Title: "advanced item check"}
		if err := repo.Insert(ctx, model); err != nil {
			t.Fatalf("Insert failed: %v", err)
		}

		// To check the extra fields, we need to query the raw DAO
		hashKey := FromModelAdvanced(ctx, model).HKey
		sortKey := FromModelAdvanced(ctx, model).SKey
		retrievedDAO, err := dynamodb.SelectOne[AdvancedTestDAO](ctx, repo, hashKey, sortKey)
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

// runTestSuite contains the core CRUD test logic, runnable for any repository configuration.
func runTestSuite[D any](t *testing.T, ctx context.Context, repo *dynamodb.Repository[TestUserModel, D], suiteName string) {
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

		// 3. Insert another, then Select with filter
		testModel2 := TestUserModel{ID: "B", Title: "second test"}
		if err := repo.Insert(ctx, testModel2); err != nil {
			t.Fatalf("Insert (2) failed: %v", err)
		}
		all, err := repo.Select(ctx, TestUserModel{}, "model.Title = :title", "title:S:second test")
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}
		if len(all) != 1 {
			t.Fatalf("Select got %d results, want 1", len(all))
		}
		if all[0] != testModel2 {
			t.Errorf("Select retrieved wrong model. Got %v, want %v", all[0], testModel2)
		}

		// 4. Update
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

		// 5. Delete
		if err := repo.Delete(ctx, TestUserModel{ID: "A"}); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		retrieved, err = repo.SelectOne(ctx, TestUserModel{ID: "A"})
		if err == nil && !retrieved.IsEmpty() {
			t.Errorf("SelectOne after delete should have been empty, but got: %v", retrieved)
		}

		// 6. Clean up second item
		if err := repo.Delete(ctx, TestUserModel{ID: "B"}); err != nil {
			t.Fatalf("Delete (2) failed: %v", err)
		}
	})
}
